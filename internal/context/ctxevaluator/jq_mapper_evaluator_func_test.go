// Copyright SnapSoft 2026
// SPDX-License-Identifier: MIT

package ctxevaluator

import (
	"reflect"
	"strings"
	"testing"

	"github.com/snapsoft/terraform-provider-context/internal/context/ctxmodel"
)

func mapper(name, runCondition, function string) ctxmodel.ContextMapperFunctionModel {
	return ctxmodel.ContextMapperFunctionModel{Name: name, RunCondition: runCondition, Function: function}
}

func defaultParams() map[string]any {
	return map[string]any{
		"stack": []any{
			map[string]any{"name": "production", "type": "ns", "vars": map[string]any{"environment": "prod"}},
			map[string]any{"name": "#hidden", "type": "ns", "vars": map[string]any{}},
			map[string]any{"name": "assets", "type": "lb", "vars": map[string]any{"resource_type": "s3-bucket"}},
		},
		"vars": map[string]any{
			"environment":   "prod",
			"prefix":        "acme",
			"team":          "platform",
			"resource_type": "s3-bucket",
		},
	}
}

func TestEvaluateJqMappers_NoMappersReturnsEmptyContext(t *testing.T) {
	t.Parallel()

	got, err := EvaluateJqMappers(nil, defaultParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m, ok := got.(map[string]any)
	if !ok {
		t.Fatalf("got %T, want map[string]any", got)
	}
	if len(m) != 0 {
		t.Errorf("expected empty context, got %v", m)
	}
}

func TestEvaluateJqMappers_SingleMapperSetsOutputs(t *testing.T) {
	t.Parallel()

	mappers := []ctxmodel.ContextMapperFunctionModel{
		mapper("set id", "", `(.context.outputs += {id: ([.vars.prefix, .vars.environment] | join("-"))}).context`),
	}

	got, err := EvaluateJqMappers(mappers, defaultParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := map[string]any{"outputs": map[string]any{"id": "acme-prod"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

func TestEvaluateJqMappers_ChainsContextBetweenMappers(t *testing.T) {
	t.Parallel()

	mappers := []ctxmodel.ContextMapperFunctionModel{
		mapper("set id", "", `(.context.outputs += {id: "first"}).context`),
		mapper("add tag using id", "", `(.context.outputs.tags += {Name: .context.outputs.id}).context`),
		mapper("add another tag", "", `(.context.outputs.tags += {Team: .vars.team}).context`),
	}

	got, err := EvaluateJqMappers(mappers, defaultParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := map[string]any{
		"outputs": map[string]any{
			"id":   "first",
			"tags": map[string]any{"Name": "first", "Team": "platform"},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}
}

// outputsOf extracts the `outputs` object from an evaluated context, failing the test if the shape is unexpected.
func outputsOf(t *testing.T, got any) map[string]any {
	t.Helper()
	context, ok := got.(map[string]any)
	if !ok {
		t.Fatalf("context = %#v, want map[string]any", got)
	}
	outputs, ok := context["outputs"].(map[string]any)
	if !ok {
		t.Fatalf("context.outputs = %#v, want map[string]any", context["outputs"])
	}
	return outputs
}

func TestEvaluateJqMappers_LaterMapperOverridesEarlier(t *testing.T) {
	t.Parallel()

	mappers := []ctxmodel.ContextMapperFunctionModel{
		mapper("first", "", `(.context.outputs += {id: "first"}).context`),
		mapper("second", "", `(.context.outputs += {id: "second"}).context`),
	}

	got, err := EvaluateJqMappers(mappers, defaultParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	outputs := outputsOf(t, got)
	if outputs["id"] != "second" {
		t.Errorf("id = %v, want \"second\"", outputs["id"])
	}
}

func TestEvaluateJqMappers_RunCondition(t *testing.T) {
	t.Parallel()

	t.Run("true condition runs mapper", func(t *testing.T) {
		t.Parallel()
		mappers := []ctxmodel.ContextMapperFunctionModel{
			mapper("conditional", `.vars.environment == "prod"`, `(.context.outputs += {id: "ran"}).context`),
		}
		got, err := EvaluateJqMappers(mappers, defaultParams())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if outputsOf(t, got)["id"] != "ran" {
			t.Errorf("expected mapper to run, got %v", got)
		}
	})

	t.Run("false condition skips mapper and keeps previous context", func(t *testing.T) {
		t.Parallel()
		mappers := []ctxmodel.ContextMapperFunctionModel{
			mapper("first", "", `(.context.outputs += {id: "kept"}).context`),
			mapper("skipped", `.vars.environment == "sbx"`, `(.context.outputs += {id: "overwritten"}).context`),
		}
		got, err := EvaluateJqMappers(mappers, defaultParams())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if outputsOf(t, got)["id"] != "kept" {
			t.Errorf("expected skipped mapper to not change context, got %v", got)
		}
	})

	t.Run("condition can reference previously computed context", func(t *testing.T) {
		t.Parallel()
		mappers := []ctxmodel.ContextMapperFunctionModel{
			mapper("first", "", `(.context.outputs += {id: "x"}).context`),
			mapper("only if id set", `.context.outputs.id != null`, `(.context.outputs.tags += {HasId: "yes"}).context`),
		}
		got, err := EvaluateJqMappers(mappers, defaultParams())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		tags, ok := outputsOf(t, got)["tags"].(map[string]any)
		if !ok || tags["HasId"] != "yes" {
			t.Errorf("expected HasId tag, got %v", tags)
		}
	})

	t.Run("condition on missing var compares against null", func(t *testing.T) {
		t.Parallel()
		// Mirrors the real-world `.vars.context_config_default_id_generation_disabled != "true"` guard.
		mappers := []ctxmodel.ContextMapperFunctionModel{
			mapper("guarded", `.vars.disabled != "true"`, `(.context.outputs += {id: "ran"}).context`),
		}
		got, err := EvaluateJqMappers(mappers, defaultParams())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if outputsOf(t, got)["id"] != "ran" {
			t.Errorf("expected mapper to run when var is absent, got %v", got)
		}
	})
}

func TestEvaluateJqMappers_StackAccess(t *testing.T) {
	t.Parallel()

	// Mirrors the landing-zone LabelIdOverride mapper: skip '#'-prefixed namespaces, include label name and resource_type.
	fn := `(.context.outputs += {id: (
		[.vars.prefix] +
		[.vars.environment] +
		[.stack[] | select(.type == "ns" and (.name | startswith("#") | not)) | .name] +
		[.stack[] | select(.type == "lb") | .name] +
		[.vars.resource_type]
		| map(select(. != null and . != "")) | join("-")
	)}).context`

	got, err := EvaluateJqMappers([]ctxmodel.ContextMapperFunctionModel{mapper("id", "", fn)}, defaultParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	id := outputsOf(t, got)["id"]
	if id != "acme-prod-production-assets-s3-bucket" {
		t.Errorf("id = %v, want acme-prod-production-assets-s3-bucket", id)
	}
}

func TestEvaluateJqMappers_StackElementVarsAreAccessible(t *testing.T) {
	t.Parallel()

	fn := `(.context.outputs += {id: .stack[0].vars.environment, rt: .stack[2].vars.resource_type}).context`
	got, err := EvaluateJqMappers([]ctxmodel.ContextMapperFunctionModel{mapper("stack vars", "", fn)}, defaultParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	outputs := outputsOf(t, got)
	if outputs["id"] != "prod" {
		t.Errorf("stack[0].vars.environment = %v, want \"prod\"", outputs["id"])
	}
	if outputs["rt"] != "s3-bucket" {
		t.Errorf("stack[2].vars.resource_type = %v, want \"s3-bucket\"", outputs["rt"])
	}
}

func TestEvaluateJqMappers_ExtraParamsAreNotMutated(t *testing.T) {
	t.Parallel()

	params := defaultParams()
	mappers := []ctxmodel.ContextMapperFunctionModel{
		mapper("set id", "", `(.context.outputs += {id: "x"}).context`),
	}
	if _, err := EvaluateJqMappers(mappers, params); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, found := params["context"]; found {
		t.Error("extraInputParams must not receive the internal 'context' key")
	}
	if !reflect.DeepEqual(params, defaultParams()) {
		t.Errorf("extraInputParams were mutated: %#v", params)
	}
}

func TestEvaluateJqMappers_CanReturnNonObject(t *testing.T) {
	t.Parallel()

	// The evaluator itself is agnostic of the result type; type checking happens downstream in EvaluateContextOutput.
	got, err := EvaluateJqMappers([]ctxmodel.ContextMapperFunctionModel{mapper("str", "", `"just-a-string"`)}, defaultParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "just-a-string" {
		t.Errorf("got %v, want \"just-a-string\"", got)
	}
}

func TestEvaluateJqMappers_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		mappers     []ctxmodel.ContextMapperFunctionModel
		wantErrPart string
	}{
		{
			name:        "invalid jq syntax in function",
			mappers:     []ctxmodel.ContextMapperFunctionModel{mapper("bad-fn", "", `.context | |`)},
			wantErrPart: `mapper function "bad-fn" has an invalid jq syntax in the 'function'`,
		},
		{
			name:        "invalid jq syntax in run_condition",
			mappers:     []ctxmodel.ContextMapperFunctionModel{mapper("bad-cond", `.vars.x ==`, `.context`)},
			wantErrPart: `mapper function "bad-cond" has an invalid jq syntax in the 'run_condition'`,
		},
		{
			name:        "run_condition returns non boolean",
			mappers:     []ctxmodel.ContextMapperFunctionModel{mapper("str-cond", `"yes"`, `.context`)},
			wantErrPart: `invalid return type for "str-cond" run condition`,
		},
		{
			name:        "run_condition returns null",
			mappers:     []ctxmodel.ContextMapperFunctionModel{mapper("null-cond", `.vars.missing`, `.context`)},
			wantErrPart: `invalid return type for "null-cond" run condition`,
		},
		{
			name:        "run_condition yields no value",
			mappers:     []ctxmodel.ContextMapperFunctionModel{mapper("empty-cond", `empty`, `.context`)},
			wantErrPart: `mapper function "empty-cond" run condition is returned with no value`,
		},
		{
			name:        "run_condition evaluation error",
			mappers:     []ctxmodel.ContextMapperFunctionModel{mapper("err-cond", `.vars | keys | .[0] | .x`, `.context`)},
			wantErrPart: `mapper function "err-cond" run condition has evaluation error`,
		},
		{
			name:        "function yields no value",
			mappers:     []ctxmodel.ContextMapperFunctionModel{mapper("empty-fn", "", `empty`)},
			wantErrPart: `mapper function "empty-fn" is returned with no value`,
		},
		{
			name:        "function evaluation error",
			mappers:     []ctxmodel.ContextMapperFunctionModel{mapper("err-fn", "", `.vars.prefix | .x`)},
			wantErrPart: `mapper function "err-fn" has evaluation error`,
		},
		{
			name:        "function explicit error",
			mappers:     []ctxmodel.ContextMapperFunctionModel{mapper("explicit", "", `error("boom")`)},
			wantErrPart: "boom",
		},
		{
			name: "error in second mapper reports that mapper",
			mappers: []ctxmodel.ContextMapperFunctionModel{
				mapper("ok", "", `.context`),
				mapper("second-bad", "", `error("second")`),
			},
			wantErrPart: `mapper function "second-bad"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := EvaluateJqMappers(tt.mappers, defaultParams())
			if err == nil {
				t.Fatalf("expected error, got nil (result %#v)", got)
			}
			if got != nil {
				t.Errorf("expected nil result on error, got %#v", got)
			}
			if !strings.Contains(err.Error(), tt.wantErrPart) {
				t.Errorf("error %q does not contain %q", err.Error(), tt.wantErrPart)
			}
		})
	}
}

func TestEvaluateJqMappers_NilExtraParams(t *testing.T) {
	t.Parallel()

	got, err := EvaluateJqMappers([]ctxmodel.ContextMapperFunctionModel{mapper("x", "", `(.context += {a: 1}).context`)}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := map[string]any{"a": 1}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %#v, want %#v", got, want)
	}
}
