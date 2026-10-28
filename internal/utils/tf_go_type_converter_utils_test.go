// Copyright SnapSoft 2026
// SPDX-License-Identifier: MIT

package utils

import (
	"math/big"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestConvertTfAttrValueToGoType(t *testing.T) {
	t.Parallel()

	strList, _ := types.ListValue(types.StringType, []attr.Value{types.StringValue("a"), types.StringValue("b")})
	strSet, _ := types.SetValue(types.StringType, []attr.Value{types.StringValue("x")})
	tuple, _ := types.TupleValue([]attr.Type{types.StringType, types.BoolType}, []attr.Value{types.StringValue("s"), types.BoolValue(true)})
	strMap, _ := types.MapValue(types.StringType, map[string]attr.Value{"k": types.StringValue("v")})
	obj, _ := types.ObjectValue(
		map[string]attr.Type{"name": types.StringType, "enabled": types.BoolType, "list": types.ListType{ElemType: types.StringType}},
		map[string]attr.Value{"name": types.StringValue("n"), "enabled": types.BoolValue(false), "list": strList},
	)

	tests := []struct {
		name string
		in   attr.Value
		want any
	}{
		{"null string is nil", types.StringNull(), nil},
		{"unknown string is nil", types.StringUnknown(), nil},
		{"null map is nil", types.MapNull(types.StringType), nil},
		{"bool", types.BoolValue(true), true},
		{"number", types.NumberValue(big.NewFloat(42.5)), 42.5},
		{"string", types.StringValue("hello"), "hello"},
		{"list of strings", strList, []any{"a", "b"}},
		{"set of strings", strSet, []any{"x"}},
		{"tuple", tuple, []any{"s", true}},
		{"map of strings", strMap, map[string]any{"k": "v"}},
		{"nested object", obj, map[string]any{"name": "n", "enabled": false, "list": []any{"a", "b"}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ConvertTfAttrValueToGoType(tt.in)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestConvertTfAttrValueToGoType_UnsupportedType(t *testing.T) {
	t.Parallel()

	_, err := ConvertTfAttrValueToGoType(types.DynamicValue(types.StringValue("x")))
	if err == nil {
		t.Fatal("expected error for unsupported dynamic type, got nil")
	}
}

func TestConvertGoTypeToTfAttrValue(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	t.Run("nil becomes null string", func(t *testing.T) {
		t.Parallel()
		got, err := ConvertGoTypeToTfAttrValue(ctx, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !got.IsNull() || !got.Type(ctx).Equal(types.StringType) {
			t.Errorf("got %v (type %v), want null string", got, got.Type(ctx))
		}
	})

	t.Run("scalars", func(t *testing.T) {
		t.Parallel()
		cases := []struct {
			in   any
			want attr.Value
		}{
			{true, types.BoolValue(true)},
			{"str", types.StringValue("str")},
			{1.5, types.NumberValue(big.NewFloat(1.5))},
		}
		for _, c := range cases {
			got, err := ConvertGoTypeToTfAttrValue(ctx, c.in)
			if err != nil {
				t.Fatalf("unexpected error for %v: %v", c.in, err)
			}
			if !got.Equal(c.want) {
				t.Errorf("ConvertGoTypeToTfAttrValue(%v) = %v, want %v", c.in, got, c.want)
			}
		}
	})

	t.Run("slice becomes tuple", func(t *testing.T) {
		t.Parallel()
		got, err := ConvertGoTypeToTfAttrValue(ctx, []any{"a", true})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		tuple, ok := got.(types.Tuple)
		if !ok {
			t.Fatalf("got %T, want types.Tuple", got)
		}
		elems := tuple.Elements()
		if len(elems) != 2 || !elems[0].Equal(types.StringValue("a")) || !elems[1].Equal(types.BoolValue(true)) {
			t.Errorf("unexpected tuple elements: %v", elems)
		}
	})

	t.Run("map becomes object", func(t *testing.T) {
		t.Parallel()
		got, err := ConvertGoTypeToTfAttrValue(ctx, map[string]any{"name": "n", "nested": map[string]any{"ok": true}})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		obj, ok := got.(types.Object)
		if !ok {
			t.Fatalf("got %T, want types.Object", got)
		}
		attrs := obj.Attributes()
		if !attrs["name"].Equal(types.StringValue("n")) {
			t.Errorf("name = %v, want \"n\"", attrs["name"])
		}
		nested, ok := attrs["nested"].(types.Object)
		if !ok {
			t.Fatalf("nested is %T, want types.Object", attrs["nested"])
		}
		if !nested.Attributes()["ok"].Equal(types.BoolValue(true)) {
			t.Errorf("nested.ok = %v, want true", nested.Attributes()["ok"])
		}
	})

	t.Run("unsupported go type errors", func(t *testing.T) {
		t.Parallel()
		if _, err := ConvertGoTypeToTfAttrValue(ctx, 42); err == nil {
			t.Error("expected error for int, got nil")
		}
		if _, err := ConvertGoTypeToTfAttrValue(ctx, []any{struct{}{}}); err == nil {
			t.Error("expected error for slice with unsupported element, got nil")
		}
		if _, err := ConvertGoTypeToTfAttrValue(ctx, map[string]any{"bad": struct{}{}}); err == nil {
			t.Error("expected error for map with unsupported value, got nil")
		}
	})
}

func TestConvertRoundTrip(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	original := map[string]any{
		"name":    "vpc",
		"enabled": true,
		"count":   2.0,
		"tags":    map[string]any{"Env": "prod"},
		"list":    []any{"a", "b"},
	}

	tfValue, err := ConvertGoTypeToTfAttrValue(ctx, original)
	if err != nil {
		t.Fatalf("to tf: %v", err)
	}
	back, err := ConvertTfAttrValueToGoType(tfValue)
	if err != nil {
		t.Fatalf("to go: %v", err)
	}
	if !reflect.DeepEqual(back, original) {
		t.Errorf("round trip mismatch:\n got  %#v\n want %#v", back, original)
	}
}
