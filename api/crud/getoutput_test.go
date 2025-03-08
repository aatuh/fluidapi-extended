package crud

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/pakkasys/fluidapi-extended/api/types"
)

// expectPanic is a helper to test for panics.
func expectPanic(t *testing.T, f func(), expected string) {
	t.Helper()
	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("expected panic but got none")
		} else {
			msg := fmt.Sprint(r)
			if !strings.Contains(msg, expected) {
				t.Errorf("expected panic message to contain %q, got %q",
					expected, msg)
			}
		}
	}()
	f()
}

func TestStructToAPIFields(t *testing.T) {
	// Non-struct type.
	t.Run("NonStructType", func(t *testing.T) {
		expectPanic(t, func() {
			structToAPIFields(reflect.TypeOf(123), types.APIFields{}, 0)
		}, "is not a struct")
	})

	// Interface type.
	t.Run("InterfaceType", func(t *testing.T) {
		iface := reflect.TypeOf((*error)(nil)).Elem()
		expectPanic(t, func() {
			structToAPIFields(iface, types.APIFields{}, 0)
		}, "type is a null interface")
	})

	// Duplicate alias in allFields.
	t.Run("DuplicateAlias", func(t *testing.T) {
		type Dummy struct {
			Field string `json:"f1"`
		}
		all := types.APIFields{
			{APIName: "f1", Alias: "dup"},
			{APIName: "f2", Alias: "dup"},
		}
		expectPanic(t, func() {
			structToAPIFields(reflect.TypeOf(Dummy{}), all, 0)
		}, "duplicate alias")
	})

	// Duplicate API name in allFields.
	t.Run("DuplicateAPIName", func(t *testing.T) {
		type Dummy struct {
			Field string `json:"f1"`
		}
		all := types.APIFields{
			{APIName: "dup"},
			{APIName: "dup"},
		}
		expectPanic(t, func() {
			structToAPIFields(reflect.TypeOf(Dummy{}), all, 0)
		}, "duplicate API name")
	})

	// Alias field not found.
	t.Run("AliasNotFound", func(t *testing.T) {
		type Dummy struct {
			Field string `alias:"notfound"`
		}
		expectPanic(t, func() {
			structToAPIFields(reflect.TypeOf(Dummy{}), types.APIFields{}, 0)
		}, "alias field notfound")
	})

	// Simple field with json tag.
	t.Run("SimpleFieldDefault", func(t *testing.T) {
		type Simple struct {
			Field1 string `json:"field1" validate:"v1,v2" required:"true"`
		}
		out := structToAPIFields(reflect.TypeOf(Simple{}), types.APIFields{}, 0)
		if len(out) != 1 {
			t.Fatalf("expected 1 field, got %d", len(out))
		}
		f := out[0]
		if f.APIName != "field1" {
			t.Errorf("expected APIName 'field1', got %q", f.APIName)
		}
		if !f.Required {
			t.Errorf("expected Required true")
		}
		if f.Type != "string" {
			t.Errorf("expected Type 'string', got %q", f.Type)
		}
		if len(f.Validate) != 2 || f.Validate[0] != "v1" ||
			f.Validate[1] != "v2" {
			t.Errorf("unexpected Validate rules: %v", f.Validate)
		}
	})

	// Override validate rules from allFields.
	t.Run("FieldOverrideValidate", func(t *testing.T) {
		type Override struct {
			Field1 int `json:"field1" validate:"v3,v4"`
		}
		all := types.APIFields{
			{APIName: "field1", Validate: []string{"old"}},
		}
		out := structToAPIFields(reflect.TypeOf(Override{}), all, 0)
		if len(out) != 1 {
			t.Fatalf("expected 1 field, got %d", len(out))
		}
		if len(out[0].Validate) != 2 || out[0].Validate[0] != "v3" ||
			out[0].Validate[1] != "v4" {
			t.Errorf("expected validate override, got %v", out[0].Validate)
		}
		if out[0].Type != "int" {
			t.Errorf("expected Type 'int', got %q", out[0].Type)
		}
	})

	// Field with type override.
	t.Run("FieldWithTypeOverride", func(t *testing.T) {
		type CustomType struct {
			Field1 int `json:"field1" type:"custom"`
		}
		out := structToAPIFields(reflect.TypeOf(CustomType{}), types.APIFields{}, 0)
		if out[0].Type != "custom" {
			t.Errorf("expected Type 'custom', got %q", out[0].Type)
		}
	})

	// Invalid required tag.
	t.Run("InvalidRequiredTag", func(t *testing.T) {
		type BadRequired struct {
			Field1 string `json:"field1" required:"notabool"`
		}
		expectPanic(t, func() {
			structToAPIFields(reflect.TypeOf(BadRequired{}), types.APIFields{}, 0)
		}, "invalid value for required tag")
	})

	// Test 10: Nested struct.
	t.Run("NestedStruct", func(t *testing.T) {
		type Nested struct {
			Inner string `json:"inner"`
		}
		type Outer struct {
			Nested Nested `json:"nested"`
		}
		out := structToAPIFields(reflect.TypeOf(Outer{}), types.APIFields{}, 0)
		if len(out) != 1 {
			t.Fatalf("expected 1 field, got %d", len(out))
		}
		nested := out[0].Nested
		if len(nested) != 1 || nested[0].APIName != "inner" {
			t.Errorf("expected nested field 'inner', got %v", nested)
		}
	})

	// Test 11: Slice of struct.
	t.Run("SliceOfStruct", func(t *testing.T) {
		type Child struct {
			Value int `json:"value"`
		}
		type Container struct {
			Children []Child `json:"children"`
		}
		out := structToAPIFields(reflect.TypeOf(Container{}), types.APIFields{}, 0)
		if len(out) != 1 {
			t.Fatalf("expected 1 field, got %d", len(out))
		}
		if len(out[0].Nested) != 1 ||
			out[0].Nested[0].APIName != "value" {
			t.Errorf("expected nested field 'value', got %v",
				out[0].Nested)
		}
	})

	// Test 12: Skip field without json tag.
	t.Run("SkipNoJSONTag", func(t *testing.T) {
		type NoJSON struct {
			Ignored string
			Field   int `json:"field"`
		}
		out := structToAPIFields(reflect.TypeOf(NoJSON{}), types.APIFields{}, 0)
		if len(out) != 1 {
			t.Errorf("expected 1 field, got %d", len(out))
		}
		if out[0].APIName != "field" {
			t.Errorf("expected APIName 'field', got %q", out[0].APIName)
		}
	})

	// Test 13: External field missing in allFields.
	t.Run("ExtTagMissingField", func(t *testing.T) {
		type ExtStruct struct {
			Field string `json:"field" ext:"true"`
		}
		expectPanic(t, func() {
			structToAPIFields(reflect.TypeOf(ExtStruct{}), types.APIFields{}, 0)
		}, "APIField not found for external field")
	})

	// Test 14: Alias usage.
	t.Run("AliasUsage", func(t *testing.T) {
		type AliasStruct struct {
			Field string `alias:"a_field" validate:"a_rule"`
		}
		all := types.APIFields{
			{APIName: "ignored", Alias: "a_field", Validate: []string{"old_rule"}},
		}
		out := structToAPIFields(reflect.TypeOf(AliasStruct{}), all, 0)
		if len(out) != 1 {
			t.Fatalf("expected 1 field, got %d", len(out))
		}
		if out[0].APIName != "ignored" {
			t.Errorf("expected APIName 'ignored', got %q", out[0].APIName)
		}
		if len(out[0].Validate) != 1 ||
			out[0].Validate[0] != "a_rule" {
			t.Errorf("expected validate override, got %v", out[0].Validate)
		}
	})

	// Test 15: Pointer field should yield underlying type.
	t.Run("PointerField", func(t *testing.T) {
		type PtrStruct struct {
			Field *int `json:"field"`
		}
		out := structToAPIFields(reflect.TypeOf(PtrStruct{}), types.APIFields{}, 0)
		if len(out) != 1 {
			t.Fatalf("expected 1 field, got %d", len(out))
		}
		expected := "int"
		if out[0].Type != expected {
			t.Errorf("expected Type %q, got %q", expected, out[0].Type)
		}
	})
}

// Offensive test suite: trying to break the code with unexpected inputs.
func TestStructToAPIFields_Offensive(t *testing.T) {
	// Empty JSON tag (should skip the field)
	t.Run("EmptyJSONTag", func(t *testing.T) {
		type EmptyJSON struct {
			Field string `json:""`
		}
		out := structToAPIFields(reflect.TypeOf(EmptyJSON{}), types.APIFields{}, 0)
		if len(out) != 0 {
			t.Errorf("expected 0 fields (empty json tag skips field), got %d", len(out))
		}
	})

	// JSON tag with "-" (dash) should create a field with APIName "-"
	t.Run("DashJSONTag", func(t *testing.T) {
		type DashField struct {
			Field string `json:"-"`
		}
		out := structToAPIFields(reflect.TypeOf(DashField{}), types.APIFields{}, 0)
		if len(out) != 1 {
			t.Fatalf("expected 1 field, got %d", len(out))
		}
		if out[0].APIName != "-" {
			t.Errorf("expected APIName '-' but got %q", out[0].APIName)
		}
		if out[0].Type != "string" {
			t.Errorf("expected Type 'string' but got %q", out[0].Type)
		}
	})

	// Unexported field with JSON tag should be processed.
	t.Run("UnexportedField", func(t *testing.T) {
		type unexported struct {
			field string `json:"field"`
		}
		out := structToAPIFields(reflect.TypeOf(unexported{}), types.APIFields{}, 0)
		if len(out) != 1 {
			t.Fatalf("expected 1 field, got %d", len(out))
		}
		if out[0].APIName != "field" {
			t.Errorf("expected APIName 'field' but got %q", out[0].APIName)
		}
	})

	// Anonymous (embedded) struct field with a tag.
	t.Run("AnonymousEmbeddedField", func(t *testing.T) {
		type embedded struct {
			Inner string `json:"inner"`
		}
		type Outer struct {
			embedded `json:"embedded"`
		}
		out := structToAPIFields(reflect.TypeOf(Outer{}), types.APIFields{}, 0)
		if len(out) != 1 {
			t.Fatalf("expected 1 field, got %d", len(out))
		}
		// The embedded field becomes a nested struct.
		if out[0].APIName != "embedded" {
			t.Errorf("expected APIName 'embedded', got %q", out[0].APIName)
		}
		if len(out[0].Nested) != 1 || out[0].Nested[0].APIName != "inner" {
			t.Errorf("expected nested field 'inner', got %+v", out[0].Nested)
		}
	})

	// Self-referential struct should eventually cause infinite recursion.
	t.Run("SelfReferentialStruct", func(t *testing.T) {
		type Node struct {
			Next *Node `json:"next"`
		}
		done := make(chan struct{})
		go func() {
			defer func() {
				// If it panics, we mark done (even if it's a stack overflow,
				// it may crash the goroutine).
				recover()
				done <- struct{}{}
			}()
			_ = structToAPIFields(reflect.TypeOf(Node{}), types.APIFields{}, 0)
			done <- struct{}{}
		}()
		select {
		case <-done:
			// Finished (or panicked) before timeout. This indicates the function
			// does not protect against cycles.
		case <-time.After(100 * time.Millisecond):
			t.Error("self-referential struct caused infinite recursion or hang")
		}
	})

	// Field with double pointer type (**int)
	t.Run("DoublePointerField", func(t *testing.T) {
		type DoublePtr struct {
			Field **int `json:"field"`
		}
		out := structToAPIFields(reflect.TypeOf(DoublePtr{}), types.APIFields{}, 0)
		if len(out) != 1 {
			t.Fatalf("expected 1 field, got %d", len(out))
		}
		// The function dereferences one level so **int becomes *int.
		expectedType := "*int"
		if out[0].Type != expectedType {
			t.Errorf("expected Type %q, got %q", expectedType, out[0].Type)
		}
	})

	// Field with malformed validate tag (e.g. extra commas)
	t.Run("MalformedValidateTag", func(t *testing.T) {
		type MalformedValidate struct {
			Field string `json:"field" validate:",,"`
		}
		out := structToAPIFields(reflect.TypeOf(MalformedValidate{}), types.APIFields{}, 0)
		if len(out) != 1 {
			t.Fatalf("expected 1 field, got %d", len(out))
		}
		// Expect three empty strings in the Validate slice.
		if len(out[0].Validate) != 3 {
			t.Errorf("expected 3 validate rules (even if empty), got %v", out[0].Validate)
		}
	})

	// Field with extra whitespace in tags.
	t.Run("WhitespaceInTags", func(t *testing.T) {
		type WhitespaceTags struct {
			Field string `json:"  field  " validate:"  v1  ,  v2  " required:" true "`
		}
		out := structToAPIFields(reflect.TypeOf(WhitespaceTags{}), types.APIFields{}, 0)
		if len(out) != 1 {
			t.Fatalf("expected 1 field, got %d", len(out))
		}
		// Expect the whitespace to be preserved because there is no trim.
		if out[0].APIName != "  field  " {
			t.Errorf("expected APIName with whitespace, got %q", out[0].APIName)
		}
		if len(out[0].Validate) != 2 ||
			out[0].Validate[0] != "  v1  " || out[0].Validate[1] != "  v2  " {
			t.Errorf("unexpected validate rules: %v", out[0].Validate)
		}
		if !out[0].Required {
			t.Errorf("expected Required true, got false")
		}
	})

	// Field with ext tag set to "false" (should not trigger panic).
	t.Run("ExtTagFalse", func(t *testing.T) {
		type ExtFalse struct {
			Field string `json:"field" ext:"false"`
		}
		// No APIField provided in allFields, but ext is not "true" so it should work.
		out := structToAPIFields(reflect.TypeOf(ExtFalse{}), types.APIFields{}, 0)
		if len(out) != 1 {
			t.Fatalf("expected 1 field, got %d", len(out))
		}
		if out[0].APIName != "field" {
			t.Errorf("expected APIName 'field', got %q", out[0].APIName)
		}
	})

	// Field of interface type should be processed normally.
	t.Run("InterfaceField", func(t *testing.T) {
		type InterfaceField struct {
			Field interface{} `json:"field"`
		}
		out := structToAPIFields(reflect.TypeOf(InterfaceField{}), types.APIFields{}, 0)
		if len(out) != 1 {
			t.Fatalf("expected 1 field, got %d", len(out))
		}
		// reflect.TypeOf(interface{}) prints "interface {}"
		expectedType := "interface {}"
		if out[0].Type != expectedType {
			t.Errorf("expected Type %q, got %q", expectedType, out[0].Type)
		}
	})

	// Duplicate alias test with case sensitivity: "dup" vs "DUP" should be allowed.
	t.Run("DuplicateAliasCaseSensitive", func(t *testing.T) {
		type DupAlias struct {
			Field1 string `alias:"dup"`
			Field2 string `alias:"DUP"`
		}
		all := types.APIFields{
			{APIName: "f1", Alias: "dup"},
			{APIName: "f2", Alias: "DUP"},
		}
		out := structToAPIFields(reflect.TypeOf(DupAlias{}), all, 0)
		if len(out) != 2 {
			t.Fatalf("expected 2 fields, got %d", len(out))
		}
	})

	// Duplicate APIName with case sensitivity.
	t.Run("DuplicateAPINameCaseSensitive", func(t *testing.T) {
		type DupAPIName struct {
			Field1 string `json:"dup"`
			Field2 string `json:"DUP"`
		}
		all := types.APIFields{
			{APIName: "dup"},
			{APIName: "DUP"},
		}
		out := structToAPIFields(reflect.TypeOf(DupAPIName{}), all, 0)
		if len(out) != 2 {
			t.Fatalf("expected 2 fields, got %d", len(out))
		}
	})

	// JSON tag with multiple options.
	t.Run("JSONTagMultipleOptions", func(t *testing.T) {
		type MultiOption struct {
			Field string `json:"field,omitempty,someoption" validate:"v1,v2"`
		}
		out := structToAPIFields(reflect.TypeOf(MultiOption{}), types.APIFields{}, 0)
		if len(out) != 1 {
			t.Fatalf("expected 1 field, got %d", len(out))
		}
		if out[0].APIName != "field" {
			t.Errorf("expected APIName 'field', got %q", out[0].APIName)
		}
		if len(out[0].Validate) != 2 ||
			out[0].Validate[0] != "v1" || out[0].Validate[1] != "v2" {
			t.Errorf("unexpected validate rules: %v", out[0].Validate)
		}
	})

	// Nested slice of slices should not recursively process inner slice elements.
	t.Run("NestedSliceOfSlices", func(t *testing.T) {
		type Inner struct {
			Val int `json:"val"`
		}
		type Outer struct {
			Matrix [][]Inner `json:"matrix"`
		}
		out := structToAPIFields(reflect.TypeOf(Outer{}), types.APIFields{}, 0)
		if len(out) != 1 {
			t.Fatalf("expected 1 field, got %d", len(out))
		}
		// Since Matrix is [][]Inner, only one level of slice is processed.
		if out[0].Nested != nil {
			t.Errorf("expected no nested fields for a slice of slices, got %+v", out[0].Nested)
		}
	})
}
