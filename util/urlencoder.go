package util

import (
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	maxRecursionDepth = 10   // Maximum allowed depth for nested structures
	maxSliceSize      = 1000 // Maximum allowed size for slices
)

// Matches a string with a word followed by "[" and a number in decimal
// (base 10) and "]" e.g. "mySlice[0]" matches as "mySlice" and "0"
const sliceRegexp = `(\w+)\[(\d+)\]`

// Tag for JSON encoding
const jsonTag = "json"

type URLEncoder struct{}

func NewURLEncoder() *URLEncoder {
	return &URLEncoder{}
}

// Encodes data into URL values to support the following syntax:
// someKey=value
// someStruct.field=value
// someSlice[0]=value
// someStruct[0].key=value
//
// It will preserve the order of the fields in the struct.
// It will return an error if a "json" tag is not found for a struct field.
//
// Parameters:
//   - data: Data to encode
//
// Returns:
//   - url.Values: URL values
//   - error: Error
func (e URLEncoder) Encode(data map[string]any) (url.Values, error) {
	values := url.Values{}
	for key, value := range data {
		err := encodeURL(&values, key, reflect.ValueOf(value))
		if err != nil {
			return nil, err
		}
	}

	return values, nil
}

// Decodes data from URL values to support the following syntax:
// someKey=value
// someStruct.field=value
// someSlice[0]=value
// someStruct[0].key=value
//
// It will preserve the order of the fields in the struct.
//
// Parameters:
//   - values: URL values
//
// Returns:
//   - Decoded data as a map
//   - Error if decoding fails
func (e URLEncoder) Decode(values url.Values) (map[string]any, error) {
	return decodeURL(values)
}

// TODO: Test order preservation
func decodeURL(values url.Values) (map[string]any, error) {
	urlData := make(map[string]any)
	depth := 0
	for key, value := range values {
		var err error
		depth, err = setNestedMapValue(urlData, key, value[0], depth)
		if err != nil {
			return nil, err
		}
	}
	convertMinSlicesToRegularSlices(urlData)
	return urlData, nil
}

// TODO: Test order preservation
func encodeURL(values *url.Values, fieldTag string, v reflect.Value) error {
	return encodeValue(values, fieldTag, v)
}

func encodeValue(values *url.Values, fieldTag string, v reflect.Value) error {
	switch v.Kind() {
	case reflect.Ptr, reflect.Interface:
		return encodePointer(values, fieldTag, v)
	case reflect.String:
		return encodeString(values, fieldTag, v)
	case reflect.Int, reflect.Int32, reflect.Int64:
		return encodeInt(values, fieldTag, v)
	case reflect.Bool:
		return encodeBool(values, fieldTag, v)
	case reflect.Slice:
		return encodeSlice(values, fieldTag, v)
	case reflect.Struct:
		return encodeStruct(values, fieldTag, v)
	default:
		return fmt.Errorf(
			"value type not supported by URL encoding: %s",
			v.Kind(),
		)
	}
}

func encodePointer(values *url.Values, fieldTag string, v reflect.Value) error {
	if !v.IsNil() {
		return encodeValue(values, fieldTag, v.Elem())
	}
	return nil
}

func encodeString(values *url.Values, fieldTag string, v reflect.Value) error {
	values.Set(fieldTag, v.String())
	return nil
}

func encodeInt(values *url.Values, fieldTag string, v reflect.Value) error {
	values.Set(fieldTag, fmt.Sprintf("%d", v.Int()))
	return nil
}

func encodeBool(values *url.Values, fieldTag string, v reflect.Value) error {
	values.Set(fieldTag, strconv.FormatBool(v.Bool()))
	return nil
}

func encodeSlice(values *url.Values, fieldTag string, v reflect.Value) error {
	for j := 0; j < v.Len(); j++ {
		sliceElem := v.Index(j)
		newFieldTag := fmt.Sprintf("%s[%d]", fieldTag, j)
		if err := encodeValue(values, newFieldTag, sliceElem); err != nil {
			return err
		}
	}
	return nil
}

func encodeStruct(values *url.Values, fieldTag string, v reflect.Value) error {
	if v.Type() == reflect.TypeOf(time.Time{}) {
		values.Set(fieldTag, v.Interface().(time.Time).Format(time.RFC3339))
		return nil
	}
	for i := 0; i < v.NumField(); i++ {
		if err := encodeStructField(values, fieldTag, v, i); err != nil {
			return err
		}
	}
	return nil
}

func encodeStructField(
	values *url.Values,
	fieldTag string,
	v reflect.Value,
	i int,
) error {
	field := v.Field(i)
	fieldType := v.Type().Field(i)

	if fieldType.Anonymous {
		if err := encodeValue(values, fieldTag, field); err != nil {
			return err
		}
		return nil
	}

	newFieldTag := fieldType.Tag.Get(jsonTag)
	if newFieldTag == "-" || newFieldTag == "" {
		return fmt.Errorf(
			"cannot encode field %q because it has no json tag",
			fieldType.Name,
		)
	}

	if fieldTag != "" {
		newFieldTag = fieldTag + "." + newFieldTag
	}
	if err := encodeValue(values, newFieldTag, field); err != nil {
		return err
	}

	return nil
}

func setNestedMapValue(
	current map[string]any,
	key string,
	value any,
	depth int,
) (int, error) {
	if depth > maxRecursionDepth {
		return 0, fmt.Errorf(
			"exceeded maximum recursion depth of %d",
			maxRecursionDepth,
		)
	}
	depth++

	var parts []string
	if key != "" {
		parts = strings.Split(key, ".")
	}
	for i := range parts {
		part := parts[i]
		if i == len(parts)-1 {
			return 0, setFinalValue(current, part, value)
		}
		var err error
		current, err = getIntermediateValue(current, part)
		if err != nil {
			return 0, err
		}
	}
	return depth, nil
}

func setFinalValue(current map[string]any, part string, value any) error {
	reg := regexp.MustCompile(sliceRegexp)
	if sliceIndex := reg.FindStringSubmatch(part); sliceIndex != nil {
		return setSliceValue(current, sliceIndex, value)
	}
	current[part] = value
	return nil
}

func setSliceValue(
	current map[string]any,
	sliceIndex []string,
	value any,
) error {
	sliceName, idx, err := parseSliceIndex(sliceIndex)
	if err != nil {
		return err
	}
	slice, err := getOrCreateSlice(current, sliceName)
	if err != nil {
		return err
	}
	slice.set(idx, value)
	current[sliceName] = slice // Use MinSlice to handle slice elements safely
	return nil
}

func getIntermediateValue(
	current map[string]any,
	part string,
) (map[string]any, error) {
	reg := regexp.MustCompile(sliceRegexp)
	if sliceIndex := reg.FindStringSubmatch(part); sliceIndex != nil {
		return createMapIntoSlice(sliceIndex, current)
	}
	// Create a map with the part name if it doesn't exist
	if _, ok := current[part]; !ok {
		current[part] = make(map[string]any)
	}
	return getMap(current, part)
}

func getMap(current map[string]any, part string) (map[string]any, error) {
	retMap, ok := current[part]
	if !ok {
		return nil, fmt.Errorf("expected map[string]any, got %T", current[part])
	}
	cast, ok := retMap.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("expected map[string]any, got %T", retMap)
	}
	return cast, nil
}

func createMapIntoSlice(
	sliceIndex []string,
	current map[string]any,
) (map[string]any, error) {
	sliceName, idx, err := parseSliceIndex(sliceIndex)
	if err != nil {
		return nil, err
	}
	slice, err := getOrCreateSlice(current, sliceName)
	if err != nil {
		return nil, err
	}
	// Ensure the element at idx is a map and initialize if necessary
	elem, exists := slice.get(idx)
	if !exists {
		elem = make(map[string]any)
		slice.set(idx, elem)
	}
	// Eensure elem is a map
	castedElem, ok := elem.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("expected map[string]any, got %T", elem)
	}
	current[sliceName] = slice
	return castedElem, nil
}

func parseSliceIndex(sliceIndex []string) (string, int, error) {
	if len(sliceIndex) != 3 {
		return "", 0, fmt.Errorf("invalid slice index: %s", sliceIndex)
	}
	// Matched slice name and index, e.g. "mySlice" and "0"
	sliceName, index := sliceIndex[1], sliceIndex[2]
	// Index must be an integer
	idx, err := strconv.Atoi(index)
	if err != nil {
		return "", 0, fmt.Errorf("invalid index: %s", index)
	}
	return sliceName, idx, nil
}

func getOrCreateSlice(
	current map[string]any,
	sliceName string,
) (*minSlice, error) {
	if _, ok := current[sliceName]; !ok {
		current[sliceName] = newMinSlice()
	}
	minSlice, ok := current[sliceName].(*minSlice)
	if !ok {
		return nil, fmt.Errorf("expected *minSlice, got %T", current[sliceName])
	}
	if len(minSlice.elements) >= maxSliceSize {
		return nil, fmt.Errorf(
			"exceeded maximum slice size of %d",
			maxSliceSize,
		)
	}
	return minSlice, nil
}

// MinSlices keeps track of slice elements with minimal length
type minSlice struct {
	elements map[int]any
}

func newMinSlice() *minSlice {
	return &minSlice{elements: make(map[int]any)}
}

func (s *minSlice) set(index int, value any) {
	s.elements[index] = value
}

func (s *minSlice) get(index int) (any, bool) {
	value, exists := s.elements[index]
	return value, exists
}

func (s *minSlice) toSlice() []any {
	slice := make([]any, 0, len(s.elements))
	for _, value := range s.elements {
		slice = append(slice, value)
	}
	return slice
}

// Converts all MinSlice instances in the map to regular slices recursively.
func convertMinSlicesToRegularSlices(data map[string]any) {
	for key, value := range data {
		switch v := value.(type) {
		case *minSlice:
			data[key] = v.toSlice()
		case map[string]any:
			convertMinSlicesToRegularSlices(v)
		}
	}
}
