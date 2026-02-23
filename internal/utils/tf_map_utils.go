// Copyright SnapSoft 2026
// SPDX-License-Identifier: MIT

package utils

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func MergeTfMaps(map1, map2 types.Map) (types.Map, error) {
	if map1.IsNull() {
		map1 = types.Map{}
	}

	if map2.IsNull() {
		map2 = types.Map{}
	}

	mergedMap := make(map[string]attr.Value)
	for k, v := range map1.Elements() {
		mergedMap[k] = v
	}

	for k, v := range map2.Elements() {
		mergedMap[k] = v
	}

	mapValue, diags := types.MapValue(types.StringType, mergedMap)
	if diags.HasError() {
		var errorDetails = diags.Errors()[0].Detail()
		return mapValue, fmt.Errorf("the following values cannot be converted to terraform map %q\n%s", mergedMap, errorDetails)
	}
	return mapValue, nil
}

func ConvertGoMapToTfMap(goMap map[string]string) (types.Map, error) {
	var attrMap = make(map[string]attr.Value)
	for k, v := range goMap {
		attrMap[k] = types.StringValue(v)
	}

	tfMap, diags := types.MapValue(types.StringType, attrMap)
	if diags.HasError() {
		var errorDetails = diags.Errors()[0].Detail()
		return tfMap, fmt.Errorf("the following values cannot be converted to terraform map %q\n%s", goMap, errorDetails)
	}

	return tfMap, nil
}
