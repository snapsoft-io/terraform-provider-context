// Copyright SnapSoft 2026
// SPDX-License-Identifier: MIT

package utils

import "strings"

// ApplyCasing concatenates parts using the specified casing strategy.
// Supported strategies: "camelCase", "PascalCase", "snake_case", "kebab-case".
// If casing is unrecognized or empty, defaults to "kebab-case".
func ApplyCasing(parts []string, casing string) string {
	nonEmpty := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			nonEmpty = append(nonEmpty, p)
		}
	}
	if len(nonEmpty) == 0 {
		return ""
	}
	switch casing {
	case "camelCase":
		return camelCase(nonEmpty)
	case "PascalCase":
		return pascalCase(nonEmpty)
	case "snake_case":
		return snakeCase(nonEmpty)
	default:
		return kebabCase(nonEmpty)
	}
}

func capitalizeFirst(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
}

func camelCase(parts []string) string {
	result := strings.ToLower(parts[0])
	for _, p := range parts[1:] {
		result += capitalizeFirst(p)
	}
	return result
}

func pascalCase(parts []string) string {
	result := ""
	for _, p := range parts {
		result += capitalizeFirst(p)
	}
	return result
}

func snakeCase(parts []string) string {
	lower := make([]string, len(parts))
	for i, p := range parts {
		lower[i] = strings.ToLower(p)
	}
	return strings.Join(lower, "_")
}

func kebabCase(parts []string) string {
	lower := make([]string, len(parts))
	for i, p := range parts {
		lower[i] = strings.ToLower(p)
	}
	return strings.Join(lower, "-")
}
