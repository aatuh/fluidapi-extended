package inputlogic

import (
	"net/http"

	"github.com/pakkasys/fluidapi-extended/util"
	"github.com/stretchr/testify/mock"
)

// MockObjectPicker is a mock implementation of the ObjectPicker interface.
type MockObjectPicker[T any] struct {
	mock.Mock
}

func (m *MockObjectPicker[T]) PickObject(
	r *http.Request,
	w http.ResponseWriter,
	obj T,
) (*T, error) {
	args := m.Called(r, w, obj)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*T), args.Error(1)
}

// MockOutputHandler is a mock implementation of the OutputHandler interface.
type MockOutputHandler struct {
	mock.Mock
}

func (m *MockOutputHandler) ProcessOutput(
	w http.ResponseWriter,
	r *http.Request,
	out any,
	outError error,
	statusCode int,
) error {
	args := m.Called(w, r, out, outError, statusCode)
	return args.Error(0)
}

// MockLogger is a mock implementation of the ILogger interface.
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Trace(messages ...any) {
	m.Called(messages)
}

func (m *MockLogger) Error(messages ...any) {
	m.Called(messages)
}

// TODO: Get working without helper
// MockHelper helps to mock a value receiver object.
type MockHelper struct {
	mock.Mock
}

// MockValidatedInput is a mock implementation of the ValidatedInput interface.
type MockValidatedInput struct {
	helper *MockHelper
}

func (m MockValidatedInput) Validate() []util.FieldError {
	if m.helper == nil {
		panic("MockHelper is not initialized.")
	}
	args := m.helper.Called()
	return args.Get(0).([]util.FieldError)
}

// Factory function to create MockValidatedInput with initialized helper
func NewMockValidatedInput(helper *MockHelper) MockValidatedInput {
	return MockValidatedInput{helper: helper}
}
