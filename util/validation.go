package util

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

const (
	errorFmt = "Validation failed on rule %q"
)

// FieldError represents a field-level validation error
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrorData contains a list of field-level validation errors
type ValidationErrorData struct {
	Errors []FieldError `json:"errors"`
}

// Validation is a validator using the validator/v10 package.
type Validation struct {
	validate *validator.Validate
}

// NewValidation creates a new ValidatorService instance.
func NewValidation() *Validation {
	return &Validation{
		validate: validator.New(),
	}
}

// Validate validates an object and returns a slice of FieldError if fails.
func (vs *Validation) Validate(obj any) []FieldError {
	err := vs.validate.Struct(obj)
	if err != nil {
		return vs.parseValidationErrors(err)
	}
	return nil
}

// parseValidationErrors parses validation errors into a slice of FieldError.
func (vs *Validation) parseValidationErrors(err error) []FieldError {
	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return []FieldError{
			{
				Field:   "unknown",
				Message: err.Error(),
			},
		}
	}

	var fieldErrors []FieldError
	for _, valErr := range validationErrors {
		fieldErrors = append(fieldErrors, FieldError{
			Field:   valErr.Field(),
			Message: fmt.Sprintf(errorFmt, valErr.Tag()),
		})
	}

	return fieldErrors
}
