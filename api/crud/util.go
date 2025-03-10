package crud

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/pakkasys/fluidapi-extended/api"
	"github.com/pakkasys/fluidapi-extended/api/types"
	"github.com/pakkasys/fluidapi/endpoint"
)

// ---------------------------------------------------------------------
// Constants and Helper Field Names
// ---------------------------------------------------------------------

const (
	FieldSelectors = "selectors"
	FieldUpdates   = "updates"
	FieldCount     = "count"
	FieldValue     = "value"
	FieldPredicate = "predicate"
)

// ---------------------------------------------------------------------
// Basic Types and Interfaces
// ---------------------------------------------------------------------

// CreateErrorMapping returns a map of masked error IDs to ExpectedErrors.
//
// Example:
//
//	expectedErrors := api.ExpectedErrors{
//	    {
//	        ID:       "error_id",
//	        MaskedID: "masked_id",
//	    },
//	}
//
// Output:
//
//	    "masked_id": {
//	        ID:       "error_id",
//	        MaskedID: "masked_id",
//	    },
//	}
func CreateErrorMapping(
	expectedErrors api.ExpectedErrors,
) map[string]api.ExpectedError {
	mappedErrors := map[string]api.ExpectedError{}
	for i := range expectedErrors {
		mappedErrors[expectedErrors[i].MaskedID] = expectedErrors[i]
	}
	return mappedErrors
}

// getAPIFieldToDBColumnMapping builds a mapping from API field names to DB
// fields.
//
// Example:
//
//	apiFields := []types.APIField{
//	    {
//	        APIName:  "name",
//	        DBColumn: "name",
//	    },
//	}
//	tableName := "users"
//
// Output:
//
//	mapping := map[string]endpoint.DBField{
//	    "name": {
//	        Table:  "users",
//	        Column: "name",
//	    },
//	}
func getAPIFieldToDBColumnMapping(
	apiFields []types.APIField,
	tableName string,
) map[string]endpoint.DBField {
	mapping := make(map[string]endpoint.DBField)
	for _, def := range apiFields {
		mapping[def.APIName] = endpoint.DBField{
			Table:  tableName,
			Column: def.DBColumn,
		}
	}
	return mapping
}

// dbEntityToMap creates a map from a database entity based on APIFields.
// The map will have the same keys as the APIFields, but the values will
// be the values from the database entity.
// It will modify the output map in place.
//
// Example:
//
//	type User struct {
//		ID int `db:"id"`
//		Name string `db:"name"`
//	}
//
//	entity := User{
//		ID: 123,
//		Name: "Alice"
//	}
//	apiFields := APIFields{
//		{APIName: "user_id", DBColumn: "id"},
//		{APIName: "user_name", DBColumn: "name"},
//	}
//	output := make(map[string]any)
//
//	Output:
//
//	map["user_id": 123, "user_name": "Alice"]
//
// Accepted Input:
//   - A struct or pointer to a struct with valid `db` tags.
//   - An APIFields slice with DBColumn values matching the struct's tags.
//
// Error Cases:
//   - Passing a nil or non-struct value.
//   - Missing expected `db` tag for an entity field.
func dbEntityToMap(
	entity any,
	apiFields types.APIFields,
	output map[string]any,
) (*map[string]any, error) {
	v := reflect.ValueOf(entity)
	if !v.IsValid() || (v.Kind() == reflect.Ptr && v.IsNil()) {
		return nil, fmt.Errorf("dbEntityToMap: received nil entity")
	}
	v = reflect.Indirect(v)
	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf(
			"dbEntityToMap: expected a struct but got %v", v.Kind(),
		)
	}
	t := v.Type()
	dbFieldMap := make(map[string]reflect.Value)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		dbTag := field.Tag.Get("db")
		if dbTag != "" {
			dbFieldMap[dbTag] = v.Field(i)
		}
	}
	for _, apiField := range apiFields {
		val, ok := dbFieldMap[apiField.DBColumn]
		if !ok {
			return nil, fmt.Errorf(
				"dbEntityToMap: no matching db field for %s (API field: %s)",
				apiField.DBColumn,
				apiField.APIName,
			)
		}
		output[apiField.APIName] = val.Interface()
	}
	return &output, nil
}

// structToDBEntity creates an entity to value map from an input struct.
// It expects that the struct has an inner key from which it will create the
// entity.
// It expects that the input has a field (or JSON tag) matching inputKey, which
// must be a struct.
// For each types.APIField, it finds a matching field in the inner struct (by JSON
// tag or field name).
//
// Example:
//
//	type InnerData struct {
//	    UserName string `json:"user_name"`
//	    Age      int    `json:"age"`
//	}
//
//	type Input struct {
//	    Data InnerData `json:"data"`
//	}
//
//	input := Input{
//	    Data: InnerData{
//	        UserName: "Alice",
//	        Age:      30,
//	    },
//	}
//
//	inputKey := "data"
//
//	apiFields := APIFields{
//	    {APIName: "user_name", DBColumn: "username"},
//	    {APIName: "age",       DBColumn: "user_age"},
//	}
//
//	dbOptionFn := func(field string, value any) utildatabase.Option[Entity] {
//	    // return an option for the given field and value
//	}
//
//	dbEntityFn := func(opts ...utildatabase.Option[Entity]) Entity {
//	    // create and return an entity using the options
//	}
//
// Accepted Input:
//   - A struct or pointer to a struct with a field (or JSON tag) matching inputKey.
//   - The identified field must be a struct.
//   - An APIFields slice with APIName values matching the inner struct fields.
//
// Error Cases:
//   - Input is not a non-nil struct or pointer to a struct.
//   - The field identified by inputKey is missing.
//   - The field identified by inputKey is not a struct.
func structToDBEntity[Struct any, Entity any](
	obj *Struct, inputKey string, apiFields types.APIFields,
) (map[string]any, error) {
	// Ensure input is a non-nil struct or pointer to a struct.
	v := reflect.ValueOf(obj)
	v = reflect.Indirect(v)
	if !v.IsValid() || v.Kind() != reflect.Struct {
		return nil, fmt.Errorf(
			"structToDBEntity: input must be a non-nil struct or pointer to a struct",
		)
	}
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// Locate the inner object using inputKey.
	var inner reflect.Value
	found := false
	typ := v.Type()
	for i := range v.NumField() {
		field := typ.Field(i)
		// First check the JSON tag.
		if tag := field.Tag.Get("json"); tag != "" {
			tagName := tag
			if idx := strings.Index(tag, ","); idx != -1 {
				tagName = tag[:idx]
			}
			if tagName == inputKey {
				inner = v.Field(i)
				found = true
				break
			}
		}
		// Otherwise compare the field name.
		if field.Name == inputKey {
			inner = v.Field(i)
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf(
			"structToDBEntity: inputKey %q not found in input", inputKey,
		)
	}

	// Dereference pointer if necessary and check that inner is a struct.
	if inner.Kind() == reflect.Ptr {
		inner = inner.Elem()
	}
	if inner.Kind() != reflect.Struct {
		return nil, fmt.Errorf(
			"structToDBEntity: inner field %q must be a struct", inputKey,
		)
	}

	// Iterate over the types.APIFields and match each one against the inner struct.
	//var opts []utildatabase.Option[Entity]
	colsAndValues := map[string]any{}
	innerType := inner.Type()
	for _, apiField := range apiFields {
		var fieldValue reflect.Value
		matched := false
		for i := 0; i < inner.NumField(); i++ {
			field := innerType.Field(i)
			// Check JSON tag first.
			var fieldName string
			if tag := field.Tag.Get("json"); tag != "" {
				fieldName = tag
				if idx := strings.Index(tag, ","); idx != -1 {
					fieldName = tag[:idx]
				}
			} else {
				fieldName = field.Name
			}
			// Check if the field name matches the API name.
			if fieldName == apiField.APIName {
				fieldValue = inner.Field(i)
				matched = true
				break
			}
		}
		if matched {
			colsAndValues[apiField.DBColumn] = fieldValue.Interface()
		}
	}
	return colsAndValues, nil
}

// toGenericGetOutput creates an output map from a slice of entities.
func toGenericGetOutput[Entity any, Output any](
	entities []Entity,
	count int,
	apiFields types.APIFields,
	pluralField string,
	countField string,
	obj *Output,
) (*Output, error) {
	outputMap := make(map[string]any)
	for _, field := range apiFields {
		switch field.APIName {
		case pluralField:
			var objects []map[string]any
			for _, e := range entities {
				mapped, err := dbEntityToMap(e, field.Nested, map[string]any{})
				if err != nil {
					return nil, err
				}
				objects = append(objects, *mapped)
			}
			outputMap[pluralField] = objects
		case countField:
			outputMap[countField] = count
		default:
			return nil, fmt.Errorf(
				"toGenericGetOutput: unknown output field: %s", field.APIName,
			)
		}
	}

	err := mapstructure.Decode(outputMap, obj)
	if err != nil {
		return nil, err
	}

	return obj, nil
}

// genericGetAPIFields creates types.APIFields for a get request.
func genericGetAPIFields(
	apiFields types.APIFields,
	predicates map[string]endpoint.Predicates,
	orderable []string,
) types.APIFields {
	mustMatchPredicates(predicates, apiFields)
	mustMatchFields(orderable, apiFields)
	return types.APIFields{
		selectorFieldsEntry(apiFields, predicates),
		orderFieldsEntry(apiFields, orderable),
		pageFieldEntry(),
		countFieldEntry(),
	}
}

// genericGetOutputAPIFields creates types.APIFields for a get request.
func genericGetOutputAPIFields(
	outputKey string, apiFields types.APIFields,
) types.APIFields {
	return types.APIFields{
		{
			APIName: outputKey,
			Nested:  apiFields,
		},
		{
			APIName: FieldCount,
		},
	}
}

// genericUpdateAPIFields creates APIFields for an update request.
func genericUpdateAPIFields(
	selectableAPIFields types.APIFields,
	predicates map[string]endpoint.Predicates,
	updatableAPIFields types.APIFields,
) types.APIFields {
	return types.APIFields{
		selectorFieldsEntry(selectableAPIFields, predicates),
		updatesEntry(updatableAPIFields),
	}
}

// genericDeleteAPIFields creates APIFields for a delete request.
func genericDeleteAPIFields(
	selectableAPIFields types.APIFields,
	predicates map[string]endpoint.Predicates,
) types.APIFields {
	mustMatchPredicates(predicates, selectableAPIFields)
	return types.APIFields{
		selectorFieldsEntry(selectableAPIFields, predicates),
	}
}

// mustMatchPredicates returns a map of predicates that matches the APIFields.
// It panics if a predicate is not found for a field.
func mustMatchPredicates(
	predicates map[string]endpoint.Predicates,
	apiFields types.APIFields,
) map[string]endpoint.Predicates {
	for _, field := range apiFields {
		if predicates[field.APIName] == nil {
			panic(fmt.Sprintf(
				"mustMatchPredicates: predicate not found for field: %s",
				field.APIName,
			))
		}
	}
	return predicates
}

// mustMatchFields returns a slice of fields that matches the APIFields.
// It panics if a field is not found.
func mustMatchFields(fields []string, apiFields types.APIFields) {
	for _, field := range fields {
		field, found := apiFields.GetAPIField(field)
		if !found {
			panic(fmt.Sprintf(
				"mustMatchFields: unknown field %s", field.APIName,
			))
		}
	}
}

// selectorFieldsEntry creates an APIField for selector fields.
func selectorFieldsEntry(
	from types.APIFields,
	fieldPredicates map[string]endpoint.Predicates,
) types.APIField {
	return types.APIField{
		APIName: FieldSelectors,
		Nested:  mustSelectorFields(from, fieldPredicates),
	}
}

// mustSelectorFields creates APIFields for selector fields. It panics if a
// predicate is not found for a field.
func mustSelectorFields(
	from types.APIFields,
	fieldPredicates map[string]endpoint.Predicates,
) types.APIFields {
	var fields []types.APIField
	for _, field := range from {
		predicates, ok := fieldPredicates[field.APIName]
		if !ok {
			panic(fmt.Sprintf(
				"mustSelectorFields: no predicates found for field %q",
				field.APIName,
			))
		}
		fields = append(fields, selectorField(field, predicates))
	}
	return fields

}

// selectorField creates an APIField for a selector field.
func selectorField(from types.APIField, predicates endpoint.Predicates) types.APIField {
	return types.APIField{
		APIName:  from.APIName,
		DBColumn: from.DBColumn,
		Nested: []types.APIField{
			{
				APIName:  FieldValue,
				Validate: from.Validate,
				Type:     from.Type,
				Required: true,
			},
			{
				APIName: FieldPredicate,
				Validate: []string{
					"string",
					fmt.Sprintf(
						"oneof=%s", strings.Join(predicates.StrSlice(), " "),
					),
				},
				Type:     "string",
				Required: true,
			},
		},
	}
}

// orderFieldsEntry creates an APIField for ordering fields.
func orderFieldsEntry(
	from types.APIFields,
	orderableFields []string,
) types.APIField {
	return types.APIField{
		APIName: "orders",
		Nested:  orderFields(from, orderableFields),
	}
}

// orderFields creates APIFields for ordering fields.
func orderFields(from types.APIFields, orderableFields []string) types.APIFields {
	fields := []types.APIField{}
	for _, field := range from.MustGetAPIFields(orderableFields) {
		fields = append(fields, types.APIField{
			APIName: field.APIName,
			Validate: []string{
				"string",
				fmt.Sprintf(
					"oneof=%s %s %s %s",
					endpoint.DirectionAsc,
					endpoint.DirectionDesc,
					endpoint.DirectionAscending,
					endpoint.DirectionDescending,
				),
			},
			Type: "string",
		})
	}
	return fields
}

// pageFieldEntry creates an APIField for pagination.
func pageFieldEntry() types.APIField {
	return types.APIField{
		APIName: "page",
		Nested: []types.APIField{
			{
				APIName:  "offset",
				Validate: []string{"int64", "min=0"},
				Type:     "int64",
			},
			{
				APIName:  "limit",
				Validate: []string{"int64", "min=1", "max=1000"},
				Default:  int64(1),
				Type:     "int64",
			},
		},
	}
}

// updatesEntry creates an APIField for updating fields.
func updatesEntry(from types.APIFields) types.APIField {
	return types.APIField{
		APIName: FieldUpdates,
		Nested:  updates(from),
	}
}

// updates creates APIFields for updating fields.
func updates(from types.APIFields) types.APIFields {
	var fields []types.APIField
	for _, field := range from {
		fields = append(fields, update(field))
	}
	return fields

}

// update creates an APIField for updating fields.
func update(from types.APIField) types.APIField {
	return types.APIField{
		APIName:  from.APIName,
		DBColumn: from.DBColumn,
	}
}

// countFieldEntry creates an APIField for counting fields.
func countFieldEntry() types.APIField {
	return types.APIField{
		APIName:  FieldCount,
		Validate: []string{"bool"},
		Type:     "bool",
	}
}

// mustApplyErrorMapping applies a mapping of error IDs to new IDs by modifying
// the masked error IDs in the default errors. If an error ID is not found in
// the mapping, it panics.
func mustApplyErrorMapping(
	defaultErrors api.ExpectedErrors,
	mappedErrors map[string]api.ExpectedError,
) api.ExpectedErrors {
	newDefaultErrors := defaultErrors
	for i := range mappedErrors {
		defaultError := defaultErrors.GetByID(mappedErrors[i].ID)
		if defaultError == nil {
			panic(fmt.Sprintf(
				"mustApplyErrorMapping: mapped error ID %q not found in default errors",
				mappedErrors[i].ID,
			))
		}
		newDefaultErrors = newDefaultErrors.With(mappedErrors[i])
	}
	return newDefaultErrors
}
