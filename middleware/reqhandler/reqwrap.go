package reqhandler

import (
	"bytes"
	"io"
	"net/http"
)

// ReqWrap wraps an http.Request, capturing its body for multiple reads and
// inspection.
type ReqWrap struct {
	*http.Request        // Embedded request.
	BodyContent   []byte // Captured request body.
}

// NewReqWrap creates a new ReqWrap instance and captures the request body,
// enforcing a maximum size to prevent excessive memory usage. If the request
// body is larger than the maximum size, an error is returned.
//
// Parameters:
//   - r: The original http.Request.
//
// Returns:
//   - *ReqWrap: The wrapped request.
//   - error: Any error encountered during body reading.
func NewReqWrap(r *http.Request, maxRequestBodySize int64) (*ReqWrap, error) {
	bodyBytes, err := io.ReadAll(io.LimitReader(r.Body, maxRequestBodySize))
	if err != nil {
		return nil, err
	}
	// Replace the body so it can be read again.
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	return &ReqWrap{
		Request:     r,
		BodyContent: bodyBytes,
	}, nil
}
