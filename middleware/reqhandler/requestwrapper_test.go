package reqhandler

import (
	"bytes"
	"errors"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// errorReader is an implementation of io.ReadCloser that returns an error when
// read.
type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("read error")
}

func (e *errorReader) Close() error {
	return nil
}

// TestNewRequestWrapper tests the NewRequestWrapper function.
func TestNewRequestWrapper(t *testing.T) {
	bodyContent := []byte("test body content")
	req := httptest.NewRequest(
		"POST",
		"http://example.com",
		bytes.NewBuffer(bodyContent),
	)

	wrappedReq, err := NewRequestWrapper(req)

	assert.Nil(t, err, "NewRequestWrapper should not return an error")
	assert.NotNil(t, wrappedReq, "NewRequestWrapper should return a non-nil value")
	assert.Equal(t, bodyContent, wrappedReq.GetBodyContent(), "Captured body content should match original request body")

	readBody, err := io.ReadAll(wrappedReq.Body)
	assert.Nil(t, err, "Reading body from wrapped request should not return an error")
	assert.Equal(t, bodyContent, readBody, "Body content should match the original request body")
}

// TestNewRequestWrapper_ReadBodyError tests NewRequestWrapper function when
// reading the body returns an error.
func TestNewRequestWrapper_ReadBodyError(t *testing.T) {
	req := httptest.NewRequest("POST", "http://example.com", &errorReader{})

	wrappedReq, err := NewRequestWrapper(req)

	assert.Error(t, err, "Expected an error when body read fails")
	assert.Nil(t, wrappedReq, "RequestWrapper should be nil when an error occurs")
}

// TestGetBodyContent tests the GetBodyContent function.
func TestGetBodyContent(t *testing.T) {
	bodyContent := []byte("sample content")
	req := httptest.NewRequest(
		"POST",
		"http://example.com",
		bytes.NewBuffer(bodyContent),
	)

	wrappedReq, err := NewRequestWrapper(req)
	assert.Nil(t, err)

	assert.Equal(t, bodyContent, wrappedReq.GetBodyContent(), "GetBodyContent should return the correct body content")
}

// TestGetBody tests the GetBody function.
func TestGetBody(t *testing.T) {
	bodyContent := []byte("test content for GetBody")
	req := httptest.NewRequest(
		"POST",
		"http://example.com",
		bytes.NewBuffer(bodyContent),
	)

	wrappedReq, err := NewRequestWrapper(req)
	assert.Nil(t, err)

	// Get a function to read the body again
	getBodyFunc, err := wrappedReq.GetBody()
	assert.Nil(t, err, "GetBody should not return an error")

	// Use the returned function to get a ReadCloser and read the body
	bodyReader, err := getBodyFunc()
	assert.Nil(t, err, "GetBody function should return a valid ReadCloser")

	readContent, err := io.ReadAll(bodyReader)
	assert.Nil(t, err, "Reading body from GetBody function should not return an error")
	assert.Equal(t, bodyContent, readContent, "Body content should match the original request body")
}

// TestReadBodyMultipleTimes tests the ability to read the request body
// multiple times.
func TestReadBodyMultipleTimes(t *testing.T) {
	bodyContent := []byte("content for multiple reads")
	req := httptest.NewRequest(
		"POST",
		"http://example.com",
		bytes.NewBuffer(bodyContent),
	)

	wrappedReq, err := NewRequestWrapper(req)
	assert.Nil(t, err)

	// Read body the first time
	readBody, err := io.ReadAll(wrappedReq.Body)
	assert.Nil(t, err)
	assert.Equal(t, bodyContent, readBody, "First read of body content should match original content")

	// Read body again
	wrappedReq.Body = io.NopCloser(bytes.NewBuffer(wrappedReq.BodyContent))
	readBody, err = io.ReadAll(wrappedReq.Body)
	assert.Nil(t, err)
	assert.Equal(t, bodyContent, readBody, "Second read of body content should match original content")
}
