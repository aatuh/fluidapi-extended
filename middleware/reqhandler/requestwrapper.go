package reqhandler

import (
	"bytes"
	"io"
	"net/http"
)

// RequestWrapper wraps an http.Request, capturing its information for
// inspection and modification.
type RequestWrapper struct {
	*http.Request        // Embedded *http.Request for full request capabilities
	BodyContent   []byte // Captured body content of the request
}

// NewRequestWrapper creates a new instance of RequestWrapper and captures the
// body content to allow multiple reads.
// This is particularly useful for scenarios where the request body needs
// to be processed and logged or validated multiple times.
//
// Parameters:
// - r: The original http.Request to wrap.
//
// Returns:
// - A pointer to the newly created RequestWrapper instance.
// - An error if reading the request body fails.
func NewRequestWrapper(r *http.Request) (*RequestWrapper, error) {
	// Capture the body content so that it can be read multiple times.
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	// Reassign a new ReadCloser with the body data so it can be read again.
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	return &RequestWrapper{
		Request:     r,
		BodyContent: bodyBytes,
	}, nil
}

// GetBodyContent provides a way to read the captured body content.
// This allows direct access to the body bytes without consuming the original
// body.
//
// Returns:
// - The captured body content as a slice of bytes.
func (rw *RequestWrapper) GetBodyContent() []byte {
	return rw.BodyContent
}

// GetBody provides a method to get the body as a reader that can be used
// multiple times.
// This is particularly useful for forwarding or logging requests where
// the body needs to be read more than once.
//
// Returns:
// - A function that returns an io.ReadCloser that reads the captured body.
// - An error if the creation of the reader fails.
func (rw *RequestWrapper) GetBody() (func() (io.ReadCloser, error), error) {
	return func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewBuffer(rw.BodyContent)), nil
	}, nil
}
