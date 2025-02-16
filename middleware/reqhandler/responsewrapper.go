package reqhandler

import (
	"net/http"
)

// ResponseWrapper wraps an http.ResponseWriter, capturing response data written
// to it.
type ResponseWrapper struct {
	// Embedded ResponseWriter for response writing capabilities
	http.ResponseWriter
	// Captured headers to be written
	Headers http.Header
	// Captured status code to be written
	StatusCode int
	// Captured response body
	Body []byte
	// Indicates if headers have been written
	headerWritten bool
}

// NewResponseWrapper creates a new instance of ResponseWrapper.
// It is used to wrap the ResponseWriter to capture and inspect the response.
//
// Parameters:
// - w: The original ResponseWriter to wrap.
//
// Returns:
// - A pointer to the newly created ResponseWrapper instance.
func NewResponseWrapper(w http.ResponseWriter) *ResponseWrapper {
	return &ResponseWrapper{
		ResponseWriter: w,
		Headers:        make(http.Header),
		StatusCode:     http.StatusOK, // Default status code
		headerWritten:  false,
	}
}

// Header overrides the Header method of the http.ResponseWriter interface.
// It returns the captured headers without modifying the underlying
// ResponseWriter's headers.
//
// Returns:
// - The captured http.Header that can be modified before writing.
func (rw *ResponseWrapper) Header() http.Header {
	return rw.Headers
}

// WriteHeader captures the status code to be written, delaying its execution.
// It ensures that headers are only written once and applies the captured
// headers to the underlying ResponseWriter.
//
// Parameters:
// - statusCode: The HTTP status code to write.
func (rw *ResponseWrapper) WriteHeader(statusCode int) {
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
// - data: The response body data to write.
//
// Returns:
// - The number of bytes written.
// - An error if writing to the underlying ResponseWriter fails.
func (rw *ResponseWrapper) Write(data []byte) (int, error) {
	// Ensure headers and status code are written before writing the body.
	if !rw.headerWritten {
		rw.WriteHeader(rw.StatusCode)
	}

	// Append the data to the response body buffer.
	rw.Body = append(rw.Body, data...)

	// Write the data to the underlying ResponseWriter.
	return rw.ResponseWriter.Write(data)
}
