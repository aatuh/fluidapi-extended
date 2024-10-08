package urlencoder

import (
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestMinSlice tests the methods of MinSlice.
func TestMinSlice(t *testing.T) {
	// Case: Initialize a new MinSlice
	slice := newMinSlice()
	assert.NotNil(t, slice, "expected a new MinSlice to be initialized")
	assert.Equal(
		t,
		0,
		len(slice.elements),
		"expected the initial MinSlice to have no elements",
	)

	// Case: Set and Get elements in MinSlice
	slice.set(0, "first")
	value, exists := slice.get(0)
	assert.True(t, exists, "expected element at index 0 to exist")
	assert.Equal(t, "first", value, "expected value at index 0 to be 'first'")

	slice.set(5, "fifth")
	value, exists = slice.get(5)
	assert.True(t, exists, "expected element at index 5 to exist")
	assert.Equal(t, "fifth", value, "expected value at index 5 to be 'fifth'")

	// Case: Get an element that does not exist
	value, exists = slice.get(10)
	assert.False(t, exists, "expected element at index 10 to not exist")
	assert.Nil(t, value, "expected value at index 10 to be nil")

	// Case: Update an existing element
	slice.set(5, "newFifth")
	value, exists = slice.get(5)
	assert.True(t, exists, "expected element at index 5 to exist after update")
	assert.Equal(
		t,
		"newFifth",
		value,
		"expected updated value at index 5 to be 'newFifth'",
	)

	// Case: Convert MinSlice to a regular slice
	slice.set(2, "second")
	regularSlice := slice.toSlice()
	expectedSlice := []any{"first", "second", "newFifth"}
	assert.ElementsMatch(
		t,
		expectedSlice,
		regularSlice,
		"expected MinSlice to convert to regular slice correctly",
	)

}

// TestDecodeURL tests the DecodeURL function.
func TestDecodeURL(t *testing.T) {
	// Case: Simple key-value pair
	values := url.Values{}
	values.Set("simpleKey", "simpleValue")

	result, err := DecodeURL(values)
	assert.Nil(t, err)

	expected := map[string]any{
		"simpleKey": "simpleValue",
	}
	assert.Equal(t, expected, result)

	// Case: Nested key-value pair
	values = url.Values{}
	values.Set("level1.level2.key", "nestedValue")

	result, err = DecodeURL(values)
	assert.Nil(t, err)

	expected = map[string]any{
		"level1": map[string]any{
			"level2": map[string]any{
				"key": "nestedValue",
			},
		},
	}
	assert.Equal(t, expected, result)

	// Case: Slice of structs
	values = url.Values{}
	values.Set("mySlice[0].key", "value")
	values.Set("mySlice[1].key", "value")

	result, err = DecodeURL(values)
	assert.Nil(t, err)

	expected = map[string]any{
		"mySlice": []any{
			map[string]any{"key": "value"},
			map[string]any{"key": "value"},
		},
	}
	assert.Equal(t, expected, result)

	// Case: Slice elements
	values = url.Values{}
	values.Set("mySlice[0]", "sliceValue1")
	values.Set("mySlice[1]", "sliceValue2")

	result, err = DecodeURL(values)
	assert.Nil(t, err)

	expected = map[string]any{"mySlice": []any{"sliceValue1", "sliceValue2"}}
	assert.ElementsMatch(t, expected["mySlice"], result["mySlice"])

	// Case: Overwrite existing values
	values = url.Values{}
	values.Set("level1.level2.key", "nestedValue")
	values.Set("level1.level2.key", "newValue")

	result, err = DecodeURL(values)
	assert.Nil(t, err)

	expected = map[string]any{
		"level1": map[string]any{
			"level2": map[string]any{
				"key": "newValue",
			},
		},
	}
	assert.Equal(t, expected, result)

	// Case: Slice with string index
	values = url.Values{}
	values.Set("invalidSlice[abc]", "invalidValue")

	result, err = DecodeURL(values)
	assert.Nil(t, err)
	expected = map[string]any{"invalidSlice[abc]": "invalidValue"}
	assert.Equal(t, expected, result)

	// Case: Complex nested and slice structure
	values = url.Values{}
	values.Set("complex.level1[0]", "value1")
	values.Set("complex.level1[1]", "value2")
	values.Set("complex.level2.key3", "value3")

	result, err = DecodeURL(values)
	assert.Nil(t, err)

	expected = map[string]any{
		"complex": map[string]any{
			"level1": []any{"value1", "value2"},
			"level2": map[string]any{"key3": "value3"},
		},
	}
	complexMap, ok := expected["complex"].(map[string]any)
	if !ok {
		t.Errorf("expected complex to be a map")
	}
	assert.ElementsMatch(
		t,
		complexMap["level1"],
		result["complex"].(map[string]any)["level1"],
	)

	// Case: Triggering error with type mismatch
	values = url.Values{}
	values.Set("myMap.key", "mapValue")    // Initializes myMap as a map
	values.Set("myMap[0]", "invalidValue") // Treat myMap as a slice, error

	result, err = DecodeURL(values)
	assert.NotNil(t, err)
	assert.Nil(t, result)
}

// TestEncodeValue tests the encodeValue function.
func TestEncodeValue(t *testing.T) {
	values := &url.Values{}

	// Case: Encoding a pointer
	str := "test pointer"
	v := reflect.ValueOf(&str)
	err := encodeValue(values, "pointerField", v)
	assert.Nil(t, err)
	assert.Equal(t, url.Values{"pointerField": {"test pointer"}}, *values)

	// Case: Encoding a string
	values = &url.Values{}
	v = reflect.ValueOf("test string")
	err = encodeValue(values, "stringField", v)
	assert.Nil(t, err)
	assert.Equal(t, url.Values{"stringField": {"test string"}}, *values)

	// Case: Encoding an integer
	values = &url.Values{}
	v = reflect.ValueOf(42)
	err = encodeValue(values, "intField", v)
	assert.Nil(t, err)
	assert.Equal(t, url.Values{"intField": {"42"}}, *values)

	// Case: Encoding a boolean
	values = &url.Values{}
	v = reflect.ValueOf(true)
	err = encodeValue(values, "boolField", v)
	assert.Nil(t, err)
	assert.Equal(t, url.Values{"boolField": {"true"}}, *values)

	// Case: Encoding a slice
	values = &url.Values{}
	v = reflect.ValueOf([]string{"apple", "banana"})
	err = encodeValue(values, "sliceField", v)
	assert.Nil(t, err)
	assert.Equal(
		t,
		url.Values{"sliceField[0]": {"apple"}, "sliceField[1]": {"banana"}},
		*values,
	)

	// Case: Encoding a struct
	values = &url.Values{}
	type ExampleStruct struct {
		Field string `json:"field"`
	}
	v = reflect.ValueOf(ExampleStruct{Field: "structField"})
	err = encodeValue(values, "structField", v)
	assert.Nil(t, err)
	assert.Equal(t, url.Values{"structField.field": {"structField"}}, *values)

	// Case: Unsupported type
	values = &url.Values{}
	v = reflect.ValueOf(map[string]string{"key": "value"}) // Unsupported type
	err = encodeValue(values, "unsupportedField", v)
	assert.NotNil(t, err)
	assert.EqualError(t, err, "value type not supported by URL encoding: map")
}

// TestEncodePointer tests the encodePointer function.
func TestEncodePointer(t *testing.T) {
	values := &url.Values{}
	strPtr := "pointer string"
	v := reflect.ValueOf(&strPtr)

	err := encodePointer(values, "pointerField", v)
	assert.Nil(t, err)
	assert.Equal(t, url.Values{"pointerField": {"pointer string"}}, *values)

	// Case: Nil pointer
	values = &url.Values{}
	var nilPtr *string
	v = reflect.ValueOf(nilPtr)

	err = encodePointer(values, "nilPointerField", v)
	assert.Nil(t, err)
	assert.Equal(t, url.Values{}, *values) // No value should be added
}

// TestEncodeString tests the encodeString function.
func TestEncodeString(t *testing.T) {
	values := &url.Values{}
	v := reflect.ValueOf("test string")
	err := encodeString(values, "stringField", v)
	assert.Nil(t, err)
	assert.Equal(t, url.Values{"stringField": {"test string"}}, *values)
}

// TestEncodeInt tests the encodeInt function.
func TestEncodeInt(t *testing.T) {
	values := &url.Values{}
	v := reflect.ValueOf(123)
	err := encodeInt(values, "intField", v)
	assert.Nil(t, err)
	assert.Equal(t, url.Values{"intField": {"123"}}, *values)
}

// TestEncodeBool tests the encodeBool function.
func TestEncodeBool(t *testing.T) {
	values := &url.Values{}
	v := reflect.ValueOf(true)
	err := encodeBool(values, "boolField", v)
	assert.Nil(t, err)
	assert.Equal(t, url.Values{"boolField": {"true"}}, *values)
}

// TestEncodeSlice tests the encodeSlice function.
func TestEncodeSlice(t *testing.T) {
	values := &url.Values{}
	v := reflect.ValueOf([]int{1, 2, 3})
	err := encodeSlice(values, "sliceField", v)
	assert.Nil(t, err)
	expected := url.Values{
		"sliceField[0]": {"1"},
		"sliceField[1]": {"2"},
		"sliceField[2]": {"3"},
	}
	assert.Equal(t, expected, *values)

	// Case: Empty slice
	values = &url.Values{}
	v = reflect.ValueOf([]string{})
	err = encodeSlice(values, "emptySliceField", v)
	assert.Nil(t, err)
	assert.Equal(t, url.Values{}, *values)

	// Case: Slice with unsupported element type
	type UnsupportedType struct {
		Field string
	}
	values = &url.Values{}
	v = reflect.ValueOf([]UnsupportedType{{Field: "value"}})
	err = encodeSlice(values, "unsupportedSliceField", v)
	assert.NotNil(t, err)
	assert.EqualError(
		t,
		err,
		"cannot encode field \"Field\" because it has no json tag",
	)
}

// TestEncodeStruct tests the encodeStruct function.
func TestEncodeStruct(t *testing.T) {
	values := &url.Values{}

	// Case: Encoding a struct with a time.Time field
	type StructWithTime struct {
		Timestamp time.Time `json:"timestamp"`
	}

	structWithTime := StructWithTime{
		Timestamp: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	v := reflect.ValueOf(structWithTime)

	err := encodeStruct(values, "timeField", v)
	assert.Nil(t, err)
	expected := url.Values{"timeField.timestamp": {"2023-01-01T00:00:00Z"}}
	assert.Equal(t, expected, *values)

	// Case: Struct with multiple fields
	type MultiFieldStruct struct {
		Title  string `json:"title"`
		Author string `json:"author"`
	}

	values = &url.Values{}
	multiFieldStruct := MultiFieldStruct{
		Title:  "Go Programming",
		Author: "John Doe",
	}
	v = reflect.ValueOf(multiFieldStruct)

	err = encodeStruct(values, "", v)
	assert.Nil(t, err)
	expected = url.Values{
		"title":  {"Go Programming"},
		"author": {"John Doe"},
	}
	assert.Equal(t, expected, *values)
}

// TestEncodeStructField tests the encodeStructField function
func TestEncodeStructField(t *testing.T) {
	type EmbeddedStruct struct {
		Alive bool `json:"alive"`
	}

	// Case: Anonymous field encoding
	type AnonymousField struct {
		EmbeddedStruct
	}

	values := &url.Values{}
	anonymousStruct := AnonymousField{EmbeddedStruct{Alive: true}}
	err := encodeStructField(values, "", reflect.ValueOf(anonymousStruct), 0)
	assert.Nil(t, err)

	expected := url.Values{
		"alive": {"true"},
	}
	assert.Equal(t, expected, *values)

	// Case: Error with unsupported anonymous field type
	type UnsupportedAnonymous struct {
		UnsupportedField int
	}
	type StructWithInvalidAnonymous struct {
		UnsupportedAnonymous        // Anonymous field with unsupported type
		Name                 string `json:"name"`
	}

	values = &url.Values{}
	structWithInvalidAnonymous := StructWithInvalidAnonymous{
		UnsupportedAnonymous: UnsupportedAnonymous{UnsupportedField: 42},
		Name:                 "John Doe",
	}

	// Encoding the anonymous field
	err = encodeStructField(
		values,
		"",
		reflect.ValueOf(structWithInvalidAnonymous),
		0,
	)
	assert.NotNil(t, err)
	assert.EqualError(
		t,
		err,
		"cannot encode field \"UnsupportedField\" because it has no json tag",
	)

	// Define a simple struct type for further tests
	type SimpleStruct struct {
		Name  string `json:"name"`
		Age   int    `json:"age"`
		Alive bool   `json:"alive"`
	}

	// Case: Field with a JSON tag
	values = &url.Values{}
	simpleStruct := SimpleStruct{Name: "John", Age: 30, Alive: true}

	// Encode "Name" field
	err = encodeStructField(values, "", reflect.ValueOf(simpleStruct), 0)
	assert.Nil(t, err)

	expected = url.Values{
		"name": {"John"},
	}
	assert.Equal(t, expected, *values)

	// Case: Error when field without a JSON tag is encountered
	type StructWithoutJSON struct {
		Field string
	}

	values = &url.Values{}
	structWithoutJSON := StructWithoutJSON{Field: "value"}
	err = encodeStructField(values, "", reflect.ValueOf(structWithoutJSON), 0)
	assert.NotNil(t, err)
	assert.EqualError(
		t,
		err,
		"cannot encode field \"Field\" because it has no json tag",
	)

	// Case: Field with a JSON tag of "-"
	type StructWithIgnoredField struct {
		IgnoredField string `json:"-"`
		ValidField   string `json:"valid_field"`
	}

	values = &url.Values{}
	structWithIgnoredField := StructWithIgnoredField{
		IgnoredField: "ignore",
		ValidField:   "valid",
	}

	// Encode "IgnoredField"
	err = encodeStructField(
		values,
		"",
		reflect.ValueOf(structWithIgnoredField),
		0,
	)
	assert.NotNil(t, err)
	assert.EqualError(
		t,
		err,
		"cannot encode field \"IgnoredField\" because it has no json tag",
	)

	// Case: Nested field encoding
	type NestedStruct struct {
		Inner SimpleStruct `json:"inner"`
	}

	values = &url.Values{}
	nestedStruct := NestedStruct{
		Inner: SimpleStruct{Name: "Alice", Age: 25, Alive: true},
	}

	// Encode "Inner" field
	err = encodeStructField(
		values,
		"",
		reflect.ValueOf(nestedStruct),
		0,
	)
	assert.Nil(t, err)

	expected = url.Values{
		"inner.name":  {"Alice"},
		"inner.age":   {"25"},
		"inner.alive": {"true"},
	}
	assert.Equal(t, expected, *values)

	// Case: Error with unsupported field type with a JSON tag
	type StructWithUnsupportedField struct {
		UnsupportedField func() `json:"unsupported_field"`
		ValidField       string `json:"valid_field"`
	}

	values = &url.Values{}
	structWithUnsupportedField := StructWithUnsupportedField{
		UnsupportedField: func() {},
		ValidField:       "valid",
	}

	// Encode "UnsupportedField"
	err = encodeStructField(
		values,
		"",
		reflect.ValueOf(structWithUnsupportedField),
		0,
	)
	assert.NotNil(t, err)
	assert.EqualError(t, err, "value type not supported by URL encoding: func")
}

// TestSetNestedMapValue tests the setNestedMapValue function.
func TestSetNestedMapValue(t *testing.T) {
	// Case: Empty key
	currentMap := map[string]any{}
	depth, err := setNestedMapValue(currentMap, "", "value", 0)
	assert.Nil(t, err)
	assert.Equal(t, currentMap, map[string]any{})
	assert.Equal(t, 1, depth)

	// Case: Simple key-value pair
	currentMap = map[string]any{}
	key := "simpleKey"
	value := "simpleValue"

	// Attempt to set a simple key-value pair
	_, err = setNestedMapValue(currentMap, key, value, 0)
	assert.Nil(t, err, "expected no error for simple key-value pair")
	assert.Equal(
		t,
		"simpleValue",
		currentMap["simpleKey"],
		"expected value to be set correctly",
	)

	// Case: Invalid intermediate value type
	currentMap = map[string]any{}
	currentMap["level1"] = "notAMap"
	key = "level1.level2.key"
	value = "nestedValue"

	_, err = setNestedMapValue(currentMap, key, value, 0)
	assert.NotNil(
		t,
		err,
		"expected an error for invalid intermediate value type",
	)
	expectedError := "expected map[string]any, got string"
	assert.EqualError(
		t,
		err,
		expectedError,
		"expected error message for invalid intermediate value type",
	)

	// Case: Max recursion depth exceeded
	currentMap = map[string]any{}
	_, err = setNestedMapValue(currentMap, "", "value", maxRecursionDepth+1)
	assert.NotNil(t, err)
	assert.EqualError(
		t,
		err,
		"exceeded maximum recursion depth of "+strconv.Itoa(maxRecursionDepth),
	)
}

// TestSetFinalValue tests the setFinalValue function.
func TestSetFinalValue(t *testing.T) {
	currentMap := map[string]any{}

	// Case: Set a regular key-value pair
	err := setFinalValue(currentMap, "simpleKey", "simpleValue")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expected := map[string]any{"simpleKey": "simpleValue"}
	if !reflect.DeepEqual(currentMap, expected) {
		t.Errorf("expected %v, got %v", expected, currentMap)
	}

	// Case: Set a value in a slice
	err = setFinalValue(currentMap, "mySlice[1]", "sliceValue")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	convertMinSlicesToRegularSlices(currentMap)

	expected = map[string]any{
		"simpleKey": "simpleValue",
		"mySlice":   []any{"sliceValue"},
	}

	if !reflect.DeepEqual(currentMap, expected) {
		t.Errorf("expected %v, got %v", expected, currentMap)
	}

	// Case: Ensure it overwrites existing value
	err = setFinalValue(currentMap, "simpleKey", "newValue")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	convertMinSlicesToRegularSlices(currentMap)

	expected["simpleKey"] = "newValue"
	if !reflect.DeepEqual(currentMap, expected) {
		t.Errorf("expected %v, got %v", expected, currentMap)
	}
}

// TestSetSliceValue tests the setSliceValue function.
func TestSetSliceValue(t *testing.T) {
	// Case: Valid slice index and value
	currentMap := map[string]any{}
	sliceIndex := []string{"", "mySlice", "1"}
	value := "testValue"

	// Attempt to set a value at a valid slice index
	err := setSliceValue(currentMap, sliceIndex, value)
	assert.Nil(t, err, "expected no error for valid slice index and value")

	// Ensure the value is set correctly in the slice
	storedSlice, ok := currentMap["mySlice"].(*minSlice)
	assert.True(t, ok, "expected stored slice to be of type *minSlice")
	storedValue, exists := storedSlice.get(1)
	assert.True(t, exists, "expected value to exist at index 1")
	assert.Equal(
		t,
		value,
		storedValue,
		"expected the value to be set correctly in the slice",
	)

	// Case: Invalid slice index format
	sliceIndex = []string{"", "mySlice", "abc"}

	// Attempt to set a value with an invalid slice index format
	err = setSliceValue(currentMap, sliceIndex, value)
	assert.NotNil(t, err, "expected an error for invalid slice index format")
	expectedError := "invalid index: abc"
	assert.EqualError(
		t,
		err,
		expectedError,
		"expected error message for invalid index format",
	)

	// Case: Invalid slice type in current map
	currentMap = map[string]any{}
	currentMap["invalidSlice"] = "notASlice" // Set an invalid type

	// Attempt to set an invalid value type
	sliceIndex = []string{"", "invalidSlice", "0"}
	err = setSliceValue(currentMap, sliceIndex, value)
	assert.NotNil(t, err, "expected an error from getOrCreateSlice")
	expectedError = "expected *minSlice, got string"
	assert.EqualError(
		t,
		err,
		expectedError,
		"expected error message for incorrect slice type",
	)
}

// TestGetIntermediateValue tests the getIntermediateValue function.
func TestGetIntermediateValue(t *testing.T) {
	currentMap := map[string]any{}

	// Case: Get or create a regular map
	part := "subMap"

	result, err := getIntermediateValue(currentMap, part)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expected := make(map[string]any)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}

	// Ensure the map is created correctly in the parent map
	if _, ok := currentMap[part]; !ok {
		t.Errorf("expected key %q to be created in parent map", part)
	}

	// Case: Delegate to getSliceValue for slice parts
	part = "mySlice[0]"
	result, err = getIntermediateValue(currentMap, part)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expectedSlice := make(map[string]any)
	if !reflect.DeepEqual(result, expectedSlice) {
		t.Errorf("expected %v, got %v", expectedSlice, result)
	}

	// Ensure the slice value is created correctly
	if _, ok := currentMap["mySlice"]; !ok {
		t.Errorf("expected slice key %q to be created in parent map", "mySlice")
	}
}

func TestGetMap(t *testing.T) {
	// Case: Key exists and is a map[string]any
	currentMap := map[string]any{
		"validMap": map[string]any{
			"key": "value",
		},
	}

	result, err := getMap(currentMap, "validMap")
	assert.Nil(t, err)
	assert.Equal(t, currentMap["validMap"], result)

	// Case: Key does not exist
	_, err = getMap(currentMap, "nonExistentKey")
	if err == nil {
		t.Fatalf("expected an error, got nil")
	}

	expectedError := "expected map[string]any, got <nil>"
	if err.Error() != expectedError {
		t.Errorf("expected error %q, got %q", expectedError, err.Error())
	}

	// Case: Key exists but is not a map[string]any
	currentMap["notAMap"] = "some string"
	_, err = getMap(currentMap, "notAMap")
	if err == nil {
		t.Fatalf("expected an error, got nil")
	}

	expectedError = "expected map[string]any, got string"
	if err.Error() != expectedError {
		t.Errorf("expected error %q, got %q", expectedError, err.Error())
	}
}

// TestCreateMapIntoSlice tests the TestCreateMapIntoSlice function.
func TestCreateMapIntoSlice(t *testing.T) {
	// Case: Valid slice index creation
	currentMap := map[string]any{}
	sliceIndex := []string{"", "mySlice", "1"}

	// Attempt to create a map into slice with valid slice index
	result, err := createMapIntoSlice(sliceIndex, currentMap)
	assert.Nil(t, err, "expected no error for valid slice index creation")
	assert.NotNil(t, result, "expected a new map to be created")
	assert.IsType(
		t,
		map[string]any{},
		result,
		"expected result to be of type map[string]any",
	)

	// Ensure the slice is created correctly in the map
	storedSlice, ok := currentMap["mySlice"].(*minSlice)
	assert.True(t, ok, "expected stored slice to be of type *minSlice")
	storedElem, exists := storedSlice.get(1)
	assert.True(t, exists, "expected element to exist at index 1")
	assert.Equal(
		t,
		result,
		storedElem,
		"expected created map to be stored in the slice",
	)

	// Case: Invalid slice index format
	sliceIndex = []string{"", "mySlice", "abc"}
	_, err = createMapIntoSlice(sliceIndex, currentMap)
	assert.NotNil(t, err, "expected an error for invalid slice index format")
	expectedError := "invalid index: abc"
	assert.EqualError(
		t,
		err,
		expectedError,
		"expected error message for invalid index format",
	)

	// Case: Handling incorrect element type at index
	currentMap = map[string]any{}
	slice, _ := getOrCreateSlice(currentMap, "testSlice")
	slice.set(0, "notAMap")

	sliceIndex = []string{"", "testSlice", "0"}
	_, err = createMapIntoSlice(sliceIndex, currentMap)
	assert.NotNil(
		t,
		err,
		"expected an error for incorrect element type at index",
	)
	expectedError = "expected map[string]any, got string"
	assert.EqualError(
		t,
		err,
		expectedError,
		"expected error message for incorrect element type",
	)

	// Case: Error returned by getOrCreateSlice
	currentMap = map[string]any{}
	currentMap["invalidSlice"] = "notASlice" // Set an invalid type

	sliceIndex = []string{"", "invalidSlice", "0"}
	_, err = createMapIntoSlice(sliceIndex, currentMap)
	assert.NotNil(t, err, "expected an error from getOrCreateSlice")
	expectedError = "expected *minSlice, got string"
	assert.EqualError(
		t,
		err,
		expectedError,
		"expected error message for incorrect slice type",
	)
}

// TestParseSliceIndex tests the ParseSliceIndex function.
func TestParseSliceIndex(t *testing.T) {
	sliceIndex := []string{"", "mySlice", "1"}

	// Case: Valid slice index parsing
	sliceName, index, err := parseSliceIndex(sliceIndex)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if sliceName != "mySlice" || index != 1 {
		t.Errorf(
			"expected slice name %q and index 1, got %q and %d", "mySlice",
			sliceName,
			index,
		)
	}

	// Case: Invalid slice index format
	sliceIndex = []string{"", "mySlice", "abc"}
	_, _, err = parseSliceIndex(sliceIndex)
	if err == nil {
		t.Fatalf("expected an error, got nil")
	}

	// Case: Wrong number of slice index parts
	sliceIndex = []string{"", "mySlice"}
	_, _, err = parseSliceIndex(sliceIndex)
	if err == nil {
		t.Fatalf("expected an error, got nil")
	}
}

// TestGetOrCreateSlice tests the GetOrCreateSlice function.
func TestGetOrCreateSlice(t *testing.T) {
	// Case: Creating a new slice
	currentMap := map[string]any{}
	sliceName := "newSlice"

	// Attempt to create a new slice
	slice, err := getOrCreateSlice(currentMap, sliceName)
	assert.Nil(t, err, "expected no error when creating a new slice")
	assert.NotNil(t, slice, "expected slice to be created")
	assert.IsType(t, &minSlice{}, slice, "expected a *minSlice type")

	// Ensure the slice is created in the map
	storedSlice, ok := currentMap[sliceName].(*minSlice)
	assert.True(t, ok, "expected stored slice to be of type *minSlice")
	assert.Equal(
		t,
		storedSlice,
		slice,
		"expected the created slice to be stored in the map",
	)

	// Case: Retrieving an existing slice
	existingSlice := newMinSlice()
	currentMap["existingSlice"] = existingSlice

	// Attempt to retrieve the existing slice
	retrievedSlice, err := getOrCreateSlice(currentMap, "existingSlice")
	assert.Nil(t, err, "expected no error when retrieving an existing slice")
	assert.Equal(
		t,
		existingSlice,
		retrievedSlice,
		"expected the retrieved slice to match the existing slice",
	)

	// Case: Handling incorrect type
	currentMap["incorrectType"] = "notASlice"

	// Attempt to retrieve a slice where the value is not a *minSlice
	_, err = getOrCreateSlice(currentMap, "incorrectType")
	assert.NotNil(t, err, "expected an error when the value is not a *minSlice")
	expectedError := "expected *minSlice, got string"
	assert.EqualError(
		t,
		err,
		expectedError,
		"expected error message for incorrect type",
	)

	// Case: Too big slice
	tooBigSlice := newMinSlice()
	// Simulate the slice reaching its maximum size
	for i := 0; i < maxSliceSize; i++ {
		tooBigSlice.set(i, fmt.Sprintf("value%d", i))
	}
	currentMap["tooBigSlice"] = tooBigSlice

	// Attempt to add an element beyond the maxSliceSize
	_, err = getOrCreateSlice(currentMap, "tooBigSlice")
	assert.NotNil(t, err, "expected an error when exceeding maximum slice size")
	expectedError = fmt.Sprintf(
		"exceeded maximum slice size of %d",
		maxSliceSize,
	)
	assert.EqualError(
		t,
		err,
		expectedError,
		"expected error message for too big slice",
	)

}
