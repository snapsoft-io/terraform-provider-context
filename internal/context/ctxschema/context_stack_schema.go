// Copyright SnapSoft 2026
// SPDX-License-Identifier: MIT

package ctxschema

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/snapsoft/terraform-provider-context/internal/context/ctxmodel"
)

type ContextStackElementSchema struct {
	Name    types.String                   `tfsdk:"name"`
	LabelId types.String                   `tfsdk:"label_id"`
	Vars    types.Map                      `tfsdk:"vars"`
	Mappers *[]ContextMapperFunctionSchema `tfsdk:"mappers"`
}

type ContextStackSchema []ContextStackElementSchema

func (s *ContextStackSchema) Add(name types.String, contextLabel ctxmodel.ContextLabel, vars types.Map, mappers *[]ContextMapperFunctionSchema) {
	*s = append(*s, ContextStackElementSchema{
		Name:    name,
		LabelId: types.StringValue(contextLabel.String()),
		Vars:    vars,
		Mappers: mappers,
	})
}

func (s *ContextStackSchema) AddWithLabel(contextLabel ctxmodel.ContextLabel) {
	*s = append(*s, ContextStackElementSchema{
		Name:    types.StringNull(),
		LabelId: types.StringValue(contextLabel.String()),
		Vars:    types.MapNull(types.StringType),
		Mappers: nil,
	})
}

func (s *ContextStackSchema) AddWithNameLabelVars(name types.String, contextLabel ctxmodel.ContextLabel, vars types.Map) {
	*s = append(*s, ContextStackElementSchema{
		Name:    name,
		LabelId: types.StringValue(contextLabel.String()),
		Vars:    vars,
		Mappers: nil,
	})
}

func (s *ContextStackSchema) ToAnyGoType(ctx context.Context) []any {
	stack := make([]any, 0, len(*s))
	for _, stackElement := range *s {
		vars := make(map[string]string, 0)
		if !stackElement.Vars.IsNull() {
			for k, v := range stackElement.Vars.Elements() {
				vars[k] = v.String()
			}
		}
		newStackElement := map[string]any{
			"name":     stackElement.Name.ValueString(),
			"label_id": stackElement.LabelId.ValueString(),
			"vars":     vars,
		}
		stack = append(stack, newStackElement)
	}

	return stack
}

func (s *ContextStackSchema) IsLastElementInValidPlace() bool {
	if len(*s) <= 1 {
		return true
	}

	stackLength := len(*s)
	lastStackElement, err := ctxmodel.ParseContextClassEnum((*s)[stackLength-1].LabelId.ValueString())
	if err != nil {
		return false
	}

	for i := stackLength - 2; i >= 0; i-- {
		parsedContextClass, err := ctxmodel.ParseContextClassEnum((*s)[i].LabelId.ValueString())
		if err != nil || !lastStackElement.IsPredecessorAllowed(parsedContextClass) {
			return false
		}

		if parsedContextClass != ctxmodel.ContextLabelNamespace {
			break
		}
	}

	return true
}

func (s *ContextStackSchema) GetStackVarsInTopDownOrder(ctx context.Context) map[string]string {
	vars := make(map[string]string)
	for _, stackElement := range *s {
		if !stackElement.Vars.IsNull() {
			var stackElementMap map[string]string
			stackElement.Vars.ElementsAs(ctx, &stackElementMap, false)
			for key, value := range stackElementMap {
				vars[key] = value
			}
		}
	}

	return vars
}

func (s *ContextStackSchema) GetStackMappersInBottomUpOrder() *[]ctxmodel.ContextMapperFunctionModel {
	var mappers = make([]ctxmodel.ContextMapperFunctionModel, 0)
	for _, stackElement := range *s {
		if stackElement.Mappers != nil {
			var mapperModelList = make([]ctxmodel.ContextMapperFunctionModel, 0)
			for _, mapper := range *stackElement.Mappers {
				mapperModelList = append(mapperModelList, *mapper.ToModel())
			}

			mappers = append(mapperModelList, mappers...)
		}
	}

	return &mappers
}
