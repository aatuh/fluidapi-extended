package middleware

import (
	"fmt"
	"net/http"

	"github.com/pakkasys/fluidapi/endpoint"
)

type StackBuilder struct {
	stack endpoint.Stack
}

func NewStackBuilder(
	requestIDFn func(r *http.Request) string,
	panicHandlerLoggerFn func(r *http.Request) func(messages ...any),
	requestLoggerFn func(r *http.Request) func(messages ...any),
) *StackBuilder {
	return &StackBuilder{
		stack: endpoint.Stack{
			*RequestHandlerMiddlewareWrapper(
				requestIDFn,
				panicHandlerLoggerFn,
				requestLoggerFn,
			),
		},
	}
}

func (b *StackBuilder) Build() endpoint.Stack {
	return b.stack
}

func (b *StackBuilder) MustAddMiddleware(
	wrapper ...endpoint.Wrapper,
) StackBuilder {
	for i := range wrapper {
		success := b.stack.InsertAfterID(
			RequestHandlerMiddlewareID,
			wrapper[i],
		)
		if !success {
			panic(fmt.Sprintf("Failed to add middleware: %s", wrapper[i].ID))
		}
	}
	return *b
}
