package sqlitequery

import (
	"github.com/pakkasys/fluidapi/database"
)

// Builder is a query builder for MySQL.
type Builder struct{}

// Insert delegates to the package-level Insert function.
func (b *Builder) Insert(
	tableName string,
	insertedValues database.InsertedValues,
) (string, []any) {
	return Insert(tableName, insertedValues)
}

// InsertMany delegates to the package-level InsertMany function.
func (b *Builder) InsertMany(
	tableName string,
	insertedValues []database.InsertedValues,
) (string, []any) {
	return InsertMany(tableName, insertedValues)
}

// UpsertMany delegates to the package-level UpsertMany function.
func (b *Builder) UpsertMany(
	tableName string,
	insertedValues []database.InsertedValues,
	updateProjections []database.Projection,
) (string, []any) {
	return UpsertMany(tableName, insertedValues, updateProjections)
}

// Get delegates to the package-level Get function.
func (b *Builder) Get(
	tableName string,
	options *database.GetOptions,
) (string, []any) {
	return Get(tableName, options)
}

// Count delegates to the package-level Count function.
func (b *Builder) Count(
	tableName string,
	options *database.CountOptions,
) (string, []any) {
	return Count(tableName, options)
}

// UpdateQuery delegates to the package-level UpdateQuery function.
func (b *Builder) UpdateQuery(
	tableName string,
	updateFields []database.UpdateField,
	selectors []database.Selector,
) (string, []any) {
	return UpdateQuery(tableName, updateFields, selectors)
}

// Delete delegates to the package-level Delete function.
func (b *Builder) Delete(
	tableName string,
	selectors []database.Selector,
	opts *database.DeleteOptions,
) (string, []any) {
	return Delete(tableName, selectors, opts)
}

// CreateDatabaseQuery delegates to the package-level CreateDatabaseQuery function.
func (b *Builder) CreateDatabaseQuery(
	dbName string,
	ifNotExists bool,
	charset string,
	collate string,
) (string, []any, error) {
	return CreateDatabaseQuery(dbName, ifNotExists, charset, collate)
}

// CreateTableQuery delegates to the package-level CreateTableQuery function.
func (b *Builder) CreateTableQuery(
	tableName string,
	ifNotExists bool,
	columns []database.ColumnDefinition,
	constraints []string,
	options database.TableOptions,
) (string, []any, error) {
	return CreateTableQuery(
		tableName,
		ifNotExists,
		columns,
		constraints,
		options,
	)
}

// UseDatabaseQuery delegates to the package-level UseDatabaseQuery function.
func (b *Builder) UseDatabaseQuery(dbName string) (string, []any, error) {
	return UseDatabaseQuery(dbName)
}

// SetVariableQuery delegates to the package-level SetVariableQuery function.
func (b *Builder) SetVariableQuery(
	variable string,
	value string,
) (string, []any, error) {
	return SetVariableQuery(variable, value)
}

// AdvisoryUnlock delegates to the package-level AdvisoryUnlock function.
func (b *Builder) AdvisoryUnlock(lockName string) (string, []any, error) {
	return AdvisoryUnlock(lockName)
}

// AdvisoryLock delegates to the package-level AdvisoryLock function.
func (b *Builder) AdvisoryLock(
	lockName string,
	timeout int,
) (string, []any, error) {
	return AdvisoryLock(lockName, timeout)
}
