package crud

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/pakkasys/fluidapi-extended/api/types"
)

// MustValidateGetOutput validates output for a GET endpoint. It expects that
// type T contains a "count" field and a field whose JSON tag equals outputKey,
// and that the type of that field is acceptable per the rules of
// MustStructToAPIFields.
func MustValidateGetOutput[T any](apiFields types.APIFields, outputKey string) {
	t := reflect.TypeOf((*T)(nil)).Elem()

	// Check that T has a "count" field.
	countFound := false
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if strings.Split(field.Tag.Get("json"), ",")[0] == FieldCount {
			countFound = true
			break
		}
	}
	if !countFound {
		panic(fmt.Sprintf(
			"MustValidateGetOutput: type %s must have a field with JSON tag %q",
			t.Name(),
			FieldCount,
		))
	}

	// Find the field with JSON tag matching outputKey.
	var targetField reflect.StructField
	found := false
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if strings.Split(field.Tag.Get("json"), ",")[0] == outputKey {
			targetField = field
			found = true
			break
		}
	}
	if !found {
		panic(fmt.Sprintf(
			"MustValidateGetOutput: type %s must have a field with JSON tag \"%s\"",
			t.Name(),
			outputKey,
		))
	}

	// Determine the underlying type.
	var underlyingType reflect.Type
	switch targetField.Type.Kind() {
	case reflect.Slice:
		underlyingType = targetField.Type.Elem()
	default:
		panic(fmt.Sprintf(
			"MustValidateGetOutput: field for %s must be a slice of structs",
			outputKey,
		))
	}

	// Validate the underlying type using our reflection-based helper.
	mustStructToAPIFieldsFromType(underlyingType, apiFields)
}

// MustStructToAPIFields is a generic function that returns a slice of APIFields
// for any struct type T. For each field, if an APIField with a matching JSON
// name exists in allFields, that APIField is used. If the struct field has a
// "validate" tag, its rules override any validations from the APIField.
// Additionally, if a field is tagged with `required:"true"`, then it must
// be found in allFields or else the function panics. It will try to find the
// "required" and "type" tags and these will also override APIField settings.
func MustStructToAPIFields[T any](allFields types.APIFields) types.APIFields {
	apiFields := structToAPIFields(
		reflect.TypeOf((*T)(nil)).Elem(), allFields, 0,
	)
	err := validateAPIFieldTypes[T](apiFields)
	if err != nil {
		panic(err)
	}
	return apiFields
}

// mustStructToAPIFieldsFromType is a generic function that returns a slice of
// APIFields for any struct type T. For each field, if an APIField with a
// matching JSON name exists in allFields, that APIField is used. If the struct
// field has a "validate" tag, its rules override any validations from the
// APIField. Additionally, if a field is tagged with `required:"true"`, then
// it must be found in allFields or else the function panics. It will try to find
// the "required" and "type" tags and these will also override APIField settings.
func mustStructToAPIFieldsFromType(
	t reflect.Type, allFields types.APIFields,
) types.APIFields {
	apiFields := structToAPIFields(t, allFields, 0)
	err := validateAPIFieldTypesWithType(t, apiFields)
	if err != nil {
		panic(err)
	}
	return apiFields
}

// structToAPIFields processes the given type recursively.
func structToAPIFields(
	t reflect.Type, allFields types.APIFields, depth int,
) types.APIFields {
	// Check if t is empty an type.
	if t.Kind() == reflect.Interface {
		panic("structToAPIFields: type is a null interface")
	}
	// Check if t is not a struct type.
	if t.Kind() != reflect.Struct {
		panic(fmt.Sprintf("structToAPIFields: type %q is not a struct", t.Name()))
	}

	depth++
	if depth > 100 {
		panic(fmt.Sprintf("structToAPIFields: type %q is too deep", t.Name()))
	}

	// Build a lookup map from APIName to APIField.
	fieldsByName := make(map[string]types.APIField)
	fieldsByAlias := make(map[string]types.APIField)
	for _, f := range allFields {
		if f.Alias != "" {
			// Check if the alias is already in the map.
			if _, ok := fieldsByAlias[f.Alias]; ok {
				panic(fmt.Sprintf(
					"structToAPIFields: type %s has duplicate alias %q",
					t.Name(),
					f.Alias,
				))
			}
			fieldsByAlias[f.Alias] = f
		} else {
			// Check if the APIName is already in the map.
			if _, ok := fieldsByName[f.APIName]; ok {
				panic(fmt.Sprintf(
					"structToAPIFields: type %s has duplicate API name %q",
					t.Name(),
					f.APIName,
				))
			}
			fieldsByName[f.APIName] = f
		}
	}

	var inputFields types.APIFields
	// Iterate over each field of the struct type.
	for i := range t.NumField() {
		field := t.Field(i)
		// Extract the alias tag and use it if present.
		// Alias fields are useful when there are multiple fields with the same
		// JSON name.
		jsonTag := field.Tag.Get("alias")
		useAlias := false
		if jsonTag != "" {
			useAlias = true
		} else {
			// Extract the JSON tag.
			jsonTag = field.Tag.Get("json")
			if jsonTag == "" {
				// Skip fields without JSON tags.
				continue
			}
		}
		// Split tag options (e.g. "my_field,omitempty").
		parts := strings.Split(jsonTag, ",")
		fieldName := parts[0]

		// Get the "validate" tag.
		validateTag := field.Tag.Get("validate")
		var validateRules []string
		if validateTag != "" {
			validateRules = strings.Split(validateTag, ",")
		}

		// Get the "required" tag.
		requiredTag := field.Tag.Get("required")
		var required bool
		if requiredTag != "" {
			parsed, err := strconv.ParseBool(strings.TrimSpace(requiredTag))
			if err != nil {
				panic(fmt.Sprintf(
					"structToAPIFields: invalid value for required tag on field %s: %v",
					fieldName,
					err),
				)
			}
			required = parsed
		}

		// Get the "ext" tag
		extTag := field.Tag.Get("ext")

		// Lookup the APIField.
		var apiField types.APIField
		var exists bool
		if useAlias {
			apiField, exists = fieldsByAlias[fieldName]
			if !exists {
				panic(fmt.Sprintf(
					"structToAPIFields: alias field %s not found in fieldsByAlias",
					fieldName,
				))
			}
		} else {
			apiField, exists = fieldsByName[fieldName]
		}
		if exists {
			// Override validations if the struct provides them.
			if len(validateRules) > 0 {
				apiField.Validate = validateRules
			}
		} else {
			// Field not found in allFields.
			if extTag == "true" {
				panic(fmt.Sprintf(
					"structToAPIFields: APIField not found for external field: %s",
					fieldName,
				))
			}
			apiField = types.APIField{
				APIName:  fieldName,
				Validate: validateRules,
			}
		}

		// Override the APIField.Required if the "required" tag is present.
		if requiredTag != "" {
			apiField.Required = required
		}

		// Override the APIField.Type if the "type" tag is present.
		typeTag := field.Tag.Get("type")
		if typeTag != "" {
			apiField.Type = typeTag
		} else if apiField.Type == "" {
			// Take type from the struct field and if it is a pointer type,
			// dereference it.
			if field.Type.Kind() == reflect.Ptr {
				field.Type = field.Type.Elem()
				apiField.Type = field.Type.String()
			} else {
				apiField.Type = field.Type.String()
			}
		}

		// Process potential nested types.
		fieldType := field.Type
		// Dereference pointer types.
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}

		// Check if the field is a slice.
		if fieldType.Kind() == reflect.Slice {
			elemType := fieldType.Elem()
			// Dereference element pointer if necessary.
			if elemType.Kind() == reflect.Ptr {
				elemType = elemType.Elem()
			}
			if elemType.Kind() == reflect.Struct {
				// Recursively process the slice element type.
				apiField.Nested = structToAPIFields(elemType, allFields, depth)
			}
		} else if fieldType.Kind() == reflect.Struct {
			// Recursively process the nested struct.
			apiField.Nested = structToAPIFields(fieldType, allFields, depth)
		}

		inputFields = append(inputFields, apiField)
	}
	return inputFields
}

// getJSONTag extracts the JSON key from a struct field’s tag.
func getJSONTag(field reflect.StructField) string {
	tag := field.Tag.Get("json")
	if tag == "" {
		return field.Name // fallback to field name if no json tag exists
	}
	parts := strings.Split(tag, ",")
	return parts[0]
}

// findFieldByJSONTag returns the first field in t whose JSON tag matches jsonName.
func findFieldByJSONTag(t reflect.Type, jsonName string) *reflect.StructField {
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if getJSONTag(f) == jsonName {
			return &f
		}
	}
	return nil
}

// ValidateAPIFields is a generic function that validates any struct type T
// against a slice of APIField definitions using a custom type map.
func validateAPIFieldTypes[T any](apiFields []types.APIField) error {
	var t T
	rt := reflect.TypeOf(t)
	// If T is a pointer, use its element type.
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}
	if rt.Kind() != reflect.Struct {
		return fmt.Errorf("type %s is not a struct", rt.Name())
	}
	return validateAPIFieldTypesForType(rt, apiFields)
}

func validateAPIFieldTypesWithType(
	rt reflect.Type, apiFields []types.APIField,
) error {
	// If T is a pointer, use its element type.
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}
	if rt.Kind() != reflect.Struct {
		return fmt.Errorf("type %s is not a struct", rt.Name())
	}
	return validateAPIFieldTypesForType(rt, apiFields)
}

// validateAPIFieldTypesForType recursively checks that each APIField in apiFields
// matches the corresponding struct field in t.
// func validateAPIFieldTypesForType(t reflect.Type, apiFields []APIField, customTypes map[string]string) error {
func validateAPIFieldTypesForType(
	t reflect.Type, apiFields []types.APIField,
) error {
	// Ensure that t is a struct
	if t.Kind() != reflect.Struct && t.Kind() != reflect.Slice {
		return fmt.Errorf("type %s is not a struct or slice", t.Name())
	}

	for _, apiField := range apiFields {
		underlyingType := t
		if t.Kind() == reflect.Slice {
			underlyingType = t.Elem()
		}
		// Find the matching struct field by comparing json tags.
		field := findFieldByJSONTag(underlyingType, apiField.APIName)
		if field == nil {
			// If the struct does not have a matching field, skip validation.
			continue
		}

		// Get the field type and dereference pointer types.
		ft := field.Type
		if ft.Kind() == reflect.Ptr {
			ft = ft.Elem()
		}

		// If the APIField has nested APIFields, then we expect the field to be a struct.
		if len(apiField.Nested) > 0 {
			if ft.Kind() != reflect.Struct && ft.Kind() != reflect.Slice {
				return fmt.Errorf(
					"validateAPIFieldTypesForType: field %s is expected to be a struct or slice for nested API fields, got %s",
					field.Name,
					ft.Kind().String(),
				)
			}
			if err := validateAPIFieldTypesForType(ft, apiField.Nested); err != nil {
				return fmt.Errorf(
					"validateAPIFieldTypesForType: nested type validation error on field %s: %w",
					field.Name,
					err,
				)
			}
			// Skip further type-checking for nested fields.
			continue
		}

		// // Determine the “actual” type name for comparison.
		// var actualType string
		// if custom, ok := customTypes[ft.String()]; ok {
		// 	// If the field’s Go type (e.g. "uuid.UUID") is in the custom type map,
		// 	// use the custom name (e.g. "uuid").
		// 	actualType = custom
		// } else if _, ok := builtInTypes[ft.String()]; ok {
		// 	// Otherwise, if the type is one of our recognized built-in types, use it.
		// 	actualType = ft.String()
		// } else {
		// 	// If the type isn’t recognized in either list, skip checking.
		// 	continue
		// }

		// // Compare the determined type against the expected APIField type.
		// if apiField.Type == "" {
		// 	return fmt.Errorf("field %s type is empty", field.Name)
		// } else if actualType != apiField.Type {
		// 	return fmt.Errorf(
		// 		"field %s type mismatch: expected %s, got %s",
		// 		field.Name,
		// 		apiField.Type,
		// 		actualType,
		// 	)
		// }
	}
	return nil
}
