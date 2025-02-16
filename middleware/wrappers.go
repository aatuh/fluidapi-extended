package middleware

import (
	"net/http"

	"github.com/pakkasys/fluidapi-extended/middleware/cors"
	"github.com/pakkasys/fluidapi-extended/middleware/inputlogic"
	"github.com/pakkasys/fluidapi-extended/middleware/reqhandler"
	"github.com/pakkasys/fluidapi/endpoint"
)

const (
	CORSMiddlewareID           = "cors"
	InputLogicMiddlewareID     = "inputlogic"
	RequestHandlerMiddlewareID = "request_handler"
)

// CORSWrapper creates a new wrapper with the CORSMiddleware.
//
//   - allowedOrigins: The list of allowed origins
//   - allowedMethods: The list of allowed methods
//   - allowedHeaders: The list of allowed headers
func CORSWrapper(
	allowedOrigins []string,
	allowedMethods []string,
	allowedHeaders []string,
) *endpoint.Wrapper {
	return &endpoint.Wrapper{
		ID: CORSMiddlewareID,
		Middleware: cors.Middleware(
			allowedOrigins,
			allowedMethods,
			allowedHeaders,
		),
	}
}

// InputLogicMiddlewareWrapper wraps the callback and returns a new
// MiddlewareWrapper for the InputLogic middleware.
//
//   - callback: The function that handles the request.
func InputLogicMiddlewareWrapper(
	callback inputlogic.Callback,
) *endpoint.Wrapper {
	return &endpoint.Wrapper{
		ID:         InputLogicMiddlewareID,
		Middleware: inputlogic.Middleware(callback),
	}
}

// RequestHandlerMiddlewareWrapper creates a new MiddlewareWrapper for the
// Request Handler middleware.
//
//   - traceIDFn: A function that generates a unique trace ID.
//   - panicHandlerLoggerFn: A function that logs panic details.
//   - requestLoggerFn: A function that logs request start/completion.
func RequestHandlerMiddlewareWrapper(
	traceIDFn func(r *http.Request) string,
	panicHandlerLoggerFn func(r *http.Request) func(messages ...any),
	requestLoggerFn func(r *http.Request) func(messages ...any),
) *endpoint.Wrapper {
	return &endpoint.Wrapper{
		ID: RequestHandlerMiddlewareID,
		Middleware: reqhandler.Middleware(
			traceIDFn,
			panicHandlerLoggerFn,
			requestLoggerFn,
		),
	}
}
