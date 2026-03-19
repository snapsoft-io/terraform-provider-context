// Copyright SnapSoft 2026
// SPDX-License-Identifier: MIT

package ctxvalidator

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var (
	_ validator.String = idCasingValidator{}
)

var validIdCasingValues = []string{"camelCase", "PascalCase", "snake_case", "kebab-case"}

func IdCasingValidator() validator.String {
	return &idCasingValidator{}
}

type idCasingValidator struct{}

func (v idCasingValidator) Description(_ context.Context) string {
	return "id_casing must be one of: 'camelCase', 'PascalCase', 'snake_case', 'kebab-case'"
}

func (v idCasingValidator) MarkdownDescription(_ context.Context) string {
	return "`id_casing` must be one of: `camelCase`, `PascalCase`, `snake_case`, `kebab-case`"
}

func (v idCasingValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	val := req.ConfigValue.ValueString()
	for _, valid := range validIdCasingValues {
		if val == valid {
			return
		}
	}

	resp.Diagnostics.AddAttributeError(
		req.Path,
		"Invalid id_casing value",
		fmt.Sprintf("The value %q is not valid. id_casing must be one of: camelCase, PascalCase, snake_case, kebab-case", val),
	)
}
