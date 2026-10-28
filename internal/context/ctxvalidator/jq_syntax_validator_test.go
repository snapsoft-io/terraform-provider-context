// Copyright SnapSoft 2026
// SPDX-License-Identifier: MIT

package ctxvalidator

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var mapperAttrTypes = map[string]attr.Type{
	"name":          types.StringType,
	"run_condition": types.StringType,
	"function":      types.StringType,
}

func mapperObject(t *testing.T, name string, runCondition types.String, function string) attr.Value {
	t.Helper()
	obj, diags := types.ObjectValue(mapperAttrTypes, map[string]attr.Value{
		"name":          types.StringValue(name),
		"run_condition": runCondition,
		"function":      types.StringValue(function),
	})
	if diags.HasError() {
		t.Fatalf("failed to build mapper object: %v", diags)
	}
	return obj
}

func mapperList(t *testing.T, elements ...attr.Value) types.List {
	t.Helper()
	list, diags := types.ListValue(types.ObjectType{AttrTypes: mapperAttrTypes}, elements)
	if diags.HasError() {
		t.Fatalf("failed to build mapper list: %v", diags)
	}
	return list
}

func runMapperValidator(t *testing.T, value types.List) validator.ListResponse {
	t.Helper()
	req := validator.ListRequest{
		Path:        path.Root("mappers"),
		ConfigValue: value,
	}
	resp := validator.ListResponse{}
	JqSyntaxValidator().ValidateList(t.Context(), req, &resp)
	return resp
}

func TestJqSyntaxValidator_Descriptions(t *testing.T) {
	t.Parallel()

	v := JqSyntaxValidator()
	if v.Description(t.Context()) == "" {
		t.Error("Description must not be empty")
	}
	if v.MarkdownDescription(t.Context()) == "" {
		t.Error("MarkdownDescription must not be empty")
	}
}

func TestJqSyntaxValidator_SkipsNullAndUnknown(t *testing.T) {
	t.Parallel()

	elemType := types.ObjectType{AttrTypes: mapperAttrTypes}
	if resp := runMapperValidator(t, types.ListNull(elemType)); resp.Diagnostics.HasError() {
		t.Errorf("null list must not error, got %v", resp.Diagnostics)
	}
	if resp := runMapperValidator(t, types.ListUnknown(elemType)); resp.Diagnostics.HasError() {
		t.Errorf("unknown list must not error, got %v", resp.Diagnostics)
	}
}

func TestJqSyntaxValidator_Valid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		list types.List
	}{
		{"empty list", mapperList(t)},
		{"function only with null run_condition", mapperList(t, mapperObject(t, "a", types.StringNull(), ".context"))},
		{"function only with empty run_condition", mapperList(t, mapperObject(t, "a", types.StringValue(""), ".context"))},
		{"function and run_condition", mapperList(t, mapperObject(t, "a", types.StringValue(`.vars.env == "sbx"`), `(.context.outputs.tags += {A: "1"}).context`))},
		{
			"complex real-world mapper",
			mapperList(t, mapperObject(t, "LabelIdOverride", types.StringValue(`.vars.disabled != "true"`), `
				(.context.outputs += {id: (
					[.vars.prefix] + [.vars.environment] +
					[.stack[] | select(.type == "ns" and (.name | startswith("#") | not)) | .name] +
					[.stack[] | select(.type == "lb") | .name] +
					(if (.vars.resource_type // "" | startswith("#")) then [] else [.vars.resource_type] end)
					| map(select(. != null and . != "")) | join("-")
				)}).context`)),
		},
		{"multiple valid mappers", mapperList(t, mapperObject(t, "a", types.StringNull(), ".context"), mapperObject(t, "b", types.StringValue("true"), ".context"))},
		// gojq parses an empty program as the identity filter, so it is syntactically valid.
		{"empty function is accepted as identity", mapperList(t, mapperObject(t, "empty", types.StringNull(), ""))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if resp := runMapperValidator(t, tt.list); resp.Diagnostics.HasError() {
				t.Errorf("unexpected diagnostics: %v", resp.Diagnostics)
			}
		})
	}
}

func TestJqSyntaxValidator_Invalid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		list        types.List
		wantSummary string
		wantDetail  string
	}{
		{
			name:        "invalid function",
			list:        mapperList(t, mapperObject(t, "broken-fn", types.StringNull(), ".context | |")),
			wantSummary: "Invalid jq syntax in the mapper function",
			wantDetail:  `Mapper function "broken-fn" has an invalid jq syntax in the 'function'`,
		},
		{
			name:        "invalid run_condition",
			list:        mapperList(t, mapperObject(t, "broken-cond", types.StringValue(".vars.x =="), ".context")),
			wantSummary: "Invalid jq syntax in the mapper run condition",
			wantDetail:  `Mapper function "broken-cond" has an invalid jq syntax in the 'run_condition'`,
		},
		{
			name:        "run_condition checked before function",
			list:        mapperList(t, mapperObject(t, "both", types.StringValue(".vars.x =="), ".context | |")),
			wantSummary: "Invalid jq syntax in the mapper run condition",
			wantDetail:  `"both"`,
		},
		{
			name:        "second mapper invalid is reported",
			list:        mapperList(t, mapperObject(t, "ok", types.StringNull(), ".context"), mapperObject(t, "second", types.StringNull(), "{")),
			wantSummary: "Invalid jq syntax in the mapper function",
			wantDetail:  `"second"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			resp := runMapperValidator(t, tt.list)
			if !resp.Diagnostics.HasError() {
				t.Fatal("expected error diagnostics, got none")
			}
			if n := resp.Diagnostics.ErrorsCount(); n != 1 {
				t.Errorf("expected exactly 1 error (validator stops at first failure), got %d", n)
			}
			d := resp.Diagnostics.Errors()[0]
			if d.Summary() != tt.wantSummary {
				t.Errorf("summary = %q, want %q", d.Summary(), tt.wantSummary)
			}
			if !strings.Contains(d.Detail(), tt.wantDetail) {
				t.Errorf("detail %q does not contain %q", d.Detail(), tt.wantDetail)
			}
			withPath, ok := d.(interface{ Path() path.Path })
			if !ok || !withPath.Path().Equal(path.Root("mappers")) {
				t.Errorf("expected diagnostic scoped to 'mappers' path")
			}
		})
	}
}

func TestValidIdCasingValues(t *testing.T) {
	t.Parallel()

	want := []string{"camelCase", "PascalCase", "snake_case", "kebab-case"}
	if len(ValidIdCasingValues) != len(want) {
		t.Fatalf("ValidIdCasingValues = %v, want %v", ValidIdCasingValues, want)
	}
	for i, v := range want {
		if ValidIdCasingValues[i] != v {
			t.Errorf("ValidIdCasingValues[%d] = %q, want %q", i, ValidIdCasingValues[i], v)
		}
	}
}
