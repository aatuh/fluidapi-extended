package mysql

import (
	"reflect"
	"strings"
	"testing"

	"github.com/pakkasys/fluidapi/database"
	"github.com/stretchr/testify/assert"
)

// TODO: implement

// import (
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// )

// type TestCreateEntity struct {
// 	ID   int
// 	Name string
// 	Age  int
// }

// // TestGetInsertQueryColumnNames_MultipleColumns tests getInsertQueryColumnNames
// // with multiple columns.
// func TestGetInsertQueryColumnNames_MultipleColumns(t *testing.T) {
// 	// Multiple columns
// 	columns := []string{"id", "name", "age"}

// 	result := getInsertQueryColumnNames(columns)

// 	// Expected result
// 	expectedResult := "`id`, `name`, `age`"

// 	assert.Equal(t, expectedResult, result)
// }

// // TestGetInsertQueryColumnNames_SingleColumn tests getInsertQueryColumnNames
// // with a single column.
// func TestGetInsertQueryColumnNames_SingleColumn(t *testing.T) {
// 	// Single column
// 	columns := []string{"id"}

// 	result := getInsertQueryColumnNames(columns)

// 	// Expected result
// 	expectedResult := "`id`"

// 	assert.Equal(t, expectedResult, result)
// }

// // TestGetInsertQueryColumnNames_EmptyColumns tests getInsertQueryColumnNames
// // with an empty list of columns.
// func TestGetInsertQueryColumnNames_EmptyColumns(t *testing.T) {
// 	// Empty columns
// 	columns := []string{}

// 	result := getInsertQueryColumnNames(columns)

// 	// Expected result is an empty string
// 	expectedResult := ""

// 	assert.Equal(t, expectedResult, result)
// }

// // TestInsertQuery_NormalOperation tests insertQuery with a standard entity.
// func TestInsertQuery_NormalOperation(t *testing.T) {
// 	// Inserter function that returns two columns and values
// 	inserter := func(entity *TestCreateEntity) ([]string, []any) {
// 		return []string{"id", "name"}, []any{1, "Alice"}
// 	}

// 	query, values := Insert(
// 		&TestCreateEntity{ID: 1, Name: "Alice"},
// 		"user",
// 		inserter,
// 	)

// 	expectedQuery := "INSERT INTO `user` (`id`, `name`) VALUES (?, ?)"
// 	expectedValues := []any{1, "Alice"}

// 	assert.Equal(t, expectedQuery, query)
// 	assert.Equal(t, expectedValues, values)
// }

// // TestInsertQuery_SingleColumnEntity tests insertQuery with an entity that has
// // only one column.
// func TestInsertQuery_SingleColumnEntity(t *testing.T) {
// 	// Inserter function that returns a single column and value
// 	inserter := func(entity *TestCreateEntity) ([]string, []any) {
// 		return []string{"id"}, []any{1}
// 	}

// 	query, values := Insert(&TestCreateEntity{ID: 1}, "user", inserter)

// 	expectedQuery := "INSERT INTO `user` (`id`) VALUES (?)"
// 	expectedValues := []any{1}

// 	assert.Equal(t, expectedQuery, query)
// 	assert.Equal(t, expectedValues, values)
// }

// // TestInsertQuery_NoColumns tests insertQuery with an entity that has no
// // columns.
// func TestInsertQuery_NoColumns(t *testing.T) {
// 	// Inserter function that returns no columns or values
// 	inserter := func(entity *TestCreateEntity) ([]string, []any) {
// 		return []string{}, []any{}
// 	}

// 	query, values := Insert(&TestCreateEntity{}, "user", inserter)

// 	expectedQuery := "INSERT INTO `user` () VALUES ()"
// 	expectedValues := []any{}

// 	assert.Equal(t, expectedQuery, query)
// 	assert.Equal(t, expectedValues, values)
// }

// TestBuildBaseCountQuery_NoSelectorsNoJoins tests BuildBaseCountQuery with no
// selectors or joins.
func TestBuildBaseCountQuery_NoSelectorsNoJoins(t *testing.T) {
	tableName := "test_table"
	dbOptions := &database.CountOptions{}

	query, whereValues := Count(tableName, dbOptions)

	expectedQuery := "SELECT COUNT(*) FROM `test_table`"
	expectedValues := []any{}

	assert.Equal(t, expectedQuery, query)
	assert.ElementsMatch(t, expectedValues, whereValues)
}

// TestBuildBaseCountQuery_WithSelectors tests BuildBaseCountQuery only
// selectors.
func TestBuildBaseCountQuery_WithSelectors(t *testing.T) {
	tableName := "test_table"
	dbOptions := &database.CountOptions{
		Selectors: []database.Selector{
			{Table: "test_table", Field: "id", Predicate: "=", Value: 1},
		},
	}

	query, whereValues := Count(tableName, dbOptions)

	expectedQuery := "SELECT COUNT(*) FROM `test_table`  WHERE `test_table`.`id` = ?"
	expectedValues := []any{1}

	assert.Equal(t, expectedQuery, query)
	assert.ElementsMatch(t, expectedValues, whereValues)
}

// TestBuildBaseCountQuery_WithJoins tests BuildBaseCountQuery with joins only.
func TestBuildBaseCountQuery_WithJoins(t *testing.T) {
	tableName := "test_table"
	dbOptions := &database.CountOptions{
		Joins: []database.Join{
			{
				Type:  database.JoinTypeInner,
				Table: "other_table",
				OnLeft: database.ColumnSelector{
					Table:   "test_table",
					Columnn: "id",
				},
				OnRight: database.ColumnSelector{
					Table:   "other_table",
					Columnn: "ref_id",
				},
			},
		},
	}
	query, whereValues := Count(tableName, dbOptions)

	expectedQuery := "SELECT COUNT(*) FROM `test_table` INNER JOIN `other_table` ON `test_table`.`id` = `other_table`.`ref_id`"
	expectedValues := []any{}

	assert.Equal(t, expectedQuery, query)
	assert.ElementsMatch(t, expectedValues, whereValues)
}

// TestBuildBaseCountQuery_WithSelectorsAndJoins tests BuildBaseCountQuery with
// both selectors and joins.
func TestBuildBaseCountQuery_WithSelectorsAndJoins(t *testing.T) {
	tableName := "test_table"
	dbOptions := &database.CountOptions{
		Selectors: []database.Selector{
			{Table: "test_table", Field: "id", Predicate: "=", Value: 1},
		},
		Joins: []database.Join{
			{
				Type:  database.JoinTypeInner,
				Table: "other_table",
				OnLeft: database.ColumnSelector{
					Table:   "test_table",
					Columnn: "id",
				},
				OnRight: database.ColumnSelector{
					Table:   "other_table",
					Columnn: "ref_id",
				},
			},
		},
	}

	query, whereValues := Count(tableName, dbOptions)

	expectedQuery := "SELECT COUNT(*) FROM `test_table` INNER JOIN `other_table` ON `test_table`.`id` = `other_table`.`ref_id` WHERE `test_table`.`id` = ?"
	expectedValues := []any{1}

	assert.Equal(t, expectedQuery, query)
	assert.ElementsMatch(t, expectedValues, whereValues)
}

// TestGetLimitOffsetClauseFromPage_NoPage tests the case where no page is
// provided.
func TestGetLimitOffsetClauseFromPage_NoPage(t *testing.T) {
	var p *database.Page = nil
	limitOffsetClause := GetLimitOffsetClauseFromPage(p)
	assert.Equal(t, "", limitOffsetClause)
}

// TestGetLimitOffsetClauseFromPage_WithPage tests the case where a page with
// limit and offset is provided.
func TestGetLimitOffsetClauseFromPage_WithPage(t *testing.T) {
	p := &database.Page{Limit: 10, Offset: 20}

	limitOffsetClause := GetLimitOffsetClauseFromPage(p)

	expected := "LIMIT 10 OFFSET 20"
	assert.Equal(t, expected, limitOffsetClause)
}

// TestGetLimitOffsetClauseFromPage_ZeroLimit tests the case where limit is 0.
func TestGetLimitOffsetClauseFromPage_ZeroLimit(t *testing.T) {
	p := &database.Page{Limit: 0, Offset: 20}

	limitOffsetClause := GetLimitOffsetClauseFromPage(p)

	expected := "LIMIT 0 OFFSET 20"
	assert.Equal(t, expected, limitOffsetClause)
}

// TestGetLimitOffsetClauseFromPage_ZeroOffset tests the case where offset is 0.
func TestGetLimitOffsetClauseFromPage_ZeroOffset(t *testing.T) {
	p := &database.Page{Limit: 10, Offset: 0}

	limitOffsetClause := GetLimitOffsetClauseFromPage(p)

	expected := "LIMIT 10 OFFSET 0"
	assert.Equal(t, expected, limitOffsetClause)
}

// TestGetLimitOffsetClauseFromPage_ZeroLimitAndOffset tests the case where both
// limit and offset are 0.
func TestGetLimitOffsetClauseFromPage_ZeroLimitAndOffset(t *testing.T) {
	p := &database.Page{Limit: 0, Offset: 0}

	limitOffsetClause := GetLimitOffsetClauseFromPage(p)

	expected := "LIMIT 0 OFFSET 0"
	assert.Equal(t, expected, limitOffsetClause)
}

// import (
// 	"testing"

// 	"github.com/pakkasys/fluidapi/database/clause"
// 	"github.com/stretchr/testify/assert"
// )

// type TestUpsertEntity struct {
// 	ID   int
// 	Name string
// 	Age  int
// }

// // TestUpsertManyQuery_NormalOperation tests upsertManyQuery with multiple
// // entities and projections.
// func TestUpsertManyQuery_NormalOperation(t *testing.T) {
// 	// Test entities and projections
// 	entities := []*TestUpsertEntity{
// 		{ID: 1, Name: "Alice", Age: 30},
// 		{ID: 2, Name: "Bob", Age: 25},
// 	}
// 	projections := []clause.Projection{
// 		{Column: "name", Alias: "test"},
// 		{Column: "age", Alias: "test"},
// 	}

// 	// Inserter function for the entities
// 	inserter := func(e *TestUpsertEntity) ([]string, []any) {
// 		return []string{"id", "name", "age"}, []any{e.ID, e.Name, e.Age}
// 	}

// 	query, values := UpsertMany(entities, "user", inserter, projections)

// 	expectedQuery := "INSERT INTO `user` (`id`, `name`, `age`) VALUES (?, ?, ?), (?, ?, ?) ON DUPLICATE KEY UPDATE `name` = VALUES(`name`), `age` = VALUES(`age`)"
// 	expectedValues := []any{1, "Alice", 30, 2, "Bob", 25}

// 	assert.Equal(t, expectedQuery, query)
// 	assert.Equal(t, expectedValues, values)
// }

// // TestUpsertManyQuery_SingleEntity tests upsertManyQuery with a single entity.
// func TestUpsertManyQuery_SingleEntity(t *testing.T) {
// 	// Test entity and projections
// 	entities := []*TestUpsertEntity{
// 		{ID: 1, Name: "Alice"},
// 	}
// 	projections := []clause.Projection{
// 		{Column: "name", Alias: "test"},
// 	}

// 	// Inserter function for the entity
// 	inserter := func(e *TestUpsertEntity) ([]string, []any) {
// 		return []string{"id", "name"}, []any{e.ID, e.Name}
// 	}

// 	query, values := UpsertMany(entities, "user", inserter, projections)

// 	expectedQuery := "INSERT INTO `user` (`id`, `name`) VALUES (?, ?) ON DUPLICATE KEY UPDATE `name` = VALUES(`name`)"
// 	expectedValues := []any{1, "Alice"}

// 	assert.Equal(t, expectedQuery, query)
// 	assert.Equal(t, expectedValues, values)
// }

// // TestUpsertManyQuery_EmptyEntities tests upsertManyQuery with no entities.
// func TestUpsertManyQuery_EmptyEntities(t *testing.T) {
// 	// Test with no entities
// 	entities := []*TestUpsertEntity{}
// 	projections := []clause.Projection{
// 		{Column: "name", Alias: "test"},
// 	}

// 	// Inserter function (not used here since entities is empty)
// 	inserter := func(e *TestUpsertEntity) ([]string, []any) {
// 		return []string{}, []any{}
// 	}

// 	query, values := UpsertMany(entities, "user", inserter, projections)

// 	assert.Equal(t, "", query)
// 	assert.Equal(t, []any(nil), values)
// }

// // TestUpsertManyQuery_MissingUpdateProjections tests upsertManyQuery with no
// // update projections.
// func TestUpsertManyQuery_MissingUpdateProjections(t *testing.T) {
// 	// Test entities
// 	entities := []*TestUpsertEntity{
// 		{ID: 1, Name: "Alice"},
// 		{ID: 2, Name: "Bob"},
// 	}

// 	// Inserter function for the entities
// 	inserter := func(e *TestUpsertEntity) ([]string, []any) {
// 		return []string{"id", "name"}, []any{e.ID, e.Name}
// 	}

// 	// Call the function with an empty projection list
// 	query, values := UpsertMany(
// 		entities,
// 		"user",
// 		inserter,
// 		[]clause.Projection{},
// 	)

// 	assert.Equal(
// 		t,
// 		"INSERT INTO `user` (`id`, `name`) VALUES (?, ?), (?, ?)",
// 		query,
// 	)
// 	assert.Equal(t, []any{1, "Alice", 2, "Bob"}, values)
// }

// TestProjectionsToStrings_NoProjections tests the case where no projections
// are provided.
func TestProjectionsToStrings_NoProjections(t *testing.T) {
	projections := []database.Projection{}
	projectionStrings := projectionsToStrings(projections)
	assert.Equal(t, []string{"*"}, projectionStrings)
}

// TestProjectionsToStrings_SingleProjection tests the case where a single
// projection is provided.
func TestProjectionsToStrings_SingleProjection(t *testing.T) {
	projections := []database.Projection{
		{Table: "user", Column: "name"},
	}

	projectionStrings := projectionsToStrings(projections)

	expected := []string{"`user`.`name`"}
	assert.Equal(t, expected, projectionStrings)
}

// TestProjectionsToStrings_MultipleProjections tests the case where multiple
// projections are provided.
func TestProjectionsToStrings_MultipleProjections(t *testing.T) {
	projections := []database.Projection{
		{Table: "user", Column: "name"},
		{Table: "orders", Column: "order_id"},
	}

	projectionStrings := projectionsToStrings(projections)

	expected := []string{"`user`.`name`", "`orders`.`order_id`"}
	assert.Equal(t, expected, projectionStrings)
}

// TestProjectionsToStrings_EmptyFields tests the case where a projection has
// empty fields.
func TestProjectionsToStrings_EmptyFields(t *testing.T) {
	projections := []database.Projection{
		{Table: "", Column: ""},
	}

	projectionStrings := projectionsToStrings(projections)

	expected := []string{"``"}
	assert.Equal(t, expected, projectionStrings)
}

// TestJoinClause_NoJoins tests the case where no joins are provided.
func TestJoinClause_NoJoins(t *testing.T) {
	joins := []database.Join{}
	joinClause := joinClause(joins)
	assert.Equal(t, "", joinClause)
}

// TestJoinClause_SingleJoin tests the case where a single join is provided.
func TestJoinClause_SingleJoin(t *testing.T) {
	joins := []database.Join{
		{
			Type:  database.JoinTypeInner,
			Table: "orders",
			OnLeft: database.ColumnSelector{
				Table:   "user",
				Columnn: "id",
			},
			OnRight: database.ColumnSelector{
				Table:   "orders",
				Columnn: "user_id",
			},
		},
	}

	joinClause := joinClause(joins)

	expected := "INNER JOIN `orders` ON `user`.`id` = `orders`.`user_id`"
	assert.Equal(t, expected, joinClause)
}

// TestJoinClause_MultipleJoins tests the case where multiple joins are
// provided.
func TestJoinClause_MultipleJoins(t *testing.T) {
	joins := []database.Join{
		{
			Type:  database.JoinTypeInner,
			Table: "order",
			OnLeft: database.ColumnSelector{
				Table:   "user",
				Columnn: "id",
			},
			OnRight: database.ColumnSelector{
				Table:   "order",
				Columnn: "user_id",
			},
		},
		{
			Type:  database.JoinTypeLeft,
			Table: "payments",
			OnLeft: database.ColumnSelector{
				Table:   "user",
				Columnn: "id",
			},
			OnRight: database.ColumnSelector{
				Table:   "payments",
				Columnn: "user_id",
			},
		},
	}

	joinClause := joinClause(joins)

	// Expect multiple JOIN clauses
	expected := "INNER JOIN `order` ON `user`.`id` = `order`.`user_id` LEFT JOIN `payments` ON `user`.`id` = `payments`.`user_id`"
	assert.Equal(t, expected, joinClause)
}

// TestJoinClause_EmptyFields tests the case where a join has empty fields.
func TestJoinClause_EmptyFields(t *testing.T) {
	joins := []database.Join{
		{
			Type:  database.JoinTypeInner,
			Table: "",
			OnLeft: database.ColumnSelector{
				Table:   "",
				Columnn: "",
			},
			OnRight: database.ColumnSelector{
				Table:   "",
				Columnn: "",
			},
		},
	}

	joinClause := joinClause(joins)

	// Expect a malformed JOIN clause with empty fields
	expected := "INNER JOIN `` ON ``.`` = ``.``"
	assert.Equal(t, expected, joinClause)
}

// TestWhereClause_NoSelectors tests the case where no selectors are provided.
func TestWhereClause_NoSelectors(t *testing.T) {
	selectors := []database.Selector{}

	whereClause, whereValues := whereClause(selectors)

	// Expect an empty string and no values since there are no selectors
	assert.Equal(t, "", whereClause)
	assert.Empty(t, whereValues)
}

// TestWhereClause_SingleSelector tests the case where a single selector is
// provided.
func TestWhereClause_SingleSelector(t *testing.T) {
	selectors := []database.Selector{
		{Table: "user", Field: "id", Predicate: "=", Value: 1},
	}

	whereClause, whereValues := whereClause(selectors)

	expectedClause := "WHERE `user`.`id` = ?"
	assert.Equal(t, expectedClause, whereClause)
	assert.Equal(t, []any{1}, whereValues)
}

// TestWhereClause_MultipleSelectors tests the case where multiple selectors are
// provided.
func TestWhereClause_MultipleSelectors(t *testing.T) {
	selectors := []database.Selector{
		{Table: "user", Field: "id", Predicate: "=", Value: 1},
		{Table: "user", Field: "age", Predicate: ">", Value: 18},
	}

	whereClause, whereValues := whereClause(selectors)

	expectedClause := "WHERE `user`.`id` = ? AND `user`.`age` > ?"
	assert.Equal(t, expectedClause, whereClause)
	assert.Equal(t, []any{1, 18}, whereValues)
}

// TestWhereClause_DifferentPredicates tests the case where different predicates
// are provided.
func TestWhereClause_DifferentPredicates(t *testing.T) {
	selectors := []database.Selector{
		{Table: "user", Field: "name", Predicate: "LIKE", Value: "%Alice%"},
		{Table: "user", Field: "age", Predicate: "<", Value: 30},
	}

	whereClause, whereValues := whereClause(selectors)

	// Expect a WHERE clause with different predicates
	expectedClause := "WHERE `user`.`name` LIKE ? AND `user`.`age` < ?"
	assert.Equal(t, expectedClause, whereClause)
	assert.Equal(t, []any{"%Alice%", 30}, whereValues)
}

// TestBuildBaseGetQuery_NoOptions tests the case where no options are provided.
func TestBuildBaseGetQuery_NoOptions(t *testing.T) {
	getOptions := database.GetOptions{}

	query, whereValues := Get("user", &getOptions)

	expectedQuery := "SELECT * FROM `user`"
	assert.Equal(t, expectedQuery, query)
	assert.Empty(t, whereValues)
}

// TestBuildBaseGetQuery_WithSelectors tests the case where selectors are
// provided.
func TestBuildBaseGetQuery_WithSelectors(t *testing.T) {
	getOptions := database.GetOptions{}
	getOptions.Selectors = []database.Selector{
		{Table: "user", Field: "id", Predicate: "=", Value: 1},
	}

	query, whereValues := Get("user", &getOptions)

	expectedQuery := "SELECT * FROM `user` WHERE `user`.`id` = ?"
	assert.Equal(t, expectedQuery, query)
	assert.Equal(t, []any{1}, whereValues)
}

// TestBuildBaseGetQuery_WithOrders tests the case where orders are provided.
func TestBuildBaseGetQuery_WithOrders(t *testing.T) {
	getOptions := database.GetOptions{}
	getOptions.Orders = []database.Order{
		{Table: "user", Field: "name", Direction: "ASC"},
	}

	query, whereValues := Get("user", &getOptions)

	expectedQuery := "SELECT * FROM `user` ORDER BY `user`.`name` ASC"
	assert.Equal(t, expectedQuery, query)
	assert.Empty(t, whereValues)
}

// TestBuildBaseGetQuery_WithProjections tests the case where projections are
// provided.
func TestBuildBaseGetQuery_WithProjections(t *testing.T) {
	getOptions := database.GetOptions{}
	getOptions.Projections = []database.Projection{
		{Table: "user", Column: "name", Alias: "user_name"},
	}

	query, whereValues := Get("user", &getOptions)

	expectedQuery := "SELECT `user`.`name` AS `user_name` FROM `user`"
	assert.Equal(t, expectedQuery, query)
	assert.Empty(t, whereValues)
}

// TestBuildBaseGetQuery_WithJoins tests the case where joins are provided.
func TestBuildBaseGetQuery_WithJoins(t *testing.T) {
	getOptions := database.GetOptions{}
	getOptions.Joins = []database.Join{
		{
			Type:  database.JoinTypeInner,
			Table: "order",
			OnLeft: database.ColumnSelector{
				Table:   "user",
				Columnn: "id",
			},
			OnRight: database.ColumnSelector{
				Table:   "order",
				Columnn: "user_id",
			},
		},
	}

	query, whereValues := Get("user", &getOptions)

	expectedQuery := "SELECT * FROM `user` INNER JOIN `order` ON `user`.`id` = `order`.`user_id`"
	assert.Equal(t, expectedQuery, query)
	assert.Empty(t, whereValues)
}

// TestBuildBaseGetQuery_WithLock tests the case where the lock option is set.
func TestBuildBaseGetQuery_WithLock(t *testing.T) {
	getOptions := database.GetOptions{}
	getOptions.Lock = true

	query, whereValues := Get("user", &getOptions)

	expectedQuery := "SELECT * FROM `user` FOR UPDATE"
	assert.Equal(t, expectedQuery, query)
	assert.Empty(t, whereValues)
}

// TestBuildBaseGetQuery_WithPage tests the case where pagination is provided.
func TestBuildBaseGetQuery_WithPage(t *testing.T) {
	getOptions := database.GetOptions{}
	getOptions.Page = &database.Page{Offset: 10, Limit: 20}

	query, whereValues := Get("user", &getOptions)

	expectedQuery := "SELECT * FROM `user` LIMIT 20 OFFSET 10"
	assert.Equal(t, expectedQuery, query)
	assert.Empty(t, whereValues)
}

// TestUpdateQuery_SingleUpdate tests the case where a single update is
// provided.
func TestUpdateQuery_SingleUpdate(t *testing.T) {
	updates := []database.UpdateField{
		{Field: "name", Value: "Alice"},
	}
	selectors := []database.Selector{
		{Table: "user", Field: "id", Predicate: "=", Value: 1},
	}

	query, values := UpdateQuery("user", updates, selectors)

	expectedQuery := "UPDATE `user` SET name = ? WHERE `user`.`id` = ?"
	expectedValues := []any{"Alice", 1}

	assert.Equal(t, expectedQuery, query)
	assert.Equal(t, expectedValues, values)
}

// TestUpdateQuery_MultipleUpdates tests the case where multiple updates are
// provided.
func TestUpdateQuery_MultipleUpdates(t *testing.T) {
	updates := []database.UpdateField{
		{Field: "name", Value: "Alice"},
		{Field: "age", Value: 30},
	}
	selectors := []database.Selector{
		{Table: "user", Field: "id", Predicate: "=", Value: 1},
	}

	query, values := UpdateQuery("user", updates, selectors)

	expectedQuery := "UPDATE `user` SET name = ?, age = ? WHERE `user`.`id` = ?"
	expectedValues := []any{"Alice", 30, 1}

	assert.Equal(t, expectedQuery, query)
	assert.Equal(t, expectedValues, values)
}

// TestUpdateQuery_NoUpdates tests the case where no updates are provided.
func TestUpdateQuery_NoUpdates(t *testing.T) {
	updates := []database.UpdateField{}
	selectors := []database.Selector{
		{Table: "user", Field: "id", Predicate: "=", Value: 1},
	}

	query, values := UpdateQuery("user", updates, selectors)

	expectedQuery := "UPDATE `user` SET  WHERE `user`.`id` = ?"
	expectedValues := []any{1}

	assert.Equal(t, expectedQuery, query)
	assert.Equal(t, expectedValues, values)
}

// TestUpdateQuery_NoSelectors tests the case where no selectors are provided.
func TestUpdateQuery_NoSelectors(t *testing.T) {
	updates := []database.UpdateField{
		{Field: "name", Value: "Alice"},
	}

	selectors := []database.Selector{} // No selectors

	query, values := UpdateQuery("user", updates, selectors)

	expectedQuery := "UPDATE `user` SET name = ?"
	expectedValues := []any{"Alice"}

	assert.Equal(t, expectedQuery, query)
	assert.Equal(t, expectedValues, values)
}

// TestUpdateQuery_EmptyFields tests the case where updates and selectors have
// empty fields.
func TestUpdateQuery_EmptyFields(t *testing.T) {
	updates := []database.UpdateField{
		{Field: "", Value: "Unknown"},
	}
	selectors := []database.Selector{
		{Table: "", Field: "", Predicate: "=", Value: nil},
	}

	query, values := UpdateQuery("user", updates, selectors)

	expectedQuery := "UPDATE `user` SET  = ? WHERE `` IS NULL"
	expectedValues := []any{"Unknown"}

	assert.Equal(t, expectedQuery, query)
	assert.Equal(t, expectedValues, values)
}

// TestGetWhereClause_NoConditions tests the case where no conditions are
// provided.
func TestGetWhereClause_NoConditions(t *testing.T) {
	whereColumns := []string{}

	whereClause := getWhereClause(whereColumns)

	expectedWhereClause := ""
	assert.Equal(t, expectedWhereClause, whereClause)
}

// TestGetWhereClause_SingleCondition tests the case where a single condition is
// provided.
func TestGetWhereClause_SingleCondition(t *testing.T) {
	whereColumns := []string{"`user`.`id` = ?"}

	whereClause := getWhereClause(whereColumns)

	expectedWhereClause := "WHERE `user`.`id` = ?"
	assert.Equal(t, expectedWhereClause, whereClause)
}

// TestGetWhereClause_MultipleConditions tests the case where multiple
// conditions are provided.
func TestGetWhereClause_MultipleConditions(t *testing.T) {
	whereColumns := []string{"`user`.`id` = ?", "`user`.`age` > 18"}

	whereClause := getWhereClause(whereColumns)

	expectedWhereClause := "WHERE `user`.`id` = ? AND `user`.`age` > 18"
	assert.Equal(t, expectedWhereClause, whereClause)
}

// TestGetSetClause_SingleUpdate tests the case where a single update is
// provided.
func TestGetSetClause_SingleUpdate(t *testing.T) {
	updates := []database.UpdateField{
		{Field: "name", Value: "Alice"},
	}

	setClause, values := getSetClause(updates)

	expectedSetClause := "name = ?"
	expectedValues := []any{"Alice"}

	assert.Equal(t, expectedSetClause, setClause)
	assert.Equal(t, expectedValues, values)
}

// TestGetSetClause_MultipleUpdates tests the case where multiple updates are
// provided.
func TestGetSetClause_MultipleUpdates(t *testing.T) {
	updates := []database.UpdateField{
		{Field: "name", Value: "Alice"},
		{Field: "age", Value: 30},
	}

	setClause, values := getSetClause(updates)

	expectedSetClause := "name = ?, age = ?"
	expectedValues := []any{"Alice", 30}

	assert.Equal(t, expectedSetClause, setClause)
	assert.Equal(t, expectedValues, values)
}

// TestGetSetClause_NoUpdates tests the case where no updates are provided.
func TestGetSetClause_NoUpdates(t *testing.T) {
	updates := []database.UpdateField{}

	setClause, values := getSetClause(updates)

	expectedSetClause := ""
	expectedValues := []any{}

	assert.Equal(t, expectedSetClause, setClause)
	assert.Equal(t, expectedValues, values)
}

// TestGetSetClause_EmptyField tests the case where an update has an empty
// field.
func TestGetSetClause_EmptyField(t *testing.T) {
	updates := []database.UpdateField{
		{Field: "", Value: "Unknown"},
	}

	setClause, values := getSetClause(updates)

	expectedSetClause := " = ?"
	expectedValues := []any{"Unknown"}

	assert.Equal(t, expectedSetClause, setClause)
	assert.Equal(t, expectedValues, values)
}

// TestWriteDeleteOptions_WithLimitAndOrders tests writeDeleteOptions with both
// limit and orders.
func TestWriteDeleteOptions_WithLimitAndOrders(t *testing.T) {
	// Create a DeleteOptions with a limit and orders
	orders := []database.Order{
		{Table: "user", Field: "name", Direction: "ASC"},
		{Table: "user", Field: "age", Direction: "DESC"},
	}
	opts := database.DeleteOptions{Limit: 10, Orders: orders}

	// Create a string builder for the SQL query
	builder := strings.Builder{}
	builder.WriteString("DELETE FROM `user` WHERE id = 1")

	writeDeleteOptions(&builder, &opts)

	expectedSQL := "DELETE FROM `user` WHERE id = 1 ORDER BY `user`.`name` ASC, `user`.`age` DESC LIMIT 10"

	assert.Equal(t, expectedSQL, builder.String())
}

// TestWriteDeleteOptions_WithOnlyOrders tests writeDeleteOptions with only
// orders and no limit.
func TestWriteDeleteOptions_WithOnlyOrders(t *testing.T) {
	// Create a DeleteOptions with only orders
	orders := []database.Order{
		{Table: "user", Field: "name", Direction: "ASC"},
	}
	opts := database.DeleteOptions{Limit: 0, Orders: orders}

	// Create a string builder for the SQL query
	builder := strings.Builder{}
	builder.WriteString("DELETE FROM `user` WHERE id = 1")

	writeDeleteOptions(&builder, &opts)

	expectedSQL := "DELETE FROM `user` WHERE id = 1 ORDER BY `user`.`name` ASC"

	assert.Equal(t, expectedSQL, builder.String())
}

// TestWriteDeleteOptions_WithOnlyLimit tests writeDeleteOptions with only a
// limit and no orders.
func TestWriteDeleteOptions_WithOnlyLimit(t *testing.T) {
	// Create a DeleteOptions with only a limit
	opts := database.DeleteOptions{Limit: 5, Orders: nil}

	// Create a string builder for the SQL query
	builder := strings.Builder{}
	builder.WriteString("DELETE FROM `user` WHERE id = 1")

	writeDeleteOptions(&builder, &opts)

	expectedSQL := "DELETE FROM `user` WHERE id = 1 LIMIT 5"

	assert.Equal(t, expectedSQL, builder.String())
}

// TestWriteDeleteOptions_WithNoOptions tests writeDeleteOptions with no limit
// and no orders.
func TestWriteDeleteOptions_WithNoOptions(t *testing.T) {
	// Create an empty DeleteOptions with no limit and no orders
	opts := database.DeleteOptions{Limit: 0, Orders: nil}

	// Create a string builder for the SQL query
	builder := strings.Builder{}
	builder.WriteString("DELETE FROM `user` WHERE id = 1")

	writeDeleteOptions(&builder, &opts)

	expectedSQL := "DELETE FROM `user` WHERE id = 1"

	assert.Equal(t, expectedSQL, builder.String())
}

// TestColumnSelectorString_NormalCase tests the normal operation of the String
// method.
func TestColumnSelectorString_NormalCase(t *testing.T) {
	selector := database.ColumnSelector{
		Table:   "users",
		Columnn: "id",
	}

	result := columnSelectorToString(selector)
	expected := "`users`.`id`"

	assert.Equal(t, expected, result)
}

// TestColumnSelectorString_EmptyTable tests the case where the Table is empty.
func TestColumnSelectorString_EmptyTable(t *testing.T) {
	selector := database.ColumnSelector{
		Table:   "",
		Columnn: "id",
	}

	result := columnSelectorToString(selector)
	expected := "``.`id`"

	assert.Equal(t, expected, result)
}

// TestColumnSelectorString_EmptyColumnn tests the case where the Columnn is empty.
func TestColumnSelectorString_EmptyColumnn(t *testing.T) {
	selector := database.ColumnSelector{
		Table:   "users",
		Columnn: "",
	}

	result := columnSelectorToString(selector)
	expected := "`users`.``"

	assert.Equal(t, expected, result)
}

// TestColumnSelectorString_EmptyTableAndColumnn tests the case where both the
// Table and Columnn are empty.
func TestColumnSelectorString_EmptyTableAndColumnn(t *testing.T) {
	selector := database.ColumnSelector{
		Table:   "",
		Columnn: "",
	}

	result := columnSelectorToString(selector)
	expected := "``.``"

	assert.Equal(t, expected, result)
}

// TestGetOrderClauseFromOrders_NoOrders tests the case where no orders are
// provided.
func TestGetOrderClauseFromOrders_NoOrders(t *testing.T) {
	orders := []database.Order{}
	orderClause := getOrderClauseFromOrders(orders)
	assert.Equal(t, "", orderClause)
}

// TestGetOrderClauseFromOrders_WithoutTable tests the case where there is no
// table in the order.
func TestGetOrderClauseFromOrders_WithoutTable(t *testing.T) {
	orders := []database.Order{
		{Field: "name", Direction: "ASC"},
	}

	orderClause := getOrderClauseFromOrders(orders)

	expected := "ORDER BY `name` ASC"
	assert.Equal(t, expected, orderClause)
}

// TestGetOrderClauseFromOrders_SingleOrder tests the case where a single order
// is provided.
func TestGetOrderClauseFromOrders_SingleOrder(t *testing.T) {
	orders := []database.Order{
		{Table: "user", Field: "name", Direction: "ASC"},
	}

	orderClause := getOrderClauseFromOrders(orders)

	expected := "ORDER BY `user`.`name` ASC"
	assert.Equal(t, expected, orderClause)
}

// TestGetOrderClauseFromOrders_MultipleOrders tests the case where multiple
// orders are provided.
func TestGetOrderClauseFromOrders_MultipleOrders(t *testing.T) {
	orders := []database.Order{
		{Table: "user", Field: "name", Direction: "ASC"},
		{Table: "user", Field: "age", Direction: "DESC"},
	}

	orderClause := getOrderClauseFromOrders(orders)

	expected := "ORDER BY `user`.`name` ASC, `user`.`age` DESC"
	assert.Equal(t, expected, orderClause)
}

// TestGetOrderClauseFromOrders_EmptyFields tests the case where orders have
// empty fields.
func TestGetOrderClauseFromOrders_EmptyFields(t *testing.T) {
	orders := []database.Order{
		{Table: "", Field: "", Direction: "ASC"},
	}

	orderClause := getOrderClauseFromOrders(orders)

	// Expect an ORDER BY clause with empty table and field
	expected := "ORDER BY `` ASC"
	assert.Equal(t, expected, orderClause)
}

// TestProjectionString_NoTableNoAlias tests the case where the Projection has
// no table and no alias.
func TestProjectionString_NoTableNoAlias(t *testing.T) {
	projection := database.Projection{
		Table:  "",
		Column: "column_name",
		Alias:  "",
	}

	result := projectionToString(projection)

	expected := "`column_name`"
	assert.Equal(t, expected, result)
}

// TestProjectionString_WithTableNoAlias tests the case where the Projection has
// a table but no alias.
func TestProjectionString_WithTableNoAlias(t *testing.T) {
	projection := database.Projection{
		Table:  "table_name",
		Column: "column_name",
		Alias:  "",
	}

	result := projectionToString(projection)

	expected := "`table_name`.`column_name`"
	assert.Equal(t, expected, result)
}

// TestProjectionString_WithTableAndAlias tests the case where the Projection
// has both a table and an alias.
func TestProjectionString_WithTableAndAlias(t *testing.T) {
	projection := database.Projection{
		Table:  "table_name",
		Column: "column_name",
		Alias:  "alias_name",
	}

	result := projectionToString(projection)

	expected := "`table_name`.`column_name` AS `alias_name`"
	assert.Equal(t, expected, result)
}

// TestProjectionString_NoTableWithAlias tests the case where the Projection has
// no table but has an alias.
func TestProjectionString_NoTableWithAlias(t *testing.T) {
	projection := database.Projection{
		Table:  "",
		Column: "column_name",
		Alias:  "alias_name",
	}

	result := projectionToString(projection)

	expected := "`column_name` AS `alias_name`"
	assert.Equal(t, expected, result)
}

// TestProjectionString_EmptyFields tests the case where all fields are empty.
func TestProjectionString_EmptyFields(t *testing.T) {
	projection := database.Projection{
		Table:  "",
		Column: "",
		Alias:  "",
	}

	result := projectionToString(projection)

	expected := "``"
	assert.Equal(t, expected, result)
}

// TestGetByField_Found tests the case where the selector with the given field
// is found.
func TestGetByField_Found(t *testing.T) {
	selectors := database.Selectors{
		{Table: "user", Field: "id", Predicate: "=", Value: 1},
		{Table: "user", Field: "name", Predicate: "=", Value: "Alice"},
	}

	selector := selectors.GetByField("name")

	assert.NotNil(t, selector)
	assert.Equal(t, "user", selector.Table)
	assert.Equal(t, "name", selector.Field)
	assert.Equal(t, "=", string(selector.Predicate))
	assert.Equal(t, "Alice", selector.Value)
}

// TestGetByField_NotFound tests the case where the selector with the given
// field is not found.
func TestGetByField_NotFound(t *testing.T) {
	selectors := database.Selectors{
		{Table: "user", Field: "id", Predicate: "=", Value: 1},
		{Table: "user", Field: "name", Predicate: "=", Value: "Alice"},
	}

	selector := selectors.GetByField("age")

	assert.Nil(t, selector)
}

// TestGetByFields_Found tests the case where the selectors with the given
// fields are found.
func TestGetByFields_Found(t *testing.T) {
	selectors := database.Selectors{
		{Table: "user", Field: "id", Predicate: "=", Value: 1},
		{Table: "user", Field: "name", Predicate: "=", Value: "Alice"},
		{Table: "user", Field: "age", Predicate: ">", Value: 25},
	}

	resultSelectors := selectors.GetByFields("name", "age")

	assert.Len(t, resultSelectors, 2)
	assert.Equal(t, "name", resultSelectors[0].Field)
	assert.Equal(t, "age", resultSelectors[1].Field)
}

// TestGetByFields_NotFound tests the case where none of the selectors with the
// given fields are found.
func TestGetByFields_NotFound(t *testing.T) {
	selectors := database.Selectors{
		{Table: "user", Field: "id", Predicate: "=", Value: 1},
		{Table: "user", Field: "name", Predicate: "=", Value: "Alice"},
	}

	resultSelectors := selectors.GetByFields("age", "address")

	assert.Len(t, resultSelectors, 0)
}

// TestGetByFields_PartialFound tests the case where some selectors with the
// given fields are found.
func TestGetByFields_PartialFound(t *testing.T) {
	selectors := database.Selectors{
		{Table: "user", Field: "id", Predicate: "=", Value: 1},
		{Table: "user", Field: "name", Predicate: "=", Value: "Alice"},
		{Table: "user", Field: "age", Predicate: ">", Value: 25},
	}

	resultSelectors := selectors.GetByFields("name", "address")

	assert.Len(t, resultSelectors, 1)
	assert.Equal(t, "name", resultSelectors[0].Field)
}

// TestProcessSelectors_NoSelectors tests the case where no selectors are
// provided.
func TestProcessSelectors_NoSelectors(t *testing.T) {
	selectors := []database.Selector{}

	whereColumns, whereValues := processSelectors(selectors)

	// Expect no columns and no values
	assert.Empty(t, whereColumns)
	assert.Empty(t, whereValues)
}

// TestProcessSelectors_SingleSelector tests the case where a single selector is
// provided.
func TestProcessSelectors_SingleSelector(t *testing.T) {
	selectors := []database.Selector{
		{Table: "user", Field: "id", Predicate: "=", Value: 1},
	}

	whereColumns, whereValues := processSelectors(selectors)

	expectedColumns := []string{"`user`.`id` = ?"}
	expectedValues := []any{1}

	assert.Equal(t, expectedColumns, whereColumns)
	assert.Equal(t, expectedValues, whereValues)
}

// TestProcessSelectors_MultipleSelectors tests the case where multiple
// selectors are provided.
func TestProcessSelectors_MultipleSelectors(t *testing.T) {
	selectors := []database.Selector{
		{Table: "user", Field: "id", Predicate: "=", Value: 1},
		{Table: "user", Field: "age", Predicate: ">", Value: 18},
	}

	whereColumns, whereValues := processSelectors(selectors)

	expectedColumns := []string{"`user`.`id` = ?", "`user`.`age` > ?"}
	expectedValues := []any{1, 18}

	assert.Equal(t, expectedColumns, whereColumns)
	assert.Equal(t, expectedValues, whereValues)
}

// TestProcessSelectors_WithInPredicate tests the case where a selector with
// "IN" predicate is provided.
func TestProcessSelectors_WithInPredicate(t *testing.T) {
	selectors := []database.Selector{
		{Table: "user", Field: "id", Predicate: "IN", Value: []int{1, 2, 3}},
	}

	whereColumns, whereValues := processSelectors(selectors)

	expectedColumns := []string{"`user`.`id` IN (?,?,?)"}
	expectedValues := []any{1, 2, 3}

	assert.Equal(t, expectedColumns, whereColumns)
	assert.Equal(t, expectedValues, whereValues)
}

// TestProcessSelectors_WithNilValue tests the case where a selector with a nil
// value is provided.
func TestProcessSelectors_WithNilValue(t *testing.T) {
	selectors := []database.Selector{
		{Table: "user", Field: "deleted_at", Predicate: "=", Value: nil},
	}

	whereColumns, whereValues := processSelectors(selectors)

	expectedColumns := []string{"`user`.`deleted_at` IS NULL"}

	assert.Equal(t, expectedColumns, whereColumns)
	assert.Empty(t, whereValues) // No values since it's a NULL condition
}

// TestProcessSelectors_WithDifferentPredicates tests the case where different
// predicates are provided.
func TestProcessSelectors_WithDifferentPredicates(t *testing.T) {
	selectors := []database.Selector{
		{Table: "user", Field: "name", Predicate: "LIKE", Value: "%Alice%"},
		{Table: "user", Field: "age", Predicate: "<", Value: 30},
	}

	whereColumns, whereValues := processSelectors(selectors)

	expectedColumns := []string{"`user`.`name` LIKE ?", "`user`.`age` < ?"}
	expectedValues := []any{"%Alice%", 30}

	assert.Equal(t, expectedColumns, whereColumns)
	assert.Equal(t, expectedValues, whereValues)
}

// TestProcessSelectors_EmptyTableField tests the case where a selector with an
// empty table and field is provided.
func TestProcessSelectors_EmptyTableField(t *testing.T) {
	selectors := []database.Selector{
		{Table: "", Field: "", Predicate: "=", Value: 1},
	}

	whereColumns, whereValues := processSelectors(selectors)

	expectedColumns := []string{"`` = ?"}
	expectedValues := []any{1}

	assert.Equal(t, expectedColumns, whereColumns)
	assert.Equal(t, expectedValues, whereValues)
}

// TestProcessSelector_WithInPredicate tests the processSelector function with
// an "IN" predicate.
func TestProcessSelector_WithInPredicate(t *testing.T) {
	selector := database.Selector{
		Table:     "user",
		Field:     "id",
		Predicate: "IN",
		Value:     []int{1, 2, 3},
	}

	column, values := processSelector(selector)

	expectedColumn := "`user`.`id` IN (?,?,?)"
	expectedValues := []any{1, 2, 3}

	assert.Equal(t, expectedColumn, column)
	assert.Equal(t, expectedValues, values)
}

// TestProcessSelector_WithDefaultPredicate tests the processSelector function
// with a default predicate.
func TestProcessSelector_WithDefaultPredicate(t *testing.T) {
	selector := database.Selector{
		Table:     "user",
		Field:     "name",
		Predicate: "=",
		Value:     "Alice",
	}

	column, values := processSelector(selector)

	expectedColumn := "`user`.`name` = ?"
	expectedValues := []any{"Alice"}

	assert.Equal(t, expectedColumn, column)
	assert.Equal(t, expectedValues, values)
}

// TestProcessInSelector_WithSliceValue tests the processInSelector function
// with a slice value.
func TestProcessInSelector_WithSliceValue(t *testing.T) {
	selector := database.Selector{
		Table:     "user",
		Field:     "id",
		Predicate: "IN",
		Value:     []int{1, 2, 3},
	}

	column, values := processInSelector(selector)

	expectedColumn := "`user`.`id` IN (?,?,?)"
	expectedValues := []any{1, 2, 3}

	assert.Equal(t, expectedColumn, column)
	assert.Equal(t, expectedValues, values)
}

// TestProcessInSelector_WithNonSliceValue tests the processInSelector function
// with a non-slice value.
func TestProcessInSelector_WithNonSliceValue(t *testing.T) {
	selector := database.Selector{
		Table:     "user",
		Field:     "id",
		Predicate: "IN",
		Value:     1,
	}

	column, values := processInSelector(selector)

	expectedColumn := "`user`.`id` IN (?)"
	expectedValues := []any{1}

	assert.Equal(t, expectedColumn, column)
	assert.Equal(t, expectedValues, values)
}

// TestProcessInSelector_EmptySlice tests the processInSelector function with an
// empty slice value.
func TestProcessInSelector_EmptySlice(t *testing.T) {
	selector := database.Selector{
		Table:     "user",
		Field:     "id",
		Predicate: "IN",
		Value:     []int{},
	}

	column, values := processInSelector(selector)

	expectedColumn := "`user`.`id` IN ()"
	expectedValues := []any{}

	assert.Equal(t, expectedColumn, column)
	assert.Equal(t, expectedValues, values)
}

// TestProcessInSelector_WithStringSliceValue tests the processInSelector
// function with a slice of strings.
func TestProcessInSelector_WithStringSliceValue(t *testing.T) {
	selector := database.Selector{
		Table:     "user",
		Field:     "name",
		Predicate: "IN",
		Value:     []string{"Alice", "Bob", "Charlie"},
	}

	column, values := processInSelector(selector)

	expectedColumn := "`user`.`name` IN (?,?,?)"
	expectedValues := []any{"Alice", "Bob", "Charlie"}

	assert.Equal(t, expectedColumn, column)
	assert.Equal(t, expectedValues, values)
}

// TestProcessInSelector_NilValue tests the processInSelector function with a
// nil value.
func TestProcessInSelector_NilValue(t *testing.T) {
	selector := database.Selector{
		Table:     "user",
		Field:     "id",
		Predicate: "IN",
		Value:     nil,
	}

	column, values := processInSelector(selector)

	expectedColumn := "`user`.`id` IN (?)"
	expectedValues := []any{nil}

	assert.Equal(t, expectedColumn, column)
	assert.Equal(t, expectedValues, values)
}

// TestProcessDefaultSelector_WithValue tests the processDefaultSelector
// function with a non-nil value.
func TestProcessDefaultSelector_WithValue(t *testing.T) {
	selector := database.Selector{
		Table:     "user",
		Field:     "name",
		Predicate: "=",
		Value:     "Alice",
	}

	column, values := processDefaultSelector(selector)

	expectedColumn := "`user`.`name` = ?"
	expectedValues := []any{"Alice"}

	assert.Equal(t, expectedColumn, column)
	assert.Equal(t, expectedValues, values)
}

// TestProcessDefaultSelector_WithoutTable tests the processDefaultSelector
// function without an empty field value.
func TestProcessDefaultSelector_WithoutTable(t *testing.T) {
	selector := database.Selector{
		Table:     "",
		Field:     "name",
		Predicate: "=",
		Value:     "Alice",
	}

	column, values := processDefaultSelector(selector)

	expectedColumn := "`name` = ?"
	expectedValues := []any{"Alice"}

	assert.Equal(t, expectedColumn, column)
	assert.Equal(t, expectedValues, values)
}

// TestProcessDefaultSelector_NilValue tests the processDefaultSelector function
// with a nil value and "=" predicate.
func TestProcessDefaultSelector_NilValue(t *testing.T) {
	selector := database.Selector{
		Table:     "user",
		Field:     "name",
		Predicate: "=",
		Value:     nil,
	}

	column, values := processDefaultSelector(selector)

	expectedColumn := "`user`.`name` IS NULL"
	assert.Equal(t, expectedColumn, column)
	assert.Nil(t, values)
}

// TestProcessNullSelector_NotEquals tests the processNullSelector function with
// a "!=" predicate.
func TestProcessNullSelector_NotEquals(t *testing.T) {
	selector := database.Selector{
		Table:     "user",
		Field:     "name",
		Predicate: "!=",
		Value:     nil,
	}

	column, values := processNullSelector(selector)

	expectedColumn := "`user`.`name` IS NOT NULL"
	assert.Equal(t, expectedColumn, column)
	assert.Nil(t, values)
}

// TestProcessNullSelector_InvalidPredicate tests the processNullSelector
// function with an unsupported predicate.
func TestProcessNullSelector_InvalidPredicate(t *testing.T) {
	selector := database.Selector{
		Table:     "user",
		Field:     "name",
		Predicate: ">",
		Value:     nil,
	}

	column, values := processNullSelector(selector)

	// Expect an empty column and nil values because of non-supported predicate
	assert.Equal(t, "", column)
	assert.Nil(t, values)
}

// TestProcessNullSelector_WithoutTable tests the processNullSelector function
// without a table.
func TestProcessNullSelector_WithoutTable(t *testing.T) {
	selector := database.Selector{
		Table:     "",
		Field:     "name",
		Predicate: "=",
		Value:     nil,
	}

	column, values := processNullSelector(selector)

	expectedColumn := "`name` IS NULL"
	assert.Equal(t, expectedColumn, column)
	assert.Nil(t, values)
}

// TestBuildNullClause_WithTable tests the buildNullClause function when a table
// is provided.
func TestBuildNullClause_WithTable(t *testing.T) {
	// Test case where a table is provided
	selector := database.Selector{
		Table: "user",
		Field: "name",
	}

	clause := buildNullClause(selector, "IS")
	expectedClause := "`user`.`name` IS NULL"

	assert.Equal(t, expectedClause, clause)
}

// TestBuildNullClause_WithoutTable tests the buildNullClause function when no
// table is provided.
func TestBuildNullClause_WithoutTable(t *testing.T) {
	// Test case where no table is provided
	selector := database.Selector{
		Table: "",
		Field: "name",
	}

	clause := buildNullClause(selector, "IS NOT")
	expectedClause := "`name` IS NOT NULL"

	assert.Equal(t, expectedClause, clause)
}

// TestBuildNullClause_EmptyField tests the buildNullClause function with an
// empty field.
func TestBuildNullClause_EmptyField(t *testing.T) {
	// Test case with an empty field
	selector := database.Selector{
		Table: "user",
		Field: "",
	}

	clause := buildNullClause(selector, "IS")
	expectedClause := "`user`.`` IS NULL"

	assert.Equal(t, expectedClause, clause)
}

// TestCreatePlaceholdersAndValues tests the createPlaceholdersAndValues
// function.
func TestCreatePlaceholdersAndValues(t *testing.T) {
	// Test case 1: Slice with multiple values
	values := []int{1, 2, 3}
	value := reflect.ValueOf(values)

	placeholders, actualValues := createPlaceholdersAndValues(value)
	expectedPlaceholders := "?,?,?"
	expectedValues := []any{1, 2, 3}

	assert.Equal(t, expectedPlaceholders, placeholders)
	assert.Equal(t, expectedValues, actualValues)

	// Test case 2: Empty slice
	emptyValues := []int{}
	value = reflect.ValueOf(emptyValues)

	placeholders, actualValues = createPlaceholdersAndValues(value)
	expectedPlaceholders = ""
	expectedValues = []any{}

	assert.Equal(t, expectedPlaceholders, placeholders)
	assert.Equal(t, expectedValues, actualValues)

	// Test case 3: Slice with a single value
	singleValue := []string{"test"}
	value = reflect.ValueOf(singleValue)

	placeholders, actualValues = createPlaceholdersAndValues(value)
	expectedPlaceholders = "?"
	expectedValues = []any{"test"}

	assert.Equal(t, expectedPlaceholders, placeholders)
	assert.Equal(t, expectedValues, actualValues)
}

// TestCreatePlaceholders tests the createPlaceholders function.
func TestCreatePlaceholders(t *testing.T) {
	// Test case 1: Creating placeholders for 3 values
	placeholders := createPlaceholders(3)
	expectedPlaceholders := "?,?,?"

	assert.Equal(t, expectedPlaceholders, placeholders)

	// Test case 2: Creating placeholders for 1 value
	placeholders = createPlaceholders(1)
	expectedPlaceholders = "?"

	assert.Equal(t, expectedPlaceholders, placeholders)

	// Test case 3: Creating placeholders for 0 values
	placeholders = createPlaceholders(0)
	expectedPlaceholders = ""

	assert.Equal(t, expectedPlaceholders, placeholders)
}
