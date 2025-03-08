package endpoint

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pakkasys/fluidapi-extended/api/repository"
	"github.com/pakkasys/fluidapi-extended/client"
	"github.com/pakkasys/fluidapi-extended/middleware"
	"github.com/pakkasys/fluidapi-extended/util"
	"github.com/pakkasys/fluidapi/core"
	"github.com/pakkasys/fluidapi/database"
	"github.com/pakkasys/fluidapi/endpoint"
)

// APIToDBFields maps API fields to database fields.
type APIToDBFields map[string]endpoint.DBField

// Selector and Update interfaces for custom predicates and updates.
type Selector interface {
	GetValue() any
	GetPredicate() endpoint.Predicate
}

type Update interface {
	ToUpdates() (endpoint.Updates, error)
}

// Input types for the generic endpoints.
type GetInput struct {
	Selectors endpoint.Selectors `json:"selectors"`
	Orders    endpoint.Orders    `json:"orders"`
	Page      *endpoint.Page     `json:"page"`
	Count     bool               `json:"count"`
}

type UpdateInput struct {
	Selectors endpoint.Selectors `json:"selectors"`
	Updates   endpoint.Updates   `json:"updates"`
	Upsert    bool               `json:"upsert"`
}

type DeleteInput struct {
	Selectors endpoint.Selectors `json:"selectors"`
}

// Output types.
type UpdateOutput struct {
	Count int64 `json:"count"`
}

type DeleteOutput struct {
	Count int64 `json:"count"`
}

// Options to override default expected errors.
type GenericEndpointOptions struct {
	ExpectedErrors *middleware.ExpectedErrors
}

// Error variables.
var (
	NeedAtLeastOneUpdateError   = core.NewAPIError("NEED_AT_LEAST_ONE_UPDATE")
	NeedAtLeastOneSelectorError = core.NewAPIError("NEED_AT_LEAST_ONE_SELECTOR")
)

var GenericErrors = middleware.ExpectedErrors{
	{ID: MapToObjectDecodingError.ID, Status: http.StatusBadRequest, PublicData: true},
	{ID: util.InvalidInputError.ID, Status: http.StatusBadRequest, PublicData: true},
	{ID: ValidationError.ID, Status: http.StatusBadRequest, PublicData: true},
}

var (
	CreateErrors = middleware.ExpectedErrors{
		{ID: util.DuplicateEntryError.ID, Status: http.StatusBadRequest, PublicData: false},
		{ID: util.ForeignConstraintError.ID, Status: http.StatusBadRequest, PublicData: false},
	}.With(GenericErrors...)

	GetErrors = middleware.ExpectedErrors{
		{ID: endpoint.InvalidPredicateError.ID, Status: http.StatusBadRequest, PublicData: true},
		{ID: endpoint.PredicateNotAllowedError.ID, Status: http.StatusBadRequest, PublicData: true},
		{ID: endpoint.InvalidSelectorFieldError.ID, Status: http.StatusBadRequest, PublicData: true},
		{ID: endpoint.InvalidOrderFieldError.ID, Status: http.StatusBadRequest, PublicData: true},
		{ID: endpoint.MaxPageLimitExceededError.ID, Status: http.StatusBadRequest, PublicData: true},
		{ID: util.NoRowsError.ID, Status: http.StatusNotFound, PublicData: true},
	}.With(GenericErrors...)

	UpdateErrors = middleware.ExpectedErrors{
		{ID: endpoint.InvalidPredicateError.ID, Status: http.StatusBadRequest, PublicData: true},
		{ID: endpoint.InvalidSelectorFieldError.ID, Status: http.StatusBadRequest, PublicData: true},
		{ID: endpoint.PredicateNotAllowedError.ID, Status: http.StatusBadRequest, PublicData: true},
		{ID: NeedAtLeastOneSelectorError.ID, Status: http.StatusBadRequest, PublicData: true},
		{ID: NeedAtLeastOneUpdateError.ID, Status: http.StatusBadRequest, PublicData: true},
		{ID: endpoint.InvalidOrderFieldError.ID, Status: http.StatusBadRequest, PublicData: true},
		{ID: util.DuplicateEntryError.ID, Status: http.StatusBadRequest, PublicData: false},
		{ID: util.ForeignConstraintError.ID, Status: http.StatusBadRequest, PublicData: false},
	}.With(GenericErrors...)

	DeleteErrors = middleware.ExpectedErrors{
		{ID: endpoint.InvalidPredicateError.ID, Status: http.StatusBadRequest, PublicData: true},
		{ID: endpoint.InvalidSelectorFieldError.ID, Status: http.StatusBadRequest, PublicData: true},
		{ID: endpoint.PredicateNotAllowedError.ID, Status: http.StatusBadRequest, PublicData: true},
		{ID: NeedAtLeastOneSelectorError.ID, Status: http.StatusBadRequest, PublicData: true},
	}.With(GenericErrors...)
)

// GenericEndpointHandler builds the endpoint definition for any operation.
func GenericEndpointHandler[Input any](
	url string,
	method string,
	inputHandler InputHandler,
	inputFactory func() Input,
	expectedErrors middleware.ExpectedErrors,
	invokeFn GenericInvoke[Input],
	loggerFactoryFn LoggerFactoryFn,
) *EndpointHandler[Input] {
	handler := &GenericHandler[Input]{
		InvokeFn: invokeFn,
	}
	return NewEndpointHandler(
		url,
		method,
		inputHandler,
		inputFactory,
		handler.Handle,
		expectedErrors,
		loggerFactoryFn,
	)
}

// SendRequest sends a request to the target host. It parses the input first.
func SendRequest[Input any, Output any](
	ctx context.Context, host string, url string, method string, input *Input,
) (*client.Response[util.APIOutput[Output]], error) {
	parsedInput, err := client.ParseInput(method, input)
	if err != nil {
		return nil, err
	}
	return SendParsedRequest[Input, Output](ctx, host, url, method, parsedInput)
}

// SendParsedRequest sends a request to the target host using the parsed input.
func SendParsedRequest[Input any, Output any](
	ctx context.Context,
	host string,
	url string,
	method string,
	parsedInput *client.ParsedInput,
) (*client.Response[util.APIOutput[Output]], error) {
	// Set options
	opts := &client.SendOptions{
		Headers: parsedInput.Headers,
		Cookies: parsedInput.Cookies,
	}
	if method != http.MethodGet {
		opts.Body = parsedInput.Body
	}

	// Add query parameters to the URL
	urlValues, err := util.NewURLEncoder().Encode(parsedInput.URLParameters)
	if err != nil {
		return nil, err
	}

	// Add query parameters to the URL
	var fullURL string
	if len(urlValues) == 0 {
		fullURL = url
	} else {
		fullURL = fmt.Sprintf("%s?%s", url, urlValues.Encode())
	}

	// Send request
	return client.Send[util.APIOutput[Output]](
		host, fullURL, method, opts,
	)
}

// GenericCreateDefinition builds the endpoint definition for a create operation.
func GenericCreateDefinition[Input any, Entity database.Mutator](
	url string,
	inputHandler InputHandler,
	inputFactory func() Input,
	getConnectionFn repository.ConnFn,
	entityFactoryFn CreateEntityFactoryFn[Input, Entity],
	toOutputFn ToCreateOutputFn[Entity],
	beforeCallback func(ctx context.Context, entity Entity, input *Input) error,
	loggerFactoryFn LoggerFactoryFn,
	mutatorRepo repository.MutatorRepo[Entity],
	txManager repository.TxManager[Entity],
	options ...GenericEndpointOptions,
) *EndpointHandler[Input] {
	expectedErrors := CreateErrors
	if len(options) > 0 && options[0].ExpectedErrors != nil {
		expectedErrors = *options[0].ExpectedErrors
	}
	handler := &CreateHandler[Entity, Input]{
		createInvokeFn: CreateInvoke[Entity],
		toOutputFn:     toOutputFn,
		connFn:         getConnectionFn,
		entityFactoryFn: func(ctx context.Context, input *Input) (Entity, error) {
			return entityFactoryFn(ctx, input)
		},
		beforeCallback: beforeCallback,
		mutatorRepo:    mutatorRepo,
		txManager:      txManager,
	}
	return NewEndpointHandler(
		url,
		http.MethodPost,
		inputHandler,
		inputFactory,
		handler.Handle,
		expectedErrors,
		loggerFactoryFn,
	)
}

// CreateInvoke wraps the create database operation in a transaction.
func CreateInvoke[Entity database.Mutator](
	ctx context.Context,
	connFn repository.ConnFn,
	entity Entity,
	mutatorRepo repository.MutatorRepo[Entity],
	txManager repository.TxManager[Entity],
) (Entity, error) {
	return txManager.WithTransaction(
		ctx,
		connFn,
		func(ctx context.Context, tx database.Tx) (Entity, error) {
			return mutatorRepo.Insert(tx, entity)
		},
	)
}

// GenericGetDefinition builds the endpoint definition for a get operation.
func GenericGetDefinition[Entity database.Getter, Output any](
	url string,
	inputHandler InputHandler,
	apiToDBFields APIToDBFields,
	toOutputFn ToGetOutputFn[Entity, Output],
	connFn repository.ConnFn,
	entityFactoryFn repository.GetterFactoryFn[Entity],
	beforeCallback func(
		ctx context.Context, entity Entity, input *GetInput,
	) error,
	loggerFactoryFn LoggerFactoryFn,
	readerRepo repository.ReaderRepo[Entity],
	txManager repository.TxManager[Entity],
) *EndpointHandler[GetInput] {
	parseInputFn := func(
		input *GetInput,
	) (*ParsedGetEndpointInput, error) {
		return ParseGetEndpointInput(
			apiToDBFields,
			input.Selectors,
			input.Orders,
			input.Page,
			100,
			input.Count,
		)
	}
	handler := &GetHandler[Entity, GetInput, Output]{
		parseInputFn:    parseInputFn,
		getInvokeFn:     GetInvoke[Entity],
		toOutputFn:      toOutputFn,
		connFn:          connFn,
		entityFactoryFn: entityFactoryFn,
		beforeCallback:  beforeCallback,
		readerRepo:      readerRepo,
		txManager:       txManager,
	}
	return NewEndpointHandler(
		url,
		http.MethodGet,
		inputHandler,
		func() GetInput { return GetInput{} },
		handler.Handle,
		GetErrors,
		loggerFactoryFn,
	)
}

// GetInvoke executes the get operation.
func GetInvoke[Getter database.Getter](
	ctx context.Context,
	parsedInput *ParsedGetEndpointInput,
	connFn repository.ConnFn,
	entityFactoryFn repository.GetterFactoryFn[Getter],
	readerRepo repository.ReaderRepo[Getter],
	_ repository.TxManager[Getter],
) ([]Getter, int, error) {
	conn, err := connFn()
	if err != nil {
		return nil, 0, err
	}
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return nil, 0, err
	}
	if parsedInput.Count {
		count, err := readerRepo.Count(
			tx,
			parsedInput.Selectors,
			parsedInput.Page,
			entityFactoryFn,
		)
		if err != nil {
			return nil, 0, err
		}
		return nil, count, nil
	}
	entities, err := readerRepo.GetMany(
		tx,
		entityFactoryFn,
		&database.GetOptions{
			Selectors: parsedInput.Selectors,
			Orders:    parsedInput.Orders,
			Page:      parsedInput.Page,
		},
	)
	if err != nil {
		return nil, 0, err
	}
	return entities, len(entities), nil
}

// ParseGetEndpointInput translates API parameters to DB parameters.
func ParseGetEndpointInput(
	apiToDBFields APIToDBFields,
	selectors endpoint.Selectors,
	orders endpoint.Orders,
	inputPage *endpoint.Page,
	maxPage int,
	count bool,
) (*ParsedGetEndpointInput, error) {
	dbOrders, err := orders.TranslateToDBOrders(apiToDBFields)
	if err != nil {
		return nil, err
	}
	if inputPage == nil {
		inputPage = &endpoint.Page{Offset: 0, Limit: maxPage}
	}
	dbSelectors, err := selectors.ToDBSelectors(apiToDBFields)
	if err != nil {
		return nil, err
	}
	return &ParsedGetEndpointInput{
		Orders:    dbOrders,
		Selectors: dbSelectors,
		Page:      inputPage.ToDBPage(),
		Count:     count,
	}, nil
}

// GenericUpdateDefinition builds the endpoint definition for an update
// operation.
func GenericUpdateDefinition(
	url string,
	inputHandler InputHandler,
	apiToDBFields APIToDBFields,
	connFn repository.ConnFn,
	entityFactoryFn UpdateEntityFactoryFn,
	beforeCallback func(
		ctx context.Context,
		entity database.Mutator,
		input *UpdateInput,
	) error,
	loggerFactoryFn LoggerFactoryFn,
) *EndpointHandler[UpdateInput] {
	parseInputFn := func(
		input *UpdateInput,
	) (*ParsedUpdateEndpointInput, error) {
		return ParseUpdateEndpointInput(
			apiToDBFields,
			input.Selectors,
			input.Updates,
			input.Upsert,
		)
	}
	handler := &UpdateHandler[UpdateInput]{
		parseInputFn:    parseInputFn,
		updateInvokeFn:  UpdateInvoke,
		toOutputFn:      ToUpdateOutput,
		connFn:          connFn,
		entityFactoryFn: entityFactoryFn,
		beforeCallback:  beforeCallback,
	}
	return NewEndpointHandler(
		url,
		http.MethodPatch,
		inputHandler,
		func() UpdateInput { return UpdateInput{} },
		handler.Handle,
		UpdateErrors,
		loggerFactoryFn,
	)
}

// UpdateInvoke executes the update operation.
func UpdateInvoke(
	ctx context.Context,
	parsedInput *ParsedUpdateEndpointInput,
	connFn repository.ConnFn,
	entity database.Mutator,
	mutatorRepo repository.MutatorRepo[database.Mutator],
	txManager repository.TxManager[*int64],
) (int64, error) {
	count, err := txManager.WithTransaction(
		ctx,
		connFn,
		func(ctx context.Context, tx database.Tx) (*int64, error) {
			c, err := mutatorRepo.Update(
				tx,
				entity,
				parsedInput.Selectors,
				parsedInput.UpdateFields,
			)
			return &c, err
		})
	if err != nil {
		return 0, err
	}
	return *count, nil
}

// ParseUpdateEndpointInput translates API update input into DB update input.
func ParseUpdateEndpointInput(
	apiToDBFields APIToDBFields,
	selectors endpoint.Selectors,
	updates endpoint.Updates,
	upsert bool,
) (*ParsedUpdateEndpointInput, error) {
	dbSelectors, err := selectors.ToDBSelectors(apiToDBFields)
	if err != nil {
		return nil, err
	}
	if len(dbSelectors) == 0 {
		return nil, NeedAtLeastOneSelectorError
	}
	dbUpdates, err := updates.ToDBUpdates(apiToDBFields)
	if err != nil {
		return nil, err
	}
	if len(dbUpdates) == 0 {
		return nil, NeedAtLeastOneUpdateError
	}
	return &ParsedUpdateEndpointInput{
		Selectors:    dbSelectors,
		UpdateFields: dbUpdates,
		Upsert:       upsert,
	}, nil
}

// ToUpdateOutput wraps the update count.
func ToUpdateOutput(count int64) (any, error) {
	return &UpdateOutput{Count: count}, nil
}

// GenericDeleteDefinition builds the endpoint definition for a delete
// operation.
func GenericDeleteDefinition(
	url string,
	inputHandler InputHandler,
	apiToDBFields APIToDBFields,
	connFn repository.ConnFn,
	entityFactoryFn DeleteEntityFactoryFn,
	beforeCallback func(
		ctx context.Context,
		entity database.Mutator,
		input *DeleteInput,
	) error,
	loggerFactoryFn LoggerFactoryFn,
) *EndpointHandler[DeleteInput] {
	parseInputFn := func(
		input *DeleteInput,
	) (*ParsedDeleteEndpointInput, error) {
		return ParseDeleteEndpointInput(
			apiToDBFields,
			input.Selectors,
			nil,
			0,
		)
	}
	handler := &DeleteHandler[DeleteInput]{
		parseInputFn:    parseInputFn,
		deleteInvokeFn:  DeleteInvoke[database.Mutator],
		toOutputFn:      ToDeleteOutput,
		connFn:          connFn,
		entityFactoryFn: entityFactoryFn,
		beforeCallback:  beforeCallback,
	}
	return NewEndpointHandler(
		url,
		http.MethodDelete,
		inputHandler,
		func() DeleteInput { return DeleteInput{} },
		handler.Handle,
		DeleteErrors,
		loggerFactoryFn,
	)
}

// DeleteInvoke executes the delete operation.
func DeleteInvoke[Entity database.Mutator](
	ctx context.Context,
	parsedInput *ParsedDeleteEndpointInput,
	connFn repository.ConnFn,
	entity Entity,
	mutatorRepo repository.MutatorRepo[database.Mutator],
	txManager repository.TxManager[*int64],
) (int64, error) {
	count, err := txManager.WithTransaction(
		ctx,
		connFn,
		func(ctx context.Context, tx database.Tx) (*int64, error) {
			c, err := mutatorRepo.Delete(
				tx,
				entity,
				parsedInput.Selectors,
				parsedInput.DeleteOpts,
			)
			return &c, err
		})
	if err != nil {
		return 0, err
	}
	return *count, nil
}

// ParseDeleteEndpointInput translates API delete input into DB delete input.
func ParseDeleteEndpointInput(
	apiToDBFields APIToDBFields,
	selectors endpoint.Selectors,
	orders endpoint.Orders,
	limit int,
) (*ParsedDeleteEndpointInput, error) {
	dbSelectors, err := selectors.ToDBSelectors(apiToDBFields)
	if err != nil {
		return nil, err
	}
	if len(dbSelectors) == 0 {
		return nil, NeedAtLeastOneSelectorError
	}
	dbOrders, err := orders.TranslateToDBOrders(apiToDBFields)
	if err != nil {
		return nil, err
	}
	return &ParsedDeleteEndpointInput{
		Selectors: dbSelectors,
		DeleteOpts: &database.DeleteOptions{
			Limit:  limit,
			Orders: dbOrders,
		},
	}, nil
}

// ToDeleteOutput wraps the delete count.
func ToDeleteOutput(count int64) (any, error) {
	return &DeleteOutput{Count: count}, nil
}
