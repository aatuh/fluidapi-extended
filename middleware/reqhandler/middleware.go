package reqhandler

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"time"

	"github.com/pakkasys/fluidapi-extended/util"
	"github.com/pakkasys/fluidapi/core"
)

const maxDumpPartSize = 1024 * 1024 // 1MB

var (
	// These keys are used to store values in the request context.
	responseDataKey = util.NewDataKey()
	requestDataKey  = util.NewDataKey()
	requestIDKey    = util.NewDataKey()
)

type requestLog struct {
	StartTime     time.Time `json:"start_time"`     // Start time of the request.
	RemoteAddress string    `json:"remote_address"` // Remote IP address of the client making the request.
	Protocol      string    `json:"protocol"`       // Protocol used in the request (e.g., HTTP/1.1).
	HTTPMethod    string    `json:"http_method"`    // HTTP method used for the request.
	URL           string    `json:"url"`            // Full URL of the request.
}

type requestMetadata struct {
	TimeStart     time.Time // Time when the request started.
	TraceID       string    // Unique identifier for the request.
	RemoteAddress string    // Remote IP address of the request.
	Protocol      string    // Protocol used in the request (e.g., HTTP/1.1).
	HTTPMethod    string    // HTTP method used for the request (e.g., GET).
	URL           string    // URL of the request.
}

// Middleware has these functionalities:
//   - Context: injects a new context.
//   - Response wrapper: wraps the response writer and request.
//   - Trace ID: attaches request metadata (including a generated ID).
//   - Panic handler: recovers from panics and logs details.
//   - Request log: logs the start and completion of the request.
//
// The provided functions are used as follows:
//   - requestIDFn: generates a unique trace ID.
//   - panicHandlerLoggerFn: logs panic details.
//   - requestLoggerFn: logs request start/completion.
func Middleware(
	traceIDFn func(r *http.Request) string,
	panicHandlerLoggerFn func(r *http.Request) func(messages ...any),
	requestLoggerFn func(r *http.Request) func(messages ...any),
) core.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Context middleware: attach a new context.
			ctx := util.NewContext(r.Context())
			r = r.WithContext(ctx)

			// Panic handler.
			defer func() {
				if rec := recover(); rec != nil {
					handlePanic(w, r, rec, panicHandlerLoggerFn)
				}
			}()

			// Response wrapper middleware: wrap response and request.
			rw := NewResponseWrapper(w)
			reqWrapper, err := NewRequestWrapper(r)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError),
					http.StatusInternalServerError)
				return
			}
			setResponseWrapper(r, rw)
			setRequestWrapper(r, reqWrapper)

			// Attach request metadata.
			reqMeta := &requestMetadata{
				TimeStart:     time.Now().UTC(),
				TraceID:       traceIDFn(r),
				RemoteAddress: RequestIPAddress(r),
				Protocol:      r.Proto,
				HTTPMethod:    r.Method,
				URL:           fmt.Sprintf("%s%s", r.Host, r.URL.Path),
			}
			util.SetContextValue(r.Context(), requestIDKey, reqMeta)

			// Request log middleware: log that the request has started.
			logRequestStart(r, reqMeta, requestLoggerFn)

			// Call the next handler with the wrapped response and request.
			next.ServeHTTP(rw, reqWrapper.Request)

			// After the handler returns, log that the request has completed.
			requestLoggerFn(r)("Request completed")
		})
	}
}

// GetRequestMetadata retrieves the request metadata from the request context.
func GetRequestMetadata(ctx context.Context) *requestMetadata {
	return util.GetContextValue[*requestMetadata](
		ctx,
		requestIDKey,
		nil,
	)
}

// setResponseWrapper saves the response wrapper in the request context.
func setResponseWrapper(r *http.Request, rw *ResponseWrapper) {
	util.SetContextValue(r.Context(), responseDataKey, rw)
}

// setRequestWrapper saves the request wrapper in the request context.
func setRequestWrapper(r *http.Request, rw *RequestWrapper) {
	util.SetContextValue(r.Context(), requestDataKey, rw)
}

// getResponseWrapper retrieves the response wrapper from the request context.
func getResponseWrapper(r *http.Request) *ResponseWrapper {
	return util.GetContextValue[*ResponseWrapper](
		r.Context(),
		responseDataKey,
		nil,
	)
}

// logRequestStart logs the beginning of the request.
func logRequestStart(r *http.Request, meta *requestMetadata,
	requestLoggerFn func(r *http.Request) func(messages ...any)) {
	if meta == nil {
		requestLoggerFn(r)("Request started", "Request metadata not found")
		return
	}
	entry := requestLog{
		StartTime:     time.Now().UTC(),
		RemoteAddress: meta.RemoteAddress,
		Protocol:      meta.Protocol,
		HTTPMethod:    meta.HTTPMethod,
		URL:           meta.URL,
	}
	requestLoggerFn(r)("Request started", entry)
}

// responseData is a simplified copy of the response details.
type responseData struct {
	StatusCode int
	Headers    map[string][]string
	Body       string
}

// requestDumpData holds a dump of request/response info for panic logging.
type requestDumpData struct {
	StatusCode int `json:"status_code"`
	Request    struct {
		URL     string              `json:"url"`
		Params  string              `json:"params"`
		Headers map[string][]string `json:"headers"`
		Body    string              `json:"body"`
	} `json:"request"`
	Response struct {
		Headers map[string][]string `json:"headers"`
		Body    string              `json:"body"`
	} `json:"response"`
}

// panicData holds the data that will be logged when a panic occurs.
type panicData struct {
	Err         any             `json:"err"`
	RequestDump requestDumpData `json:"request_dump"`
	StackTrace  []string        `json:"stack_trace"`
}

// handlePanic recovers from a panic, logs the details, and sends a 500.
func handlePanic(
	w http.ResponseWriter,
	r *http.Request,
	err any,
	panicHandlerLoggerFn func(r *http.Request) func(messages ...any),
) {
	var rd responseData
	rw := getResponseWrapper(r)
	if rw != nil {
		rd = responseData{
			StatusCode: rw.StatusCode,
			Headers:    limitHeaders(rw.Header(), maxDumpPartSize),
			Body:       string(rw.Body),
		}
	}

	pd := panicData{
		Err:         fmt.Sprintf("%v", err),
		RequestDump: *createRequestDumpData(rd, r),
		StackTrace:  stackTraceSlice(),
	}
	panicHandlerLoggerFn(r)("Panic", pd)
	http.Error(
		w,
		http.StatusText(http.StatusInternalServerError),
		http.StatusInternalServerError,
	)
}

// stackTraceSlice returns a slice of strings representing the call stack.
func stackTraceSlice() []string {
	var stackTrace []string
	for skip := 0; ; skip++ {
		pc, file, line, ok := runtime.Caller(skip)
		if !ok {
			break
		}
		fn := runtime.FuncForPC(pc)
		entry := fmt.Sprintf("%s:%d %s", file, line, fn.Name())
		stackTrace = append(stackTrace, entry)
	}
	return stackTrace
}

// createRequestDumpData constructs a dump of the request and response.
func createRequestDumpData(rd responseData, r *http.Request) *requestDumpData {
	reqBody, err := readBodyWithLimit(r.Body, maxDumpPartSize)
	if err != nil {
		reqBody = "Error reading request body"
	}
	dump := &requestDumpData{StatusCode: rd.StatusCode}
	dump.Request.URL = r.URL.String()
	dump.Request.Params = limitQueryParameters(r.URL.RawQuery, maxDumpPartSize)
	dump.Request.Headers = limitHeaders(r.Header, maxDumpPartSize)
	dump.Request.Body = reqBody
	dump.Response.Headers = limitHeaders(rd.Headers, maxDumpPartSize)
	dump.Response.Body = rd.Body
	return dump
}

// readBodyWithLimit reads up to maxSize bytes from the body.
func readBodyWithLimit(body io.ReadCloser, maxSize int64) (string, error) {
	if body == nil {
		return "", nil
	}
	defer body.Close()
	limitedReader := io.LimitReader(body, maxSize)
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(limitedReader)
	if err != nil {
		return "", err
	}
	if buf.Len() == int(maxSize) {
		return buf.String() + "... (truncated)", nil
	}
	return buf.String(), nil
}

// limitHeaders truncates header values longer than maxSize.
func limitHeaders(
	headers map[string][]string,
	maxSize int,
) map[string][]string {
	limited := make(map[string][]string)
	for key, values := range headers {
		var limitedVals []string
		for _, val := range values {
			if len(val) > maxSize {
				limitedVals = append(
					limitedVals,
					val[:maxSize]+"... (truncated)",
				)
			} else {
				limitedVals = append(limitedVals, val)
			}
		}
		limited[key] = limitedVals
	}
	return limited
}

// limitQueryParameters truncates query parameters if they exceed maxSize.
func limitQueryParameters(params string, maxSize int) string {
	if len(params) > maxSize {
		return params[:maxSize] + "... (truncated)"
	}
	return params
}
