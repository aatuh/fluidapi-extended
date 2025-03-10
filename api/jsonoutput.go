package api

import (
	"encoding/json"
	"net/http"

	"github.com/pakkasys/fluidapi/core"
)

// Constants for HTTP headers and content types.
const (
	contentType     = "Content-Type"
	applicationJSON = "application/json"
)

// Error represents a generic error.
var Error = core.NewAPIError("ERROR")

// APIOutput represents the output of a client request.
type APIOutput[T any] struct {
	Payload *T             `json:"payload,omitempty"`
	Error   *core.APIError `json:"error,omitempty"`
}

// TODO: Replace logger funcs with events +global
// ILogger represents a logger interface.
type ILogger interface {
	Trace(message ...any)
	Error(message ...any)
}

// LoggerFn represents a function that returns a logger.
type LoggerFn func(r *http.Request) ILogger

// JSONOutput represents the output of a client request.
type JSONOutput struct {
	loggerFn    LoggerFn
	errorOrigin string
}

// NewJSONOutput returns a new JSONOutput.
//
// Parameters:
//   - loggerFn: The logger function.
//   - errorOrigin: The origin of the error.
//
// Returns:
//   - JSONOutput: The new JSONOutput.
func NewJSONOutput(loggerFn LoggerFn, errorOrigin string) JSONOutput {
	return JSONOutput{
		loggerFn:    loggerFn,
		errorOrigin: errorOrigin,
	}
}

// Create marshals the output to JSON and writes it to the response.
// If the logger is not nil, it will log the output.
//
// Parameters:
//   - w: The response writer.
//   - r: The request.
//   - out: The output data.
//   - outError: The output error.
//   - status: The HTTP status code.
//
// Returns:
//   - error: An error if the request fails.
func (o JSONOutput) Create(
	w http.ResponseWriter, r *http.Request, out any, outError error, status int,
) error {
	output, err := o.jsonOutput(w, out, outError, status)
	if err != nil {
		if o.loggerFn != nil {
			o.loggerFn(r).Error("Error handling output JSON", err)
		}
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}
	if o.loggerFn != nil {
		o.loggerFn(r).Trace("Client output", output)
	}
	return nil
}

// jsonOutput marshals the output to JSON and writes it to the response.
func (o JSONOutput) jsonOutput(
	w http.ResponseWriter, outputData any, outputError error, statusCode int,
) (*APIOutput[any], error) {
	output := APIOutput[any]{
		Payload: &outputData,
		Error:   o.handleError(outputError),
	}

	jsonData, err := json.Marshal(output)
	if err != nil {
		return nil, err
	}

	w.Header().Set(contentType, applicationJSON)
	w.WriteHeader(statusCode)

	_, err = w.Write(jsonData)
	if err != nil {
		return nil, err
	}

	return &output, nil
}

// handleError returns the API error if it's a *core.APIError or a generic error
// if it's not. Returns nil if the error is nil.
func (o JSONOutput) handleError(outputError error) *core.APIError {
	if outputError == nil {
		return nil
	}
	switch errType := outputError.(type) {
	case *core.APIError:
		newErr := *errType
		newErr.Origin = o.errorOrigin
		return &newErr
	default:
		return Error
	}
}
