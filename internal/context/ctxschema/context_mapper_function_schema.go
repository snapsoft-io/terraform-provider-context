// Copyright SnapSoft 2026
// SPDX-License-Identifier: MIT

package ctxschema

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/snapsoft/terraform-provider-context/internal/context/ctxmodel"
)

type ContextMapperFunctionSchema struct {
	Name         types.String `tfsdk:"name"`
	RunCondition types.String `tfsdk:"run_condition"`
	Function     types.String `tfsdk:"function"`
}

func (cm *ContextMapperFunctionSchema) ToModel() *ctxmodel.ContextMapperFunctionModel {
	return &ctxmodel.ContextMapperFunctionModel{
		Name:         cm.Name.ValueString(),
		RunCondition: cm.RunCondition.ValueString(),
		Function:     cm.Function.ValueString(),
	}
}
