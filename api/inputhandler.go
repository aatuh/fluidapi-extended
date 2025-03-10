package api

import (
	"fmt"
	"net/http"

	"github.com/pakkasys/fluidapi-extended/api/types"
	"github.com/pakkasys/fluidapi/core"
)

// ValidationError represents a validation error
var ValidationError = core.NewAPIError("VALIDATION_ERROR")

// FieldError represents a field-level validation error
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrorData contains a list of field-level validation errors
type ValidationErrorData struct {
	Errors []FieldError `json:"errors"`
}

// MapInputHandler handles the input of a request as a map.
type MapInputHandler struct {
	apiFields     types.APIFields
	conversionMap map[string]func(any) any
	customRules   map[string]func(any) error
}

// NewMapInputHandler creates a new MapInputHandler.
//
// Parameters:
//   - apiFields: The APIFields to use for validation.
//   - conversionMap: A map of conversion functions for fields.
//   - customRules: A map of custom validation rules for fields.
//
// Returns:
//   - *MapInputHandler: The new MapInputHandler.
func NewMapInputHandler(
	apiFields types.APIFields,
	conversionMap map[string]func(any) any,
	customRules map[string]func(any) error,
) *MapInputHandler {
	inputHandler := &MapInputHandler{
		apiFields:     apiFields,
		conversionMap: conversionMap,
		customRules:   customRules,
	}

	// Validate APIFields.
	// TODO: Could cache results from these calls.
	err := inputHandler.testValidationRules(apiFields)
	if err != nil {
		panic(err)
	}
	_, err = inputHandler.mapFieldConfigFromAPIFields(apiFields)
	if err != nil {
		panic(err)
	}

	return inputHandler
}

// Handle processes the request input by creating a map presentation from it and
// validating it.
//
// Parameters:
//   - w: The HTTP response writer.
//   - r: The HTTP request.
//
// Returns:
//   - map[string]any: The map presentation of the input.
//   - error: Any error that occurred during processing.
func (h *MapInputHandler) Handle(
	w http.ResponseWriter, r *http.Request,
) (map[string]any, error) {
	input, err := h.pickMap(r, h.apiFields)
	if err != nil {
		return nil, err
	}
	if err := h.validateMap(input, h.apiFields); err != nil {
		return nil, ValidationError.WithData(ValidationErrorData{
			Errors: []FieldError{
				{
					Field:   "input",
					Message: err.Error(),
				},
			},
		})
	}

	return input, nil
}

// pickMap picks a map from the request using the provided APIFields.
func (h *MapInputHandler) pickMap(
	r *http.Request, apiFields types.APIFields,
) (map[string]any, error) {
	// Convert the APIFields to a MapFieldConfig.
	mapFieldConfig, err := h.mapFieldConfigFromAPIFields(apiFields)
	if err != nil {
		return nil, err
	}

	// Pick the map.
	return NewObjectPicker(NewURLEncoder(), h.conversionMap).
		PickMap(r, mapFieldConfig)
}

// getValidator returns a new instance of the getValidator.
func (h *MapInputHandler) getValidator() *Validate {
	return NewValidate(h.customRules)
}

// mapFieldConfigFromAPIFields converts an APIFields to a MapFieldConfig.
func (h *MapInputHandler) mapFieldConfigFromAPIFields(
	apiFields types.APIFields,
) (*MapFieldConfig, error) {
	cfg := &MapFieldConfig{
		Fields: make(map[string]*MapFieldConfig),
	}
	// Convert each field to a MapFieldConfig.
	for _, field := range apiFields {
		// Convert field to MapFieldConfig.
		// var expectedType string
		// if len(field.Validate) != 0 {
		// 	expectedType = field.Type
		// 	if expectedType == "" {
		// 		return nil, fmt.Errorf(
		// 			"type must be set for field %q", field.APIName,
		// 		)
		// 	}
		// }
		fieldCfg := &MapFieldConfig{
			Source:       field.Source,
			ExpectedType: field.Type,
			DefaultValue: field.Default,
			Optional:     !field.Required,
		}
		// Convert any nested fields recursively.
		if len(field.Nested) > 0 {
			fields, err := h.mapFieldConfigFromAPIFields(field.Nested)
			if err != nil {
				return nil, err
			}
			fieldCfg.Fields = fields.Fields

		}
		cfg.Fields[field.APIName] = fieldCfg
	}

	return cfg, nil
}

// TODO: Bug if same named fields (in nested fields?).
// validateMap validates an input map against the provided APIFields.
func (h *MapInputHandler) validateMap(
	input map[string]any, apiFields types.APIFields,
) error {
	// Ensure required fields are present and validate each value.
	for _, field := range apiFields {
		val, exists := input[field.APIName]
		if !exists {
			if field.Required {
				return fmt.Errorf("field %q is required", field.APIName)
			}
			continue
		}
		// If the field is nested, validate recursively.
		if field.Nested != nil {
			switch v := val.(type) {
			case map[string]any:
				if err := h.validateMap(v, field.Nested); err != nil {
					return fmt.Errorf("field %q: %w", field.APIName, err)
				}
			case []any:
				for _, item := range v {
					err := h.validateMap(item.(map[string]any), field.Nested)
					if err != nil {
						return fmt.Errorf("field %q: %w", field.APIName, err)
					}
				}
			default:
				return fmt.Errorf(
					"field %q is not an object or an array", field.APIName,
				)
			}
		}
		// For non-object fields, run the validation function.
		if field.Validate != nil {
			validate, err := h.getValidator().FromRules(field.Validate)
			if err != nil {
				return fmt.Errorf(
					"validation rule error for field %q: %w",
					field.APIName,
					err,
				)
			}
			if err := validate(val); err != nil {
				return fmt.Errorf(
					"validation error for field %q: %w", field.APIName, err,
				)
			}
		}
	}
	return nil
}

// testValidationRules tests that the validation rules are valid.
func (h *MapInputHandler) testValidationRules(
	apiFields types.APIFields,
) error {
	for _, field := range apiFields {
		if field.Nested != nil {
			if err := h.testValidationRules(field.Nested); err != nil {
				return fmt.Errorf("field %q: %w", field.APIName, err)
			}
		}
		if field.Validate != nil {
			_, err := h.getValidator().FromRules(field.Validate)
			if err != nil {
				return fmt.Errorf(
					"validation rule error for field %q: %w",
					field.APIName,
					err,
				)
			}
		}
	}
	return nil
}
