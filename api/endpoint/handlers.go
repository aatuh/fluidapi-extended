package endpoint

import (
	"context"
	"net/http"

	"github.com/pakkasys/fluidapi-extended/api/repository"
	"github.com/pakkasys/fluidapi/database"
)

// GenericInvoke is the logic callback for a generic endpoint.
type GenericInvoke[Input any] func(
	w http.ResponseWriter, r *http.Request, i *Input,
) (any, error)

// GenericHandler is the handler for a generic endpoint.
type GenericHandler[Input any] struct {
	InvokeFn GenericInvoke[Input]
}

// Handle processes the generic endpoint.
func (h *GenericHandler[Input]) Handle(
	w http.ResponseWriter, r *http.Request, i *Input,
) (any, error) {
	output, err := h.InvokeFn(w, r, i)
	if err != nil {
		return nil, err
	}
	return output, nil
}

// Parsed inputs for various endpoints.
type ParsedGetEndpointInput struct {
	Selectors database.Selectors
	Orders    []database.Order
	Page      *database.Page
	Count     bool
}

// Output and invoke funcs for the create endpoint.
type CreateInvokeFn[Entity database.Mutator] func(
	ctx context.Context,
	connFn repository.ConnFn,
	entity Entity,
	mutatorRepo repository.MutatorRepo[Entity],
	txManager repository.TxManager[Entity],
) (Entity, error)

type CreateEntityFactoryFn[Input any, Entity database.Mutator] func(
	ctx context.Context, input *Input,
) (Entity, error)

type ToCreateOutputFn[Entity any] func(entity Entity) (any, error)

// CreateHandler is the handler for the create endpoint.
type CreateHandler[Entity database.Mutator, Input any] struct {
	entityFactoryFn CreateEntityFactoryFn[Input, Entity]
	createInvokeFn  CreateInvokeFn[Entity]
	toOutputFn      ToCreateOutputFn[Entity]
	connFn          repository.ConnFn
	beforeCallback  func(ctx context.Context, entity Entity, input *Input) error
	mutatorRepo     repository.MutatorRepo[Entity]
	txManager       repository.TxManager[Entity]
}

// Handle processes the create endpoint.
func (h *CreateHandler[Mutator, Input]) Handle(
	w http.ResponseWriter, r *http.Request, i *Input,
) (any, error) {
	entity, err := h.entityFactoryFn(r.Context(), i)
	if err != nil {
		return nil, err
	}
	// Call the optional callback if provided.
	if h.beforeCallback != nil {
		if err := h.beforeCallback(r.Context(), entity, i); err != nil {
			return nil, err
		}
	}
	createdEntity, err := h.createInvokeFn(
		r.Context(),
		h.connFn,
		entity,
		h.mutatorRepo,
		h.txManager,
	)
	if err != nil {
		return nil, err
	}
	return h.toOutputFn(createdEntity)
}

// Invoke and output funcs for the get endpoint.
type GetInvokeFn[Entity database.Getter] func(
	ctx context.Context,
	parsedInput *ParsedGetEndpointInput,
	connFn repository.ConnFn,
	entityFactoryFn repository.GetterFactoryFn[Entity],
	readerRepo repository.ReaderRepo[Entity],
	txManager repository.TxManager[Entity],
) ([]Entity, int, error)

type ToGetOutputFn[Entity any, Output any] func(
	entities []Entity, count int,
) (*Output, error)

// GetHandler is the handler for the get endpoint.
type GetHandler[Entity database.Getter, Input any, Output any] struct {
	parseInputFn    func(input *Input) (*ParsedGetEndpointInput, error)
	getInvokeFn     GetInvokeFn[Entity]
	toOutputFn      ToGetOutputFn[Entity, Output]
	connFn          repository.ConnFn
	entityFactoryFn repository.GetterFactoryFn[Entity]
	beforeCallback  func(ctx context.Context, entity Entity, input *Input) error
	readerRepo      repository.ReaderRepo[Entity]
	txManager       repository.TxManager[Entity]
}

// Handle processes the get endpoint.
func (h *GetHandler[Entity, Input, Output]) Handle(
	w http.ResponseWriter, r *http.Request, i *Input,
) (any, error) {
	parsedInput, err := h.parseInputFn(i)
	if err != nil {
		return nil, err
	}
	// Instantiate an entity instance.
	entity := h.entityFactoryFn()
	// If the before callback is provided, call it.
	if h.beforeCallback != nil {
		if err := h.beforeCallback(r.Context(), entity, i); err != nil {
			return nil, err
		}
	}
	entities, count, err := h.getInvokeFn(
		r.Context(),
		parsedInput,
		h.connFn,
		h.entityFactoryFn,
		h.readerRepo,
		h.txManager,
	)
	if err != nil {
		return nil, err
	}
	return h.toOutputFn(entities, count)
}

type ParsedUpdateEndpointInput struct {
	Selectors    database.Selectors
	UpdateFields []database.UpdateField
	Upsert       bool
}

// Invoke and output funcs for the update endpoint.
type ToUpdateOutputFn func(count int64) (any, error)
type UpdateEntityFactoryFn func() database.Mutator

type UpdateInvokeFn func(
	ctx context.Context,
	parsedInput *ParsedUpdateEndpointInput,
	connFn repository.ConnFn,
	updater database.Mutator,
	mutatorRepo repository.MutatorRepo[database.Mutator],
	txManager repository.TxManager[*int64],
) (int64, error)

// UpdateHandler is the handler for the update endpoint.
type UpdateHandler[Input any] struct {
	parseInputFn    func(input *Input) (*ParsedUpdateEndpointInput, error)
	updateInvokeFn  UpdateInvokeFn
	toOutputFn      ToUpdateOutputFn
	connFn          repository.ConnFn
	entityFactoryFn UpdateEntityFactoryFn
	beforeCallback  func(
		ctx context.Context,
		entity database.Mutator,
		input *Input,
	) error
	mutatorRepo repository.MutatorRepo[database.Mutator]
	txManager   repository.TxManager[*int64]
}

// Handle processes the update endpoint.
func (h *UpdateHandler[Input]) Handle(
	w http.ResponseWriter, r *http.Request, i *Input,
) (any, error) {
	parsedInput, err := h.parseInputFn(i)
	if err != nil {
		return nil, err
	}
	// Create the updater entity.
	entity := h.entityFactoryFn()
	// Call the optional callback.
	if h.beforeCallback != nil {
		if err := h.beforeCallback(r.Context(), entity, i); err != nil {
			return nil, err
		}
	}
	count, err := h.updateInvokeFn(
		r.Context(), parsedInput, h.connFn, entity, h.mutatorRepo, h.txManager,
	)
	if err != nil {
		return nil, err
	}
	return h.toOutputFn(count)
}

type ParsedDeleteEndpointInput struct {
	Selectors  database.Selectors
	DeleteOpts *database.DeleteOptions
}

// Invoke and output funcs for the delete endpoint.
type ToDeleteOutputFn func(count int64) (any, error)
type DeleteEntityFactoryFn func() database.Mutator
type DeleteInvokeFn func(
	ctx context.Context,
	parsedInput *ParsedDeleteEndpointInput,
	connFn repository.ConnFn,
	entity database.Mutator,
	mutatorRepo repository.MutatorRepo[database.Mutator],
	txManager repository.TxManager[*int64],
) (int64, error)

// DeleteHandler is the handler for the delete endpoint.
type DeleteHandler[Input any] struct {
	parseInputFn    func(input *Input) (*ParsedDeleteEndpointInput, error)
	deleteInvokeFn  DeleteInvokeFn
	toOutputFn      ToDeleteOutputFn
	connFn          repository.ConnFn
	entityFactoryFn DeleteEntityFactoryFn
	beforeCallback  func(
		ctx context.Context,
		entity database.Mutator,
		input *Input,
	) error
	mutatorRepo repository.MutatorRepo[database.Mutator]
	txManager   repository.TxManager[*int64]
}

// Handle processes the delete endpoint.
func (h *DeleteHandler[Input]) Handle(
	w http.ResponseWriter, r *http.Request, i *Input,
) (any, error) {
	parsedInput, err := h.parseInputFn(i)
	if err != nil {
		return nil, err
	}
	// Create the deleter entity.
	entity := h.entityFactoryFn()
	// Call the optional callback.
	if h.beforeCallback != nil {
		if err := h.beforeCallback(r.Context(), entity, i); err != nil {
			return nil, err
		}
	}
	count, err := h.deleteInvokeFn(
		r.Context(),
		parsedInput,
		h.connFn,
		entity,
		h.mutatorRepo,
		h.txManager,
	)
	if err != nil {
		return nil, err
	}
	return h.toOutputFn(count)
}
