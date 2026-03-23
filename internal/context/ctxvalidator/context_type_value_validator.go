// Copyright SnapSoft 2026
// SPDX-License-Identifier: MIT

package ctxvalidator

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/snapsoft/terraform-provider-context/internal/context/ctxmodel"
)

var (
	_ validator.String = contextTypeValueValidator{}
)

func ContextTypeValueValidator() validator.String {
	return &contextTypeValueValidator{}
}

type contextTypeValueValidator struct{}

func (v contextTypeValueValidator) Description(ctx context.Context) string {
	return "Context type must be set to one of the following accepted values: 'ns' or 'lb'"
}

func (v contextTypeValueValidator) MarkdownDescription(ctx context.Context) string {
	return "Context type must be set to one of the following accepted values: `ns` or `lb`"
}

func (v contextTypeValueValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	var value = req.ConfigValue.ValueString()
	if _, err := ctxmodel.ParseContextClassEnum(value); err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Context type validation error",
			fmt.Sprintf("The following context type is not valid %q", value),
		)

		return
	}
}
