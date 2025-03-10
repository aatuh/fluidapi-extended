package api

import (
	"net/http"

	"github.com/pakkasys/fluidapi/core"
)

// InternalServerError represents an internal server error.
var InternalServerError = core.NewAPIError("INTERNAL_SERVER_ERROR")

// ErrorHandler handles errors and maps them to appropriate HTTP responses.
type ErrorHandler struct {
	expectedErrs []ExpectedError // Expected errors to handle.
}

// NewErrorHandler creates a new ErrorHandler.
//
// Parameters:
//   - expectedErrs: A slice of ExpectedError objects that define how to handle
//     specific errors.
//
// Returns:
//   - *ErrorHandler: A new ErrorHandler.
func NewErrorHandler(expectedErrs []ExpectedError) *ErrorHandler {
	return &ErrorHandler{
		expectedErrs: expectedErrs,
	}
}

// Handle processes an error and returns the corresponding HTTP status code and
// API error. It checks if the error is an *apierror.Error and handles it
// accordingly
//
// Parameters:
//   - err: The error to handle.
//
// Returns:
//   - int: The HTTP status code.
//   - *core.APIError: The mapped API error.
func (e ErrorHandler) Handle(err error) (int, *core.APIError) {
	apiError, ok := err.(*core.APIError)
	if !ok {
		return http.StatusInternalServerError, InternalServerError
	}
	return e.handleAPIError(apiError)
}

// handleAPIError maps an API error to an HTTP status code and API error.
func (e *ErrorHandler) handleAPIError(
	apiError *core.APIError,
) (int, *core.APIError) {
	expectedError := e.getExpectedError(apiError)
	if expectedError == nil {
		return http.StatusInternalServerError, InternalServerError
	}
	return expectedError.maskAPIError(apiError)
}

// getExpectedError finds the ExpectedError that matches the given API error.
// It returns nil if no match is found.
func (e *ErrorHandler) getExpectedError(
	apiError *core.APIError,
) *ExpectedError {
	for i := range e.expectedErrs {
		if apiError.ID == e.expectedErrs[i].ID &&
			apiError.Origin == e.expectedErrs[i].Origin {

			return &e.expectedErrs[i]
		}
	}
	return nil
}

// ExpectedError represents an expected error configuration.
// It defines how to handle specific errors that are anticipated.
type ExpectedError struct {
	ID         string // The ID of the expected error.
	MaskedID   string // An optional ID to mask the original error ID in the response.
	Status     int    // The HTTP status code to return for this error.
	PublicData bool   // Whether to include the error data in the response.
	Origin     string // The orign of the error.
}

// NewExpectedError creates a new ExpectedError.
//
// Parameters:
//   - id: The ID of the expected error.
//   - status: The HTTP status code to return for this error.
//   - origin: The origin of the error.
//
// Returns:
//   - ExpectedError: The new ExpectedError.
func NewExpectedError(id string, status int, origin string) ExpectedError {
	return ExpectedError{
		ID:     id,
		Status: status,
		Origin: origin,
	}
}

// WithMaskedID returns a new ExpectedError with the given masked ID.
//
// Parameters:
//   - maskedID: The masked ID to use in the response.
//
// Returns:
//   - ExpectedError: The new ExpectedError.
func (e ExpectedError) WithMaskedID(maskedID string) ExpectedError {
	e.MaskedID = maskedID
	return e
}

// WithPublicData returns a new ExpectedError with the public data flag set.
//
// Returns:
//   - ExpectedError: The new ExpectedError.
func (e ExpectedError) WithPublicData(isPublic bool) ExpectedError {
	e.PublicData = isPublic
	return e
}

// maskAPIError masks the ID and data of the given API error based on the
// configuration of the ExpectedError.
func (e *ExpectedError) maskAPIError(
	apiError *core.APIError,
) (int, *core.APIError) {
	// If a masked ID is defined, use it. Otherwise, use the original ID.
	var useErrorID string
	if e.MaskedID != "" {
		useErrorID = e.MaskedID
	} else {
		useErrorID = e.ID
	}

	// If the error data is public, use it. Otherwise, use nil.
	var useData any
	if e.PublicData {
		useData = apiError.Data
	} else {
		useData = nil
	}

	return e.Status, core.NewAPIError(useErrorID).WithData(useData)
}

// ExpectedErrors is a slice of ExpectedError.
type ExpectedErrors []ExpectedError

// With returns a new slice with the errors appended to the slice.
//
// Parameters:
//   - errs: The errors to append.
//
// Returns:
//   - ExpectedErrors: The new slice with the errors appended.
func (e ExpectedErrors) With(errs ...ExpectedError) ExpectedErrors {
	newSlice := append([]ExpectedError{}, e...)
	return append(newSlice, errs...)
}

// WithOrigin makes all errors in the slice have the given origin and returns
// a new slice with the origin set for all errors.
//
// Parameters:
//   - origin: The origin to set for all errors.
//
// Returns:
//   - ExpectedErrors: The new slice with the origin set for all errors.
func (e ExpectedErrors) WithOrigin(origin string) ExpectedErrors {
	newSlice := append([]ExpectedError{}, e...)
	for i := range newSlice {
		newSlice[i].Origin = origin
	}
	return newSlice
}

// GetByID returns the ExpectedError with the given ID, or nil if not found.
//
// Parameters:
//   - id: The ID of the expected error.
//
// Returns:
//   - *ExpectedError: The ExpectedError with the given ID, or nil if not found.
func (e ExpectedErrors) GetByID(id string) *ExpectedError {
	for i := range e {
		if e[i].ID == id {
			return &e[i]
		}
	}
	return nil
}
