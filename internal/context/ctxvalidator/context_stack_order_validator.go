// Copyright SnapSoft 2026
// SPDX-License-Identifier: MIT

package ctxvalidator

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/snapsoft/terraform-provider-context/internal/context/ctxmodel"
	"github.com/snapsoft/terraform-provider-context/internal/context/ctxschema"
)

var (
	_ validator.List = contextStackOrderValidator{}
)

func ContextStackOrderValidator(labelId ctxmodel.ContextLabel) validator.List {
	return &contextStackOrderValidator{LabelId: labelId}
}

type contextStackOrderValidator struct {
	LabelId ctxmodel.ContextLabel
}

func (v contextStackOrderValidator) Description(ctx context.Context) string {
	return "Context labels must be in the following order 'example module'->'root module'->'component module'->'item', and only 'namespace' can be repeated, and must be placed between 'root module' and 'item'"
}

func (v contextStackOrderValidator) MarkdownDescription(ctx context.Context) string {
	return "Context labels must be in the following order `example module`->`root module`->`component module`->`item`, and only `namespace` can be repeated, and must be placed between `root module` and `item`"
}

func (v contextStackOrderValidator) ValidateList(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	var stackElements ctxschema.ContextStackSchema
	diags := req.ConfigValue.ElementsAs(ctx, &stackElements, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	stackElements.AddWithLabel(v.LabelId)

	if !stackElements.IsLastElementInValidPlace() {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid context label order",
			"Context labels elements must be in the following order 'example module'->'root module'->'component module'->'item', and only 'namespace' can be repeated, and must be placed between 'root module' and 'item'",
		)

		return
	}
}
