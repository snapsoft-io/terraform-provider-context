// Copyright SnapSoft 2026
// SPDX-License-Identifier: MIT

package utils

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func mustStringMap(t *testing.T, m map[string]string) types.Map {
	t.Helper()
	elems := make(map[string]attr.Value, len(m))
	for k, v := range m {
		elems[k] = types.StringValue(v)
	}
	val, diags := types.MapValue(types.StringType, elems)
	if diags.HasError() {
		t.Fatalf("failed to build test map: %v", diags)
	}
	return val
}

func tfMapToGo(t *testing.T, m types.Map) map[string]string {
	t.Helper()
	res := make(map[string]string, len(m.Elements()))
	for k, v := range m.Elements() {
		s, ok := v.(types.String)
		if !ok {
			t.Fatalf("element %q is %T, want types.String", k, v)
		}
		res[k] = s.ValueString()
	}
	return res
}

func TestMergeTfMaps(t *testing.T) {
	t.Parallel()

	t.Run("second map wins on conflict", func(t *testing.T) {
		t.Parallel()
		m1 := mustStringMap(t, map[string]string{"Name": "first", "Only1": "a"})
		m2 := mustStringMap(t, map[string]string{"Name": "second", "Only2": "b"})

		got, err := MergeTfMaps(m1, m2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := map[string]string{"Name": "second", "Only1": "a", "Only2": "b"}
		assertStringMapEqual(t, tfMapToGo(t, got), want)
	})

	t.Run("null first map is treated as empty", func(t *testing.T) {
		t.Parallel()
		m2 := mustStringMap(t, map[string]string{"a": "1"})
		got, err := MergeTfMaps(types.MapNull(types.StringType), m2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertStringMapEqual(t, tfMapToGo(t, got), map[string]string{"a": "1"})
	})

	t.Run("null second map is treated as empty", func(t *testing.T) {
		t.Parallel()
		m1 := mustStringMap(t, map[string]string{"a": "1"})
		got, err := MergeTfMaps(m1, types.MapNull(types.StringType))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertStringMapEqual(t, tfMapToGo(t, got), map[string]string{"a": "1"})
	})

	t.Run("both null yields empty known map", func(t *testing.T) {
		t.Parallel()
		got, err := MergeTfMaps(types.MapNull(types.StringType), types.MapNull(types.StringType))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.IsNull() || got.IsUnknown() {
			t.Fatalf("expected known non-null map, got null=%v unknown=%v", got.IsNull(), got.IsUnknown())
		}
		if len(got.Elements()) != 0 {
			t.Errorf("expected empty map, got %v", got.Elements())
		}
	})

	t.Run("non-string element types produce an error", func(t *testing.T) {
		t.Parallel()
		bad, diags := types.MapValue(types.BoolType, map[string]attr.Value{"flag": types.BoolValue(true)})
		if diags.HasError() {
			t.Fatalf("failed to build bool map: %v", diags)
		}
		if _, err := MergeTfMaps(bad, types.MapNull(types.StringType)); err == nil {
			t.Error("expected error when merging a non-string typed map, got nil")
		}
	})
}

func TestConvertGoMapToTfMap(t *testing.T) {
	t.Parallel()

	t.Run("converts all entries", func(t *testing.T) {
		t.Parallel()
		in := map[string]string{"Name": "x", "Env": "prod", "Empty": ""}
		got, err := ConvertGoMapToTfMap(in)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !got.ElementType(t.Context()).Equal(types.StringType) {
			t.Errorf("element type = %v, want StringType", got.ElementType(t.Context()))
		}
		assertStringMapEqual(t, tfMapToGo(t, got), in)
	})

	t.Run("nil map yields empty known map", func(t *testing.T) {
		t.Parallel()
		got, err := ConvertGoMapToTfMap(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.IsNull() || len(got.Elements()) != 0 {
			t.Errorf("expected empty known map, got null=%v elems=%v", got.IsNull(), got.Elements())
		}
	})
}

func assertStringMapEqual(t *testing.T, got, want map[string]string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("map length = %d, want %d (got=%v want=%v)", len(got), len(want), got, want)
	}
	for k, wv := range want {
		gv, ok := got[k]
		if !ok {
			t.Errorf("missing key %q (got=%v)", k, got)
			continue
		}
		if gv != wv {
			t.Errorf("key %q = %q, want %q", k, gv, wv)
		}
	}
}
