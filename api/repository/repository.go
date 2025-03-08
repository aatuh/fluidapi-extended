package repository

import (
	"context"

	"github.com/pakkasys/fluidapi/database"
)

// ConnFn returns a database connection.
type ConnFn func() (database.DB, error)

// GetterFactoryFn returns a database.Getter factory function.
type GetterFactoryFn[Entity database.Getter] func() Entity

// ReaderRepo defines retrieval-related operations.
type ReaderRepo[T database.Getter] interface {
	// GetOne retrieves a single record from the DB.
	GetOne(
		preparer database.Preparer,
		entityFactoryFn GetterFactoryFn[T],
		getOptions *database.GetOptions,
	) (T, error)

	// GetMany retrieves multiple records from the DB.
	GetMany(
		preparer database.Preparer,
		entityFactoryFn GetterFactoryFn[T],
		getOptions *database.GetOptions,
	) ([]T, error)

	// Count returns a record count.
	Count(
		preparer database.Preparer,
		selectors database.Selectors,
		page *database.Page,
		entityFactoryFn GetterFactoryFn[T],
	) (int, error)
}

// DefaultReaderRepo is a concrete implementation of ReaderRepo.
type DefaultReaderRepo[T database.Getter] struct {
	QueryBuilder database.QueryBuilder
	ErrorChecker database.ErrorChecker
}

// NewDefaultReaderRepo returns a new DefaultReaderRepo.
func NewDefaultReaderRepo[T database.Getter](
	qb database.QueryBuilder,
	ec database.ErrorChecker,
) *DefaultReaderRepo[T] {
	return &DefaultReaderRepo[T]{
		QueryBuilder: qb,
		ErrorChecker: ec,
	}
}

// GetOne retrieves a single record from the DB.
func (r *DefaultReaderRepo[T]) GetOne(
	preparer database.Preparer,
	entityFactoryFn GetterFactoryFn[T],
	getOptions *database.GetOptions,
) (T, error) {
	return database.Get(preparer, getOptions, entityFactoryFn,
		r.QueryBuilder, r.ErrorChecker)
}

// GetMany retrieves multiple records from the DB.
func (r *DefaultReaderRepo[T]) GetMany(
	preparer database.Preparer,
	entityFactoryFn GetterFactoryFn[T],
	getOptions *database.GetOptions,
) ([]T, error) {
	return database.GetMany(preparer, getOptions, entityFactoryFn,
		r.QueryBuilder, r.ErrorChecker)
}

// Count returns a record count.
func (r *DefaultReaderRepo[T]) Count(
	preparer database.Preparer,
	selectors database.Selectors,
	page *database.Page,
	entityFactoryFn GetterFactoryFn[T],
) (int, error) {
	return database.Count(
		preparer,
		&database.CountOptions{
			Selectors: selectors,
			Page:      page,
		},
		entityFactoryFn,
		r.QueryBuilder,
	)
}

// MutatorRepo defines mutation-related operations.
type MutatorRepo[T database.Mutator] interface {
	Insert(preparer database.Preparer, mutator T) (T, error)
	Update(
		preparer database.Preparer,
		updater T,
		selectors database.Selectors,
		updateFields database.UpdateFields,
	) (int64, error)
	Delete(
		preparer database.Preparer,
		deleter T,
		selectors database.Selectors,
		deleteOpts *database.DeleteOptions,
	) (int64, error)
}

// DefaultMutatorRepo is the concrete implementation for mutation
// operations.
type DefaultMutatorRepo[T database.Mutator] struct {
	QueryBuilder database.QueryBuilder
	ErrorChecker database.ErrorChecker
}

// NewDefaultMutatorRepo returns a new DefaultMutatorRepo.
func NewDefaultMutatorRepo[T database.Mutator](
	queryBuilder database.QueryBuilder,
	errorChecker database.ErrorChecker,
) *DefaultMutatorRepo[T] {
	return &DefaultMutatorRepo[T]{
		QueryBuilder: queryBuilder,
		ErrorChecker: errorChecker,
	}
}

// Insert inserts a record into the DB.
func (r *DefaultMutatorRepo[T]) Insert(
	preparer database.Preparer, mutator T,
) (T, error) {
	_, err := database.Insert(
		preparer, mutator, r.QueryBuilder, r.ErrorChecker,
	)
	if err != nil {
		var zero T
		return zero, err
	}
	return mutator, nil
}

// Update performs the DB update operation.
func (r *DefaultMutatorRepo[T]) Update(
	preparer database.Preparer,
	updater T,
	selectors database.Selectors,
	updateFields database.UpdateFields,
) (int64, error) {
	return database.Update(
		preparer, updater, selectors, updateFields,
		r.QueryBuilder, r.ErrorChecker,
	)
}

// Delete performs the DB delete operation.
func (r *DefaultMutatorRepo[T]) Delete(
	preparer database.Preparer,
	deleter T,
	selectors database.Selectors,
	deleteOpts *database.DeleteOptions,
) (int64, error) {
	return database.Delete(
		preparer, deleter, selectors, deleteOpts,
		r.QueryBuilder,
	)
}

// TxManager is an interface for transaction management.
type TxManager[T any] interface {
	WithTransaction(
		ctx context.Context,
		connFn ConnFn,
		callback func(ctx context.Context, tx database.Tx) (T, error),
	) (T, error)
}

// DefaultTxManager is the default transaction manager.
type DefaultTxManager[T any] struct{}

// NewDefaultTxManager returns a new DefaultTxManager.
func NewDefaultTxManager[T any]() *DefaultTxManager[T] {
	return &DefaultTxManager[T]{}
}

// WithTransaction wraps a function call in a DB transaction.
func (t *DefaultTxManager[T]) WithTransaction(
	ctx context.Context,
	connFn ConnFn,
	callback func(ctx context.Context, tx database.Tx) (T, error),
) (T, error) {
	conn, err := connFn()
	if err != nil {
		var zero T
		return zero, err
	}
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		var zero T
		return zero, err
	}
	return database.Transaction(ctx, tx, callback)
}
