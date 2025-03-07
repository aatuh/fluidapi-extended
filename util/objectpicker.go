package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/pakkasys/fluidapi/core"
)

var InvalidInputError = core.NewAPIError("INVALID_INPUT")

const (
	sourceTag     = "source"
	structTag     = "json"
	sourceURL     = "url"
	sourceBody    = "body"
	sourceHeader  = "header"
	sourceHeaders = "headers"
	sourceCookie  = "cookie"
	sourceCookies = "cookies"
)

type BodyData map[string]any
type URLData map[string]any

// URLDecoder interface for decoding URL values.
type URLDecoder interface {
	Decode(values url.Values) (map[string]any, error)
}

// MapFieldConfig lets define the mapping for extracting a map input.
//   - Source: if set at a level, it will override any field-level sources in
//     the children.
//   - ExpectedType: if set for a leaf field, the value is converted to that
//     type. For example, "int64", "int", "float64", etc.
type MapFieldConfig struct {
	Source       string
	ExpectedType string
	DefaultValue any
	Optional     bool
	Fields       map[string]*MapFieldConfig
}

// ObjectPicker is generic over T which may be a struct or a map.
type ObjectPicker struct {
	urlDecoder    URLDecoder
	conversionMap map[string]func(any) any
}

// NewObjectPicker returns a new ObjectPicker.
func NewObjectPicker(
	urlDecoder URLDecoder, conversionMap map[string]func(any) any,
) *ObjectPicker {
	return &ObjectPicker{
		urlDecoder:    urlDecoder,
		conversionMap: conversionMap,
	}
}

// PickMap extracts a map[string]any from the request based on the provided
// MapFieldConfig. It supports both nested and flat extraction.
// If a source is set at a higher (parent) level, then child fields will use
// that source.
func (o *ObjectPicker) PickMap(
	r *http.Request, config *MapFieldConfig,
) (map[string]any, error) {
	// Load body and URL data since a map may draw from either.
	bodyData, err := o.bodyToMap(r)
	if err != nil {
		return nil, InvalidInputError
	}
	urlData, err := o.urlDecoder.Decode(r.URL.Query())
	if err != nil {
		return nil, InvalidInputError
	}

	// Do not override top-level if not explicitly provided.
	return o.extractMap(config, config.Source, r, urlData, bodyData), nil
}

// determineDefaultSource selects the default source based on HTTP method.
func (o *ObjectPicker) determineDefaultSource(httpMethod string) string {
	switch httpMethod {
	case http.MethodGet:
		return sourceURL
	case http.MethodPost, http.MethodPut, http.MethodPatch,
		http.MethodDelete:
		return sourceBody
	default:
		return sourceBody
	}
}

// getValueFromSource retrieves a value for a field from a specific source.
func (o *ObjectPicker) getValueFromSource(
	r *http.Request,
	field string,
	source string,
	urlData URLData,
	bodyData BodyData,
) any {
	switch source {
	case sourceURL:
		if val, exists := urlData[field]; exists {
			return val
		}
	case sourceBody:
		if val, exists := bodyData[field]; exists {
			return val
		}
	case sourceHeader, sourceHeaders:
		if val := r.Header.Get(field); val != "" {
			return val
		}
	case sourceCookie, sourceCookies:
		if cookie, err := r.Cookie(field); err == nil {
			return cookie.Value
		}
	default:
		if len(source) != 0 {
			panic(fmt.Sprintf("Unknown input source: %s", source))
		}
	}
	return ""
}

// bodyToMap decodes the request body (if any) into a map.
func (o *ObjectPicker) bodyToMap(r *http.Request) (BodyData, error) {
	body, err := o.getBody(r)
	if err != nil {
		return nil, err
	}
	if len(body) == 0 {
		return nil, nil
	}

	var m BodyData
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber() // to preserve numeric precision
	if err = decoder.Decode(&m); err != nil {
		return nil, err
	}
	return m, nil
}

// getBody reads and restores the request body.
func (o *ObjectPicker) getBody(r *http.Request) ([]byte, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	r.Body = io.NopCloser(bytes.NewBuffer(body))
	return body, nil
}

// extractMap recursively extracts fields from the request data using a config.
// The effective source is determined as follows:
//   - If a parentSource was explicitly provided (non-empty), it takes
//     precedence.
//   - Else if the field defines a source, use that.
//   - Otherwise, use the default source determined by the request.
//
// extractMap recursively extracts fields from the request data using a config.
func (o *ObjectPicker) extractMap(
	config *MapFieldConfig,
	parentSource string,
	r *http.Request,
	urlData URLData,
	bodyData BodyData,
) map[string]any {
	res := make(map[string]any)
	if config.Fields == nil {
		return res
	}

	for key, fieldCfg := range config.Fields {
		// Determine effectiveSource for this field.
		var effectiveSource string
		if parentSource != "" {
			// Parent explicitly defined a source; always use that.
			effectiveSource = parentSource
		} else if fieldCfg.Source != "" {
			// No parent's source, but the field itself defines one.
			effectiveSource = fieldCfg.Source
		} else {
			// No parent's source and none on the field; use the default one.
			effectiveSource = o.determineDefaultSource(r.Method)
		}

		if fieldCfg.Fields != nil {
			// Process nested fields.
			raw := o.getValueFromSource(r, key, effectiveSource, urlData, bodyData)
			switch v := raw.(type) {
			case map[string]any:
				// The field is a nested object.
				res[key] = o.extractMapFromRaw(v, fieldCfg, effectiveSource)
			case []any:
				// The field is a slice of nested objects.
				var list []any
				for _, elem := range v {
					if m, ok := elem.(map[string]any); ok {
						list = append(list, o.extractMapFromRaw(m, fieldCfg, effectiveSource))
					} else {
						// If the element is not a map, you can either skip it or add it as-is.
						list = append(list, elem)
					}
				}
				res[key] = list
			default:
				// Fall back to flat composite keys (e.g. "parent.child")
				nested := make(map[string]any)
				for subKey := range fieldCfg.Fields {
					compositeKey := key + "." + subKey
					val := o.getValueFromSource(
						r, compositeKey, effectiveSource, urlData, bodyData,
					)
					val = convertValue(
						val,
						fieldCfg.Fields[subKey].ExpectedType,
						o.conversionMap,
					)
					// If value is missing and field is optional, skip it.
					if (val == nil || val == "") && fieldCfg.Fields[subKey].Optional {
						continue
					}
					// If value is missing, try default value.
					if (val == nil || val == "") && fieldCfg.Fields[subKey].DefaultValue != nil {
						val = fieldCfg.Fields[subKey].DefaultValue
					}
					if val != nil && val != "" {
						nested[subKey] = val
					}
				}
				if len(nested) > 0 {
					res[key] = nested
				}
			}
		} else {
			// Leaf field.
			val := o.getValueFromSource(
				r,
				key,
				effectiveSource,
				urlData,
				bodyData,
			)
			// If value is missing and the field is optional, skip it.
			if (val == nil || val == "") && fieldCfg.Optional {
				continue
			}
			// If value is missing, check for a default value.
			if (val == nil || val == "") && fieldCfg.DefaultValue != nil {
				val = fieldCfg.DefaultValue
			} else {
				// Otherwise convert the value as needed.
				val = convertValue(val, fieldCfg.ExpectedType, o.conversionMap)
			}
			if val != nil && val != "" {
				res[key] = val
			}
		}
	}
	return res
}

// extractMapFromRaw applies nested config on an already parsed raw map.

func (o *ObjectPicker) extractMapFromRaw(
	raw map[string]any, config *MapFieldConfig, parentSource string,
) map[string]any {
	res := make(map[string]any)
	for key, fieldCfg := range config.Fields {
		if fieldCfg.Fields != nil {
			// Check if the field is a nested object.
			if subRaw, ok := raw[key].(map[string]any); ok {
				nested := o.extractMapFromRaw(subRaw, fieldCfg, parentSource)
				if len(nested) > 0 {
					res[key] = nested
				}
			} else if arr, ok := raw[key].([]any); ok {
				// If it's a slice, iterate over each element.
				var list []any
				for _, elem := range arr {
					if m, ok := elem.(map[string]any); ok {
						list = append(
							list,
							o.extractMapFromRaw(m, fieldCfg, parentSource),
						)
					} else {
						list = append(list, elem)
					}
				}
				if len(list) > 0 {
					res[key] = list
				}
			} else {
				// If nested data is missing, assign default if provided.
				if !fieldCfg.Optional && fieldCfg.DefaultValue != nil {
					res[key] = fieldCfg.DefaultValue
				}
			}
		} else {
			// Leaf field extraction.
			if val, ok := raw[key]; ok {
				res[key] = convertValue(
					val, fieldCfg.ExpectedType, o.conversionMap,
				)
			} else {
				if !fieldCfg.Optional && fieldCfg.DefaultValue != nil {
					res[key] = fieldCfg.DefaultValue
				}
			}
		}
	}
	return res
}

// convertValue converts a value to the expected type if needed.
// If a conversion function exists in convMap for the expectedType,
// it is applied.
func convertValue(
	val any, expectedType string, convMap map[string]func(any) any,
) any {
	if expectedType == "" {
		return val
	}

	// Check for a custom conversion function in the global conversion map.
	if convMap != nil {
		if convFunc, ok := convMap[expectedType]; ok {
			if converted := convFunc(val); converted != nil {
				return converted
			}
		}
	}

	// Attempt to convert based on the expected type.
	switch expectedType {
	case "int64":
		// Check if already int64.
		if v, ok := val.(int64); ok {
			return v
		}
		// Handle json.Number.
		if num, ok := val.(json.Number); ok {
			if intVal, err := num.Int64(); err == nil {
				return intVal
			}
		}
		// Handle float64.
		if f, ok := val.(float64); ok {
			return int64(f)
		}
		// Handle string.
		if s, ok := val.(string); ok {
			if intVal, err := strconv.ParseInt(s, 10, 64); err == nil {
				return intVal
			}
		}
	case "int":
		// Check if already int.
		if v, ok := val.(int); ok {
			return v
		}
		// Use int64 conversion as intermediary.
		if converted := convertValue(val, "int64", convMap); converted != val {
			if int64Val, ok := converted.(int64); ok {
				return int(int64Val)
			}
		}
	case "float64":
		// Check if already float64.
		if v, ok := val.(float64); ok {
			return v
		}
		// Handle json.Number.
		if num, ok := val.(json.Number); ok {
			if floatVal, err := num.Float64(); err == nil {
				return floatVal
			}
		}
		// Handle string.
		if s, ok := val.(string); ok {
			if floatVal, err := strconv.ParseFloat(s, 64); err == nil {
				return floatVal
			}
		}
	case "bool":
		// Check if already bool.
		if v, ok := val.(bool); ok {
			return v
		}
		// Handle string.
		if s, ok := val.(string); ok {
			if boolVal, err := strconv.ParseBool(s); err == nil {
				return boolVal
			}
		}
	}

	return val
}
