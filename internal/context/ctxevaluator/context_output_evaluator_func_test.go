// Copyright SnapSoft 2026
// SPDX-License-Identifier: MIT

package ctxevaluator

import (
	"reflect"
	"strings"
	"testing"
)

func TestEvaluateContextOutput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		in       any
		wantId   string
		wantTags map[string]string
	}{
		{
			name:     "empty context yields empty outputs",
			in:       map[string]any{},
			wantId:   "",
			wantTags: nil,
		},
		{
			name:     "nil outputs yields empty outputs",
			in:       map[string]any{"outputs": nil},
			wantId:   "",
			wantTags: nil,
		},
		{
			name:     "outputs with only id",
			in:       map[string]any{"outputs": map[string]any{"id": "my-id"}},
			wantId:   "my-id",
			wantTags: nil,
		},
		{
			name:     "outputs with only tags",
			in:       map[string]any{"outputs": map[string]any{"tags": map[string]any{"Name": "n", "Env": "prod"}}},
			wantId:   "",
			wantTags: map[string]string{"Name": "n", "Env": "prod"},
		},
		{
			name:     "outputs with id and tags",
			in:       map[string]any{"outputs": map[string]any{"id": "x", "tags": map[string]any{"A": "1"}}},
			wantId:   "x",
			wantTags: map[string]string{"A": "1"},
		},
		{
			name:     "empty tags map yields empty non-nil tags",
			in:       map[string]any{"outputs": map[string]any{"tags": map[string]any{}}},
			wantId:   "",
			wantTags: map[string]string{},
		},
		{
			name:     "non-string id is ignored",
			in:       map[string]any{"outputs": map[string]any{"id": 42}},
			wantId:   "",
			wantTags: nil,
		},
		{
			name:     "non-map tags are ignored",
			in:       map[string]any{"outputs": map[string]any{"tags": "not-a-map"}},
			wantId:   "",
			wantTags: nil,
		},
		{
			name:     "other keys in context are ignored",
			in:       map[string]any{"meta": "x", "outputs": map[string]any{"id": "y", "extra": true}},
			wantId:   "y",
			wantTags: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := EvaluateContextOutput(tt.in)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got == nil {
				t.Fatal("expected non-nil outputs")
			}
			if got.Id != tt.wantId {
				t.Errorf("Id = %q, want %q", got.Id, tt.wantId)
			}
			if tt.wantTags == nil {
				if got.Tags != nil {
					t.Errorf("Tags = %v, want nil", got.Tags)
				}
			} else if !reflect.DeepEqual(got.Tags, tt.wantTags) {
				t.Errorf("Tags = %v, want %v", got.Tags, tt.wantTags)
			}
		})
	}
}

func TestEvaluateContextOutput_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		in          any
		wantErrPart string
	}{
		{"nil context", nil, "invalid context.main type"},
		{"string context", "oops", "invalid context.main type"},
		{"list context", []any{"a"}, "invalid context.main type"},
		{"outputs is not a map", map[string]any{"outputs": "oops"}, "invalid 'outputs' type"},
		{"outputs is a list", map[string]any{"outputs": []any{}}, "invalid 'outputs' type"},
		{"tag value is not a string", map[string]any{"outputs": map[string]any{"tags": map[string]any{"Count": 1}}}, "expected to be a map of strings"},
		{"tag value is nil", map[string]any{"outputs": map[string]any{"tags": map[string]any{"Nil": nil}}}, "expected to be a map of strings"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := EvaluateContextOutput(tt.in)
			if err == nil {
				t.Fatalf("expected error, got nil (result %+v)", got)
			}
			if got != nil {
				t.Errorf("expected nil outputs on error, got %+v", got)
			}
			if !strings.Contains(err.Error(), tt.wantErrPart) {
				t.Errorf("error %q does not contain %q", err.Error(), tt.wantErrPart)
			}
		})
	}
}
