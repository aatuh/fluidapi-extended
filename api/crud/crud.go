package crud

import (
	"context"

	"github.com/pakkasys/fluidapi-extended/api"
	apiendpoint "github.com/pakkasys/fluidapi-extended/api/endpoint"
	"github.com/pakkasys/fluidapi-extended/api/repository"
	"github.com/pakkasys/fluidapi-extended/api/types"
	extendeddatabase "github.com/pakkasys/fluidapi-extended/database"
	"github.com/pakkasys/fluidapi/database"
	"github.com/pakkasys/fluidapi/endpoint"
)

// ---------------------------------------------------------------------
// Common Parameters & Operation Interface
// ---------------------------------------------------------------------

// CRUDCommonParams bundles configuration shared by all CRUD operations.
type CRUDCommonParams[Entity database.CRUDEntity] struct {
	URL             string
	ConnFn          repository.ConnFn
	EntityFn        func(opts ...extendeddatabase.EntityOption[Entity]) Entity
	LoggerFactoryFn apiendpoint.LoggerFactoryFn
	TableName       string
	MutatorRepo     repository.MutatorRepo[Entity]
	ReaderRepo      repository.ReaderRepo[Entity]
	TxManager       repository.TxManager[Entity]
	ConversionRules map[string]func(any) any
	CustomRules     map[string]func(any) error
	SystemId        string
}

// ---------------------------------------------------------------------
// Create CRUD
// ---------------------------------------------------------------------

type CreateCRUD[CreateInput any, Entity database.CRUDEntity] struct {
	CRUDCommonParams[Entity]
	APIFields    types.APIFields
	InputFactory func() CreateInput
	DBOptionFn   func(
		field string, value any,
	) extendeddatabase.EntityOption[Entity]
	InputKey        string
	OutputAPIFields types.APIFields
	OutputKey       string
	BeforeCallback  func(ctx context.Context, entity Entity, input *CreateInput) error
	ErrorMapping    map[string]api.ExpectedError
}

func NewCreateCRUD[CreateInput any, Entity database.CRUDEntity](
	common CRUDCommonParams[Entity],
	apiFields types.APIFields,
	inputFactory func() CreateInput,
	dbOptionFn func(string, any) extendeddatabase.EntityOption[Entity],
	inputKey string,
	outputAPIFields types.APIFields,
	outputKey string,
	beforeCallback func(context.Context, Entity, *CreateInput) error,
	errorMapping map[string]api.ExpectedError,
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
			apiendpoint.NewErrorBuilder(c.SystemId).
				With(apiendpoint.CreateErrors()).Build(),
			c.ErrorMapping,
		)
		opts = append(opts, apiendpoint.GenericEndpointOptions{
			ExpectedErrors: &mappedErrors,
		})
	}
	return apiendpoint.GenericCreateDefinition(
		c.URL,
		api.NewMapInputHandler(
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
			var dbOpts []extendeddatabase.EntityOption[Entity]
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
		c.SystemId,
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

type GetCRUD[Entity database.CRUDEntity, Output any] struct {
	CRUDCommonParams[Entity]
	APIFields        types.APIFields
	OutputAPIFields  types.APIFields
	OutputKey        string
	OutputCountField string
	OrderableFields  []string
	BeforeCallback   func(context.Context, Entity, *apiendpoint.GetInput) error
}

func NewGetCRUD[Entity database.CRUDEntity, Output any](
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
		api.NewMapInputHandler(
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
		g.SystemId,
	)
}

func (g *GetCRUD[Entity, Output]) NewInput() any {
	return &apiendpoint.GetInput{}
}

// ---------------------------------------------------------------------
// Update CRUD
// ---------------------------------------------------------------------

type UpdateCRUD[Entity database.CRUDEntity] struct {
	CRUDCommonParams[Entity]
	APIFields      types.APIFields
	BeforeCallback func(context.Context, database.Mutator, *apiendpoint.UpdateInput) error
}

func NewUpdateCRUD[Entity database.CRUDEntity](
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
		api.NewMapInputHandler(
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
		u.SystemId,
	)
}

func (u *UpdateCRUD[Entity]) NewInput() any {
	return &apiendpoint.UpdateInput{}
}

// ---------------------------------------------------------------------
// Delete CRUD
// ---------------------------------------------------------------------

type DeleteCRUD[Entity database.CRUDEntity] struct {
	CRUDCommonParams[Entity]
	APIFields      types.APIFields
	BeforeCallback func(context.Context, database.Mutator, *apiendpoint.DeleteInput) error
}

func NewDeleteCRUD[Entity database.CRUDEntity](
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
		api.NewMapInputHandler(
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
		d.SystemId,
	)
}

func (d *DeleteCRUD[Entity]) NewInput() any {
	return &apiendpoint.DeleteInput{}
}

// ---------------------------------------------------------------------
// CRUD Config & Builder
// ---------------------------------------------------------------------

type CRUDConfig[Entity database.CRUDEntity, CreateInput any, CreateOutput any, GetOutput any] struct {
	URL                  string
	TableName            string
	EntityFn             func(...extendeddatabase.EntityOption[Entity]) Entity
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
	ErrorMapping         map[string]api.ExpectedError
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

type CRUDBuilder[Entity database.CRUDEntity, CreateInput any,
	CreateOutput any, GetOutput any] struct {
	Config     CRUDConfig[Entity, CreateInput, CreateOutput, GetOutput]
	createFlag bool
	getFlag    bool
	updateFlag bool
	deleteFlag bool
}

func NewCRUDBuilder[
	Entity database.CRUDEntity,
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

type CRUDEndpoints[Entity database.CRUDEntity, CreateInput any, CreateOutput any, GetOutput any] struct {
	Create *CreateCRUD[CreateInput, Entity]
	Get    *GetCRUD[Entity, GetOutput]
	Update *UpdateCRUD[Entity]
	Delete *DeleteCRUD[Entity]
}

func (b *CRUDBuilder[Entity, CreateInput, CreateOutput, GetOutput]) BuildCRUDEndpoints(
	systemId string,
) *CRUDEndpoints[Entity, CreateInput, CreateOutput, GetOutput] {
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
		SystemId:        systemId,
	}
	if b.createFlag {
		endpoints.Create = NewCreateCRUD(
			common,
			MustStructToAPIFields[CreateInput](b.Config.AllAPIFields),
			func() CreateInput { return *new(CreateInput) },
			extendeddatabase.WithOption[Entity],
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
