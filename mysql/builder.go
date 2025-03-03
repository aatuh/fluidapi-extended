package mysql

import (
	"github.com/pakkasys/fluidapi/database"
)

// Builder is a query builder for MySQL.
type Builder struct{}

// Insert delegates to the package-level Insert function.
func (b *Builder) Insert(
	tableName string,
	insertedValues database.InsertedValuesFn,
) (string, []any) {
	return Insert(tableName, insertedValues)
}

// InsertMany delegates to the package-level InsertMany function.
func (b *Builder) InsertMany(
	tableName string,
	insertedValues []database.InsertedValuesFn,
) (string, []any) {
	return InsertMany(tableName, insertedValues)
}

// UpsertMany delegates to the package-level UpsertMany function.
func (b *Builder) UpsertMany(
	tableName string,
	insertedValues []database.InsertedValuesFn,
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
