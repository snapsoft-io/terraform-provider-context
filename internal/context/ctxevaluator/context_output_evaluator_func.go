// Copyright SnapSoft 2026
// SPDX-License-Identifier: MIT

package ctxevaluator

import (
	"fmt"

	"github.com/snapsoft/terraform-provider-context/internal/context/ctxmodel"
)

func EvaluateContextOutput(contextMain any) (*ctxmodel.ContextOutputs, error) {
	contextObject, ok := contextMain.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid context.main type expected map of objects but got a %T", contextMain)
	}

	outputsObject, ok := contextObject["outputs"]
	if !ok || outputsObject == nil {
		return nil, fmt.Errorf("missing 'outputs' field from context.main")
	}

	outputsMap, ok := outputsObject.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid 'outputs' type in context.main, expected map of objects but got a %T", outputsObject)
	}

	contextOutputs := ctxmodel.ContextOutputs{}
	if id, ok := outputsMap["id"].(string); ok {
		contextOutputs.Id = id
	} else {
		return nil, fmt.Errorf("missing required output 'id' in context.main.outputs")
	}

	if tagsMap, ok := outputsMap["tags"].(map[string]any); ok {
		var tags = make(map[string]string)
		for tagKey, tagValue := range tagsMap {
			if tagValueString, ok := tagValue.(string); ok {
				tags[tagKey] = tagValueString
			} else {
				return nil, fmt.Errorf("'tags' field within context.main.outputs expected to be a map of strings but got a %T as map value", tagValue)
			}
		}
		contextOutputs.Tags = tags
	} else {
		return nil, fmt.Errorf("missing required output 'tags' in context.main.outputs")
	}

	return &contextOutputs, nil
}
