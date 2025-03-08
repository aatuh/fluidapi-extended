package crud

import (
	"context"

	apiendpoint "github.com/pakkasys/fluidapi-extended/api/endpoint"
	"github.com/pakkasys/fluidapi-extended/api/repository"
	"github.com/pakkasys/fluidapi-extended/api/types"
	"github.com/pakkasys/fluidapi-extended/middleware"
	"github.com/pakkasys/fluidapi/database"
	"github.com/pakkasys/fluidapi/endpoint"
)

// CRUDEntity is a helper constraint for entities that can be inserted,
// retrieved, and updated (they must implement the necessary database
// interfaces).
type CRUDEntity interface {
	database.Mutator
	database.Getter
	database.TableNamer
}

// ---------------------------------------------------------------------
// Common Parameters & Operation Interface
// ---------------------------------------------------------------------

// CRUDCommonParams bundles configuration shared by all CRUD operations.
type CRUDCommonParams[Entity CRUDEntity] struct {
	URL             string
	ConnFn          repository.ConnFn
	EntityFn        func(opts ...EntityOption[Entity]) Entity
	LoggerFactoryFn apiendpoint.LoggerFactoryFn
	TableName       string
	MutatorRepo     repository.MutatorRepo[Entity]
	ReaderRepo      repository.ReaderRepo[Entity]
	TxManager       repository.TxManager[Entity]
	ConversionRules map[string]func(any) any
	CustomRules     map[string]func(any) error
}

// ---------------------------------------------------------------------
// Create CRUD
// ---------------------------------------------------------------------

type CreateCRUD[CreateInput any, Entity CRUDEntity] struct {
	CRUDCommonParams[Entity]
	APIFields       types.APIFields
	InputFactory    func() CreateInput
	DBOptionFn      func(field string, value any) EntityOption[Entity]
	InputKey        string
	OutputAPIFields types.APIFields
	OutputKey       string
	BeforeCallback  func(ctx context.Context, entity Entity, input *CreateInput) error
	ErrorMapping    map[string]middleware.ExpectedError
}

func NewCreateCRUD[CreateInput any, Entity CRUDEntity](
	common CRUDCommonParams[Entity],
	apiFields types.APIFields,
	inputFactory func() CreateInput,
	dbOptionFn func(string, any) EntityOption[Entity],
	inputKey string,
	outputAPIFields types.APIFields,
	outputKey string,
	beforeCallback func(context.Context, Entity, *CreateInput) error,
	errorMapping map[string]middleware.ExpectedError,
) *CreateCRUD[CreateInput, Entity] {
	return &CreateCRUD[CreateInput, Entity]{
		CRUDCommonParams: common,
		APIFields:        apiFields,
		InputFactory:     inputFactory,
		DBOptionFn:       dbOptionFn,
		InputKey:         inputKey,
		OutputAPIFields:  outputAPIFields,
		OutputKey:        outputKey,
		BeforeCallback:   beforeCallback,
		ErrorMapping:     errorMapping,
	}
}

func (c *CreateCRUD[CreateInput, Entity]) EndpointHandler() *apiendpoint.EndpointHandler[CreateInput] {
	opts := []apiendpoint.GenericEndpointOptions{}
	if c.ErrorMapping != nil {
		mappedErrors := mustApplyErrorMapping(
			apiendpoint.CreateErrors,
			c.ErrorMapping,
		)
		opts = append(opts, apiendpoint.GenericEndpointOptions{
			ExpectedErrors: &mappedErrors,
		})
	}
	return apiendpoint.GenericCreateDefinition(
		c.URL,
		apiendpoint.NewMapInputHandler(
			c.APIFields, c.ConversionRules, c.CustomRules,
		),
		c.InputFactory,
		c.ConnFn,
		func(ctx context.Context, input *CreateInput) (Entity, error) {
			colsAndValues, err := structToDBEntity[CreateInput, Entity](
				input,
				c.InputKey,
				c.APIFields.MustGetAPIField(c.InputKey).Nested,
			)
			if err != nil {
				var zero Entity
				return zero, err
			}
			var dbOpts []EntityOption[Entity]
			for col, value := range colsAndValues {
				dbOpts = append(dbOpts, c.DBOptionFn(col, value))
			}
			return c.EntityFn(dbOpts...), nil
		},
		func(entity Entity) (any, error) {
			outMap, err := dbEntityToMap(
				entity,
				c.OutputAPIFields.MustGetAPIField(c.OutputKey).Nested,
				map[string]any{},
			)
			if err != nil {
				return nil, err
			}
			return map[string]any{c.OutputKey: *outMap}, nil
		},
		c.BeforeCallback,
		c.LoggerFactoryFn,
		c.MutatorRepo,
		c.TxManager,
		opts...,
	)
}

func (c *CreateCRUD[CreateInput, Entity]) NewInput() any {
	input := c.InputFactory()
	return &input
}

// ---------------------------------------------------------------------
// Get CRUD
// ---------------------------------------------------------------------

type GetCRUD[Entity CRUDEntity, Output any] struct {
	CRUDCommonParams[Entity]
	APIFields        types.APIFields
	OutputAPIFields  types.APIFields
	OutputKey        string
	OutputCountField string
	OrderableFields  []string
	BeforeCallback   func(context.Context, Entity, *apiendpoint.GetInput) error
}

func NewGetCRUD[Entity CRUDEntity, Output any](
	common CRUDCommonParams[Entity],
	apiFields types.APIFields,
	outputAPIFields types.APIFields,
	outputKey string,
	outputCountField string,
	orderableFields []string,
	beforeCallback func(context.Context, Entity, *apiendpoint.GetInput) error,
) *GetCRUD[Entity, Output] {
	return &GetCRUD[Entity, Output]{
		CRUDCommonParams: common,
		APIFields:        apiFields,
		OutputAPIFields:  outputAPIFields,
		OutputKey:        outputKey,
		OutputCountField: outputCountField,
		OrderableFields:  orderableFields,
		BeforeCallback:   beforeCallback,
	}
}

func (g *GetCRUD[Entity, Output]) EndpointHandler() *apiendpoint.EndpointHandler[apiendpoint.GetInput] {
	return apiendpoint.GenericGetDefinition(
		g.URL,
		apiendpoint.NewMapInputHandler(
			g.APIFields, g.ConversionRules, g.CustomRules,
		),
		getAPIFieldToDBColumnMapping(
			g.APIFields.MustGetAPIField(FieldSelectors).Nested,
			g.TableName,
		),
		func(entities []Entity, count int) (*Output, error) {
			return toGenericGetOutput(
				entities,
				count,
				g.OutputAPIFields,
				g.OutputKey,
				g.OutputCountField,
				new(Output),
			)
		},
		g.ConnFn,
		func() Entity { return g.EntityFn() },
		g.BeforeCallback,
		g.LoggerFactoryFn,
		g.ReaderRepo,
		g.TxManager,
	)
}

func (g *GetCRUD[Entity, Output]) NewInput() any {
	return &apiendpoint.GetInput{}
}

// ---------------------------------------------------------------------
// Update CRUD
// ---------------------------------------------------------------------

type UpdateCRUD[Entity CRUDEntity] struct {
	CRUDCommonParams[Entity]
	APIFields      types.APIFields
	BeforeCallback func(context.Context, database.Mutator, *apiendpoint.UpdateInput) error
}

func NewUpdateCRUD[Entity CRUDEntity](
	common CRUDCommonParams[Entity],
	apiFields types.APIFields,
	beforeCallback func(context.Context, database.Mutator, *apiendpoint.UpdateInput) error,
) *UpdateCRUD[Entity] {
	return &UpdateCRUD[Entity]{
		CRUDCommonParams: common,
		APIFields:        apiFields,
		BeforeCallback:   beforeCallback,
	}
}

func (u *UpdateCRUD[Entity]) EndpointHandler() *apiendpoint.EndpointHandler[apiendpoint.UpdateInput] {
	return apiendpoint.GenericUpdateDefinition(
		u.URL,
		apiendpoint.NewMapInputHandler(
			u.APIFields, u.ConversionRules, u.CustomRules,
		),
		getAPIFieldToDBColumnMapping(
			u.APIFields.MustGetAPIField(FieldSelectors).Nested,
			u.TableName,
		),
		u.ConnFn,
		func() database.Mutator { return u.EntityFn() },
		u.BeforeCallback,
		u.LoggerFactoryFn,
	)
}

func (u *UpdateCRUD[Entity]) NewInput() any {
	return &apiendpoint.UpdateInput{}
}

// ---------------------------------------------------------------------
// Delete CRUD
// ---------------------------------------------------------------------

type DeleteCRUD[Entity CRUDEntity] struct {
	CRUDCommonParams[Entity]
	APIFields      types.APIFields
	BeforeCallback func(context.Context, database.Mutator, *apiendpoint.DeleteInput) error
}

func NewDeleteCRUD[Entity CRUDEntity](
	common CRUDCommonParams[Entity],
	apiFields types.APIFields,
	beforeCallback func(context.Context, database.Mutator, *apiendpoint.DeleteInput) error,
) *DeleteCRUD[Entity] {
	return &DeleteCRUD[Entity]{
		CRUDCommonParams: common,
		APIFields:        apiFields,
		BeforeCallback:   beforeCallback,
	}
}

func (d *DeleteCRUD[Entity]) EndpointHandler() *apiendpoint.EndpointHandler[apiendpoint.DeleteInput] {
	return apiendpoint.GenericDeleteDefinition(
		d.URL,
		apiendpoint.NewMapInputHandler(
			d.APIFields, d.ConversionRules, d.CustomRules,
		),
		getAPIFieldToDBColumnMapping(
			d.APIFields.MustGetAPIField(FieldSelectors).Nested,
			d.TableName,
		),
		d.ConnFn,
		func() database.Mutator { return d.EntityFn() },
		d.BeforeCallback,
		d.LoggerFactoryFn,
	)
}

func (d *DeleteCRUD[Entity]) NewInput() any {
	return &apiendpoint.DeleteInput{}
}

// ---------------------------------------------------------------------
// CRUD Config & Builder
// ---------------------------------------------------------------------

type CRUDConfig[Entity CRUDEntity, CreateInput any, CreateOutput any, GetOutput any] struct {
	URL                  string
	TableName            string
	EntityFn             func(...EntityOption[Entity]) Entity
	Predicates           map[string]endpoint.Predicates
	Orderable            []string
	ConnFn               repository.ConnFn
	AllAPIFields         types.APIFields
	UpdateAPIFields      types.APIFields
	EntityName           string
	EntityNamePlural     string
	BeforeCreateCallback func(context.Context, Entity, *CreateInput) error
	BeforeGetCallback    func(context.Context, Entity, *apiendpoint.GetInput) error
	BeforeUpdateCallback func(context.Context, database.Mutator, *apiendpoint.UpdateInput) error
	BeforeDeleteCallback func(context.Context, database.Mutator, *apiendpoint.DeleteInput) error
	ErrorMapping         map[string]middleware.ExpectedError
	LoggerFactoryFn      apiendpoint.LoggerFactoryFn
	MutatorRepo          repository.MutatorRepo[Entity]
	ReaderRepo           repository.ReaderRepo[Entity]
	TxManager            repository.TxManager[Entity]
	ConversionRules      map[string]func(any) any
	CustomRules          map[string]func(any) error
}

type CRUDDefinitions struct {
	Create *endpoint.Definition
	Get    *endpoint.Definition
	Update *endpoint.Definition
	Delete *endpoint.Definition
}

type CRUDBuilder[Entity CRUDEntity, CreateInput any,
	CreateOutput any, GetOutput any] struct {
	Config     CRUDConfig[Entity, CreateInput, CreateOutput, GetOutput]
	createFlag bool
	getFlag    bool
	updateFlag bool
	deleteFlag bool
}

func NewCRUDBuilder[
	Entity CRUDEntity,
	CreateInput any,
	CreateOutput any,
	GetOutput any,
](config CRUDConfig[Entity, CreateInput, CreateOutput, GetOutput],
) *CRUDBuilder[Entity, CreateInput, CreateOutput, GetOutput] {
	return &CRUDBuilder[Entity, CreateInput, CreateOutput, GetOutput]{
		Config:     config,
		createFlag: true,
		getFlag:    true,
		updateFlag: true,
		deleteFlag: true,
	}
}

func (b *CRUDBuilder[Entity, CreateInput, CreateOutput,
	GetOutput]) WithCreate(enabled bool) *CRUDBuilder[Entity, CreateInput,
	CreateOutput, GetOutput] {
	b.createFlag = enabled
	return b
}

func (b *CRUDBuilder[Entity, CreateInput, CreateOutput,
	GetOutput]) WithGet(enabled bool) *CRUDBuilder[Entity, CreateInput,
	CreateOutput, GetOutput] {
	b.getFlag = enabled
	return b
}

func (b *CRUDBuilder[Entity, CreateInput, CreateOutput,
	GetOutput]) WithUpdate(enabled bool) *CRUDBuilder[Entity, CreateInput,
	CreateOutput, GetOutput] {
	b.updateFlag = enabled
	return b
}

func (b *CRUDBuilder[Entity, CreateInput, CreateOutput,
	GetOutput]) WithDelete(enabled bool) *CRUDBuilder[Entity, CreateInput,
	CreateOutput, GetOutput] {
	b.deleteFlag = enabled
	return b
}

type CRUDEndpoints[Entity CRUDEntity, CreateInput any, CreateOutput any, GetOutput any] struct {
	Create *CreateCRUD[CreateInput, Entity]
	Get    *GetCRUD[Entity, GetOutput]
	Update *UpdateCRUD[Entity]
	Delete *DeleteCRUD[Entity]
}

func (b *CRUDBuilder[Entity, CreateInput, CreateOutput, GetOutput]) BuildCRUDEndpoints() *CRUDEndpoints[Entity, CreateInput, CreateOutput, GetOutput] {
	var endpoints CRUDEndpoints[Entity, CreateInput, CreateOutput, GetOutput]
	common := CRUDCommonParams[Entity]{
		URL:             b.Config.URL,
		ConnFn:          b.Config.ConnFn,
		EntityFn:        b.Config.EntityFn,
		LoggerFactoryFn: b.Config.LoggerFactoryFn,
		TableName:       b.Config.TableName,
		MutatorRepo:     b.Config.MutatorRepo,
		ReaderRepo:      b.Config.ReaderRepo,
		TxManager:       b.Config.TxManager,
		ConversionRules: b.Config.ConversionRules,
		CustomRules:     b.Config.CustomRules,
	}
	if b.createFlag {
		endpoints.Create = NewCreateCRUD(
			common,
			MustStructToAPIFields[CreateInput](b.Config.AllAPIFields),
			func() CreateInput { return *new(CreateInput) },
			WithOption[Entity],
			b.Config.EntityName,
			MustStructToAPIFields[CreateOutput](b.Config.AllAPIFields),
			b.Config.EntityName,
			b.Config.BeforeCreateCallback,
			b.Config.ErrorMapping,
		)
	}
	if b.getFlag {
		MustValidateGetOutput[GetOutput](
			b.Config.AllAPIFields,
			b.Config.EntityNamePlural,
		)
		endpoints.Get = NewGetCRUD[Entity, GetOutput](
			common,
			genericGetAPIFields(
				b.Config.AllAPIFields, b.Config.Predicates, b.Config.Orderable,
			),
			genericGetOutputAPIFields(
				b.Config.EntityNamePlural,
				b.Config.AllAPIFields,
			),
			b.Config.EntityNamePlural,
			FieldCount,
			b.Config.Orderable,
			b.Config.BeforeGetCallback,
		)
	}
	if b.updateFlag {
		endpoints.Update = NewUpdateCRUD(
			common,
			genericUpdateAPIFields(
				b.Config.AllAPIFields,
				b.Config.Predicates,
				b.Config.UpdateAPIFields,
			),
			b.Config.BeforeUpdateCallback,
		)
	}
	if b.deleteFlag {
		endpoints.Delete = NewDeleteCRUD(
			common,
			genericDeleteAPIFields(
				b.Config.AllAPIFields,
				b.Config.Predicates,
			),
			b.Config.BeforeDeleteCallback,
		)
	}
	return &endpoints
}
