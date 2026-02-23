// Copyright SnapSoft 2026
// SPDX-License-Identifier: MIT

package utils

func Contains[T comparable](array []T, item T) bool {
	for _, v := range array {
		if v == item {
			return true
		}
	}
	return false
}
