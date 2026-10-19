package main

import (
	"testing"

	"golang.org/x/image/font/sfnt"
)

func TestBundledFontContainsPersianAndPriorityScriptGlyphs(t *testing.T) {
	font, err := sfnt.Parse(unicodeFontBytes)
	if err != nil {
		t.Fatalf("parse bundled font: %v", err)
	}
	var buffer sfnt.Buffer
	for _, char := range []rune("فارسی پچژگک𐬀𐎠𐭀") {
		glyph, err := font.GlyphIndex(&buffer, char)
		if err != nil {
			t.Fatalf("look up character %U: %v", char, err)
		}
		if glyph == 0 {
			t.Errorf("bundled font has no glyph for character %U", char)
		}
	}
}

func TestBundledFontCoversHistoricNotoScripts(t *testing.T) {
	font, err := sfnt.Parse(unicodeFontBytes)
	if err != nil {
		t.Fatalf("parse bundled font: %v", err)
	}
	var buffer sfnt.Buffer
	for _, char := range []rune{
		0x10B00, 0x103A0, 0x10B40, 0x10B60, 0x10AC0, 0x10F00, 0x10F30,
		0x10000, 0x10300, 0x10330, 0x10380, 0x10800, 0x10900, 0x12000,
		0x13000, 0x14400, 0x2C80, 0x10A00, 0x11000, 0x11580, 0x1E900,
	} {
		glyph, err := font.GlyphIndex(&buffer, char)
		if err != nil {
			t.Fatalf("look up historic-script character U+%04X: %v", char, err)
		}
		if glyph == 0 {
			t.Errorf("bundled font has no glyph for historic-script character U+%04X", char)
		}
	}
}
