package main

import "testing"

func TestAncientScriptReadings(t *testing.T) {
	if got, want := len(avestanReadings), 54; got != want {
		t.Fatalf("Avesta readings count = %d, want %d", got, want)
	}
	if got, want := len(oldPersianReadings), 36; got != want {
		t.Fatalf("Ancient Persian readings count = %d, want %d", got, want)
	}
	if got, want := len(parthianReadings), 22; got != want {
		t.Fatalf("Parthian readings count = %d, want %d", got, want)
	}
	if got, want := len(pahlaviReadings), 19; got != want {
		t.Fatalf("Pahlavi readings count = %d, want %d", got, want)
	}
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
	}
	for _, test := range tests {
		if got := phoneticFor(test.language, test.char); got != test.want {
			t.Errorf("phoneticFor(%q, U+%04X) = %q, want %q", test.language, test.char, got, test.want)
		}
	}
}
