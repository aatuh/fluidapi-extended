package util

// FieldError represents a field-level validation error
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrorData contains a list of field-level validation errors
type ValidationErrorData struct {
	Errors []FieldError `json:"errors"`
}
