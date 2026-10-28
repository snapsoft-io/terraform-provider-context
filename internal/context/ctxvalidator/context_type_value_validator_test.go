// Copyright SnapSoft 2026
// SPDX-License-Identifier: MIT

package ctxvalidator

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func runStringValidator(t *testing.T, v validator.String, value types.String) validator.StringResponse {
	t.Helper()
	req := validator.StringRequest{
		Path:        path.Root("context").AtName("stack").AtListIndex(0).AtName("type"),
		ConfigValue: value,
	}
	resp := validator.StringResponse{}
	v.ValidateString(t.Context(), req, &resp)
	return resp
}

func TestContextTypeValueValidator_Descriptions(t *testing.T) {
	t.Parallel()

	v := ContextTypeValueValidator()
	if v.Description(t.Context()) == "" {
		t.Error("Description must not be empty")
	}
	if v.MarkdownDescription(t.Context()) == "" {
		t.Error("MarkdownDescription must not be empty")
	}
}

func TestContextTypeValueValidator(t *testing.T) {
	t.Parallel()

	v := ContextTypeValueValidator()

	tests := []struct {
		name    string
		value   types.String
		wantErr bool
	}{
		{"null is skipped", types.StringNull(), false},
		{"unknown is skipped", types.StringUnknown(), false},
		{"ns is valid", types.StringValue("ns"), false},
		{"lb is valid", types.StringValue("lb"), false},
		{"empty string is invalid", types.StringValue(""), true},
		{"uppercase NS is invalid", types.StringValue("NS"), true},
		{"namespace long form is invalid", types.StringValue("namespace"), true},
		{"label long form is invalid", types.StringValue("label"), true},
		{"whitespace padded is invalid", types.StringValue(" ns "), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			resp := runStringValidator(t, v, tt.value)
			if got := resp.Diagnostics.HasError(); got != tt.wantErr {
				t.Errorf("HasError() = %v, want %v; diags=%v", got, tt.wantErr, resp.Diagnostics)
			}
			if tt.wantErr {
				if resp.Diagnostics.Errors()[0].Summary() != "Context type validation error" {
					t.Errorf("unexpected summary %q", resp.Diagnostics.Errors()[0].Summary())
				}
			}
		})
	}
}
