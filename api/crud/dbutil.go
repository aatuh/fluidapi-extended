package crud

import (
	"fmt"
	"reflect"
	"time"

	"github.com/pakkasys/fluidapi/database"
)

// EntityOption defines a functional option for configuring an entity.
type EntityOption[T any] func(T)

// MustGetSelector creates a new Selector with the given parameters. It panics
// if the provided value's type is not assignable to the expected type.
func MustGetSelector(
	tableName string,
	entity any,
	fieldName string,
	predicate database.Predicate,
	value any,
) *database.Selector {
	mustValidateDBMapping(entity, fieldName, value)
	return &database.Selector{
		Table:     tableName,
		Field:     fieldName,
		Predicate: predicate,
		Value:     value,
	}
}

// MustGetUpdate creates a new UpdateField with the given parameters. It panics
// if the provided value's type is not assignable to the expected type.
func MustGetUpdate(
	entity any,
	fieldName string,
	value any,
) *database.UpdateField {
	mustValidateDBMapping(entity, fieldName, value)
	return &database.UpdateField{
		Field: fieldName,
		Value: value,
	}
}

// WithOption is a generic functional option that sets a field on an object.
// The field parameter should match the object's struct `db` tag.
// It attempts to automatically handle differences between pointer and
// non-pointer values.
// Only single level pointer differences are handled and passing a multi-level
// pointer field will panic.
//
// Example:
//
//	type SomeEntity struct {
//	    ID   uuid.UUID `db:"id"`
//	    Data string    `db:"data"`
//	}
//
//	field = "data"
//	value = "example data"
//
//	Output:
//
//	someEntity.Data = "example data"
//
// Limitations:
//   - Only single-level pointer differences are handled.
//     Multi-level pointers (e.g. **SomeEntity) are not supported.
//   - If the field is not found in the object, or if the provided value's type
//     is not assignable (even after pointer adjustments), the function will
//     panic.
func WithOption[T any](field string, value any) EntityOption[T] {
	return func(t T) {
		v := reflect.ValueOf(t).Elem()
		typ := v.Type()
		// Ensure we're working on the underlying struct.
		if typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
			v = v.Elem()
		}
		var found bool
		for i := 0; i < typ.NumField(); i++ {
			f := typ.Field(i)
			if f.Tag.Get("db") == field {
				fieldVal := v.Field(i)
				// If the value is zero (or nil for pointers), simply return.
				if reflect.ValueOf(value).IsZero() {
					return
				}
				val := reflect.ValueOf(value)

				// If types directly match, set directly.
				if val.Type().AssignableTo(fieldVal.Type()) {
					fieldVal.Set(val)
					found = true
					break
				}

				// If the value is a pointer and non-nil, but the target is a
				// non-pointer, try setting with the dereferenced value.
				if val.Kind() == reflect.Ptr && !val.IsNil() &&
					val.Elem().Type().AssignableTo(fieldVal.Type()) {
					fieldVal.Set(val.Elem())
					found = true
					break
				}

				// If the target is a pointer but the value is not, allocate a
				// new pointer.
				if fieldVal.Type().Kind() == reflect.Ptr &&
					val.Type().AssignableTo(fieldVal.Type().Elem()) {
					ptrVal := reflect.New(val.Type())
					ptrVal.Elem().Set(val)
					fieldVal.Set(ptrVal)
					found = true
					break
				}

				panic(fmt.Sprintf(
					"WithOption: value for field %q must be assignable to type %v, got %v",
					field,
					fieldVal.Type(),
					reflect.TypeOf(value),
				))
			}
		}
		if !found {
			panic(fmt.Sprintf(
				"WithOption: field %q not found in object struct",
				field,
			))
		}
	}
}

// ScanRow uses reflection to build a slice of pointers to the object fields.
func ScanRow[T any](t *T, row database.Row) error {
	v := reflect.ValueOf(t).Elem()
	typ := v.Type()
	var pointers []any
	for i := 0; i < v.NumField(); i++ {
		if typ.Field(i).Tag.Get("db") == "" {
			continue
		}
		pointers = append(pointers, v.Field(i).Addr().Interface())
	}
	if err := row.Scan(pointers...); err != nil {
		return fmt.Errorf("ScanRow: failed to scan row: %w", err)
	}
	return nil
}

// InsertedValues uses reflection to generate slices of column names and values.
func InsertedValues[T any](t *T) ([]string, []any) {
	val := reflect.ValueOf(*t)
	typ := reflect.TypeOf(*t)
	var cols []string
	var vals []any
	for i := 0; i < val.NumField(); i++ {
		col := typ.Field(i).Tag.Get("db")
		if col == "" {
			continue
		}
		cols = append(cols, col)
		if typ.Field(i).Type == reflect.TypeOf(time.Time{}) {
			vals = append(vals, val.Field(i).Interface().(time.Time).UnixNano())
		} else {
			vals = append(vals, val.Field(i).Interface())
		}
	}
	return cols, vals
}

// mustValidateDBMapping validates that an object field name and type are valid.
// It panics if they are not.
func mustValidateDBMapping(obj any, fieldName string, value any) {
	// Get the mapping of db field names to their types.
	mapping := getDBMapping(obj)
	expectedType, ok := mapping[fieldName]
	if !ok {
		panic(fmt.Errorf("mustValidateDBMapping: unknown field %q", fieldName))
	}
	mustBeAssignable(value, fieldName, expectedType)
}

// getDBMapping builds a mapping from a struct's db tags to their field types.
func getDBMapping(obj any) map[string]reflect.Type {
	mapping := make(map[string]reflect.Type)
	typ := reflect.TypeOf(obj)
	// NumField only supports non-pointer structs.
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		dbTag := f.Tag.Get("db")
		if dbTag != "" {
			mapping[dbTag] = f.Type
		}
	}
	return mapping
}

// mustBeAssignable checks if a provided value is assignable to the expected type.
// It handles the common case where the endpoint input may be a pointer while the
// expected (database) type is a non-pointer. Specifically, if the provided value
// is a non-nil pointer, its underlying element type is checked against the expected type.
//
// Limitations:
//   - Only a single level of pointer indirection is handled.
//     Multi-level pointers (e.g. **Type) are not supported.
//   - If the value is nil, or if neither the value nor its dereferenced element
//     is assignable to the expected type, the function panics.
func mustBeAssignable(value any, fieldName string, expectedType reflect.Type) {
	if value == nil {
		panic(fmt.Sprintf(
			"mustBeAssignable: value for field %q is nil; expected type %v",
			fieldName,
			expectedType,
		))
	}
	vType := reflect.TypeOf(value)
	// Check direct assignability.
	if vType.AssignableTo(expectedType) {
		return
	}
	// If value is a pointer and non-nil, check if element type is assignable.
	if vType.Kind() == reflect.Ptr && !reflect.ValueOf(value).IsNil() {
		if vType.Elem().AssignableTo(expectedType) {
			return
		}
	}
	panic(fmt.Sprintf(
		"mustBeAssignable: value for field %q must be assignable to type %v, got %T",
		fieldName,
		expectedType,
		value,
	))
}
