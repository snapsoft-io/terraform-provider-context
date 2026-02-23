// Copyright SnapSoft 2026
// SPDX-License-Identifier: MIT

package utils

func MergeMaps(maps ...map[string]string) map[string]string {
	res := make(map[string]string)

	for _, m := range maps {
		for k, v := range m {
			res[k] = v
		}
	}

	return res
}

func ToAnyMap(m map[string]string) map[string]any {
	res := make(map[string]any, len(m))
	for k, v := range m {
		res[k] = v
	}
	return res
}
