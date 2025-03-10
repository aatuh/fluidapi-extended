package database

import (
	"github.com/pakkasys/fluidapi/database"
)

// EntityOption defines a functional option for configuring an entity.
type EntityOption[T any] func(T)

// OptionEntityFn is a function that returns an entity with the given options.
type OptionEntityFn[Entity database.CRUDEntity] func(
	opts ...EntityOption[Entity],
) Entity

// EntityQuery represents a query for an entity.
type EntityQuery[Entity database.CRUDEntity] struct {
	tableName      string
	entityFn       func() Entity
	optionEntityFn OptionEntityFn[Entity]
	selectors      database.Selectors
	updates        database.Updates
	options        []EntityOption[database.CRUDEntity]
}

// NewQuery creates a new NewEntityQuery for an entity.
//
// Parameters:
//   - entityFn: A function that returns the entity.
//
// Returns:
//   - *EntityQuery: The new EntityQuery.
func NewEntityQuery[Entity database.CRUDEntity](
	tableName string,
	entityFn func() Entity,
	optionEntityFn OptionEntityFn[Entity],
) *EntityQuery[Entity] {
	return &EntityQuery[Entity]{
		tableName:      tableName,
		entityFn:       entityFn,
		optionEntityFn: optionEntityFn,
		selectors:      database.Selectors{},
		updates:        database.Updates{},
		options:        []EntityOption[database.CRUDEntity]{},
	}
}

// AddSelector appends a selector to the query.
//
// Parameters:
//   - field: The field to select.
//   - predicate: The predicate to apply.
//   - value: The value to compare.
//
// Returns:
//   - *EntityQuery: The updated EntityQuery.
func (q *EntityQuery[Entity]) AddSelector(
	field string, predicate database.Predicate, value any,
) *EntityQuery[Entity] {
	entity := q.entityFn()
	selector := MustGetSelector(
		q.tableName, entity, field, predicate, value,
	)
	q.selectors = append(q.selectors, *selector)
	return q
}

// Selectors returns the current selectors.
//
// Returns:
//   - database.Selectors: The current selectors.
func (q *EntityQuery[Entity]) Selectors() database.Selectors {
	return q.selectors
}

// AddUpdate appends an update clause to the query.
//
// Parameters:
//   - field: The field to update.
//   - value: The value to set.
//
// Returns:
//   - *EntityQuery: The updated EntityQuery.
func (q *EntityQuery[Entity]) AddUpdate(
	field string, value any,
) *EntityQuery[Entity] {
	q.updates = append(
		q.updates,
		*MustGetUpdate(q.entityFn(), field, value),
	)
	return q
}

// Updates returns the current updates.
//
// Returns:
//   - database.Updates: The current updates.
func (q *EntityQuery[Entity]) Updates() database.Updates {
	return q.updates
}

// Option creates an entity-specific option.
//
// Parameters:
//   - field: The field to set.
//   - value: The value to set.
//
// Returns:
//   - crud.EntityOption[*T]: The entity-specific option.
func (q *EntityQuery[Entity]) AddOption(
	field string, value any,
) *EntityQuery[Entity] {
	q.options = append(
		q.options, WithOption[database.CRUDEntity](field, value),
	)
	return q
}

// Entity returns the entity that is being queried with the set options.
//
// Returns:
//   - Entity: The entity that is being queried.
func (q *EntityQuery[Entity]) Entity() Entity {
	var entityOpts []EntityOption[Entity]
	for _, opt := range q.options {
		entityOpts = append(entityOpts, func(e Entity) {
			opt(e)
		})
	}
	return q.optionEntityFn(entityOpts...)
}
