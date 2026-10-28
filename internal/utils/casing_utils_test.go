// Copyright SnapSoft 2026
// SPDX-License-Identifier: MIT

package utils

import "testing"

func TestApplyCasing(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		parts  []string
		casing string
		want   string
	}{
		{"kebab-case basic", []string{"production", "eu-west-1", "assets"}, "kebab-case", "production-eu-west-1-assets"},
		{"kebab-case lowercases", []string{"Production", "Assets"}, "kebab-case", "production-assets"},
		{"snake_case basic", []string{"production", "jobs", "function"}, "snake_case", "production_jobs_function"},
		{"snake_case lowercases and keeps hyphens", []string{"Production", "eu-west-1"}, "snake_case", "production_eu-west-1"},
		{"PascalCase basic", []string{"production", "assets"}, "PascalCase", "ProductionAssets"},
		{"PascalCase splits hyphenated parts", []string{"production", "eu-west-1", "assets"}, "PascalCase", "ProductionEuWest1Assets"},
		{"PascalCase splits underscored parts", []string{"my_app", "db"}, "PascalCase", "MyAppDb"},
		{"PascalCase normalizes inner casing", []string{"PRODUCTION", "aSSets"}, "PascalCase", "ProductionAssets"},
		{"camelCase basic", []string{"production", "assets"}, "camelCase", "productionAssets"},
		{"camelCase splits hyphenated parts", []string{"production", "eu-west-1", "assets"}, "camelCase", "productionEuWest1Assets"},
		{"camelCase first word lowercased", []string{"Production", "Assets"}, "camelCase", "productionAssets"},
		{"camelCase single part", []string{"Production"}, "camelCase", "production"},
		{"empty casing defaults to kebab", []string{"a", "b"}, "", "a-b"},
		{"unknown casing defaults to kebab", []string{"a", "b"}, "SCREAMING", "a-b"},
		{"empty parts are skipped", []string{"", "a", "", "b", ""}, "kebab-case", "a-b"},
		{"all empty parts yields empty string", []string{"", ""}, "kebab-case", ""},
		{"nil parts yields empty string", nil, "kebab-case", ""},
		{"camelCase with only separators yields empty string", []string{"-", "_"}, "camelCase", ""},
		{"PascalCase with only separators yields empty string", []string{"-", "_"}, "PascalCase", ""},
		{"single part kebab", []string{"vpc"}, "kebab-case", "vpc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := ApplyCasing(tt.parts, tt.casing); got != tt.want {
				t.Errorf("ApplyCasing(%v, %q) = %q, want %q", tt.parts, tt.casing, got, tt.want)
			}
		})
	}
}

func TestCapitalizeFirst(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"":        "",
		"a":       "A",
		"abc":     "Abc",
		"ABC":     "Abc",
		"aBC":     "Abc",
		"1abc":    "1abc",
		"-dashed": "-dashed",
	}

	for in, want := range tests {
		if got := capitalizeFirst(in); got != want {
			t.Errorf("capitalizeFirst(%q) = %q, want %q", in, got, want)
		}
	}
}
