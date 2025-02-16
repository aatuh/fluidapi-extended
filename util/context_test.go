package util

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewDataKey tests the NewDataKey function.
func TestNewDataKey(t *testing.T) {
	key1 := NewDataKey()
	key2 := NewDataKey()

	assert.NotEqual(t, key1, key2, "NewDataKey should generate unique keys")
}

// TestConcurrencyOnNewDataKey tests the concurrency safety of NewDataKey.
func TestConcurrencyOnNewDataKey(t *testing.T) {
	// Ensuring concurrency safety for NewDataKey
	var wg sync.WaitGroup
	keys := make(map[DataKey]bool)
	lock := sync.Mutex{}

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			key := NewDataKey()
			lock.Lock()
			keys[key] = true
			lock.Unlock()
		}()
	}

	wg.Wait()
	assert.Equal(t, 1000, len(keys), "All keys should be unique and have 1000 entries")
}

// TestNewContext tests the NewContext function.
func TestNewContext(t *testing.T) {
	baseCtx := context.Background()
	customCtx := NewContext(baseCtx)

	assert.True(t, IsContextSet(customCtx), "NewContext should set the custom context data")
}

// TestIsContextSet tests the IsContextSet function.
func TestIsContextSet(t *testing.T) {
	baseCtx := context.Background()
	assert.False(t, IsContextSet(baseCtx), "Base context should not have custom context set")

	customCtx := NewContext(baseCtx)
	assert.True(t, IsContextSet(customCtx), "Context should have custom context set")
}

// TestHasContextValue tests the HasContextValue function.
func TestHasContextValue(t *testing.T) {
	baseCtx := context.Background()
	customCtx := NewContext(baseCtx)

	key := NewDataKey()
	assert.False(t, HasContextValue(customCtx, key), "No value should be set initially")

	SetContextValue(customCtx, key, "test_value")
	assert.True(t, HasContextValue(customCtx, key), "HasContextValue should return true after setting a value")
}

// TestHasContextValue_NoCustomContext tests the HasContextValue function when
// no custom context is set.
func TestHasContextValue_NoCustomContext(t *testing.T) {
	baseCtx := context.Background()
	key := NewDataKey()
	assert.False(t, HasContextValue(baseCtx, key), "HasContextValue should return false if no custom context is set")
}

// TestGetContextValue tests the GetContextValue function.
func TestGetContextValue(t *testing.T) {
	baseCtx := context.Background()
	customCtx := NewContext(baseCtx)

	key := NewDataKey()
	defaultValue := "default"

	result := GetContextValue(customCtx, key, defaultValue)
	assert.Equal(t, defaultValue, result, "GetContextValue should return default value if key is not set")

	SetContextValue(customCtx, key, "actual_value")
	result = GetContextValue(customCtx, key, defaultValue)
	assert.Equal(t, "actual_value", result, "GetContextValue should return the correct value if key is set")
}

// TestGetContextValue_TypeMismatch tests the GetContextValue function with a
// type mismatch.
func TestGetContextValue_TypeMismatch(t *testing.T) {
	customCtx := NewContext(context.Background())

	key := NewDataKey()
	SetContextValue(customCtx, key, "test_value")

	// Attempting to retrieve value as a different type should return the
	// provided default value
	result := GetContextValue[int](customCtx, key, -1)
	assert.Equal(t, -1, result, "Type mismatch should return the provided default value")
}

// TestGetContextValue_NoCustomContext tests the GetContextValue function when
// no custom context is set.
func TestGetContextValue_NoCustomContext(t *testing.T) {
	baseCtx := context.Background()
	key := NewDataKey()
	defaultValue := "default"

	result := GetContextValue(baseCtx, key, defaultValue)
	assert.Equal(t, defaultValue, result, "GetContextValue should return default value if no custom context is set")
}

// TestMustGetContextValue tests the MustGetContextValue function.
func TestMustGetContextValue(t *testing.T) {
	baseCtx := context.Background()
	customCtx := NewContext(baseCtx)

	key := NewDataKey()

	assert.Panics(t, func() {
		MustGetContextValue[string](customCtx, key)
	}, "MustGetContextValue should panic if key is not set")

	SetContextValue(customCtx, key, "actual_value")
	result := MustGetContextValue[string](customCtx, key)
	assert.Equal(t, "actual_value", result, "MustGetContextValue should return the correct value if key is set")
}

// TestMustGetContextValue_TypeMismatch tests the MustGetContextValue function
// with a type mismatch.
func TestMustGetContextValue_TypeMismatch(t *testing.T) {
	customCtx := NewContext(context.Background())

	key := NewDataKey()
	SetContextValue(customCtx, key, "test_value")

	// MustGetContextValue should panic if the type doesn't match
	assert.Panics(t, func() {
		MustGetContextValue[int](customCtx, key)
	}, "MustGetContextValue should panic if type mismatch occurs")
}

// TestMustGetContextValue_WithNilValue tests the MustGetContextValue function
// with a nil value.
func TestMustGetContextValue_WithNilValue(t *testing.T) {
	customCtx := NewContext(context.Background())

	key := NewDataKey()
	SetContextValue(customCtx, key, nil)

	// Should panic if we try to MustGetContextValue with a nil set value
	assert.Panics(t, func() {
		MustGetContextValue[string](customCtx, key)
	}, "MustGetContextValue should panic if key exists but value is nil")
}

// TestMustGetContextValue_NoCustomContext tests the MustGetContextValue
// function when no custom context is set.
func TestMustGetContextValue_NoCustomContext(t *testing.T) {
	baseCtx := context.Background()
	key := NewDataKey()

	assert.Panics(t, func() {
		MustGetContextValue[string](baseCtx, key)
	}, "MustGetContextValue should panic if no custom context is set")
}

// TestSetContextValue tests the SetContextValue function.
func TestSetContextValue(t *testing.T) {
	baseCtx := context.Background()
	customCtx := NewContext(baseCtx)

	key := NewDataKey()
	SetContextValue(customCtx, key, "test_value")

	result := GetContextValue(customCtx, key, "")
	assert.Equal(t, "test_value", result, "SetContextValue should correctly set a value in the context")
}

// TestSetContextValue_PanicIfNoCustomContext tests the SetContextValue
// function.
func TestSetContextValue_PanicIfNoCustomContext(t *testing.T) {
	baseCtx := context.Background()

	key := NewDataKey()
	assert.Panics(t, func() {
		SetContextValue(baseCtx, key, "value")
	}, "SetContextValue should panic if the custom context is not set")
}

// TestSetContextValue_ConcurrentWrites tests the SetContextValue function
// with concurrent writes.
func TestSetContextValue_ConcurrentWrites(t *testing.T) {
	customCtx := NewContext(context.Background())

	key := NewDataKey()
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			SetContextValue(customCtx, key, val)
		}(i)
	}

	wg.Wait()

	// The final value is not deterministic due to concurrent writes,
	// but it should always be a valid integer between 0 and 99.
	result := MustGetContextValue[int](customCtx, key)
	assert.GreaterOrEqual(t, result, 0, "Value should be within valid range after concurrent writes")
	assert.LessOrEqual(t, result, 99, "Value should be within valid range after concurrent writes")
}

// TestSetContextValue_WithNilKey tests the SetContextValue function with a
// nil key.
func TestSetContextValue_WithNilKey(t *testing.T) {
	customCtx := NewContext(context.Background())

	// Testing setting a value with a nil key
	assert.Panics(t, func() {
		SetContextValue(customCtx, nil, "test_value")
	}, "SetContextValue should panic if provided key is nil")
}

// TestClearContextValue tests the ClearContextValue function.
func TestClearContextValue(t *testing.T) {
	baseCtx := context.Background()
	customCtx := NewContext(baseCtx)

	key := NewDataKey()
	SetContextValue(customCtx, key, "test_value")
	assert.True(t, HasContextValue(customCtx, key), "Value should be present before clearing")

	ClearContextValue(customCtx, key)
	assert.False(t, HasContextValue(customCtx, key), "Value should be removed after clearing")
}

// TestClearContextValue_NonExistentKey tests the ClearContextValue function
// with a non-existent key.
func TestClearContextValue_NonExistentKey(t *testing.T) {
	customCtx := NewContext(context.Background())

	// Clearing a value for a key that was never set should not panic or cause an error
	assert.NotPanics(t, func() {
		ClearContextValue(customCtx, NewDataKey())
	}, "ClearContextValue should not panic for a non-existent key")
}

// TestGetContextData tests the getContextData function.
func TestGetContextData(t *testing.T) {
	baseCtx := context.Background()
	cd, ok := getContextData(baseCtx)
	assert.Nil(t, cd, "getContextData should return nil for base context without custom data")
	assert.False(t, ok, "getContextData should return false for base context without custom data")

	customCtx := NewContext(baseCtx)
	cd, ok = getContextData(customCtx)
	assert.NotNil(t, cd, "getContextData should return non-nil value for custom context")
	assert.True(t, ok, "getContextData should return true for custom context")
}
