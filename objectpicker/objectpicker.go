package objectpicker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"slices"
	"strings"
	"sync"

	"github.com/pakkasys/fluidapi/core/api"

	"github.com/mitchellh/mapstructure"
)

var InvalidInputError = api.NewError[any]("INVALID_INPUT")

type FieldConfig map[string][]string
type StructRegistry map[reflect.Type]FieldConfig
type BodyData map[string]any
type URLData map[string]any
type PickedObjects map[reflect.Type]any

// URLDecoder interface for decoding URL values.
type URLDecoder interface {
	DecodeURL(values url.Values) (map[string]any, error)
}

const (
	sourceTag     = "source"
	jsonTag       = "json"
	sourceURL     = "url"
	sourceBody    = "body"
	sourceHeader  = "header"
	sourceHeaders = "headers"
	sourceCookie  = "cookie"
	sourceCookies = "cookies"
)

var (
	structRegistry = make(StructRegistry)
	registryMutex  = sync.RWMutex{}
)

type ObjectPicker[T any] struct {
	urlDecoder URLDecoder
}

func NewObjectPicker[T any](urlDecoder URLDecoder) *ObjectPicker[T] {
	return &ObjectPicker[T]{
		urlDecoder: urlDecoder,
	}
}

func (o *ObjectPicker[T]) PickObject(
	r *http.Request,
	w http.ResponseWriter,
	obj T,
) (*T, error) {
	typ := reflect.TypeOf(obj)
	fieldConfig := structRegistry[typ]

	if fieldConfig == nil {
		o.mustUpdateObjectRegistry(
			[]any{obj},
			o.determineDefaultSource(r.Method),
		)
		fieldConfig = structRegistry[typ]
	}

	var bodyData BodyData
	if o.needsSource(fieldConfig, sourceBody) {
		var err error
		bodyData, err = o.bodyToMap(r)
		if err != nil {
			return nil, InvalidInputError
		}
	}

	var urlData URLData
	if o.needsSource(fieldConfig, sourceURL) {
		var err error
		urlData, err = o.urlDecoder.DecodeURL(r.URL.Query())
		if err != nil {
			return nil, InvalidInputError
		}
	}

	pickedObject, err := o.pickObjectForObj(
		typ,
		structRegistry[typ],
		urlData,
		bodyData,
		r,
	)
	if err != nil {
		return nil, err
	}

	castObj, ok := pickedObject.(T)
	if !ok {
		return nil, fmt.Errorf(
			"failed to cast picked object: %v",
			pickedObject,
		)
	}
	return &castObj, nil
}

func (o *ObjectPicker[T]) determineDefaultSource(
	httpMethod string,
) string {
	switch httpMethod {
	case http.MethodGet:
		return sourceURL
	case http.MethodPost:
		return sourceBody
	case http.MethodPut:
		return sourceBody
	case http.MethodPatch:
		return sourceBody
	case http.MethodDelete:
		return sourceBody
	default:
		return sourceBody
	}
}

func (o *ObjectPicker[T]) mustUpdateObjectRegistry(
	objectSamples []any,
	defaultSource string,
) {
	var err error
	structRegistry, err = o.addToStructRegistry(
		structRegistry,
		objectSamples,
		defaultSource,
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to add to struct registry: %v", err))
	}
}

func (o *ObjectPicker[T]) pickObjectForObj(
	typ reflect.Type,
	fieldConfig FieldConfig,
	urlData URLData,
	bodyData BodyData,
	request *http.Request,
) (any, error) {
	ptr := reflect.New(typ)

	valueMap := o.populateValuesFromSources(
		ptr.Interface(),
		fieldConfig,
		urlData,
		bodyData,
		request,
	)

	decoderConfig := &mapstructure.DecoderConfig{
		Metadata:         nil,
		Result:           ptr.Interface(),
		TagName:          jsonTag,
		WeaklyTypedInput: true,
	}

	decoder, err := mapstructure.NewDecoder(decoderConfig)
	if err != nil {
		return nil, InvalidInputError
	}

	if err := decoder.Decode(valueMap); err != nil {
		return nil, InvalidInputError
	}

	return ptr.Elem().Interface(), nil
}

func (o *ObjectPicker[T]) addToStructRegistry(
	structRegistry StructRegistry,
	pickedObjects []any,
	defaultSource string,
) (StructRegistry, error) {
	registryMutex.Lock()
	defer registryMutex.Unlock()

	for i := range pickedObjects {
		pickedObject := pickedObjects[i]
		objectType := reflect.TypeOf(pickedObject)

		if objectType.Kind() == reflect.Ptr {
			objectType = objectType.Elem()
		}

		fieldConfig, err := o.buildFieldConfig(pickedObject, defaultSource)
		if err != nil {
			return nil, err
		}

		if _, ok := structRegistry[objectType]; ok {
			for fieldName, sources := range fieldConfig {
				fieldConfig[fieldName] = append(
					structRegistry[objectType][fieldName],
					sources...,
				)
			}
		} else {
			structRegistry[objectType] = fieldConfig
		}
	}

	return structRegistry, nil
}

func (o *ObjectPicker[T]) buildFieldConfig(
	obj any,
	defaultSource string,
) (FieldConfig, error) {
	typ := reflect.TypeOf(obj)
	if typ == nil {
		return nil, fmt.Errorf("nil type received")
	}
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected a struct type but got %s", typ.Kind())
	}

	config := make(FieldConfig)
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldName := field.Tag.Get(jsonTag)
		if fieldName == "" {
			continue
		}

		sourceTagValue := field.Tag.Get(sourceTag)
		var sources []string
		if sourceTagValue == "" {
			sources = []string{defaultSource}
		} else {
			sources = strings.Split(sourceTagValue, ",")
		}

		config[fieldName] = sources
	}

	return config, nil
}

func (o *ObjectPicker[T]) populateValuesFromSources(
	obj any,
	config FieldConfig,
	urlData URLData,
	bodyData BodyData,
	request *http.Request,
) map[string]any {
	val := reflect.ValueOf(obj).Elem()
	typ := val.Type()
	valueMap := make(map[string]any)

	for i := 0; i < val.NumField(); i++ {
		typeField := typ.Field(i)
		jsonTag := typeField.Tag.Get(jsonTag)
		if sources, ok := config[jsonTag]; ok {
			for _, source := range sources {
				fieldValue := o.getValueFromSource(
					request,
					jsonTag,
					source,
					urlData,
					bodyData,
				)
				if fieldValue != nil && fieldValue != "" {
					valueMap[jsonTag] = fieldValue
				}
			}
		}
	}

	return valueMap
}

func (o *ObjectPicker[T]) getValueFromSource(
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
			panic(fmt.Sprintf("unknown input source: %s", source))
		}
	}

	return ""
}

func (o *ObjectPicker[T]) needsSource(config FieldConfig, source string) bool {
	needsBody := false
	for _, sources := range config {
		if slices.Contains(sources, source) {
			needsBody = true
			break
		}
	}
	return needsBody
}

func (o *ObjectPicker[T]) bodyToMap(r *http.Request) (BodyData, error) {
	body, err := o.getBody(r)
	if err != nil {
		return nil, err
	}
	if len(body) == 0 {
		return nil, nil
	}

	var m BodyData
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber() // for big integers
	if err = decoder.Decode(&m); err != nil {
		return nil, err
	}
	return m, nil
}

func (o *ObjectPicker[T]) getBody(request *http.Request) ([]byte, error) {
	body, err := io.ReadAll(request.Body)
	if err != nil {
		return nil, err
	}

	defer request.Body.Close()
	request.Body = io.NopCloser(bytes.NewBuffer(body))

	return body, nil
}
