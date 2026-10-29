// Copyright SnapSoft 2026
// SPDX-License-Identifier: MIT

package utils

import (
	"reflect"
	"testing"
)

func TestMergeMaps(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		maps []map[string]string
		want map[string]string
	}{
		{
			name: "no maps yields empty non-nil map",
			maps: nil,
			want: map[string]string{},
		},
		{
			name: "single map is copied",
			maps: []map[string]string{{"a": "1"}},
			want: map[string]string{"a": "1"},
		},
		{
			name: "disjoint keys are unioned",
			maps: []map[string]string{{"a": "1"}, {"b": "2"}},
			want: map[string]string{"a": "1", "b": "2"},
		},
		{
			name: "later maps override earlier ones",
			maps: []map[string]string{{"a": "1", "b": "1"}, {"b": "2"}, {"b": "3", "c": "3"}},
			want: map[string]string{"a": "1", "b": "3", "c": "3"},
		},
		{
			name: "nil maps are tolerated",
			maps: []map[string]string{nil, {"a": "1"}, nil},
			want: map[string]string{"a": "1"},
		},
		{
			name: "empty string values override",
			maps: []map[string]string{{"a": "1"}, {"a": ""}},
			want: map[string]string{"a": ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := MergeMaps(tt.maps...)
			if got == nil {
				t.Fatal("MergeMaps returned nil, want non-nil map")
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MergeMaps() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMergeMaps_DoesNotMutateInputs(t *testing.T) {
	t.Parallel()

	a := map[string]string{"k": "a"}
	b := map[string]string{"k": "b"}
	res := MergeMaps(a, b)
	res["k"] = "mutated"
	res["new"] = "x"

	if a["k"] != "a" || len(a) != 1 {
		t.Errorf("first input mutated: %v", a)
	}
	if b["k"] != "b" || len(b) != 1 {
		t.Errorf("second input mutated: %v", b)
	}
}

func TestToAnyMap(t *testing.T) {
	t.Parallel()

	t.Run("converts values preserving keys", func(t *testing.T) {
		t.Parallel()
		got := ToAnyMap(map[string]string{"a": "1", "b": ""})
		want := map[string]any{"a": "1", "b": ""}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("ToAnyMap() = %#v, want %#v", got, want)
		}
		if _, ok := got["a"].(string); !ok {
			t.Errorf("expected values to be plain strings, got %T", got["a"])
		}
	})

	t.Run("nil input yields empty non-nil map", func(t *testing.T) {
		t.Parallel()
		got := ToAnyMap(nil)
		if got == nil || len(got) != 0 {
			t.Errorf("ToAnyMap(nil) = %#v, want empty map", got)
		}
	})
}
