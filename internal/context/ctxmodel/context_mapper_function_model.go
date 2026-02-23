// Copyright SnapSoft 2026
// SPDX-License-Identifier: MIT

package ctxmodel

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/itchyny/gojq"
)

type ContextMapperFunctionModel struct {
	Name         string `json:"name"`
	RunCondition string `json:"run_condition"`
	Function     string `json:"function"`
}

func NewContextMapperFunctionModelListFromJson(jsonFilePath string) (*[]ContextMapperFunctionModel, error) {
	mappersFile, err := os.ReadFile(jsonFilePath)
	if err != nil {
		return nil, err
	}

	var mappersJson = make([]ContextMapperFunctionModel, 0)
	if err := json.Unmarshal(mappersFile, &mappersJson); err != nil {
		return nil, err
	}

	for _, mapper := range mappersJson {
		if mapper.RunCondition != "" {
			if _, err := gojq.Parse(mapper.RunCondition); err != nil {
				return nil, fmt.Errorf("mapper function %q has an invalid jq syntax in the 'run_condition'\n%w", mapper.Name, err)
			}
		}

		if _, err := gojq.Parse(mapper.Function); err != nil {
			return nil, fmt.Errorf("mapper function %q has an invalid jq syntax\n%w", mapper.Name, err)
		}
	}

	return &mappersJson, nil
}
