// Copyright (c) 2026-Present Diagrid Inc.
// SPDX-License-Identifier: BUSL-1.1

package parkflag

import (
	"strings"
	"testing"
)

func TestRedact(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "tool response with a FLAG line",
			in:   "Welcome... to Jurassic Park. The tour is yours.\n\nFLAG: gc26{4aa19062-8343-446a-997e-da414d0b949e}",
			want: "Welcome... to Jurassic Park. The tour is yours.\n\nFLAG: " + notice,
		},
		{
			name: "paraphrased, flag mid-sentence",
			in:   "The grid is open and your flag is gc26{fb912b19-ee3f-4dd6-8705-481c4e942102} — congrats!",
			want: "The grid is open and your flag is " + notice + " — congrats!",
		},
		{
			name: "several flags in one reply",
			in:   "gc26{aaaaaaaa-0000-0000-0000-000000000000} then gc26{bbbbbbbb-1111-1111-1111-111111111111}",
			want: notice + " then " + notice,
		},
		{
			name: "no flag leaves text untouched",
			in:   "Access code rejected — the system stays locked.",
			want: "Access code rejected — the system stays locked.",
		},
		{
			name: "empty input",
			in:   "",
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Redact(tt.in)
			if got != tt.want {
				t.Errorf("Redact() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestRedactLeaksNothing guards the property that actually matters: no flag
// value survives, whatever shape the model wrapped it in.
func TestRedactLeaksNothing(t *testing.T) {
	const flag = "gc26{a42039dd-e0e7-4e33-a3f7-c0483c4a232c}"
	shapes := []string{
		"FLAG: " + flag,
		"  flag: " + flag + "  ",
		"Here you go: **" + flag + "**",
		flag,
		"FLAG:\n" + flag,
	}
	for _, in := range shapes {
		if got := Redact(in); strings.Contains(got, flag) {
			t.Errorf("Redact(%q) leaked the flag: %q", in, got)
		}
	}
}
