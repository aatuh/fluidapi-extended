package endpoint

import (
	"fmt"
	"net/http"

	"github.com/mitchellh/mapstructure"
	"github.com/pakkasys/fluidapi-extended/api"
	"github.com/pakkasys/fluidapi/core"
	"github.com/pakkasys/fluidapi/endpoint"
)

var (
	MapToObjectDecodingError = core.NewAPIError("ERROR_DECODING_MAP_TO_OBJECT")
)

type LoggerFactoryFn func(r *http.Request) api.ILogger

// EndpointHandler represents an endpoint with customizable middleware.
type EndpointHandler[Input any] struct {
	url          string
	method       string
	inputHandler InputHandler
	inputFactory func() Input
	handlerLogic func(
		w http.ResponseWriter, r *http.Request, i *Input,
	) (any, error)
	expectedErrors  []api.ExpectedError
	loggerFactoryFn func(r *http.Request) api.ILogger
	systemId        string
}

// InputHandler defines how to process the request input.
type InputHandler interface {
	Handle(w http.ResponseWriter, r *http.Request) (map[string]any, error)
}

// NewEndpointHandler creates a new EndpointHandler with optional customizations.
func NewEndpointHandler[Input any](
	url string,
	method string,
	inputHandler InputHandler,
	inputFactory func() Input,
	handlerLogic func(
		w http.ResponseWriter, r *http.Request, i *Input,
	) (any, error),
	expectedErrors []api.ExpectedError,
	loggerFactoryFn LoggerFactoryFn,
	systemId string,
) *EndpointHandler[Input] {
	return &EndpointHandler[Input]{
		url:          url,
		method:       method,
		inputHandler: inputHandler,
		inputFactory: inputFactory,
		handlerLogic: func(
			w http.ResponseWriter, r *http.Request, i *Input,
		) (any, error) {
			return handlerLogic(w, r, i)
		},
		expectedErrors:  expectedErrors,
		loggerFactoryFn: loggerFactoryFn,
		systemId:        systemId,
	}
}

// Handle executes common endpoints logic: input decoding, business logic, and
// output.
func (h *EndpointHandler[Input]) Handle(
	w http.ResponseWriter,
	r *http.Request,
) {
	dataMap, err := h.inputHandler.Handle(w, r)
	if err != nil {
		h.handleError(w, r, err, h.expectedErrors, h.systemId)
		return
	}

	blankInput := h.inputFactory()
	input, err := mapToObject(dataMap, &blankInput)
	if err != nil {
		h.handleError(w, r, err, h.expectedErrors, h.systemId)
		return
	}

	out, err := h.handlerLogic(w, r, input)
	if err != nil {
		h.handleError(w, r, err, h.expectedErrors, h.systemId)
		return
	}

	h.handleOutput(w, r, out, nil, http.StatusOK, h.systemId)
}

// Build constructs the endpoint definition using the middleware stack.
func (h *EndpointHandler[Input]) Build(
	stack *endpoint.Stack,
) *endpoint.Definition {
	return &endpoint.Definition{
		URL:    h.url,
		Method: h.method,
		Stack:  stack,
	}
}

// mapToObject decodes a map into the provided object.
func mapToObject[T any](value map[string]any, obj *T) (*T, error) {
	cfg := &mapstructure.DecoderConfig{
		Metadata:         nil,
		Result:           obj,
		TagName:          "json",
		WeaklyTypedInput: true,
	}
	decoder, err := mapstructure.NewDecoder(cfg)
	if err != nil {
		return nil, fmt.Errorf("mapToObject: error creating decoder: %v", err)
	}
	if err := decoder.Decode(value); err != nil {
		return nil, MapToObjectDecodingError.WithMessage(err.Error())
	}
	return obj, nil
}

// handleError maps errors and writes the error response.
func (h *EndpointHandler[Input]) handleError(
	w http.ResponseWriter,
	r *http.Request,
	err error,
	expectedErrs []api.ExpectedError,
	systemId string,
) {
	statusCode, outError := api.NewErrorHandler(expectedErrs).Handle(err)
	logMsg := fmt.Sprintf(
		"Error, status: %d, err: %s, out: %s", statusCode, err, outError,
	)
	if statusCode >= http.StatusInternalServerError {
		h.loggerFactoryFn(r).Error(logMsg)
	} else {
		h.loggerFactoryFn(r).Trace(logMsg)
	}
	h.handleOutput(w, r, nil, outError, statusCode, systemId)
}

// handleOutput processes and writes the endpoint response.
func (h *EndpointHandler[Input]) handleOutput(
	w http.ResponseWriter,
	r *http.Request,
	out any,
	outputError error,
	statusCode int,
	systemId string,
) {
	output := api.NewJSONOutput(h.loggerFactoryFn, systemId)
	if err := output.Create(w, r, out, outputError, statusCode); err != nil {
		if h.loggerFactoryFn != nil {
			h.loggerFactoryFn(r).Error("Output error: %s", err)
		}
		w.WriteHeader(http.StatusInternalServerError)
	}
}
