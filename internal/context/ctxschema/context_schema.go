// Copyright SnapSoft 2026
// SPDX-License-Identifier: MIT

package ctxschema

type ContextSchema struct {
	Stack ContextStackSchema `tfsdk:"stack"`
}

func NewEmptyContextModel() *ContextSchema {
	return &ContextSchema{Stack: []ContextStackElementSchema{}}
}
