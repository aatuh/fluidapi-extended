package reqhandler

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"runtime/debug"
	"strings"
)

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

// handlePanic recovers from a panic, logs details (including a stack trace),
// and sends an HTTP 500 response.
func handlePanic(
	w http.ResponseWriter,
	r *http.Request,
	err any,
	panicHandlerLoggerFn func(r *http.Request) func(messages ...any),
	maxDumpPartSize int64,
) {
	var rd responseData
	rw := getResponseWrapper(r)
	if rw != nil {
		rd = responseData{
			StatusCode: rw.StatusCode,
			Headers:    limitHeaders(rw.Header(), int(maxDumpPartSize)),
			Body:       string(rw.Body),
		}
	}

	stack := string(debug.Stack())
	pd := panicData{
		Err:         fmt.Sprintf("%v", err),
		RequestDump: *createRequestDumpData(rd, r, maxDumpPartSize),
		StackTrace:  strings.Split(stack, "\n"),
	}
	panicHandlerLoggerFn(r)("Panic", pd)
	http.Error(
		w,
		http.StatusText(http.StatusInternalServerError),
		http.StatusInternalServerError,
	)
}

// createRequestDumpData constructs a dump of the request and response.
func createRequestDumpData(
	rd responseData,
	r *http.Request,
	maxDumpPartSize int64,
) *requestDumpData {
	reqBody, err := readBodyWithLimit(r.Body, maxDumpPartSize)
	if err != nil {
		reqBody = "Error reading request body"
	}
	intMaxSize := int(maxDumpPartSize)
	dump := &requestDumpData{StatusCode: rd.StatusCode}
	dump.Request.URL = r.URL.String()
	dump.Request.Params = limitQueryParameters(r.URL.RawQuery, intMaxSize)
	dump.Request.Headers = limitHeaders(r.Header, intMaxSize)
	dump.Request.Body = reqBody
	dump.Response.Headers = limitHeaders(rd.Headers, intMaxSize)
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
	headers map[string][]string, maxSize int,
) map[string][]string {
	limited := make(map[string][]string)
	for key, values := range headers {
		var limitedVals []string
		for _, val := range values {
			if len(val) > maxSize {
				limitedVals = append(
					limitedVals, val[:maxSize]+"... (truncated)",
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
