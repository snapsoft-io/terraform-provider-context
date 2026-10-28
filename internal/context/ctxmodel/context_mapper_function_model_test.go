// Copyright SnapSoft 2026
// SPDX-License-Identifier: MIT

package ctxmodel

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTempJson(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "mappers.json")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	return path
}

func TestNewContextMapperFunctionModelListFromJson_Valid(t *testing.T) {
	t.Parallel()

	path := writeTempJson(t, `[
		{
			"name": "Enrich tags",
			"function": "(.context.outputs.tags += {Team: .vars.team}).context"
		},
		{
			"name": "Conditional",
			"run_condition": ".vars.environment == \"sbx\"",
			"function": ".context"
		}
	]`)

	mappers, err := NewContextMapperFunctionModelListFromJson(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mappers == nil {
		t.Fatal("expected non-nil mapper list")
	}
	if len(*mappers) != 2 {
		t.Fatalf("expected 2 mappers, got %d", len(*mappers))
	}

	first := (*mappers)[0]
	if first.Name != "Enrich tags" {
		t.Errorf("first.Name = %q", first.Name)
	}
	if first.RunCondition != "" {
		t.Errorf("first.RunCondition = %q, want empty", first.RunCondition)
	}
	if !strings.Contains(first.Function, ".context.outputs.tags") {
		t.Errorf("first.Function = %q", first.Function)
	}

	second := (*mappers)[1]
	if second.RunCondition != `.vars.environment == "sbx"` {
		t.Errorf("second.RunCondition = %q", second.RunCondition)
	}
}

func TestNewContextMapperFunctionModelListFromJson_EmptyList(t *testing.T) {
	t.Parallel()

	mappers, err := NewContextMapperFunctionModelListFromJson(writeTempJson(t, `[]`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mappers == nil || len(*mappers) != 0 {
		t.Fatalf("expected empty non-nil list, got %v", mappers)
	}
}

func TestNewContextMapperFunctionModelListFromJson_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		content     string
		wantErrPart string
	}{
		{
			name:        "malformed json",
			content:     `[{"name": "x", "function": `,
			wantErrPart: "unexpected end of JSON",
		},
		{
			name:        "json object instead of list",
			content:     `{"name": "x", "function": ".context"}`,
			wantErrPart: "cannot unmarshal",
		},
		{
			name:        "invalid jq in function",
			content:     `[{"name": "broken", "function": ".context | | "}]`,
			wantErrPart: `mapper function "broken" has an invalid jq syntax`,
		},
		{
			name:        "invalid jq in run_condition",
			content:     `[{"name": "broken-cond", "run_condition": ".vars.x ==", "function": ".context"}]`,
			wantErrPart: `mapper function "broken-cond" has an invalid jq syntax in the 'run_condition'`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := NewContextMapperFunctionModelListFromJson(writeTempJson(t, tt.content))
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErrPart) {
				t.Errorf("error %q does not contain %q", err.Error(), tt.wantErrPart)
			}
		})
	}
}

func TestNewContextMapperFunctionModelListFromJson_MissingFile(t *testing.T) {
	t.Parallel()

	_, err := NewContextMapperFunctionModelListFromJson(filepath.Join(t.TempDir(), "does-not-exist.json"))
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
	if !os.IsNotExist(err) {
		t.Errorf("expected a not-exist error, got %v", err)
	}
}
