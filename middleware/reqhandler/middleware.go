package reqhandler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/pakkasys/fluidapi-extended/util"
	"github.com/pakkasys/fluidapi/core"
)

const (
	headerXForwardedFor = "X-Forwarded-For"
)

var (
	// These keys are used to store values in the request context.
	responseDataKey = util.NewDataKey()
	requestDataKey  = util.NewDataKey()
	requestIDKey    = util.NewDataKey()
)

type requestLog struct {
	StartTime     time.Time `json:"start_time"`     // Time when the request started.
	RemoteAddress string    `json:"remote_address"` // Client IP address.
	Protocol      string    `json:"protocol"`       // HTTP protocol (e.g., HTTP/1.1).
	HTTPMethod    string    `json:"http_method"`    // HTTP method used.
	URL           string    `json:"url"`            // Request URL.
}

type requestMetadata struct {
	TimeStart     time.Time // Request start time.
	TraceID       string    // Unique identifier for the request.
	RemoteAddress string    // Client IP address.
	Protocol      string    // HTTP protocol.
	HTTPMethod    string    // HTTP method.
	URL           string    // Request URL.
}

// Middleware creates a request handler middleware that provides the following:
//   - Injects a new context into the request.
//   - Wraps the response writer and request for inspection.
//   - Attaches a unique trace ID and additional metadata to the context.
//   - Recovers from panics and logs detailed request/response data along with a
//     stack trace.
//   - Logs the start and completion of the request.
//
// Parameters:
//   - maxRequestBodySize: Maximum size of the request body in bytes.
//   - maxDumpPartSize: Maximum size of each panic dump part in bytes.
//   - traceIDFn: Function to generate a unique trace ID for the request.
//   - panicHandlerLoggerFn: Function that returns a logger for panic details.
//   - requestLoggerFn: Function that returns a logger for request logs.
//
// Returns:
//   - core.Middleware: The configured request handler middleware.
func Middleware(
	maxRequestBodySize int64,
	maxPanicDumpPartSize int64,
	traceIDFn func(r *http.Request) string,
	panicHandlerLoggerFn func(r *http.Request) func(messages ...any),
	requestLoggerFn func(r *http.Request) func(messages ...any),
) core.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Attach a new context.
			ctx := util.NewContext(r.Context())
			r = r.WithContext(ctx)

			// Panic recovery.
			defer func() {
				if rec := recover(); rec != nil {
					handlePanic(
						w, r, rec, panicHandlerLoggerFn, maxPanicDumpPartSize,
					)
				}
			}()

			// Wrap response and request.
			rw := NewResWrap(w)
			reqWrapper, err := NewReqWrap(r, maxRequestBodySize)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError),
					http.StatusInternalServerError)
				return
			}
			setResponseWrapper(r, rw)
			setRequestWrapper(r, reqWrapper)

			// Create and attach request metadata.
			reqMeta := &requestMetadata{
				TimeStart:     time.Now().UTC(),
				TraceID:       traceIDFn(r),
				RemoteAddress: requestIPAddress(r),
				Protocol:      r.Proto,
				HTTPMethod:    r.Method,
				URL:           fmt.Sprintf("%s%s", r.Host, r.URL.Path),
			}
			util.SetContextValue(r.Context(), requestIDKey, reqMeta)

			// Log request start.
			logRequestStart(r, reqMeta, requestLoggerFn)

			// Execute the next handler.
			next.ServeHTTP(rw, reqWrapper.Request)

			// Log request completion.
			requestLoggerFn(r)("Request completed")
		})
	}
}

// GetRequestMetadata retrieves the request metadata from the context.
//
// Parameters:
//   - ctx: The request context.
//
// Returns:
//   - *requestMetadata: The request metadata, or nil if not found.
func GetRequestMetadata(ctx context.Context) *requestMetadata {
	return util.GetContextValue[*requestMetadata](ctx, requestIDKey, nil)
}

// setResponseWrapper saves the response wrapper in the request context.
func setResponseWrapper(r *http.Request, rw *ResWrap) {
	util.SetContextValue(r.Context(), responseDataKey, rw)
}

// setRequestWrapper saves the request wrapper in the request context.
func setRequestWrapper(r *http.Request, rw *ReqWrap) {
	util.SetContextValue(r.Context(), requestDataKey, rw)
}

// getResponseWrapper retrieves the response wrapper from the request context.
func getResponseWrapper(r *http.Request) *ResWrap {
	return util.GetContextValue[*ResWrap](
		r.Context(), responseDataKey, nil,
	)
}

// logRequestStart logs the beginning of the request.
func logRequestStart(r *http.Request, meta *requestMetadata,
	requestLoggerFn func(r *http.Request) func(messages ...any)) {
	if meta == nil {
		requestLoggerFn(r)("Request started", "Request metadata not found")
		return
	}
	requestLoggerFn(r)(
		"Request started",
		requestLog{
			StartTime:     time.Now().UTC(),
			RemoteAddress: meta.RemoteAddress,
			Protocol:      meta.Protocol,
			HTTPMethod:    meta.HTTPMethod,
			URL:           meta.URL,
		},
	)
}
