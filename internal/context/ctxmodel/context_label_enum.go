// Copyright SnapSoft 2026
// SPDX-License-Identifier: MIT

package ctxmodel

import (
	"fmt"

	"github.com/snapsoft/terraform-provider-context/internal/utils"
)

type ContextLabel int

const (
	ContextLabelExampleModule ContextLabel = iota
	ContextLabelRootModule
	ContextLabelComponentModule
	ContextLabelNamespace
	ContextLabelItem
)

var contextLabeblToStringLabelId = map[ContextLabel]string{
	ContextLabelExampleModule:   "em",
	ContextLabelRootModule:      "rm",
	ContextLabelComponentModule: "cm",
	ContextLabelNamespace:       "ns",
	ContextLabelItem:            "it",
}

var stringLabelIdToContextLabel = map[string]ContextLabel{
	"em": ContextLabelExampleModule,
	"rm": ContextLabelRootModule,
	"cm": ContextLabelComponentModule,
	"ns": ContextLabelNamespace,
	"it": ContextLabelItem,
}

var contextLabelAllowedPredecessors = map[ContextLabel][]ContextLabel{
	ContextLabelExampleModule:   {},
	ContextLabelRootModule:      {ContextLabelExampleModule},
	ContextLabelComponentModule: {ContextLabelExampleModule, ContextLabelRootModule, ContextLabelNamespace},
	ContextLabelNamespace:       {ContextLabelExampleModule, ContextLabelRootModule, ContextLabelComponentModule, ContextLabelNamespace},
	ContextLabelItem:            {ContextLabelExampleModule, ContextLabelRootModule, ContextLabelComponentModule, ContextLabelNamespace},
}

func (cl ContextLabel) String() string {
	return contextLabeblToStringLabelId[cl]
}

func ParseContextClassEnum(labelId string) (ContextLabel, error) {
	var contaxtLabel, ok = stringLabelIdToContextLabel[labelId]
	if !ok {
		return contaxtLabel, fmt.Errorf("invalid context label: the following string is not a valid context label enum %q", labelId)
	}

	return contaxtLabel, nil
}

func (cl ContextLabel) IsPredecessorAllowed(predecessor ContextLabel) bool {
	return utils.Contains(contextLabelAllowedPredecessors[cl], predecessor)
}
