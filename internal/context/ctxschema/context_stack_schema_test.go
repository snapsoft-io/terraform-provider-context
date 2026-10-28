// Copyright SnapSoft 2026
// SPDX-License-Identifier: MIT

package ctxschema

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/snapsoft/terraform-provider-context/internal/context/ctxmodel"
)

func strMap(t *testing.T, m map[string]string) types.Map {
	t.Helper()
	elems := make(map[string]attr.Value, len(m))
	for k, v := range m {
		elems[k] = types.StringValue(v)
	}
	val, diags := types.MapValue(types.StringType, elems)
	if diags.HasError() {
		t.Fatalf("failed to build map: %v", diags)
	}
	return val
}

func nsElement(name string, vars types.Map, mappers *[]ContextMapperFunctionSchema, casing, prefix types.String, include types.Bool) ContextStackElementSchema {
	return ContextStackElementSchema{
		Name:                    types.StringValue(name),
		ContextType:             types.StringValue(ctxmodel.ContextTypeNamespace.String()),
		Vars:                    vars,
		Mappers:                 mappers,
		IdCasing:                casing,
		IdPrefix:                prefix,
		IncludeResourceTypeInId: include,
	}
}

func simpleNs(name string) ContextStackElementSchema {
	return nsElement(name, types.MapNull(types.StringType), nil, types.StringNull(), types.StringNull(), types.BoolNull())
}

func lbElement(name string, vars types.Map) ContextStackElementSchema {
	return ContextStackElementSchema{
		Name:                    types.StringValue(name),
		ContextType:             types.StringValue(ctxmodel.ContextTypeLabel.String()),
		Vars:                    vars,
		IdCasing:                types.StringNull(),
		IdPrefix:                types.StringNull(),
		IncludeResourceTypeInId: types.BoolNull(),
	}
}

func mapperSchema(name, cond, fn string) ContextMapperFunctionSchema {
	return ContextMapperFunctionSchema{
		Name:         types.StringValue(name),
		RunCondition: types.StringValue(cond),
		Function:     types.StringValue(fn),
	}
}

func TestNewEmptyContextModel(t *testing.T) {
	t.Parallel()

	m := NewEmptyContextModel()
	if m == nil {
		t.Fatal("expected non-nil model")
	}
	if m.Stack == nil {
		t.Fatal("expected non-nil (empty) stack so it serializes as [] rather than null")
	}
	if len(m.Stack) != 0 {
		t.Errorf("expected empty stack, got %d elements", len(m.Stack))
	}
}

func TestContextMapperFunctionSchema_ToModel(t *testing.T) {
	t.Parallel()

	s := mapperSchema("n", ".vars.x == \"1\"", ".context")
	got := s.ToModel()
	want := &ctxmodel.ContextMapperFunctionModel{Name: "n", RunCondition: ".vars.x == \"1\"", Function: ".context"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ToModel() = %+v, want %+v", got, want)
	}

	nullCond := ContextMapperFunctionSchema{Name: types.StringValue("n"), RunCondition: types.StringNull(), Function: types.StringValue(".context")}
	if got := nullCond.ToModel(); got.RunCondition != "" {
		t.Errorf("null run_condition should map to empty string, got %q", got.RunCondition)
	}
}

func TestContextStackSchema_Add(t *testing.T) {
	t.Parallel()

	var s ContextStackSchema
	mappers := &[]ContextMapperFunctionSchema{mapperSchema("m", "", ".context")}
	s.Add(types.StringValue("prod"), ctxmodel.ContextTypeNamespace, strMap(t, map[string]string{"a": "1"}), mappers, types.StringValue("snake_case"), types.StringValue("acme"), types.BoolValue(true))

	if len(s) != 1 {
		t.Fatalf("expected 1 element, got %d", len(s))
	}
	e := s[0]
	if e.Name.ValueString() != "prod" {
		t.Errorf("Name = %q", e.Name.ValueString())
	}
	if e.ContextType.ValueString() != "ns" {
		t.Errorf("ContextType = %q, want ns", e.ContextType.ValueString())
	}
	if e.Mappers != mappers {
		t.Error("Mappers pointer was not preserved")
	}
	if e.IdCasing.ValueString() != "snake_case" || e.IdPrefix.ValueString() != "acme" || !e.IncludeResourceTypeInId.ValueBool() {
		t.Errorf("id-shaping attrs not preserved: %+v", e)
	}

	s.Add(types.StringValue("second"), ctxmodel.ContextTypeNamespace, types.MapNull(types.StringType), nil, types.StringNull(), types.StringNull(), types.BoolNull())
	if len(s) != 2 || s[1].Name.ValueString() != "second" {
		t.Errorf("Add must append in order, got %+v", s)
	}
}

func TestContextStackSchema_AddWithType(t *testing.T) {
	t.Parallel()

	var s ContextStackSchema
	s.AddWithType(ctxmodel.ContextTypeLabel)

	if len(s) != 1 {
		t.Fatalf("expected 1 element, got %d", len(s))
	}
	e := s[0]
	if e.ContextType.ValueString() != "lb" {
		t.Errorf("ContextType = %q, want lb", e.ContextType.ValueString())
	}
	if !e.Name.IsNull() || !e.Vars.IsNull() || e.Mappers != nil || !e.IdCasing.IsNull() || !e.IdPrefix.IsNull() || !e.IncludeResourceTypeInId.IsNull() {
		t.Errorf("all other attributes must be null/nil, got %+v", e)
	}
}

func TestContextStackSchema_AddWithNameLabelVars(t *testing.T) {
	t.Parallel()

	var s ContextStackSchema
	vars := strMap(t, map[string]string{"resource_type": "vpc"})
	s.AddWithNameLabelVars(types.StringValue("main"), ctxmodel.ContextTypeLabel, vars)

	if len(s) != 1 {
		t.Fatalf("expected 1 element, got %d", len(s))
	}
	e := s[0]
	if e.Name.ValueString() != "main" || e.ContextType.ValueString() != "lb" {
		t.Errorf("unexpected element %+v", e)
	}
	if !e.Vars.Equal(vars) {
		t.Errorf("Vars = %v, want %v", e.Vars, vars)
	}
	if e.Mappers != nil || !e.IdCasing.IsNull() || !e.IdPrefix.IsNull() || !e.IncludeResourceTypeInId.IsNull() {
		t.Errorf("mappers/id attrs must be nil/null, got %+v", e)
	}
}

func TestContextStackSchema_ToAnyGoType(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	s := ContextStackSchema{
		nsElement("production", strMap(t, map[string]string{"environment": "prod", "owner": "team"}), nil, types.StringValue("PascalCase"), types.StringNull(), types.BoolNull()),
		simpleNs("#hidden"),
		lbElement("assets", strMap(t, map[string]string{"resource_type": "s3"})),
	}

	got := s.ToAnyGoType(ctx)
	want := []any{
		map[string]any{"name": "production", "type": "ns", "vars": map[string]any{"environment": "prod", "owner": "team"}},
		map[string]any{"name": "#hidden", "type": "ns", "vars": map[string]any{}},
		map[string]any{"name": "assets", "type": "lb", "vars": map[string]any{"resource_type": "s3"}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ToAnyGoType() =\n %#v\nwant\n %#v", got, want)
	}

	// Values must be gojq-compatible: plain strings inside map[string]any, never JSON-quoted.
	first, ok := got[0].(map[string]any)
	if !ok {
		t.Fatalf("got[0] = %#v, want map[string]any", got[0])
	}
	vars, ok := first["vars"].(map[string]any)
	if !ok {
		t.Fatalf("got[0].vars = %#v, want map[string]any", first["vars"])
	}
	if v, ok := vars["environment"].(string); !ok || v != "prod" {
		t.Errorf("vars.environment = %#v, want plain string \"prod\"", vars["environment"])
	}
}

func TestContextStackSchema_ToAnyGoType_Empty(t *testing.T) {
	t.Parallel()

	s := ContextStackSchema{}
	got := s.ToAnyGoType(t.Context())
	if got == nil || len(got) != 0 {
		t.Errorf("expected empty non-nil slice, got %#v", got)
	}
}

func TestContextStackSchema_IsLastElementInValidPlace(t *testing.T) {
	t.Parallel()

	ns := simpleNs("ns")
	lb := lbElement("lb", types.MapNull(types.StringType))
	bad := ContextStackElementSchema{Name: types.StringValue("x"), ContextType: types.StringValue("invalid")}

	tests := []struct {
		name  string
		stack ContextStackSchema
		want  bool
	}{
		{"empty stack is valid", ContextStackSchema{}, true},
		{"single namespace", ContextStackSchema{ns}, true},
		{"single label requires predecessor", ContextStackSchema{lb}, false},
		{"namespace after namespace", ContextStackSchema{ns, ns}, true},
		{"many namespaces", ContextStackSchema{ns, ns, ns, ns}, true},
		{"label after namespace", ContextStackSchema{ns, lb}, true},
		{"label after many namespaces", ContextStackSchema{ns, ns, ns, lb}, true},
		{"label after label", ContextStackSchema{ns, lb, lb}, false},
		{"namespace after label", ContextStackSchema{ns, lb, ns}, false},
		{"namespace after label deeper", ContextStackSchema{ns, lb, ns, ns}, false},
		{"invalid type as last element", ContextStackSchema{ns, bad}, false},
		{"invalid type as predecessor", ContextStackSchema{bad, ns}, false},
		{"invalid type single", ContextStackSchema{bad}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			stack := tt.stack
			if got := stack.IsLastElementInValidPlace(); got != tt.want {
				t.Errorf("IsLastElementInValidPlace() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContextStackSchema_GetStackVarsInTopDownOrder(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	t.Run("deeper elements override shallower ones", func(t *testing.T) {
		t.Parallel()
		s := ContextStackSchema{
			nsElement("root", strMap(t, map[string]string{"a": "root", "b": "root"}), nil, types.StringNull(), types.StringNull(), types.BoolNull()),
			simpleNs("no-vars"),
			nsElement("child", strMap(t, map[string]string{"b": "child", "c": "child"}), nil, types.StringNull(), types.StringNull(), types.BoolNull()),
			lbElement("leaf", strMap(t, map[string]string{"c": "leaf", "resource_type": "vpc"})),
		}
		got := s.GetStackVarsInTopDownOrder(ctx)
		want := map[string]string{"a": "root", "b": "child", "c": "leaf", "resource_type": "vpc"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("empty stack yields empty non-nil map", func(t *testing.T) {
		t.Parallel()
		s := ContextStackSchema{}
		got := s.GetStackVarsInTopDownOrder(ctx)
		if got == nil || len(got) != 0 {
			t.Errorf("got %#v, want empty map", got)
		}
	})

	t.Run("all null vars yields empty map", func(t *testing.T) {
		t.Parallel()
		s := ContextStackSchema{simpleNs("a"), simpleNs("b")}
		if got := s.GetStackVarsInTopDownOrder(ctx); len(got) != 0 {
			t.Errorf("got %v, want empty", got)
		}
	})
}

func TestContextStackSchema_GetStackMappersInBottomUpOrder(t *testing.T) {
	t.Parallel()

	t.Run("leaf mappers come first, root mappers last, order within element preserved", func(t *testing.T) {
		t.Parallel()
		rootMappers := &[]ContextMapperFunctionSchema{mapperSchema("root-1", "", "."), mapperSchema("root-2", "", ".")}
		childMappers := &[]ContextMapperFunctionSchema{mapperSchema("child-1", "", "."), mapperSchema("child-2", "", ".")}
		s := ContextStackSchema{
			nsElement("root", types.MapNull(types.StringType), rootMappers, types.StringNull(), types.StringNull(), types.BoolNull()),
			simpleNs("middle"),
			nsElement("child", types.MapNull(types.StringType), childMappers, types.StringNull(), types.StringNull(), types.BoolNull()),
		}

		got := s.GetStackMappersInBottomUpOrder()
		if got == nil {
			t.Fatal("expected non-nil result")
		}
		names := make([]string, 0, len(*got))
		for _, m := range *got {
			names = append(names, m.Name)
		}
		want := []string{"child-1", "child-2", "root-1", "root-2"}
		if !reflect.DeepEqual(names, want) {
			t.Errorf("mapper order = %v, want %v", names, want)
		}
	})

	t.Run("no mappers yields empty non-nil list", func(t *testing.T) {
		t.Parallel()
		s := ContextStackSchema{simpleNs("a"), simpleNs("b")}
		got := s.GetStackMappersInBottomUpOrder()
		if got == nil || len(*got) != 0 {
			t.Errorf("got %#v, want empty list", got)
		}
	})

	t.Run("empty mapper list on element is tolerated", func(t *testing.T) {
		t.Parallel()
		s := ContextStackSchema{
			nsElement("a", types.MapNull(types.StringType), &[]ContextMapperFunctionSchema{}, types.StringNull(), types.StringNull(), types.BoolNull()),
			nsElement("b", types.MapNull(types.StringType), &[]ContextMapperFunctionSchema{mapperSchema("only", "", ".")}, types.StringNull(), types.StringNull(), types.BoolNull()),
		}
		got := s.GetStackMappersInBottomUpOrder()
		if len(*got) != 1 || (*got)[0].Name != "only" {
			t.Errorf("got %+v, want single 'only' mapper", *got)
		}
	})

	t.Run("converts schema fields to model", func(t *testing.T) {
		t.Parallel()
		s := ContextStackSchema{
			nsElement("a", types.MapNull(types.StringType), &[]ContextMapperFunctionSchema{mapperSchema("m", ".c", ".f")}, types.StringNull(), types.StringNull(), types.BoolNull()),
		}
		got := (*s.GetStackMappersInBottomUpOrder())[0]
		want := ctxmodel.ContextMapperFunctionModel{Name: "m", RunCondition: ".c", Function: ".f"}
		if got != want {
			t.Errorf("got %+v, want %+v", got, want)
		}
	})
}

func TestContextStackSchema_GetNamespaceNames(t *testing.T) {
	t.Parallel()

	s := ContextStackSchema{
		simpleNs("production"),
		simpleNs("#hidden"),
		simpleNs(""),
		lbElement("assets", types.MapNull(types.StringType)),
	}
	got := s.GetNamespaceNames()
	want := []string{"production", "#hidden", ""}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("GetNamespaceNames() = %v, want %v", got, want)
	}

	empty := ContextStackSchema{}
	if got := empty.GetNamespaceNames(); got == nil || len(got) != 0 {
		t.Errorf("empty stack: got %#v, want empty non-nil slice", got)
	}
}

func TestContextStackSchema_GetEffectiveIdCasing(t *testing.T) {
	t.Parallel()

	t.Run("not set anywhere", func(t *testing.T) {
		t.Parallel()
		s := ContextStackSchema{simpleNs("a"), simpleNs("b")}
		if v, ok := s.GetEffectiveIdCasing(); ok || v != "" {
			t.Errorf("got (%q,%v), want (\"\",false)", v, ok)
		}
	})

	t.Run("last namespace with value wins", func(t *testing.T) {
		t.Parallel()
		s := ContextStackSchema{
			nsElement("a", types.MapNull(types.StringType), nil, types.StringValue("snake_case"), types.StringNull(), types.BoolNull()),
			nsElement("b", types.MapNull(types.StringType), nil, types.StringValue("PascalCase"), types.StringNull(), types.BoolNull()),
			simpleNs("c"),
		}
		if v, ok := s.GetEffectiveIdCasing(); !ok || v != "PascalCase" {
			t.Errorf("got (%q,%v), want (\"PascalCase\",true)", v, ok)
		}
	})

	t.Run("label element values are ignored", func(t *testing.T) {
		t.Parallel()
		lb := lbElement("lb", types.MapNull(types.StringType))
		lb.IdCasing = types.StringValue("camelCase")
		s := ContextStackSchema{
			nsElement("a", types.MapNull(types.StringType), nil, types.StringValue("snake_case"), types.StringNull(), types.BoolNull()),
			lb,
		}
		if v, ok := s.GetEffectiveIdCasing(); !ok || v != "snake_case" {
			t.Errorf("got (%q,%v), want (\"snake_case\",true)", v, ok)
		}
	})
}

func TestContextStackSchema_GetEffectiveIdPrefix(t *testing.T) {
	t.Parallel()

	t.Run("not set anywhere", func(t *testing.T) {
		t.Parallel()
		s := ContextStackSchema{simpleNs("a")}
		if v, ok := s.GetEffectiveIdPrefix(); ok || v != "" {
			t.Errorf("got (%q,%v), want (\"\",false)", v, ok)
		}
	})

	t.Run("last namespace with value wins including empty string", func(t *testing.T) {
		t.Parallel()
		s := ContextStackSchema{
			nsElement("a", types.MapNull(types.StringType), nil, types.StringNull(), types.StringValue("acme"), types.BoolNull()),
			nsElement("b", types.MapNull(types.StringType), nil, types.StringNull(), types.StringValue(""), types.BoolNull()),
		}
		// An explicitly empty prefix on a deeper namespace must be able to clear an inherited one.
		if v, ok := s.GetEffectiveIdPrefix(); !ok || v != "" {
			t.Errorf("got (%q,%v), want (\"\",true)", v, ok)
		}
	})

	t.Run("single value", func(t *testing.T) {
		t.Parallel()
		s := ContextStackSchema{
			simpleNs("root"),
			nsElement("b", types.MapNull(types.StringType), nil, types.StringNull(), types.StringValue("snap"), types.BoolNull()),
		}
		if v, ok := s.GetEffectiveIdPrefix(); !ok || v != "snap" {
			t.Errorf("got (%q,%v), want (\"snap\",true)", v, ok)
		}
	})
}

func TestContextStackSchema_GetEffectiveIncludeResourceTypeInId(t *testing.T) {
	t.Parallel()

	t.Run("not set anywhere", func(t *testing.T) {
		t.Parallel()
		s := ContextStackSchema{simpleNs("a")}
		if v, ok := s.GetEffectiveIncludeResourceTypeInId(); ok || v {
			t.Errorf("got (%v,%v), want (false,false)", v, ok)
		}
	})

	t.Run("deeper explicit false overrides shallower true", func(t *testing.T) {
		t.Parallel()
		s := ContextStackSchema{
			nsElement("a", types.MapNull(types.StringType), nil, types.StringNull(), types.StringNull(), types.BoolValue(true)),
			nsElement("b", types.MapNull(types.StringType), nil, types.StringNull(), types.StringNull(), types.BoolValue(false)),
		}
		if v, ok := s.GetEffectiveIncludeResourceTypeInId(); !ok || v {
			t.Errorf("got (%v,%v), want (false,true)", v, ok)
		}
	})

	t.Run("true is found", func(t *testing.T) {
		t.Parallel()
		s := ContextStackSchema{
			nsElement("a", types.MapNull(types.StringType), nil, types.StringNull(), types.StringNull(), types.BoolValue(true)),
			simpleNs("b"),
		}
		if v, ok := s.GetEffectiveIncludeResourceTypeInId(); !ok || !v {
			t.Errorf("got (%v,%v), want (true,true)", v, ok)
		}
	})
}
