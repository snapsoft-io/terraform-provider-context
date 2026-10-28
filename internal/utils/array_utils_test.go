// Copyright SnapSoft 2026
// SPDX-License-Identifier: MIT

package utils

import "testing"

func TestContains(t *testing.T) {
	t.Parallel()

	t.Run("strings", func(t *testing.T) {
		t.Parallel()
		arr := []string{"ns", "lb"}
		if !Contains(arr, "ns") {
			t.Error("expected to contain \"ns\"")
		}
		if !Contains(arr, "lb") {
			t.Error("expected to contain \"lb\"")
		}
		if Contains(arr, "xx") {
			t.Error("did not expect to contain \"xx\"")
		}
		if Contains(arr, "") {
			t.Error("did not expect to contain empty string")
		}
	})

	t.Run("ints", func(t *testing.T) {
		t.Parallel()
		arr := []int{1, 2, 3}
		if !Contains(arr, 2) {
			t.Error("expected to contain 2")
		}
		if Contains(arr, 4) {
			t.Error("did not expect to contain 4")
		}
	})

	t.Run("empty and nil slices", func(t *testing.T) {
		t.Parallel()
		if Contains([]string{}, "a") {
			t.Error("empty slice should not contain anything")
		}
		var nilSlice []string
		if Contains(nilSlice, "a") {
			t.Error("nil slice should not contain anything")
		}
	})
}
