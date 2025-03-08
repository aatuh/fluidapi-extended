package types

import "fmt"

// APIField holds the core mapping information.
type APIField struct {
	APIName  string
	Alias    string
	DBColumn string
	Required bool
	Default  any
	Source   string
	Validate []string
	Nested   APIFields
	Type     string
}

// SetRequired returns a copy of APIField with Required set.
func (field APIField) SetRequired(required bool) APIField {
	field.Required = required
	return field
}

// Nested returns a copy of APIField with Nested set.
func (field APIField) SetNested(nested APIFields) APIField {
	field.Nested = nested
	return field
}

// SetAlias returns a copy of APIField with Alias set.
func (field APIField) SetAlias(alias string) APIField {
	field.Alias = alias
	return field
}

// SetAPIName returns a copy of APIField with APIName set.
func (field APIField) SetAPIName(apiName string) APIField {
	field.APIName = apiName
	return field
}

// APIFields is a slice of APIField.
type APIFields []APIField

// With returns a copy of APIFields with fields appended.
func (a APIFields) With(fields ...APIField) APIFields {
	return append(a, fields...)
}

// MustGetAPIField looks up a master definition by its API name.
func (a APIFields) MustGetAPIField(field string) APIField {
	def, ok := a.GetAPIField(field)
	if !ok {
		panic(fmt.Errorf("MustGetAPIField: unknown API field %q", field))
	}
	return def
}

// MustGetAPIFields looks up master definitions by their API names.
func (a APIFields) MustGetAPIFields(fields []string) APIFields {
	var defs APIFields
	for _, field := range fields {
		def, ok := a.GetAPIField(field)
		if !ok {
			panic(fmt.Errorf("mustGetAPIFields: unknown API field %q", field))
		}
		defs = append(defs, def)
	}
	return defs
}

// GetAPIField looks up a field definition by its API name.
func (a APIFields) GetAPIField(field string) (APIField, bool) {
	for _, def := range a {
		if def.APIName == field {
			return def, true
		}
	}
	return APIField{}, false
}
