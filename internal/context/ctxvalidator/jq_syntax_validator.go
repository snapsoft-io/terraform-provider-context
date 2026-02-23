// Copyright SnapSoft 2026
// SPDX-License-Identifier: MIT

package ctxvalidator

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/itchyny/gojq"
	"github.com/snapsoft/terraform-provider-context/internal/context/ctxschema"
)

var (
	_ validator.List = jqSyntaxValidator{}
)

func JqSyntaxValidator() validator.List {
	return &jqSyntaxValidator{}
}

type jqSyntaxValidator struct{}

func (v jqSyntaxValidator) Description(ctx context.Context) string {
	return "Validates that 'run_condition' and 'function' attributes in the mapper list contains syntactically correct JQ expressions."
}

func (v jqSyntaxValidator) MarkdownDescription(ctx context.Context) string {
	return "Validates that `run_condition` and `function` attributes in the mapper list contains syntactically correct JQ expressions."
}

func (v jqSyntaxValidator) ValidateList(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	if req.ConfigValue.IsUnknown() || req.ConfigValue.IsNull() {
		return
	}

	var mapperFunctions []ctxschema.ContextMapperFunctionSchema
	diags := req.ConfigValue.ElementsAs(ctx, &mapperFunctions, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	for _, mapper := range mapperFunctions {
		var name = mapper.Name.ValueString()
		var runCondition = mapper.RunCondition.ValueString()
		var function = mapper.RunCondition.ValueString()

		if runCondition != "" {
			if _, err := gojq.Parse(runCondition); err != nil {
				resp.Diagnostics.AddAttributeError(
					req.Path,
					"Invalid jq syntax in the mapper run condition",
					fmt.Sprintf("Mapper function %q has an invalid jq syntax in the 'run_condition'\n%s", name, err),
				)

				return
			}
		}

		if _, err := gojq.Parse(function); err != nil {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Invalid jq syntax in the mapper function",
				fmt.Sprintf("Mapper function %q has an invalid jq syntax in the 'function'\n%s", name, err),
			)

			return
		}
	}
}
