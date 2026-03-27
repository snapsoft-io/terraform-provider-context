// Copyright SnapSoft 2026
// SPDX-License-Identifier: MIT

package ctxmodel

type ContextProviderConfigModel struct {
	MapperFunctions        *[]ContextMapperFunctionModel
	Vars                   map[string]string
	IdCasing               string
	IdPrefix               string
	IncludeResourceTypeInId bool
}
