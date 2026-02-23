// Copyright SnapSoft 2026
// SPDX-License-Identifier: MIT

package ctxevaluator

import (
	"fmt"

	"github.com/itchyny/gojq"
	"github.com/snapsoft/terraform-provider-context/internal/context/ctxmodel"
)

func EvaluateJqMappers(mappers []ctxmodel.ContextMapperFunctionModel, extraInputParams map[string]any) (any, error) {
	var currentContext any = make(map[string]interface{})
	queryInput := make(map[string]any, len(extraInputParams)+1)
	for k, v := range extraInputParams {
		queryInput[k] = v
	}

	for _, mapper := range mappers {
		isRunConditionMet := true
		queryInput["context"] = currentContext

		if mapper.RunCondition != "" {
			runConditionQuery, err := gojq.Parse(mapper.RunCondition)
			if err != nil {
				return nil, fmt.Errorf("mapper function %q has an invalid jq syntax in the 'run_condition'\n%w", mapper.Name, err)
			}

			runConditionIter := runConditionQuery.Run(queryInput)
			runConditionResult, ok := runConditionIter.Next()
			if !ok {
				return nil, fmt.Errorf("mapper function %q run condition is returned with no value, must return a boolean value", mapper.Name)
			}
			if err, ok := runConditionResult.(error); ok {
				return nil, fmt.Errorf("mapper function %q run condition has evaluation error\n%w", mapper.Name, err)
			}

			isRunConditionMet, ok = runConditionResult.(bool)
			if !ok {
				return nil, fmt.Errorf("invalid return type for %q run condition, must return a boolean value", mapper.Name)
			}
		}

		if isRunConditionMet {
			functionQuery, err := gojq.Parse(mapper.Function)
			if err != nil {
				return nil, fmt.Errorf("mapper function %q has an invalid jq syntax in the 'function'\n%w", mapper.Name, err)
			}

			functionIter := functionQuery.Run(queryInput)
			functionResult, ok := functionIter.Next()
			if !ok {
				return nil, fmt.Errorf("mapper function %q is returned with no value, after evaluation it must return a new context", mapper.Name)
			}
			if err, ok := functionResult.(error); ok {
				return nil, fmt.Errorf("mapper function %q has evaluation error\n%w", mapper.Name, err)
			}

			currentContext = functionResult
		}
	}

	return currentContext, nil
}
