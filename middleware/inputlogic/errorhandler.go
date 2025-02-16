package inputlogic

import (
	"net/http"

	"github.com/pakkasys/fluidapi/core"
)

var InternalServerError = core.NewAPIError("INTERNAL_SERVER_ERROR")

// ErrorHandler handles errors and maps them to appropriate HTTP responses.
type ErrorHandler struct{}

// Handle processes an error and returns the corresponding HTTP status code and
// API error. It checks if the error is an *apierror.Error and handles it
// accordingly.
func (e ErrorHandler) Handle(
	err error,
	expectedErrs []ExpectedError,
) (int, *core.APIError) {
	apiError, ok := err.(*core.APIError)
	if !ok {
		return http.StatusInternalServerError, InternalServerError
	}
	return e.handleAPIError(apiError, expectedErrs)
}

func (e *ErrorHandler) handleAPIError(
	apiError *core.APIError,
	expectedErrs []ExpectedError,
) (int, *core.APIError) {
	expectedError := e.getExpectedError(apiError, expectedErrs)
	if expectedError == nil {
		return http.StatusInternalServerError, InternalServerError
	}
	return expectedError.MaskAPIError(apiError)
}

func (e *ErrorHandler) getExpectedError(
	apiError *core.APIError,
	expectedErrs []ExpectedError,
) *ExpectedError {
	for i := range expectedErrs {
		if apiError.ID() == expectedErrs[i].ID {
			return &expectedErrs[i]
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
}

func (e *ExpectedError) MaskAPIError(
	apiError *core.APIError,
) (int, *core.APIError) {
	var useErrorID string
	if e.MaskedID != "" {
		useErrorID = e.MaskedID
	} else {
		useErrorID = e.ID
	}

	var useData any
	if e.PublicData {
		useData = apiError.Data()
	} else {
		useData = nil
	}

	return e.Status, core.NewAPIError(useErrorID).WithData(useData)
}
