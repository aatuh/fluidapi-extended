package mysql

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

// Query is a query builder for MySQL.
type Query struct{}

// Insert returns the query and values to insert an entity.
//
// Parameters:
//   - tableName: The name of the database table.
//   - insertedValues: Function used to get the columns and values to insert.
//
// Returns:
//   - string: The query.
//   - []any: The values.
func (q *Query) Insert(
	tableName string, insertedValues database.InsertedValuesFn,
) (string, []any) {
	columns, values := insertedValues()
	columnames := getInsertQueryColumnames(columns)

	valuePlaceholders := strings.TrimSuffix(
		strings.Repeat("?, ", len(values)),
		", ",
	)

	query := fmt.Sprintf(
		"INSERT INTO `%s` (%s) VALUES (%s)",
		tableName,
		columnames,
		valuePlaceholders,
	)

	return query, values
}

// InsertMany returns the query and values to insert multiple entities.
//
// Parameters:
//   - tableName: The name of the database table.
//   - insertedValues: Functions used to get the columns and values to insert.
//
// Returns:
//   - string: The query.
//   - []any: The values.
func (q *Query) InsertMany(
	tableName string, insertedValues []database.InsertedValuesFn,
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
		"INSERT INTO `%s` (%s) VALUES %s",
		tableName,
		columnames,
		strings.Join(valuePlaceholders, ", "),
	)

	return query, allValues
}

// UpsertMany creates an upsert query for a list of entities.
//
// Parameters:
//   - tableName: The name of the database table.
//   - insertedValue: The function used to get the columns and values to insert.
//   - updateProjections: The projections of the entities to update.
//
// Returns:
//   - string: The upsert query.
//   - []any: The values.
func (q *Query) UpsertMany(
	tableName string,
	insertedValues []database.InsertedValuesFn,
	updateProjections []database.Projection,
) (string, []any) {
	if len(insertedValues) == 0 {
		return "", nil
	}

	updateParts := make([]string, len(updateProjections))
	for i, proj := range updateProjections {
		updateParts[i] = fmt.Sprintf(
			"`%s` = VALUES(`%s`)",
			proj.Column,
			proj.Column,
		)
	}

	insertQueryPart, allValues := q.InsertMany(tableName, insertedValues)

	builder := strings.Builder{}
	builder.WriteString(insertQueryPart)
	if len(updateParts) != 0 {
		builder.WriteString(" ON DUPLICATE KEY UPDATE ")
		builder.WriteString(strings.Join(updateParts, ", "))
	}
	upsertQuery := builder.String()

	return upsertQuery, allValues
}

// Get returns a get query.
//
//   - tableName: The name of the database table.
//   - dbOptions: The options for the query.
//
// Returns:
//   - string: The query.
//   - []any: The values.
func (q *Query) Get(
	tableName string, opts *database.GetOptions,
) (string, []any) {
	whereClause, whereValues := whereClause(opts.Selectors)

	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf(
		"SELECT %s",
		strings.Join(projectionsToStrings(opts.Projections), ","),
	))
	builder.WriteString(fmt.Sprintf(" FROM `%s`", tableName))
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
	if opts.Lock {
		builder.WriteString(" FOR UPDATE")
	}

	return builder.String(), whereValues
}

// Count returns a count query.
//
// Parameters:
//   - tableName: The name of the database table.
//   - dbOptions: The options for the query.
//
// Returns:
//   - string: The query.
//   - []any: The values.
func (q *Query) Count(
	tableName string, opts *database.CountOptions,
) (string, []any) {
	whereClause, whereValues := whereClause(opts.Selectors)
	joinStmt := joinClause(opts.Joins)

	if opts.Page != nil {
		// Build the inner query with pagination.
		innerQuery := strings.Trim(fmt.Sprintf(
			"SELECT * FROM `%s` %s %s %s",
			tableName,
			joinStmt,
			whereClause,
			getLimitOffsetClauseFromPage(opts.Page),
		), " ")

		// Wrap the inner query in an outer COUNT(*) query.
		query := fmt.Sprintf(
			"SELECT COUNT(*) FROM (%s) AS limited_result",
			innerQuery,
		)
		return query, whereValues
	}

	// Otherwise, build a simple COUNT query without pagination.
	query := strings.Trim(fmt.Sprintf(
		"SELECT COUNT(*) FROM `%s` %s %s",
		tableName,
		joinStmt,
		whereClause,
	), " ")

	return query, whereValues
}

// UpdateQuery returns the SQL query and values for an update query.
//
// Parameters:
//   - tableName: The name of the database table.
//   - updates: The fields to update.
//   - selectors: The selectors for the entities to update.
//
// Returns:
//   - string: The query.
//   - []any: The values.
func (q *Query) UpdateQuery(
	tableName string, updates []database.Update, selectors []database.Selector,
) (string, []any) {
	whereColumns, whereValues := processSelectors(selectors)

	setClause, values := getSetClause(updates)
	values = append(values, whereValues...)

	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf(
		"UPDATE `%s` SET %s",
		tableName,
		setClause,
	))
	if len(whereColumns) != 0 {
		builder.WriteString(" " + getWhereClause(whereColumns))
	}

	return builder.String(), values
}

// Delete returns the SQL query string and the values for the query.
//
// Parameters:
//   - tableName: The name of the database table.
//   - selectors: The selectors for the entities to delete.
//   - opts: The options for the query.
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
		fmt.Sprintf("DELETE FROM `%s` %s", tableName, whereClause),
	)

	if opts != nil {
		writeDeleteOptions(&builder, opts)
	}

	return builder.String(), whereValues
}

// CreateDatabaseQuery generates a SQL query for creating a database.
// If ifNotExists is true, the query will include the IF NOT EXISTS clause.
// If charset or collate are non-empty, they are appended as options.
//
// Parameters:
//   - dbName: The name of the database to create.
//   - ifNotExists: If true, the query will include the IF NOT EXISTS clause.
//   - charset: The character set to use for the database.
//   - collate: The collation to use for the database.
//
// Returns:
//   - string: The SQL query.
//   - []any: The values.
//   - error: An error if the query could not be created.
func (q *Query) CreateDatabaseQuery(
	dbName string, ifNotExists bool, charset string, collate string,
) (string, []any, error) {
	var builder strings.Builder
	builder.WriteString("CREATE DATABASE ")
	if ifNotExists {
		builder.WriteString("IF NOT EXISTS ")
	}
	builder.WriteString(fmt.Sprintf("`%s`", dbName))
	if charset != "" {
		builder.WriteString(fmt.Sprintf(" DEFAULT CHARACTER SET %s", charset))
	}
	if collate != "" {
		builder.WriteString(fmt.Sprintf(" COLLATE = %s", collate))
	}
	builder.WriteString(";")
	return builder.String(), nil, nil
}

// CreateTableQuery generates a SQL query for creating a table.
//   - tableName: the name of the table.
//   - ifNotExists: if true, adds IF NOT EXISTS.
//   - columns: a slice of ColumnDefinition, one per column.
//   - constraints: additional table constraints as raw strings (e.g. unique
//     keys or composite primary keys).
//   - options: extra table options such as engine, charset and collate.
func (q *Query) CreateTableQuery(
	tableName string,
	ifNotExists bool,
	columns []database.ColumnDefinition,
	constraints []string,
	options database.TableOptions,
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
			def += " " + col.Extra
		}
		if col.NotNull {
			def += " NOT NULL"
		} else {
			def += " NULL"
		}
		if col.Default != nil {
			def += " DEFAULT "
			// For values like CURRENT_TIMESTAMP or NULL, no quoting is done.
			if *col.Default == "CURRENT_TIMESTAMP" || *col.Default == "NULL" {
				def += *col.Default
			} else {
				def += fmt.Sprintf("'%s'", *col.Default)
			}
		}
		if col.AutoIncrement {
			def += " AUTO_INCREMENT"
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

	// Append table options.
	if options.Engine != "" {
		builder.WriteString(fmt.Sprintf(" ENGINE = %s", options.Engine))
	}
	if options.Charset != "" {
		builder.WriteString(fmt.Sprintf(" DEFAULT CHARSET = %s", options.Charset))
	}
	if options.Collate != "" {
		builder.WriteString(fmt.Sprintf(" COLLATE = %s", options.Collate))
	}
	builder.WriteString(";")
	return builder.String(), nil, nil
}

// UseDatabaseQuery generates a SQL query to switch to a specified database.
//
// Parameters:
//   - dbName: The name of the database to switch to.
//
// Returns:
//   - string: The SQL query.
//   - []any: The values.
//   - error: An error if the query could not be created.
func (q *Query) UseDatabaseQuery(dbName string) (string, []any, error) {
	return fmt.Sprintf("USE `%s`;", dbName), nil, nil
}

// SetVariableQuery generates a SQL query to set a session variable.
// For example, to set the character set or time zone.
// If the variable is "NAMES" (case-insensitive), the function uses the
// syntax "SET NAMES 'value'"; otherwise it uses "SET variable = 'value'".
//
// Parameters:
//   - variable: The name of the variable to set.
//   - value: The value to set the variable to.
//
// Returns:
//   - string: The SQL query.
func (q *Query) SetVariableQuery(
	variable string, value string,
) (string, []any, error) {
	upperVar := strings.ToUpper(variable)
	if upperVar == "NAMES" {
		return fmt.Sprintf("SET NAMES '%s';", value), nil, nil
	}
	return fmt.Sprintf("SET %s = '%s';", variable, value), nil, nil
}

// AdvisoryLock generates the query to acquire an advisory lock in MySQL.
// The timeout parameter is specified in seconds.
//
// Parameters:
//   - lockName: The name of the lock to acquire.
//   - timeout: The timeout for the lock.
//
// Returns:
//   - string: The SQL query.
//   - []any: The values.
//   - error: An error if the query could not be created.
func (q *Query) AdvisoryLock(
	lockName string, timeout int,
) (string, []any, error) {
	return "SELECT GET_LOCK(?, ?);", []any{lockName, timeout}, nil
}

// AdvisoryUnlock generates the query to release an advisory lock in MySQL.
//
// Parameters:
//   - lockName: The name of the lock to release.
//   - timeout: The timeout for the lock.
//
// Returns:
//   - string: The SQL query.
//   - []any: The values.
//   - error: An error if the query could not be created.
func (q *Query) AdvisoryUnlock(lockName string) (string, []any, error) {
	return "SELECT RELEASE_LOCK(?);", []any{lockName}, nil
}

// getLimitOffsetClauseFromPage returns the LIMIT and OFFSET clause for a page.
func getLimitOffsetClauseFromPage(page *database.Page) string {
	if page == nil {
		return ""
	}
	return fmt.Sprintf(
		"LIMIT %d OFFSET %d",
		page.Limit,
		page.Offset,
	)
}

// getOrderClauseFromOrders returns an ORDER BY clause.
func getOrderClauseFromOrders(orders []database.Order) string {
	if len(orders) == 0 {
		return ""
	}
	orderClause := "ORDER BY"
	for _, order := range orders {
		if order.Table == "" {
			orderClause += fmt.Sprintf(
				" `%s` %s,", order.Field, order.Direction,
			)
		} else {
			orderClause += fmt.Sprintf(
				" `%s`.`%s` %s,", order.Table, order.Field, order.Direction,
			)
		}
	}
	return strings.TrimSuffix(orderClause, ",")
}

// columnSelectorToString returns the string representation of a column
// selector.
func columnSelectorToString(columnSelector database.ColumnSelector) string {
	return fmt.Sprintf(
		"`%s`.`%s`", columnSelector.Table, columnSelector.Column,
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
func projectionToString(projection database.Projection) string {
	builder := strings.Builder{}
	if projection.Table == "" {
		builder.WriteString(fmt.Sprintf("`%s`", projection.Column))
	} else {
		builder.WriteString(fmt.Sprintf(
			"`%s`.`%s`", projection.Table, projection.Column,
		))
	}
	if projection.Alias != "" {
		builder.WriteString(fmt.Sprintf(" AS `%s`", projection.Alias))
	}
	return builder.String()
}

// getInsertQueryColumnames returns the string representation of column names.
func getInsertQueryColumnames(columns []string) string {
	wrappedColumns := make([]string, len(columns))
	for i, column := range columns {
		wrappedColumns[i] = "`" + column + "`"
	}
	columnames := strings.Join(wrappedColumns, ", ")
	return columnames
}

// projectionsToStrings returns the string representations of projections.
func projectionsToStrings(projections []database.Projection) []string {
	if len(projections) == 0 {
		return []string{"*"}
	}
	projectionStrings := make([]string, len(projections))
	for i, projection := range projections {
		projectionStrings[i] = projectionToString(projection)
	}
	return projectionStrings
}

// joinClause returns the string representation of a join clause.
func joinClause(joins []database.Join) string {
	var joinClause string
	for _, join := range joins {
		if joinClause != "" {
			joinClause += " "
		}
		joinClause += fmt.Sprintf(
			"%s JOIN `%s` ON %s = %s",
			join.Type,
			join.Table,
			columnSelectorToString(join.OnLeft),
			columnSelectorToString(join.OnRight),
		)
	}
	return joinClause
}

// whereClause returns the string representation of a where clause.
func whereClause(selectors []database.Selector) (string, []any) {
	whereColumns, whereValues := processSelectors(selectors)
	var whereClause string
	if len(whereColumns) > 0 {
		whereClause = "WHERE " + strings.Join(whereColumns, " AND ")
	}
	return strings.Trim(whereClause, " "), whereValues
}

// getWhereClause returns the string representation of a where clause.
func getWhereClause(whereColumns []string) string {
	whereClause := ""
	if len(whereColumns) > 0 {
		whereClause = "WHERE " + strings.Join(whereColumns, " AND ")
	}
	return whereClause
}

// getSetClause returns the string representation of a SET clause.
func getSetClause(updates []database.Update) (string, []any) {
	setClauseParts := make([]string, len(updates))
	values := make([]any, len(updates))
	for i, update := range updates {
		setClauseParts[i] = fmt.Sprintf(
			"%s = ?",
			update.Field,
		)
		values[i] = update.Value
	}
	return strings.Join(setClauseParts, ", "), values
}

// writeDeleteOptions writes the delete options to the builder.
func writeDeleteOptions(
	builder *strings.Builder, opts *database.DeleteOptions,
) {
	orderClause := getOrderClauseFromOrders(opts.Orders)
	if orderClause != "" {
		builder.WriteString(" " + orderClause)
	}
	limit := opts.Limit
	if limit > 0 {
		builder.WriteString(fmt.Sprintf(" LIMIT %d", limit))
	}
}

// processSelector processes a selector and returns a condition and values.
func processSelector(selector database.Selector) (string, []any) {
	if selector.Predicate == in {
		return processInSelector(selector)
	}
	return processDefaultSelector(selector)
}

// processInSelector processes an IN selector and returns conditions and values.
func processInSelector(selector database.Selector) (string, []any) {
	value := reflect.ValueOf(selector.Value)
	if value.Kind() == reflect.Slice {
		placeholders, values := createPlaceholdersAndValues(value)
		column := fmt.Sprintf(
			"`%s`.`%s` %s (%s)",
			selector.Table,
			selector.Column,
			in,
			placeholders,
		)
		return column, values
	}
	// If value is not a slice, treat as a single value
	return fmt.Sprintf(
		"`%s`.`%s` %s (?)", selector.Table, selector.Column, in,
	), []any{selector.Value}
}

// processDefaultSelector processes a default selector and returns conditions
// and values.
func processDefaultSelector(selector database.Selector) (string, []any) {
	if selector.Value == nil {
		return processNullSelector(selector)
	}
	if selector.Table == "" {
		return fmt.Sprintf(
			"`%s` %s ?",
			selector.Column,
			selector.Predicate,
		), []any{selector.Value}
	} else {
		return fmt.Sprintf(
			"`%s`.`%s` %s ?",
			selector.Table,
			selector.Column,
			selector.Predicate,
		), []any{selector.Value}
	}
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
		return fmt.Sprintf("`%s` %s %s", selector.Column, clause, null)
	}
	return fmt.Sprintf(
		"`%s`.`%s` %s %s", selector.Table, selector.Column, clause, null,
	)
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
