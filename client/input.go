package client

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"
)

const (
	sourceTag = "source"
	jsonTag   = "json"

	tagURL     = "url"
	tagBody    = "body"
	tagHeader  = "header"
	tagHeaders = "headers"
	tagCookie  = "cookie"
	tagCookies = "cookies"
)

// ParseInput parses the input struct and returns the parsed data. It will
// populate the URL parameters, headers, cookies, and body based on the
// struct tags. E.g. the struct field `source: url` will be placed in the URL.
//
// Example:
//
//	type MyInput struct {
//	    ID   int    `json:"id" source:"url"`
//	    Name string `json:"name"`
//	}
//	input := &MyInput{ID: 42, Name: "example"}
//	data, err := ParseInput("GET", input)
//
// Parameters:
//   - method: The HTTP method for the request.
//   - input: The input struct or pointer to a struct to parse.
//
// Returns:
//   - *RequestData: The parsed request data.
//   - error: An error if parsing fails.
func ParseInput(method string, input any) (*RequestData, error) {
	if input == nil {
		return nil, fmt.Errorf("ParseInput: input cannot be nil")
	}
	val := reflect.ValueOf(input)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		return nil, fmt.Errorf(
			"ParseInput: input must be a pointer to a struct",
		)
	}

	requestData := NewRequestData()
	inputVal := reflect.ValueOf(input).Elem()
	inputType := inputVal.Type()

	// Iterate over the fields of the input struct
	for i := range inputVal.NumField() {
		err := requestData.PlaceField(
			inputVal.Field(i), inputType.Field(i), method,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"processField: error processing field: %v", err,
			)
		}
	}
	return requestData, nil
}

// RequestData represents request data.
type RequestData struct {
	URLParameters map[string]any
	Headers       map[string]string
	Cookies       []http.Cookie
	Body          map[string]any
}

// NewRequestData returns a new instance of RequestData.
func NewRequestData() *RequestData {
	return &RequestData{
		URLParameters: make(map[string]any),
		Headers:       make(map[string]string),
		Cookies:       make([]http.Cookie, 0),
		Body:          make(map[string]any),
	}
}

// PlaceField places the field value in the appropriate part of the request
// data. It will modify the RequestData object data in-place.
//
// Parameters:
//   - field: The field value to process.
//   - fieldInfo: The field information.
//   - method: The HTTP method for the request.
//
// Returns:
//   - error: An error if processing fails.
func (d *RequestData) PlaceField(
	field reflect.Value, fieldInfo reflect.StructField, method string,
) error {
	placement := fieldInfo.Tag.Get(sourceTag)
	if placement == "" {
		placement = d.defaultPlacement(method)
	}
	fieldName := d.determineFieldName(
		fieldInfo.Tag.Get(jsonTag), fieldInfo.Name,
	)
	if fieldName == "" {
		return fmt.Errorf(
			"ProcessField: json tag cannot be empty for field %s",
			fieldInfo.Name,
		)
	}
	return d.placeFieldValue(placement, fieldName, d.fieldValue(field))
}

// defaultPlacement uses the HTTP method to determine the default placement.
func (d *RequestData) defaultPlacement(method string) string {
	switch method {
	case http.MethodGet:
		return tagURL
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return tagBody
	default:
		return tagBody
	}
}

// determineFieldName determines the field name to use in the request based on
// the JSON tag and the field name.
func (d *RequestData) determineFieldName(
	jsonTag string, fieldName string,
) string {
	json := strings.Split(jsonTag, ",")[0]
	if json == "" {
		json = fieldName
	}
	return json
}

// fieldValue extracts the field value based on its type.
func (d *RequestData) fieldValue(field reflect.Value) any {
	switch field.Kind() {
	case reflect.Bool:
		return field.Bool()
	case reflect.String:
		return field.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return field.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return field.Uint()
	case reflect.Float32, reflect.Float64:
		return field.Float()
	default:
		return field.Interface()
	}
}

// placeFieldValue places the field value in the appropriate map or slice.
func (d *RequestData) placeFieldValue(
	placement string, field string, value any,
) error {
	switch placement {
	case tagURL:
		d.URLParameters[field] = value
	case tagBody:
		d.Body[field] = value
	case tagHeader, tagHeaders:
		d.Headers[field] = fmt.Sprintf("%v", value)
	case tagCookie, tagCookies:
		d.Cookies = append(d.Cookies, http.Cookie{
			Name:  field,
			Value: fmt.Sprintf("%v", value),
		})
	default:
		return fmt.Errorf("placeFieldValue: invalid source tag: %s", placement)
	}
	return nil
}
