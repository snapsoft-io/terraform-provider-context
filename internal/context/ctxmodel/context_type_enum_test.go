// Copyright SnapSoft 2026
// SPDX-License-Identifier: MIT

package ctxmodel

import "testing"

func TestContextType_String(t *testing.T) {
	t.Parallel()

	if got := ContextTypeNamespace.String(); got != "ns" {
		t.Errorf("ContextTypeNamespace.String() = %q, want \"ns\"", got)
	}
	if got := ContextTypeLabel.String(); got != "lb" {
		t.Errorf("ContextTypeLabel.String() = %q, want \"lb\"", got)
	}
	if got := ContextType(99).String(); got != "" {
		t.Errorf("unknown ContextType.String() = %q, want empty string", got)
	}
}

func TestParseContextClassEnum(t *testing.T) {
	t.Parallel()

	valid := map[string]ContextType{
		"ns": ContextTypeNamespace,
		"lb": ContextTypeLabel,
	}
	for in, want := range valid {
		got, err := ParseContextClassEnum(in)
		if err != nil {
			t.Errorf("ParseContextClassEnum(%q) unexpected error: %v", in, err)
		}
		if got != want {
			t.Errorf("ParseContextClassEnum(%q) = %v, want %v", in, got, want)
		}
	}

	invalid := []string{"", "NS", "LB", "namespace", "label", "xx", " ns"}
	for _, in := range invalid {
		if _, err := ParseContextClassEnum(in); err == nil {
			t.Errorf("ParseContextClassEnum(%q) expected error, got nil", in)
		}
	}
}

func TestContextType_RoundTrip(t *testing.T) {
	t.Parallel()

	for _, ct := range []ContextType{ContextTypeNamespace, ContextTypeLabel} {
		parsed, err := ParseContextClassEnum(ct.String())
		if err != nil {
			t.Fatalf("round trip of %v failed: %v", ct, err)
		}
		if parsed != ct {
			t.Errorf("round trip of %v yielded %v", ct, parsed)
		}
	}
}

func TestContextType_IsPredecessorAllowed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		ct          ContextType
		predecessor ContextType
		want        bool
	}{
		{"namespace after namespace", ContextTypeNamespace, ContextTypeNamespace, true},
		{"namespace after label is not allowed", ContextTypeNamespace, ContextTypeLabel, false},
		{"label after namespace", ContextTypeLabel, ContextTypeNamespace, true},
		{"label after label is not allowed", ContextTypeLabel, ContextTypeLabel, false},
		{"unknown type has no allowed predecessors", ContextType(99), ContextTypeNamespace, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.ct.IsPredecessorAllowed(tt.predecessor); got != tt.want {
				t.Errorf("%v.IsPredecessorAllowed(%v) = %v, want %v", tt.ct, tt.predecessor, got, tt.want)
			}
		})
	}
}

func TestContextType_RequiresPredecessor(t *testing.T) {
	t.Parallel()

	if ContextTypeNamespace.RequiresPredecessor() {
		t.Error("namespace must be allowed as the first stack element")
	}
	if !ContextTypeLabel.RequiresPredecessor() {
		t.Error("label must require a preceding namespace")
	}
}
