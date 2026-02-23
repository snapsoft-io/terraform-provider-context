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
	_ validator.String = contextLabelIdValueValidator{}
)

func ContextLabelIdValueValidator() validator.String {
	return &contextLabelIdValueValidator{}
}

type contextLabelIdValueValidator struct{}

func (v contextLabelIdValueValidator) Description(ctx context.Context) string {
	return "Context label id must be set to one of the following accepted values: 'em', 'rm', 'cm', 'it' or 'ns'"
}

func (v contextLabelIdValueValidator) MarkdownDescription(ctx context.Context) string {
	return "Context label id must be set to one of the following accepted values: `em`, `rm`, `cm`, `it` or `ns`"
}

func (v contextLabelIdValueValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	var labelId = req.ConfigValue.ValueString()
	if _, err := ctxmodel.ParseContextClassEnum(labelId); err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Context label id type validation error",
			fmt.Sprintf("The following context label id is not valid %q", labelId),
		)

		return
	}
}
