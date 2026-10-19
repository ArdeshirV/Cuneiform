package main

import "testing"

func TestAncientScriptReadings(t *testing.T) {
	tests := []struct {
		language string
		char     rune
		want     string
	}{
		{"Avesta", 0x10B00, "a"},
		{"Avesta", 0x10B35, "he"},
		{"Ancient Persian", 0x103A3, "ka"},
		{"Ancient Persian", 0x103D1, "one"},
		{"Parthian", 0x10B40, "a"},
		{"Parthian", 0x10B55, "t"},
		{"Pahlavi", 0x10B60, "a"},
		{"Pahlavi", 0x10B72, "t"},
		{"Pahlavi", 0x10B78, "one"},
	}
	for _, test := range tests {
		if got := phoneticFor(test.language, test.char); got != test.want {
			t.Errorf("phoneticFor(%q, U+%04X) = %q, want %q", test.language, test.char, got, test.want)
		}
	}
}
