package repository

import (
	"context"
	"fmt"

	"github.com/pakkasys/fluidapi/database"
)

// DefaultReaderRepo is a concrete implementation of ReaderRepo.
type DefaultReaderRepo[Entity database.Getter] struct {
	QueryBuilder database.QueryBuilder
	ErrorChecker database.ErrorChecker
	readDBOps    *database.ReadDBOps[Entity]
	dbOps        *database.DBOps
}

// DefaultReaderRepo implements the ReaderRepo interface.
var _ ReaderRepo[database.Getter] = (*DefaultReaderRepo[database.Getter])(nil)

// NewDefaultReaderRepo returns a new DefaultReaderRepo.
//
// Parameters:
//   - queryBuilder: The query builder.
//   - errorChecker: The error checker.
//
// Returns:
//   - *DefaultReaderRepo: A new DefaultReaderRepo.
func NewDefaultReaderRepo[Entity database.Getter](
	queryBuilder database.QueryBuilder,
	errorChecker database.ErrorChecker,
) *DefaultReaderRepo[Entity] {
	return &DefaultReaderRepo[Entity]{
		QueryBuilder: queryBuilder,
		ErrorChecker: errorChecker,
		readDBOps:    database.NewReadDBOps[Entity](),
		dbOps:        database.NewDBOps(),
	}
}

// GetOne retrieves a single record from the DB. It returns an error if the
// entity is not found.
//
// Parameters:
//   - preparer: The database connection or transaction to use.
//   - entityFactoryFn: A function that returns a new instance of T.
//   - getOptions: Filter and query options for the query.
//
// Returns:
//   - T: The retrieved entity of type T.
//   - error: An error if not found or on failure.
func (r *DefaultReaderRepo[Entity]) GetOne(
	preparer database.Preparer,
	entityFactoryFn GetterFactoryFn[Entity],
	getOptions *database.GetOptions,
) (Entity, error) {
	return r.readDBOps.Get(
		preparer,
		getOptions,
		entityFactoryFn,
		r.QueryBuilder,
		r.ErrorChecker,
	)
}

// GetMany retrieves multiple records from the DB.
//
// Parameters:
//   - preparer: The database connection or transaction to use.
//   - entityFactoryFn: A function that returns a new instance of T.
//   - getOptions: Filter and query options for the query.
//
// Returns:
//   - T: A slice of retrieved entities of type T.
//   - error: An error if not found or on failure.
func (r *DefaultReaderRepo[Entity]) GetMany(
	preparer database.Preparer,
	entityFactoryFn GetterFactoryFn[Entity],
	getOptions *database.GetOptions,
) ([]Entity, error) {
	return r.readDBOps.GetMany(
		preparer,
		getOptions,
		entityFactoryFn,
		r.QueryBuilder,
		r.ErrorChecker,
	)
}

// Count returns a record count.
//
// Parameters:
//   - preparer: The database connection or transaction to use.
//   - selectors: Filter and query options for the query.
//   - page: Pagination options.
//   - entityFactoryFn: A function that returns a new instance of T.
//
// Returns:
//   - int: The count of matching records.
//   - error: An error if the query fails.
func (r *DefaultReaderRepo[Entity]) Count(
	preparer database.Preparer,
	selectors database.Selectors,
	page *database.Page,
	entityFactoryFn GetterFactoryFn[Entity],
) (int, error) {
	return r.readDBOps.Count(
		preparer,
		&database.CountOptions{
			Selectors: selectors,
			Page:      page,
		},
		entityFactoryFn,
		r.QueryBuilder,
		r.ErrorChecker,
	)
}

// Query performs a custom SQL query. It returns the results as a slice of
// entities. It returns an error if the query fails.
//
// Parameters:
//   - preparer: The database connection or transaction to use.
//   - query: The SQL query to execute.
//   - parameters: The query parameters.
//   - entityFactoryFn: A function that returns a new instance of T.
//
// Returns:
//   - T: A slice of retrieved entities of type T.
//   - error: An error if not found or on failure.
func (r *DefaultReaderRepo[Entity]) Query(
	preparer database.Preparer,
	query string,
	parameters []any,
	entityFactoryFn GetterFactoryFn[Entity],
) ([]Entity, error) {
	rows, stmt, err := r.dbOps.Query(
		preparer, query, parameters, r.ErrorChecker,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	defer stmt.Close()
	return database.RowsToEntities(rows, entityFactoryFn)
}

// DefaultMutatorRepo is the concrete implementation for mutation
// operations.
type DefaultMutatorRepo[Entity database.Mutator] struct {
	QueryBuilder database.QueryBuilder
	ErrorChecker database.ErrorChecker
	mutateDBOps  *database.MutateDBOps[Entity]
	dbOps        *database.DBOps
}

// DefaultMutatorRepo implements MutatorRepo[database.Mutator].
var _ MutatorRepo[database.Mutator] = (*DefaultMutatorRepo[database.Mutator])(nil)

// NewDefaultMutatorRepo returns a new DefaultMutatorRepo.
//
// Parameters:
//   - queryBuilder: The query builder.
//   - errorChecker: The error checker.
//
// Returns:
//   - *DefaultMutatorRepo: A new DefaultMutatorRepo.
func NewDefaultMutatorRepo[Entity database.Mutator](
	queryBuilder database.QueryBuilder,
	errorChecker database.ErrorChecker,
) *DefaultMutatorRepo[Entity] {
	return &DefaultMutatorRepo[Entity]{
		QueryBuilder: queryBuilder,
		ErrorChecker: errorChecker,
		mutateDBOps:  database.NewMutateDBOps[Entity](),
		dbOps:        database.NewDBOps(),
	}
}

// Insert inserts a record into the DB.
//
// Parameters:
//   - preparer: The database connection or transaction to use.
//   - mutator: The entity to insert.
//
// Returns:
//   - T: The inserted entity.
//   - error: An error if the insertion fails.
func (r *DefaultMutatorRepo[Entity]) Insert(
	preparer database.Preparer, mutator Entity,
) (Entity, error) {
	_, err := r.mutateDBOps.Insert(
		preparer, mutator, r.QueryBuilder, r.ErrorChecker,
	)
	if err != nil {
		var zero Entity
		return zero, err
	}
	return mutator, nil
}

// Update performs the DB update operation.
//
// Parameters:
//   - preparer: The database connection or transaction to use.
//   - updater: The entity to update.
//   - selectors: Filter and query options for the query.
//   - updates: Fields and values to update.
//
// Returns:
//   - int64: The number of updated records.
//   - error: An error if the update fails.
func (r *DefaultMutatorRepo[Entity]) Update(
	preparer database.Preparer,
	updater Entity,
	selectors database.Selectors,
	updates database.Updates,
) (int64, error) {
	return r.mutateDBOps.Update(
		preparer,
		updater,
		selectors,
		updates,
		r.QueryBuilder,
		r.ErrorChecker,
	)
}

// Delete performs the DB delete operation.
//
// Parameters:
//   - preparer: The database connection or transaction to use.
//   - deleter: The entity to delete.
//   - selectors: Filter and query options for the query.
//   - deleteOpts: Options for the delete operation.
//
// Returns:
//   - int64: The number of deleted records.
//   - error: An error if the delete fails.
func (r *DefaultMutatorRepo[Entity]) Delete(
	preparer database.Preparer,
	deleter Entity,
	selectors database.Selectors,
	deleteOpts *database.DeleteOptions,
) (int64, error) {
	return r.mutateDBOps.Delete(
		preparer,
		deleter,
		selectors,
		deleteOpts,
		r.QueryBuilder,
		r.ErrorChecker,
	)
}

// DefaultRawQueryer is a concrete implementation of the RawQueryer interface.
type DefaultRawQueryer struct{}

// DefaultRawQueryer implements RawQueryer.
var _ RawQueryer = (*DefaultRawQueryer)(nil)

// NewDefaultRawQueryer creates a new instance of DefaultRawQueryer.
func NewDefaultRawQueryer() *DefaultRawQueryer {
	return &DefaultRawQueryer{}
}

// Exec executes a query using a prepared statement.
//
// Parameters:
//   - preparer: The database connection or transaction to use.
//   - query: The SQL query string to execute.
//   - parameters: The query parameters.
//
// Returns:
//   - Result: The Result of execution.
//   - error: An error if the execution fails.
func (rq *DefaultRawQueryer) Exec(
	preparer database.Preparer, query string, parameters []any,
) (database.Result, error) {
	stmt, err := preparer.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	result, err := stmt.Exec(parameters...)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ExecRaw executes a query directly on the DB without explicit preparation.
//
// Parameters:
//   - db: The database connection (must implement Exec).
//   - query: The SQL query string to execute.
//   - parameters: The query parameters.
//
// Returns:
//   - Result: The Result of execution.
//   - error: An error if the execution fails.
func (rq *DefaultRawQueryer) ExecRaw(
	db database.DB, query string, parameters []any,
) (database.Result, error) {
	result, err := db.Exec(query, parameters...)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Query prepares and executes a query that returns rows.
//
// Parameters:
//   - preparer: The database connection or transaction to use.
//   - query: The SQL query string to execute.
//   - parameters: The query parameters.
//
// Returns:
//   - Rows: The rows of the query. Must be closed by the caller.
//   - Stmt: The prepared statement. Must be closed by the caller.
//   - error: An error if the execution fails.
func (rq *DefaultRawQueryer) Query(
	preparer database.Preparer, query string, parameters []any,
) (database.Rows, database.Stmt, error) {
	stmt, err := preparer.Prepare(query)
	if err != nil {
		return nil, nil, err
	}
	rows, err := stmt.Query(parameters...)
	if err != nil {
		// Attempt to close the statement if the query fails.
		if closeErr := stmt.Close(); closeErr != nil {
			return nil, nil, fmt.Errorf(
				"query error: %w; additionally, stmt.Close error: %v",
				err,
				closeErr,
			)
		}
		return nil, nil, err
	}
	return rows, stmt, nil
}

// QueryRaw executes a query directly on the DB without preparation.
//
// Parameters:
//   - db: The database connection (must implement Query).
//   - query: The SQL query string to execute.
//   - parameters: The query parameters.
//
// Returns:
//   - Rows: The rows of the query. Must be closed by the caller.
//   - error: An error if the execution fails.
func (rq *DefaultRawQueryer) QueryRaw(
	db database.DB, query string, parameters []any,
) (database.Rows, error) {
	rows, err := db.Query(query, parameters...)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// DefaultTxManager is the default transaction manager.
type DefaultTxManager[Entity any] struct{}

// DefaultTxManager implements the TxManager interface.
var _ TxManager[any] = (*DefaultTxManager[any])(nil)

// NewDefaultTxManager returns a new DefaultTxManager.
//
// Returns:
//   - *DefaultTxManager[Entity]: The new DefaultTxManager.
func NewDefaultTxManager[Entity any]() *DefaultTxManager[Entity] {
	return &DefaultTxManager[Entity]{}
}

// WithTransaction wraps a function call in a DB transaction.
//
// Parameters:
//   - ctx: The context for the transaction.
//   - connFn: A function that returns a DB connection.
//   - callback: The function to execute within the transaction.
//
// Returns:
//   - Entity: The result of the function call.
//   - error: An error if the transaction fails.
func (t *DefaultTxManager[Entity]) WithTransaction(
	ctx context.Context,
	connFn ConnFn,
	callback func(ctx context.Context, tx database.Tx) (Entity, error),
) (Entity, error) {
	conn, err := connFn()
	if err != nil {
		var zero Entity
		return zero, err
	}
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		var zero Entity
		return zero, err
	}
	return database.Transaction(ctx, tx, callback)
}
