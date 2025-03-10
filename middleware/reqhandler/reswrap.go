package reqhandler

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
)

// ResWrap wraps an http.ResponseWriter to capture response data for inspection.
type ResWrap struct {
	http.ResponseWriter             // Embedded ResponseWriter.
	Headers             http.Header // Captured headers.
	StatusCode          int         // Captured status code.
	Body                []byte      // Captured response body.
	headerWritten       bool        // Indicates if headers have been written.
}

// NewResWrap creates a new ResWrap instance wrapping the given ResponseWriter.
//
// Parameters:
//   - w: The original http.ResponseWriter.
//
// Returns:
//   - *ResWrap: The wrapped response writer.
func NewResWrap(w http.ResponseWriter) *ResWrap {
	return &ResWrap{
		ResponseWriter: w,
		Headers:        make(http.Header),
		StatusCode:     http.StatusOK,
		headerWritten:  false,
	}
}

// Header overrides the Header method of the http.ResponseWriter interface.
// It returns the captured headers without modifying the underlying
// ResponseWriter's headers.
//
// Returns:
//   - The captured http.Header that can be modified before writing.
func (rw *ResWrap) Header() http.Header {
	return rw.Headers
}

// WriteHeader captures the status code to be written, delaying its execution.
// It ensures that headers are only written once and applies the captured
// headers to the underlying ResponseWriter.
//
// Parameters:
//   - statusCode: The HTTP status code to write.
func (rw *ResWrap) WriteHeader(statusCode int) {
	if !rw.headerWritten { // Only write headers once
		rw.StatusCode = statusCode
		// Apply the captured headers to the underlying ResponseWriter.
		for key, values := range rw.Headers {
			for _, value := range values {
				rw.ResponseWriter.Header().Add(key, value)
			}
		}
		rw.ResponseWriter.WriteHeader(statusCode)
		rw.headerWritten = true
	}
}

// Write writes the response body and ensures headers and status code are
// written.
// It captures the response body and writes the data to the underlying
// ResponseWriter after ensuring that the headers and status code are written.
//
// Parameters:
//   - data: The response body data to write.
//
// Returns:
//   - int: The number of bytes written.
//   - error: An error if writing to the underlying ResponseWriter fails.
func (rw *ResWrap) Write(data []byte) (int, error) {
	// Ensure headers and status code are written before writing the body.
	if !rw.headerWritten {
		rw.WriteHeader(rw.StatusCode)
	}

	// Append the data to the response body buffer.
	rw.Body = append(rw.Body, data...)

	// Write the data to the underlying ResponseWriter.
	return rw.ResponseWriter.Write(data)
}

// Flush forwards the flush call to the underlying ResponseWriter if supported.
func (rw *ResWrap) Flush() {
	if flusher, ok := rw.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Hijack forwards the hijack call to the underlying ResponseWriter if
// supported.
//
// Returns:
//   - net.Conn: The hijacked connection.
//   - error: If the underlying ResponseWriter does not support hijacking.
func (rw *ResWrap) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := rw.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, fmt.Errorf(
		"Hijack: underlying ResponseWriter does not support hijacking",
	)
}
