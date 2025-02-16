package inputlogic

// TODO: implement

// import (
// 	"errors"
// 	"net/http"
// 	"testing"

// 	apierror "github.com/pakkasys/fluidapi/core/api/error"
// 	"github.com/pakkasys/fluidapi/extra/endpoint/middleware/inputlogic/types"
// 	"github.com/stretchr/testify/assert"
// )

// func TestErrorHandler_Handle_NonAPIError(t *testing.T) {
// 	handler := &ErrorHandler{}

// 	// Simulate a generic error (not an *apierror.Error)
// 	err := errors.New("some error")

// 	statusCode, apiErr := handler.Handle(err, nil)

// 	assert.Equal(t, http.StatusInternalServerError, statusCode)
// 	assert.Equal(t, InternalServerError, apiErr)
// }

// func TestErrorHandler_Handle_UnexpectedAPIError(t *testing.T) {
// 	handler := &ErrorHandler{}

// 	// Simulate an *apierror.Error that is not in expectedErrors
// 	err := core.NewAPIError("UNEXPECTED_ERROR")

// 	statusCode, apiErr := handler.Handle(err, nil)

// 	assert.Equal(t, http.StatusInternalServerError, statusCode)
// 	assert.Equal(t, InternalServerError, apiErr)
// }

// func TestErrorHandler_Handle_ExpectedAPIError(t *testing.T) {
// 	handler := &ErrorHandler{}

// 	// Simulate an *apierror.Error that is expected
// 	err := core.NewAPIError("EXPECTED_ERROR")
// 	data := "Some error data"
// 	err = err.WithData(data)

// 	expectedErrors := []inputlogic.ExpectedError{
// 		{
// 			ID:         "EXPECTED_ERROR",
// 			Status:     http.StatusBadRequest,
// 			PublicData: true,
// 		},
// 	}

// 	statusCode, apiErr := handler.Handle(err, expectedErrors)

// 	assert.Equal(t, http.StatusBadRequest, statusCode)
// 	assert.Equal(t, "EXPECTED_ERROR", apiErr.ID)
// 	assert.Equal(t, "Some error data", apiErr.Data)
// }

// func TestErrorHandler_Handle_ExpectedAPIError_MaskedID(t *testing.T) {
// 	handler := &ErrorHandler{}

// 	// Simulate an *apierror.Error that is expected
// 	err := core.NewAPIError("EXPECTED_ERROR")
// 	data := "Sensitive data"
// 	err = err.WithData(data)

// 	maskedErrorID := "MASKED_ERROR"
// 	expectedErrors := []inputlogic.ExpectedError{
// 		{
// 			ID:         "EXPECTED_ERROR",
// 			MaskedID:   &maskedErrorID,
// 			Status:     http.StatusForbidden,
// 			PublicData: false,
// 		},
// 	}

// 	statusCode, apiErr := handler.Handle(err, expectedErrors)

// 	assert.Equal(t, http.StatusForbidden, statusCode)
// 	assert.Equal(t, "MASKED_ERROR", apiErr.ID)
// 	assert.Nil(t, apiErr.Data)
// }

// func TestErrorHandler_HandleAPIError_NoExpectedError(t *testing.T) {
// 	handler := &ErrorHandler{}

// 	err := core.NewAPIError("UNEXPECTED_ERROR")

// 	statusCode, apiErr := handler.handleAPIError(err, nil)

// 	assert.Equal(t, http.StatusInternalServerError, statusCode)
// 	assert.Equal(t, InternalServerError, apiErr)
// }

// func TestErrorHandler_GetExpectedError_Found(t *testing.T) {
// 	handler := &ErrorHandler{}

// 	err := core.NewAPIError("ERROR_ID")

// 	expectedErrors := []inputlogic.ExpectedError{
// 		{ID: "ERROR_ID"},
// 	}

// 	expectedError := handler.getExpectedError(err, expectedErrors)

// 	assert.NotNil(t, expectedError)
// 	assert.Equal(t, "ERROR_ID", expectedError.ID)
// }

// func TestErrorHandler_GetExpectedError_NotFound(t *testing.T) {
// 	handler := &ErrorHandler{}

// 	err := core.NewAPIError("ERROR_ID")

// 	expectedErrors := []inputlogic.ExpectedError{
// 		{ID: "OTHER_ERROR_ID"},
// 	}

// 	expectedError := handler.getExpectedError(err, expectedErrors)

// 	assert.Nil(t, expectedError)
// }

// func TestExpectedError_MaskAPIError_DataVisible(t *testing.T) {
// 	err := core.NewAPIError("ERROR_ID")
// 	data := "Error details"
// 	err = err.WithData(data)

// 	expectedError := &inputlogic.ExpectedError{
// 		ID:         "ERROR_ID",
// 		Status:     http.StatusBadRequest,
// 		PublicData: true,
// 	}

// 	statusCode, apiErr := expectedError.MaskAPIError(err)

// 	assert.Equal(t, http.StatusBadRequest, statusCode)
// 	assert.Equal(t, "ERROR_ID", apiErr.ID)
// 	assert.Equal(t, "Error details", apiErr.Data)
// }

// func TestExpectedError_MaskAPIError_DataNotVisible(t *testing.T) {
// 	err := core.NewAPIError("ERROR_ID")
// 	data := "Sensitive data"
// 	err = err.WithData(data)

// 	maskedID := "MASKED_ID"
// 	expectedError := &inputlogic.ExpectedError{
// 		ID:         "ERROR_ID",
// 		MaskedID:   &maskedID,
// 		Status:     http.StatusUnauthorized,
// 		PublicData: false,
// 	}

// 	statusCode, apiErr := expectedError.MaskAPIError(err)

// 	assert.Equal(t, http.StatusUnauthorized, statusCode)
// 	assert.Equal(t, "MASKED_ID", apiErr.ID)
// 	assert.Nil(t, apiErr.Data)
// }
