package main

import (
	"testing"
	"unicode"
)

func TestLanguagesIncludePriorityAndUnicodeScripts(t *testing.T) {
	wantFirst := []string{"Avesta", "Pahlavi", "Parthian", "Ancient Persian"}
	for i, name := range wantFirst {
		if languages[i].name != name {
			t.Fatalf("languages[%d] = %q, want %q", i, languages[i].name, name)
		}
	}

	var garay *language
	for i := range languages {
		if languages[i].name == "Garay" {
			garay = &languages[i]
			break
		}
	}
	if garay == nil {
		t.Fatal("Unicode script Garay is missing from the language list")
	}
	chars := charactersFor(*garay)
	if len(chars) < 2 {
		t.Fatalf("Garay keyboard contains only %d characters", len(chars))
	}
	for _, char := range chars {
		if unicode.Is(unicode.Scripts["Garay"], char) && unicode.IsPrint(char) {
			return
		}
	}
	t.Fatal("Garay keyboard has no printable Garay characters")
}
