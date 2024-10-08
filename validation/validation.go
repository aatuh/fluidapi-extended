package validation

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/pakkasys/fluidapi/endpoint/middleware/inputlogic"
)

const (
	errorFmt = "Validation failed on rule %q"
)

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
func (vs *Validation) Validate(obj any) []inputlogic.FieldError {
	err := vs.validate.Struct(obj)
	if err != nil {
		return vs.parseValidationErrors(err)
	}
	return nil
}

// parseValidationErrors parses validation errors into a slice of FieldError.
func (vs *Validation) parseValidationErrors(err error) []inputlogic.FieldError {
	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return []inputlogic.FieldError{
			{
				Field:   "unknown",
				Message: err.Error(),
			},
		}
	}

	var fieldErrors []inputlogic.FieldError
	for _, valErr := range validationErrors {
		fieldErrors = append(fieldErrors, inputlogic.FieldError{
			Field:   valErr.Field(),
			Message: fmt.Sprintf(errorFmt, valErr.Tag()),
		})
	}

	return fieldErrors
}
