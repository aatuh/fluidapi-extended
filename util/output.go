package util

import (
	"encoding/json"
	"net/http"

	"github.com/pakkasys/fluidapi/core"
)

const (
	contentType     = "Content-Type"
	applicationJSON = "application/json"
)

var Error = core.NewAPIError("ERROR")

type APIOutput[T any] struct {
	Payload *T             `json:"payload,omitempty"`
	Error   *core.APIError `json:"error,omitempty"`
}

// TODO: Replace logger funcs with events +global
type ILogger interface {
	Trace(messages ...any)
	Error(messages ...any)
}

type LoggerFn func(r *http.Request) ILogger

type Output struct {
	LoggerFn LoggerFn
}

func (p Output) Create(
	w http.ResponseWriter,
	r *http.Request,
	out any,
	outError error,
	status int,
) error {
	output, err := jsonOutput(w, out, outError, status)
	if err != nil {
		if p.LoggerFn != nil {
			p.LoggerFn(r).Error("Error handling output JSON: %s", err)
		}
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}
	if p.LoggerFn != nil {
		p.LoggerFn(r).Trace("Client output", output)
	}
	return nil
}

func jsonOutput(
	w http.ResponseWriter,
	outputData any,
	outputError error,
	statusCode int,
) (*APIOutput[any], error) {
	output := APIOutput[any]{
		Payload: &outputData,
		Error:   handleError(outputError),
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

func handleError(outputError error) *core.APIError {
	if outputError == nil {
		return nil
	}

	switch errType := outputError.(type) {
	case *core.APIError:
		return errType
	default:
		return Error
	}
}
