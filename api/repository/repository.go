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
type ReaderRepo[Entity database.Getter] interface {
	// GetOne retrieves a single record from the DB.
	GetOne(
		preparer database.Preparer,
		entityFactoryFn GetterFactoryFn[Entity],
		getOptions *database.GetOptions,
	) (Entity, error)

	// GetMany retrieves multiple records from the DB.
	GetMany(
		preparer database.Preparer,
		entityFactoryFn GetterFactoryFn[Entity],
		getOptions *database.GetOptions,
	) ([]Entity, error)

	// Count returns a record count.
	Count(
		preparer database.Preparer,
		selectors database.Selectors,
		page *database.Page,
		entityFactoryFn GetterFactoryFn[Entity],
	) (int, error)
}

// MutatorRepo defines mutation-related operations.
type MutatorRepo[Entity database.Mutator] interface {
	Insert(preparer database.Preparer, mutator Entity) (Entity, error)
	Update(
		preparer database.Preparer,
		updater Entity,
		selectors database.Selectors,
		updates database.Updates,
	) (int64, error)
	Delete(
		preparer database.Preparer,
		deleter Entity,
		selectors database.Selectors,
		deleteOpts *database.DeleteOptions,
	) (int64, error)
}

// RawQueryer defines generic methods for executing raw queries and commands.
type RawQueryer interface {
	// Exec executes a query using a prepared statement that does not return
	// rows.
	Exec(
		preparer database.Preparer, query string, parameters []any,
	) (database.Result, error)

	// ExecRaw executes a query directly on the DB without explicit preparation.
	ExecRaw(
		db database.DB, query string, parameters []any,
	) (database.Result, error)

	// Query prepares and executes a query that returns rows. Returns both the
	// rows and the statement. The caller is responsible for closing both.
	Query(
		preparer database.Preparer, query string, parameters []any,
	) (database.Rows, database.Stmt, error)

	// QueryRaw executes a query directly on the DB without preparation and
	// returns rows. The caller is responsible for closing the returned rows.
	QueryRaw(db database.DB, query string, parameters []any,
	) (database.Rows, error)
}

// TxManager is an interface for transaction management.
type TxManager[Entity any] interface {
	// WithTransaction wraps a function call in a DB transaction.
	WithTransaction(
		ctx context.Context,
		connFn ConnFn,
		callback func(ctx context.Context, tx database.Tx) (Entity, error),
	) (Entity, error)
}
