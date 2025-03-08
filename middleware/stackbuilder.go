package middleware

import (
	"fmt"
	"net/http"

	"github.com/pakkasys/fluidapi/endpoint"
)

// StackBuilder builds a middleware stack.
type StackBuilder struct {
	stack *endpoint.Stack
}

// NewStackBuilder returns a new instance.
func NewStackBuilder(
	requestIDFn func(r *http.Request) string,
	panicHandlerLoggerFn func(r *http.Request) func(messages ...any),
	requestLoggerFn func(r *http.Request) func(messages ...any),
) *StackBuilder {
	return &StackBuilder{
		stack: endpoint.NewStack(
			*RequestHandlerMiddlewareWrapper(
				requestIDFn,
				panicHandlerLoggerFn,
				requestLoggerFn,
			),
		),
	}
}

// Build returns the middleware stack.
func (b *StackBuilder) Build() *endpoint.Stack {
	return b.stack
}

// MustAddMiddleware adds middleware to the stack and panics if it fails.
func (b *StackBuilder) MustAddMiddleware(
	wrapper ...endpoint.Wrapper,
) *StackBuilder {
	for i := range wrapper {
		stack, success := b.stack.InsertAfter(
			RequestHandlerMiddlewareID,
			wrapper[i],
		)
		if !success {
			panic(fmt.Sprintf("Failed to add middleware: %s", wrapper[i].ID))
		}
		b.stack = stack
	}
	return b
}
