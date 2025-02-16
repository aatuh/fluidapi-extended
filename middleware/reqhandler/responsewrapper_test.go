package reqhandler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewResponseWrapper tests the NewResponseWrapper function.
func TestNewResponseWrapper(t *testing.T) {
	// Wrap the ResponseRecorder in a ResponseWrapper.
	recorder := httptest.NewRecorder()
	wrapped := NewResponseWrapper(recorder)

	assert.NotNil(t, wrapped, "NewResponseWrapper should not return nil")
	assert.Equal(t, http.StatusOK, wrapped.StatusCode, "Default status code should be 200 (OK)")
	assert.NotNil(t, wrapped.Headers, "Headers map should be initialized")
	assert.Empty(t, wrapped.Body, "Body should be empty initially")
}

// TestResponseWrapper_Header tests the Header method of ResponseWrapper.
func TestResponseWrapper_Header(t *testing.T) {
	recorder := httptest.NewRecorder()
	wrapped := NewResponseWrapper(recorder)

	wrapped.Header().Set("Content-Type", "application/json")

	assert.Equal(t, "application/json", wrapped.Header().Get("Content-Type"), "Header should be correctly set")
	assert.Equal(t, "", recorder.Header().Get("Content-Type"), "Original ResponseWriter should not be modified before WriteHeader is called")
}

// TestResponseWrapper_WriteHeader tests the WriteHeader method of
// ResponseWrapper.
func TestResponseWrapper_WriteHeader(t *testing.T) {
	recorder := httptest.NewRecorder()
	wrapped := NewResponseWrapper(recorder)

	wrapped.Header().Set("X-Test-Header", "TestValue")
	wrapped.WriteHeader(http.StatusCreated)

	assert.Equal(t, http.StatusCreated, wrapped.StatusCode, "Status code should be captured correctly")
	assert.Equal(t, "TestValue", recorder.Header().Get("X-Test-Header"), "Captured headers should be applied to the underlying ResponseWriter")
	assert.Equal(t, http.StatusCreated, recorder.Code, "Underlying ResponseWriter should receive the correct status code")
}

// TestResponseWrapper_Write tests the Write method of ResponseWrapper.
func TestResponseWrapper_Write(t *testing.T) {
	recorder := httptest.NewRecorder()
	wrapped := NewResponseWrapper(recorder)

	data := []byte("Hello, world!")
	n, err := wrapped.Write(data)

	assert.Nil(t, err, "Write should not return an error")
	assert.Equal(t, len(data), n, "Number of bytes written should match input")
	assert.Equal(t, data, wrapped.Body, "Captured body should match written data")
	assert.Equal(t, data, recorder.Body.Bytes(), "Underlying ResponseWriter should receive the correct body content")
}

// TestResponseWrapper_WriteHeader_OnlyOnce tests that WriteHeader only writes
// once.
func TestResponseWrapper_WriteHeader_OnlyOnce(t *testing.T) {
	recorder := httptest.NewRecorder()
	wrapped := NewResponseWrapper(recorder)

	// Call WriteHeader multiple times
	wrapped.WriteHeader(http.StatusCreated)
	wrapped.WriteHeader(http.StatusInternalServerError)

	assert.Equal(t, http.StatusCreated, wrapped.StatusCode, "Status code should not change after the first WriteHeader call")
	assert.Equal(t, http.StatusCreated, recorder.Code, "Underlying ResponseWriter should only receive the first status code")
}

// TestResponseWrapper_Write_AfterWriteHeader tests the Write method after
// WriteHeader has been called.
func TestResponseWrapper_Write_AfterWriteHeader(t *testing.T) {
	recorder := httptest.NewRecorder()
	wrapped := NewResponseWrapper(recorder)

	// WriteHeader explicitly before Write
	wrapped.WriteHeader(http.StatusAccepted)

	data := []byte("Hello again!")
	n, err := wrapped.Write(data)

	assert.Nil(t, err, "Write should not return an error")
	assert.Equal(t, len(data), n, "Number of bytes written should match input")
	assert.Equal(t, data, wrapped.Body, "Captured body should match written data")
	assert.Equal(t, data, recorder.Body.Bytes(), "Underlying ResponseWriter should receive the correct body content")
	assert.Equal(t, http.StatusAccepted, recorder.Code, "Status code should be set before writing the body")
}

// TestResponseWrapper_DefaultWriteHeader tests that Write automatically sets
// the default WriteHeader.
func TestResponseWrapper_DefaultWriteHeader(t *testing.T) {
	recorder := httptest.NewRecorder()
	wrapped := NewResponseWrapper(recorder)

	data := []byte("Auto header write")
	_, err := wrapped.Write(data)

	assert.Nil(t, err, "Write should not return an error")
	assert.Equal(t, http.StatusOK, wrapped.StatusCode, "Default status code should be 200 (OK) if WriteHeader is not explicitly called")
	assert.Equal(t, http.StatusOK, recorder.Code, "Underlying ResponseWriter should receive the default status code of 200 (OK)")
}

// TestResponseWrapper_HeaderWriteOnce tests that headers are written only once.
func TestResponseWrapper_HeaderWriteOnce(t *testing.T) {
	recorder := httptest.NewRecorder()
	wrapped := NewResponseWrapper(recorder)

	wrapped.Header().Set("X-Test-Header", "InitialValue")
	wrapped.WriteHeader(http.StatusOK)

	// Attempt to modify headers after WriteHeader
	wrapped.Header().Set("X-Test-Header", "ModifiedValue")

	assert.Equal(t, "InitialValue", recorder.Header().Get("X-Test-Header"), "Headers should not be modified after WriteHeader is called")
}
