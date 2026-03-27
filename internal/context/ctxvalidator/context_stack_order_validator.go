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

func ContextStackOrderValidator(contextType ctxmodel.ContextType) validator.List {
	return &contextStackOrderValidator{ContextType: contextType}
}

type contextStackOrderValidator struct {
	ContextType ctxmodel.ContextType
}

func (v contextStackOrderValidator) Description(ctx context.Context) string {
	return "Context labels must be in the following order: one or more 'namespace' entries followed by a 'label'. A label cannot appear without at least one preceding namespace."
}

func (v contextStackOrderValidator) MarkdownDescription(ctx context.Context) string {
	return "Context labels must be in the following order: one or more `namespace` entries followed by a `label`. A label cannot appear without at least one preceding namespace."
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

	stackElements.AddWithType(v.ContextType)

	if !stackElements.IsLastElementInValidPlace() {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid context label order",
			"Context labels must be in the following order: one or more 'namespace' entries followed by a 'label'. A label cannot appear without at least one preceding namespace.",
		)

		return
	}
}
