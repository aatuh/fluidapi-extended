package sqlite

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/pakkasys/fluidapi/database"
)

// Constants for comparison operators.
const (
	in    = "IN"
	is    = "IS"
	isNot = "IS NOT"
	null  = "NULL"
)

// Constants for SET statements.
const (
	SET_JOURNAL_MODE = "JOURNAL_MODE"
	SET_FOREIGN_KEYS = "FOREIGN_KEYS"
)

// Query is a query builder for MySQL.
type Query struct{}

// Insert returns the query and values to insert an entity.
//
// Parameters:
//   - tableName: The name of the table.
//   - insertedValues: A function that returns the columns and values to insert.
//
// Returns:
//   - string: The query.
//   - []any: The values.
func (q *Query) Insert(
	tableName string,
	insertedValues database.InsertedValuesFn,
) (string, []any) {
	columns, values := insertedValues()
	columnames := getInsertQueryColumnames(columns)
	valuePlaceholders := strings.TrimSuffix(
		strings.Repeat("?, ", len(values)),
		", ",
	)
	query := fmt.Sprintf(
		"INSERT INTO \"%s\" (%s) VALUES (%s)",
		tableName,
		columnames,
		valuePlaceholders,
	)
	return query, values
}

// InsertMany returns the query and values to insert multiple entities.
//
// Parameters:
//   - tableName: The name of the table.
//   - insertedValues: A slice of functions that return the columns and values
//     to insert.
//
// Returns:
//   - string: The query.
//   - []any: The values.
func (q *Query) InsertMany(
	tableName string,
	insertedValues []database.InsertedValuesFn,
) (string, []any) {
	if len(insertedValues) == 0 {
		return "", nil
	}

	columns, _ := insertedValues[0]()
	columnames := getInsertQueryColumnames(columns)

	var allValues []any
	valuePlaceholders := make([]string, len(insertedValues))
	for i, iterInsertedValues := range insertedValues {
		_, values := iterInsertedValues()
		placeholders := make([]string, len(values))
		for j := range values {
			placeholders[j] = "?"
		}
		valuePlaceholders[i] = "(" + strings.Join(placeholders, ", ") + ")"
		allValues = append(allValues, values...)
	}

	query := fmt.Sprintf(
		"INSERT INTO \"%s\" (%s) VALUES %s",
		tableName,
		columnames,
		strings.Join(valuePlaceholders, ", "),
	)
	return query, allValues
}

// UpsertMany creates an upsert query for a list of entities.
// Note: SQLite does not support partial upsert like MySQL. This
// implementation uses "INSERT OR REPLACE" which replaces the entire row.
//
// Parameters:
//   - tableName: The name of the table.
//   - insertedValues: A slice of functions that return the columns and values
//     to insert.
//   - updateProjections: A slice of functions that return the columns and
//     values to update.
//
// Returns:
//   - string: The query.
//   - []any: The values.
func (q *Query) UpsertMany(
	tableName string,
	insertedValues []database.InsertedValuesFn,
	updateProjections []database.Projection,
) (string, []any) {
	if len(insertedValues) == 0 {
		return "", nil
	}

	// Get columns from the first row; assume they're consistent for all rows.
	cols, _ := insertedValues[0]()

	// Build the multi-row VALUES placeholders and gather all parameters.
	var allValues []any
	var placeholdersArr []string

	for _, iv := range insertedValues {
		_, rowVals := iv()

		// Create a list of "?" placeholders matching the row length.
		placeholder := make([]string, len(rowVals))
		for i := range placeholder {
			placeholder[i] = "?"
		}

		placeholdersArr = append(
			placeholdersArr,
			fmt.Sprintf("(%s)", strings.Join(placeholder, ", ")),
		)
		allValues = append(allValues, rowVals...)
	}

	// Base INSERT statement (no semicolon yet).
	insertQuery := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES %s",
		tableName,
		strings.Join(cols, ", "),
		strings.Join(placeholdersArr, ", "),
	)

	// Build the DO UPDATE SET clause. Skip "id" and "created" so they're not
	// overwritten if a row with the same 'id' already exists.
	var sets []string
	for _, col := range cols {
		if col == "id" || col == "created" {
			continue
		}
		sets = append(
			sets,
			fmt.Sprintf("%s=excluded.%s", col, col),
		)
	}

	projectedColumns := []string{}
	for _, proj := range updateProjections {
		projectedColumns = append(projectedColumns, proj.Column)
	}

	// If your primary key is something else (or composite), replace (id).
	conflictClause := fmt.Sprintf(
		"ON CONFLICT(%s) DO UPDATE SET %s",
		strings.Join(projectedColumns, ", "),
		strings.Join(sets, ", "),
	)

	// Final upsert query.
	upsertQuery := fmt.Sprintf("%s %s", insertQuery, conflictClause)

	return upsertQuery, allValues
}

// Get returns a SELECT query for retrieving entities.
//
// Parameters:
//   - tableName: The name of the table.
//   - opts: The database options.
//
// Returns:
//   - string: The query.
//   - []any: The values.
func (q *Query) Get(
	tableName string,
	opts *database.GetOptions,
) (string, []any) {
	whereClause, whereValues := whereClause(opts.Selectors)

	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf(
		"SELECT %s",
		strings.Join(projectionsToStrings(opts.Projections), ","),
	))
	builder.WriteString(fmt.Sprintf(" FROM \"%s\"", tableName))
	if len(opts.Joins) != 0 {
		builder.WriteString(" " + joinClause(opts.Joins))
	}
	if whereClause != "" {
		builder.WriteString(" " + whereClause)
	}
	if len(opts.Orders) != 0 {
		builder.WriteString(" " + getOrderClauseFromOrders(opts.Orders))
	}
	if opts.Page != nil {
		builder.WriteString(" " + getLimitOffsetClauseFromPage(opts.Page))
	}
	// SQLite does not support FOR UPDATE locking.
	return builder.String(), whereValues
}

// Count returns a query to count the entities.
//
// Parameters:
//   - tableName: The name of the table.
//   - opts: The database options.
//
// Returns:
//   - string: The query.
//   - []any: The values.
func (q *Query) Count(
	tableName string,
	opts *database.CountOptions,
) (string, []any) {
	whereClause, whereValues := whereClause(opts.Selectors)
	joinStmt := joinClause(opts.Joins)

	// If pagination is provided, wrap the limited query in a subquery.
	if opts.Page != nil {
		// Build the inner query.
		innerBuilder := strings.Builder{}
		innerBuilder.WriteString(fmt.Sprintf(
			"SELECT * FROM \"%s\" %s %s %s",
			tableName,
			joinStmt,
			whereClause,
			getLimitOffsetClauseFromPage(opts.Page),
		))
		innerQuery := innerBuilder.String()

		// Wrap the inner query with an outer COUNT(*)
		query := fmt.Sprintf(
			"SELECT COUNT(*) FROM (%s) AS limited_result",
			innerQuery,
		)
		return query, whereValues
	}

	// Otherwise, build a simple query without pagination.
	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf(
		"SELECT COUNT(*) FROM \"%s\" %s %s",
		tableName,
		joinStmt,
		whereClause,
	))
	query := builder.String()
	return query, whereValues
}

// UpdateQuery returns the query and values for an UPDATE.
//
// Parameters:
//   - tableName: The name of the table.
//   - updates: The fields to update.
//   - selectors: The selectors for the entities to update.
//
// Returns:
//   - string: The query.
//   - []any: The values.
func (q *Query) UpdateQuery(
	tableName string,
	updates []database.Update,
	selectors []database.Selector,
) (string, []any) {
	whereColumns, whereValues := processSelectors(selectors)
	setClause, values := getSetClause(updates)
	values = append(values, whereValues...)

	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf(
		"UPDATE \"%s\" SET %s",
		tableName,
		setClause,
	))
	if len(whereColumns) != 0 {
		builder.WriteString(" " + getWhereClause(whereColumns))
	}

	return builder.String(), values
}

// Delete returns the query and values to delete entities.
//
// Parameters:
//   - tableName: The name of the table.
//   - selectors: The selectors for the entities to delete.
//   - opts: The delete options.
//
// Returns:
//   - string: The query.
//   - []any: The values.
func (q *Query) Delete(
	tableName string,
	selectors []database.Selector,
	opts *database.DeleteOptions,
) (string, []any) {
	whereColumns, whereValues := processSelectors(selectors)

	whereClause := ""
	if len(whereColumns) > 0 {
		whereClause = "WHERE " + strings.Join(whereColumns, " AND ")
	}

	builder := strings.Builder{}
	builder.WriteString(
		fmt.Sprintf("DELETE FROM \"%s\" %s", tableName, whereClause),
	)

	if opts != nil {
		writeDeleteOptions(&builder, opts)
	}

	return builder.String(), whereValues
}

// CreateDatabaseQuery for SQLite returns an empty string.
//
// Returns:
//   - string: The query. It is an empty string.
//   - []any: The values. It is nil.
//   - error: No error.
func (q *Query) CreateDatabaseQuery(
	_ string,
	_ bool,
	_ string,
	_ string,
) (string, []any, error) {
	return "", nil, nil
}

// CreateTableQuery returns the query and values to create a table.
//
// Parameters:
//   - tableName: The name of the table.
//   - ifNotExists: Whether to create the table if it already exists.
//   - columns: The columns to create.
//   - constraints: The constraints to add to the table.
//   - opts: The table options.
//
// Returns:
//   - string: The query.
//   - []any: The values.
//   - error: An error if the query could not be created.
func (q *Query) CreateTableQuery(
	tableName string,
	ifNotExists bool,
	columns []database.ColumnDefinition,
	constraints []string,
	opts database.TableOptions,
) (string, []any, error) {
	var builder strings.Builder

	builder.WriteString("CREATE TABLE ")
	if ifNotExists {
		builder.WriteString("IF NOT EXISTS ")
	}
	builder.WriteString(fmt.Sprintf("`%s` (\n", tableName))

	var defs []string
	// Build column definitions.
	for _, col := range columns {
		def := fmt.Sprintf("  `%s` %s", col.Name, col.Type)
		if col.Extra != "" {
			// SQLite does not support specifying character set or collation.
			def += " " + col.Extra
		}
		if col.NotNull {
			def += " NOT NULL"
		} // SQLite columns are nullable by default.
		if col.Default != nil {
			def += " DEFAULT "
			if *col.Default == "CURRENT_TIMESTAMP" || *col.Default == "NULL" {
				def += *col.Default
			} else {
				def += fmt.Sprintf("'%s'", *col.Default)
			}
		}
		if col.AutoIncrement {
			// In SQLite, the auto-increment column must be an INTEGER PRIMARY KEY.
			def += " AUTOINCREMENT"
		}
		if col.PrimaryKey {
			def += " PRIMARY KEY"
		}
		if col.Unique && !col.PrimaryKey {
			def += " UNIQUE"
		}
		defs = append(defs, def)
	}
	// Append additional table constraints if provided.
	for _, constraint := range constraints {
		defs = append(defs, "  "+constraint)
	}
	builder.WriteString(strings.Join(defs, ",\n"))
	builder.WriteString("\n)")

	// Table options like ENGINE, CHARSET, or COLLATE do not apply in
	// SQLite and are ignored.

	builder.WriteString(";")
	return builder.String(), nil, nil
}

// UseDatabaseQuery is effectively a no-op for SQLite.
// Since SQLite uses file-based databases, the concept of switching
// databases isn’t applicable.
//
// Returns:
//   - string: The query. It is an empty string.
//   - []any: The values. It is nil.
//   - error: No error.
func (q *Query) UseDatabaseQuery(_ string) (string, []any, error) {
	return "", nil, nil
}

// SetVariableQuery for SQLite might support only a subset of options.
//
// Parameters:
//   - variable: The name of the variable to set.
//   - value: The value to set the variable to.
//
// Returns:
//   - string: The query.
//   - []any: The values.
//   - error: An error if the query could not be created.
func (q *Query) SetVariableQuery(
	variable string, value string,
) (string, []any, error) {
	upperVar := strings.ToUpper(variable)
	switch upperVar {
	case SET_JOURNAL_MODE:
		if value != "DELETE" && value != "TRUNCATE" && value != "PERSIST" {
			return "", nil, fmt.Errorf(
				"Invalid value for SET_JOURNAL_MODE: %s",
				value,
			)
		}
		return fmt.Sprintf("PRAGMA journal_mode = '%s';", value), nil, nil
	case SET_FOREIGN_KEYS:
		if value != "ON" && value != "OFF" {
			return "", nil, fmt.Errorf(
				"Invalid value for SET_FOREIGN_KEYS: %s",
				value,
			)
		}
		return fmt.Sprintf("PRAGMA foreign_keys = %s;", value), nil, nil
	default:
		return "", nil, fmt.Errorf(
			"SET %s is not supported by SQLite",
			variable,
		)
	}
}

// getOrderClauseFromOrders returns an ORDER BY clause.
func getOrderClauseFromOrders(orders []database.Order) string {
	if len(orders) == 0 {
		return ""
	}

	orderClause := "ORDER BY"
	for _, order := range orders {
		if order.Table == "" {
			orderClause += fmt.Sprintf(" \"%s\" %s,", order.Field,
				order.Direction)
		} else {
			orderClause += fmt.Sprintf(" \"%s\".\"%s\" %s,", order.Table,
				order.Field, order.Direction)
		}
	}

	return strings.TrimSuffix(orderClause, ",")
}

// AdvisoryLock for SQLite returns an empty string.
func (q *Query) AdvisoryLock(lockName string, timeout int) (string, []any, error) {
	return "", nil, nil
}

// AdvisoryUnlock for SQLite returns an empty string.
func (q *Query) AdvisoryUnlock(lockName string) (string, []any, error) {
	return "", nil, nil
}

// getLimitOffsetClauseFromPage returns a LIMIT/OFFSET clause.
func getLimitOffsetClauseFromPage(page *database.Page) string {
	if page == nil {
		return ""
	}
	return fmt.Sprintf("LIMIT %d OFFSET %d", page.Limit, page.Offset)
}

// columnSelectorToString returns the string representation of a column
// selector.
func columnSelectorToString(colSel database.ColumnSelector) string {
	return fmt.Sprintf("\"%s\".\"%s\"",
		colSel.Table,
		colSel.Column,
	)
}

// processSelectors processes selectors and returns conditions and values.
func processSelectors(selectors []database.Selector) ([]string, []any) {
	var whereColumns []string
	var whereValues []any
	for _, selector := range selectors {
		col, vals := processSelector(selector)
		whereColumns = append(whereColumns, col)
		whereValues = append(whereValues, vals...)
	}
	return whereColumns, whereValues
}

// projectionToString returns the string representation of a projection.
func projectionToString(proj database.Projection) string {
	builder := strings.Builder{}

	if proj.Table == "" {
		builder.WriteString(fmt.Sprintf("\"%s\"", proj.Column))
	} else {
		builder.WriteString(fmt.Sprintf("\"%s\".\"%s\"",
			proj.Table,
			proj.Column,
		))
	}

	if proj.Alias != "" {
		builder.WriteString(fmt.Sprintf(" AS \"%s\"", proj.Alias))
	}

	return builder.String()
}

// getInsertQueryColumnames returns the string representation of column names.
func getInsertQueryColumnames(columns []string) string {
	wrappedColumns := make([]string, len(columns))
	for i, col := range columns {
		wrappedColumns[i] = "\"" + col + "\""
	}
	return strings.Join(wrappedColumns, ", ")
}

// projectionsToStrings returns the string representations of projections.
func projectionsToStrings(projections []database.Projection) []string {
	if len(projections) == 0 {
		return []string{"*"}
	}

	projStrings := make([]string, len(projections))
	for i, proj := range projections {
		projStrings[i] = projectionToString(proj)
	}
	return projStrings
}

// joinClause returns the string representation of a join clause.
func joinClause(joins []database.Join) string {
	var clause string
	for _, join := range joins {
		if clause != "" {
			clause += " "
		}
		clause += fmt.Sprintf(
			"%s JOIN \"%s\" ON %s = %s",
			join.Type,
			join.Table,
			columnSelectorToString(join.OnLeft),
			columnSelectorToString(join.OnRight),
		)
	}
	return clause
}

// whereClause returns the string representation of a WHERE clause.
func whereClause(selectors []database.Selector) (string, []any) {
	whereCols, whereVals := processSelectors(selectors)

	var clause string
	if len(whereCols) > 0 {
		clause = "WHERE " + strings.Join(whereCols, " AND ")
	}

	return strings.Trim(clause, " "), whereVals
}

// getWhereClause returns the string representation of a WHERE clause.
func getWhereClause(whereColumns []string) string {
	if len(whereColumns) > 0 {
		return "WHERE " + strings.Join(whereColumns, " AND ")
	}
	return ""
}

// getSetClause returns the string representation of a SET clause.
func getSetClause(updates []database.Update) (string, []any) {
	setParts := make([]string, len(updates))
	values := make([]any, len(updates))

	for i, update := range updates {
		setParts[i] = fmt.Sprintf("%s = ?", update.Field)
		values[i] = update.Value
	}

	return strings.Join(setParts, ", "), values
}

// writeDeleteOptions writes the delete options to the builder.
func writeDeleteOptions(
	builder *strings.Builder,
	opts *database.DeleteOptions,
) {
	orderClause := getOrderClauseFromOrders(opts.Orders)
	if orderClause != "" {
		builder.WriteString(" " + orderClause)
	}

	if opts.Limit > 0 {
		builder.WriteString(fmt.Sprintf(" LIMIT %d", opts.Limit))
	}
}

// processSelector processes a selector and returns conditions and values.
func processSelector(selector database.Selector) (string, []any) {
	if selector.Predicate == in {
		return processInSelector(selector)
	}
	return processDefaultSelector(selector)
}

// processInSelector processes an IN selector and returns conditions and values.
func processInSelector(selector database.Selector) (string, []any) {
	var col string
	if selector.Table != "" {
		col = fmt.Sprintf("\"%s\".\"%s\"", selector.Table, selector.Column)
	} else {
		col = fmt.Sprintf("\"%s\"", selector.Column)
	}

	value := reflect.ValueOf(selector.Value)
	if value.Kind() == reflect.Slice {
		placeholders, values := createPlaceholdersAndValues(value)
		return fmt.Sprintf("%s %s (%s)", col, in, placeholders), values
	}
	// If value is not a slice, treat it as a single value.
	return fmt.Sprintf("%s %s (?)", col, in), []any{selector.Value}
}

// processDefaultSelector processes a default selector and returns conditions
// and values.
func processDefaultSelector(selector database.Selector) (string, []any) {
	if selector.Value == nil {
		return processNullSelector(selector)
	}
	if selector.Table == "" {
		return fmt.Sprintf("\"%s\" %s ?", selector.Column,
			selector.Predicate), []any{selector.Value}
	}
	return fmt.Sprintf("\"%s\".\"%s\" %s ?", selector.Table,
		selector.Column, selector.Predicate), []any{selector.Value}
}

// processNullSelector processes a null selector and returns conditions and
// values.
func processNullSelector(selector database.Selector) (string, []any) {
	if selector.Predicate == "=" {
		return buildNullClause(selector, is), nil
	}
	if selector.Predicate == "!=" {
		return buildNullClause(selector, isNot), nil
	}
	return "", nil
}

// buildNullClause returns the string representation of a null clause.
func buildNullClause(selector database.Selector, clause string) string {
	if selector.Table == "" {
		return fmt.Sprintf("\"%s\" %s %s", selector.Column, clause, null)
	}
	return fmt.Sprintf("\"%s\".\"%s\" %s %s",
		selector.Table, selector.Column, clause, null)
}

// createPlaceholdersAndValues creates placeholders and values for a slice.
func createPlaceholdersAndValues(value reflect.Value) (string, []any) {
	placeholderCount := value.Len()
	placeholders := createPlaceholders(placeholderCount)
	values := make([]any, placeholderCount)
	for i := 0; i < placeholderCount; i++ {
		values[i] = value.Index(i).Interface()
	}
	return placeholders, values
}

// createPlaceholders creates placeholders for a slice.
func createPlaceholders(count int) string {
	return strings.TrimSuffix(strings.Repeat("?,", count), ",")
}
