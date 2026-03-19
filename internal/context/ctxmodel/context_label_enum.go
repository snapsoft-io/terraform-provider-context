// Copyright SnapSoft 2026
// SPDX-License-Identifier: MIT

package ctxmodel

import (
	"fmt"

	"github.com/snapsoft/terraform-provider-context/internal/utils"
)

type ContextType int

const (
	ContextTypeNamespace ContextType = iota
	ContextTypeLabel
)

var contextTypeToStringId = map[ContextType]string{
	ContextTypeNamespace: "ns",
	ContextTypeLabel:     "lb",
}

var stringIdToContextType = map[string]ContextType{
	"ns": ContextTypeNamespace,
	"lb": ContextTypeLabel,
}

var contextTypeAllowedPredecessors = map[ContextType][]ContextType{
	ContextTypeNamespace: {ContextTypeNamespace},
	ContextTypeLabel:     {ContextTypeNamespace},
}

func (ct ContextType) String() string {
	return contextTypeToStringId[ct]
}

func ParseContextClassEnum(labelId string) (ContextType, error) {
	var contextType, ok = stringIdToContextType[labelId]
	if !ok {
		return contextType, fmt.Errorf("invalid context label: the following string is not a valid context label enum %q", labelId)
	}

	return contextType, nil
}

func (ct ContextType) IsPredecessorAllowed(predecessor ContextType) bool {
	return utils.Contains(contextTypeAllowedPredecessors[ct], predecessor)
}

// RequiresPredecessor returns true for types that cannot appear as the first
// (or only) element in a context stack — i.e. they always need at least one
// preceding namespace.
func (ct ContextType) RequiresPredecessor() bool {
	return ct == ContextTypeLabel
}
