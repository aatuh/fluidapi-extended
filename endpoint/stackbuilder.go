package endpoint

import (
	"fmt"
	"net/http"

	"github.com/pakkasys/fluidapi/core/api"
	"github.com/pakkasys/fluidapi/endpoint/middleware"
	"github.com/pakkasys/fluidapi/endpoint/runner"
)

type StackBuilder struct {
	middlewareStack middleware.Stack
}

func NewStackBuilder(
	requestIDFn func() string,
	panicHandlerLoggerFn func(r *http.Request) func(messages ...any),
	requestLoggerFn func(r *http.Request) func(messages ...any),
) runner.StackBuilder {
	return &StackBuilder{
		[]api.MiddlewareWrapper{
			*middleware.ContextMiddlewareWrapper(),
			*middleware.ResponseWrapperMiddlewareWrapper(),
			*middleware.RequestIDMiddlewareWrapper(requestIDFn),
			*middleware.PanicHandlerMiddlewareWrapper(panicHandlerLoggerFn),
			*middleware.RequestLogMiddlewareWrapper(requestLoggerFn),
		},
	}
}

func (b *StackBuilder) Build() middleware.Stack {
	return b.middlewareStack
}

func (b *StackBuilder) MustAddMiddleware(
	wrapper ...api.MiddlewareWrapper,
) runner.StackBuilder {
	for i := range wrapper {
		success := b.middlewareStack.InsertAfterID(
			middleware.RequestLogMiddlewareID,
			wrapper[i],
		)
		if !success {
			panic(fmt.Sprintf("Failed to add middleware: %s", wrapper[i].ID))
		}
	}
	return b
}
