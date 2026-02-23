// Copyright SnapSoft 2026
// SPDX-License-Identifier: MIT

package utils

import (
	"context"
	"fmt"
	"math/big"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func ConvertTfAttrValueToGoType(attrValue attr.Value) (any, error) {
	if attrValue.IsNull() || attrValue.IsUnknown() {
		return nil, nil
	}

	switch value := attrValue.(type) {

	case types.Bool:
		return value.ValueBool(), nil

	case types.Number:
		floatValue, _ := value.ValueBigFloat().Float64()
		return floatValue, nil

	case types.String:
		return value.ValueString(), nil

	case types.List:
		elements := value.Elements()
		listValues := make([]any, len(elements))
		for i, e := range elements {
			goType, err := ConvertTfAttrValueToGoType(e)
			if err != nil {
				return nil, err
			}
			listValues[i] = goType
		}
		return listValues, nil

	case types.Set:
		elements := value.Elements()
		setValues := make([]any, len(elements))
		for i, e := range elements {
			goType, err := ConvertTfAttrValueToGoType(e)
			if err != nil {
				return nil, err
			}
			setValues[i] = goType
		}
		return setValues, nil

	case types.Tuple:
		elements := value.Elements()
		tupleValues := make([]any, len(elements))
		for i, e := range elements {
			goType, err := ConvertTfAttrValueToGoType(e)
			if err != nil {
				return nil, err
			}
			tupleValues[i] = goType
		}
		return tupleValues, nil

	case types.Map:
		elements := value.Elements()
		mapValues := make(map[string]any, len(elements))
		for k, v := range elements {
			goType, err := ConvertTfAttrValueToGoType(v)
			if err != nil {
				return nil, err
			}
			mapValues[k] = goType
		}
		return mapValues, nil

	case types.Object:
		attributes := value.Attributes()
		objectValues := make(map[string]any, len(attributes))
		for k, a := range attributes {
			goType, err := ConvertTfAttrValueToGoType(a)
			if err != nil {
				return nil, err
			}
			objectValues[k] = goType
		}
		return objectValues, nil

	default:
		return nil, fmt.Errorf("unsupported terraform value of type %T cannot be converted to a go type", value)
	}
}

func ConvertGoTypeToTfAttrValue(ctx context.Context, goType any) (attr.Value, error) {
	if goType == nil {
		return types.StringNull(), nil
	}

	switch value := goType.(type) {

	case bool:
		return types.BoolValue(value), nil

	case float64:
		return types.NumberValue(big.NewFloat(value)), nil

	case string:
		return types.StringValue(value), nil

	case []any:
		tupleValues := make([]attr.Value, len(value))
		tupleTypes := make([]attr.Type, len(value))
		for i, e := range value {
			attrValue, err := ConvertGoTypeToTfAttrValue(ctx, e)
			if err != nil {
				return types.StringNull(), err
			}

			tupleValues[i] = attrValue
			tupleTypes[i] = attrValue.Type(ctx)
		}

		tuple, diags := types.TupleValue(tupleTypes, tupleValues)
		if diags.HasError() {
			return tuple, fmt.Errorf("the following values cannot be converted to a terraform based tuple %q", tupleValues)
		}
		return tuple, nil

	case map[string]any:
		attributes := make(map[string]attr.Value, len(value))
		attributeTypes := make(map[string]attr.Type, len(value))
		for k, v := range value {
			attrValue, err := ConvertGoTypeToTfAttrValue(ctx, v)
			if err != nil {
				return types.StringNull(), err
			}

			attributes[k] = attrValue
			attributeTypes[k] = attrValue.Type(ctx)
		}

		object, diags := types.ObjectValue(attributeTypes, attributes)
		if diags.HasError() {
			return object, fmt.Errorf("the following attributes cannot be converted to a terraform based object %q", attributes)
		}
		return object, nil

	default:
		return types.StringNull(), fmt.Errorf("unsupported go type %T cannot be converted to a terraform value", value)
	}
}
