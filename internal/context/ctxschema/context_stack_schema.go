// Copyright SnapSoft 2026
// SPDX-License-Identifier: MIT

package ctxschema

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/snapsoft/terraform-provider-context/internal/context/ctxmodel"
)

type ContextStackElementSchema struct {
	Name                    types.String                   `tfsdk:"name"`
	ContextType             types.String                   `tfsdk:"label_id"`
	Vars                    types.Map                      `tfsdk:"vars"`
	Mappers                 *[]ContextMapperFunctionSchema `tfsdk:"mappers"`
	IdCasing                types.String                   `tfsdk:"id_casing"`
	IdPrefix                types.String                   `tfsdk:"id_prefix"`
	IncludeResourceTypeInId types.Bool                     `tfsdk:"include_resource_type_in_id"`
}

type ContextStackSchema []ContextStackElementSchema

func (s *ContextStackSchema) Add(name types.String, contextType ctxmodel.ContextType, vars types.Map, mappers *[]ContextMapperFunctionSchema, idCasing types.String, idPrefix types.String, includeResourceTypeInId types.Bool) {
	*s = append(*s, ContextStackElementSchema{
		Name:                    name,
		ContextType:             types.StringValue(contextType.String()),
		Vars:                    vars,
		Mappers:                 mappers,
		IdCasing:                idCasing,
		IdPrefix:                idPrefix,
		IncludeResourceTypeInId: includeResourceTypeInId,
	})
}

func (s *ContextStackSchema) AddWithType(contextType ctxmodel.ContextType) {
	*s = append(*s, ContextStackElementSchema{
		Name:                    types.StringNull(),
		ContextType:             types.StringValue(contextType.String()),
		Vars:                    types.MapNull(types.StringType),
		Mappers:                 nil,
		IdCasing:                types.StringNull(),
		IdPrefix:                types.StringNull(),
		IncludeResourceTypeInId: types.BoolNull(),
	})
}

func (s *ContextStackSchema) AddWithNameLabelVars(name types.String, contextType ctxmodel.ContextType, vars types.Map) {
	*s = append(*s, ContextStackElementSchema{
		Name:                    name,
		ContextType:             types.StringValue(contextType.String()),
		Vars:                    vars,
		Mappers:                 nil,
		IdCasing:                types.StringNull(),
		IdPrefix:                types.StringNull(),
		IncludeResourceTypeInId: types.BoolNull(),
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
			"label_id": stackElement.ContextType.ValueString(),
			"vars":     vars,
		}
		stack = append(stack, newStackElement)
	}

	return stack
}

func (s *ContextStackSchema) IsLastElementInValidPlace() bool {
	stackLength := len(*s)

	if stackLength == 0 {
		return true
	}

	lastStackElement, err := ctxmodel.ParseContextClassEnum((*s)[stackLength-1].ContextType.ValueString())
	if err != nil {
		return false
	}

	// An element that requires a predecessor (e.g. a label) cannot be the
	// only entry in the stack — it must be preceded by at least one namespace.
	if stackLength == 1 {
		return !lastStackElement.RequiresPredecessor()
	}

	for i := stackLength - 2; i >= 0; i-- {
		parsedContextClass, err := ctxmodel.ParseContextClassEnum((*s)[i].ContextType.ValueString())
		if err != nil || !lastStackElement.IsPredecessorAllowed(parsedContextClass) {
			return false
		}

		if parsedContextClass != ctxmodel.ContextTypeNamespace {
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

// GetNamespaceNames returns the names of all namespace elements in the stack, in order.
func (s *ContextStackSchema) GetNamespaceNames() []string {
	names := make([]string, 0)
	for _, elem := range *s {
		if elem.ContextType.ValueString() == ctxmodel.ContextTypeNamespace.String() {
			names = append(names, elem.Name.ValueString())
		}
	}
	return names
}

// GetEffectiveIdCasing returns the id_casing from the last namespace element that has it set.
// Returns the value and true if found, or empty string and false if no namespace sets it.
func (s *ContextStackSchema) GetEffectiveIdCasing() (string, bool) {
	var found string
	var ok bool
	for _, elem := range *s {
		if elem.ContextType.ValueString() == ctxmodel.ContextTypeNamespace.String() && !elem.IdCasing.IsNull() {
			found = elem.IdCasing.ValueString()
			ok = true
		}
	}
	return found, ok
}

// GetEffectiveIdPrefix returns the id_prefix from the last namespace element that has it set.
// Returns the value and true if found, or empty string and false if no namespace sets it.
func (s *ContextStackSchema) GetEffectiveIdPrefix() (string, bool) {
	var found string
	var ok bool
	for _, elem := range *s {
		if elem.ContextType.ValueString() == ctxmodel.ContextTypeNamespace.String() && !elem.IdPrefix.IsNull() {
			found = elem.IdPrefix.ValueString()
			ok = true
		}
	}
	return found, ok
}

// GetEffectiveIncludeResourceTypeInId returns the include_resource_type_in_id from the last namespace element that has it set.
// Returns the value and true if found, or false and false if no namespace sets it.
func (s *ContextStackSchema) GetEffectiveIncludeResourceTypeInId() (bool, bool) {
	var found bool
	var ok bool
	for _, elem := range *s {
		if elem.ContextType.ValueString() == ctxmodel.ContextTypeNamespace.String() && !elem.IncludeResourceTypeInId.IsNull() {
			found = elem.IncludeResourceTypeInId.ValueBool()
			ok = true
		}
	}
	return found, ok
}
