// Copyright SnapSoft 2026
// SPDX-License-Identifier: MIT

package ctxvalidator

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/snapsoft/terraform-provider-context/internal/context/ctxmodel"
)

// stackElementAttrTypes mirrors the attribute types of a context stack element as defined by the data source schemas.
var stackElementAttrTypes = map[string]attr.Type{
	"name":                        types.StringType,
	"type":                        types.StringType,
	"vars":                        types.MapType{ElemType: types.StringType},
	"mappers":                     types.ListType{ElemType: types.ObjectType{AttrTypes: mapperAttrTypes}},
	"id_casing":                   types.StringType,
	"id_prefix":                   types.StringType,
	"include_resource_type_in_id": types.BoolType,
}

func stackElement(t *testing.T, name, contextType string) attr.Value {
	t.Helper()
	obj, diags := types.ObjectValue(stackElementAttrTypes, map[string]attr.Value{
		"name":                        types.StringValue(name),
		"type":                        types.StringValue(contextType),
		"vars":                        types.MapNull(types.StringType),
		"mappers":                     types.ListNull(types.ObjectType{AttrTypes: mapperAttrTypes}),
		"id_casing":                   types.StringNull(),
		"id_prefix":                   types.StringNull(),
		"include_resource_type_in_id": types.BoolNull(),
	})
	if diags.HasError() {
		t.Fatalf("failed to build stack element: %v", diags)
	}
	return obj
}

func stackList(t *testing.T, elements ...attr.Value) types.List {
	t.Helper()
	list, diags := types.ListValue(types.ObjectType{AttrTypes: stackElementAttrTypes}, elements)
	if diags.HasError() {
		t.Fatalf("failed to build stack list: %v", diags)
	}
	return list
}

func runListValidator(t *testing.T, v validator.List, value types.List) validator.ListResponse {
	t.Helper()
	req := validator.ListRequest{
		Path:        path.Root("context").AtName("stack"),
		ConfigValue: value,
	}
	resp := validator.ListResponse{}
	v.ValidateList(t.Context(), req, &resp)
	return resp
}

func TestContextStackOrderValidator_Descriptions(t *testing.T) {
	t.Parallel()

	v := ContextStackOrderValidator(ctxmodel.ContextTypeLabel)
	if v.Description(t.Context()) == "" {
		t.Error("Description must not be empty")
	}
	if v.MarkdownDescription(t.Context()) == "" {
		t.Error("MarkdownDescription must not be empty")
	}
}

func TestContextStackOrderValidator_SkipsNullAndUnknown(t *testing.T) {
	t.Parallel()

	for _, target := range []ctxmodel.ContextType{ctxmodel.ContextTypeNamespace, ctxmodel.ContextTypeLabel} {
		v := ContextStackOrderValidator(target)
		elemType := types.ObjectType{AttrTypes: stackElementAttrTypes}

		if resp := runListValidator(t, v, types.ListNull(elemType)); resp.Diagnostics.HasError() {
			t.Errorf("null stack must not produce errors, got %v", resp.Diagnostics)
		}
		if resp := runListValidator(t, v, types.ListUnknown(elemType)); resp.Diagnostics.HasError() {
			t.Errorf("unknown stack must not produce errors, got %v", resp.Diagnostics)
		}
	}
}

func TestContextStackOrderValidator_ForNamespace(t *testing.T) {
	t.Parallel()

	v := ContextStackOrderValidator(ctxmodel.ContextTypeNamespace)

	tests := []struct {
		name    string
		stack   []attr.Value
		wantErr bool
	}{
		{"empty stack is valid for a namespace", nil, false},
		{"after one namespace", []attr.Value{stackElement(t, "a", "ns")}, false},
		{"after many namespaces", []attr.Value{stackElement(t, "a", "ns"), stackElement(t, "b", "ns"), stackElement(t, "c", "ns")}, false},
		{"after a label is invalid", []attr.Value{stackElement(t, "a", "ns"), stackElement(t, "l", "lb")}, true},
		{"after an invalid type is invalid", []attr.Value{stackElement(t, "a", "bogus")}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			resp := runListValidator(t, v, stackList(t, tt.stack...))
			if got := resp.Diagnostics.HasError(); got != tt.wantErr {
				t.Errorf("HasError() = %v, want %v; diags=%v", got, tt.wantErr, resp.Diagnostics)
			}
		})
	}
}

func TestContextStackOrderValidator_ForLabel(t *testing.T) {
	t.Parallel()

	v := ContextStackOrderValidator(ctxmodel.ContextTypeLabel)

	tests := []struct {
		name    string
		stack   []attr.Value
		wantErr bool
	}{
		{"empty stack is invalid for a label", nil, true},
		{"after one namespace", []attr.Value{stackElement(t, "a", "ns")}, false},
		{"after many namespaces", []attr.Value{stackElement(t, "a", "ns"), stackElement(t, "b", "ns")}, false},
		{"after a label is invalid", []attr.Value{stackElement(t, "a", "ns"), stackElement(t, "l", "lb")}, true},
		{"after an invalid type is invalid", []attr.Value{stackElement(t, "a", "bogus")}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			resp := runListValidator(t, v, stackList(t, tt.stack...))
			if got := resp.Diagnostics.HasError(); got != tt.wantErr {
				t.Errorf("HasError() = %v, want %v; diags=%v", got, tt.wantErr, resp.Diagnostics)
			}
		})
	}
}

func TestContextStackOrderValidator_ErrorIsAttributeScoped(t *testing.T) {
	t.Parallel()

	v := ContextStackOrderValidator(ctxmodel.ContextTypeLabel)
	resp := runListValidator(t, v, stackList(t))
	if !resp.Diagnostics.HasError() {
		t.Fatal("expected error")
	}
	diag := resp.Diagnostics.Errors()[0]
	if diag.Summary() != "Invalid context label order" {
		t.Errorf("summary = %q", diag.Summary())
	}
	withPath, ok := diag.(interface{ Path() path.Path })
	if !ok {
		t.Fatal("expected an attribute-scoped diagnostic")
	}
	if !withPath.Path().Equal(path.Root("context").AtName("stack")) {
		t.Errorf("path = %v, want context.stack", withPath.Path())
	}
}
