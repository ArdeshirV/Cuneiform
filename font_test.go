package main

import (
	"testing"

	"golang.org/x/image/font/sfnt"
)

func TestBundledFontContainsPersianGlyphs(t *testing.T) {
	font, err := sfnt.Parse(unicodeFontBytes)
	if err != nil {
		t.Fatalf("parse bundled font: %v", err)
	}
	var buffer sfnt.Buffer
	for _, char := range "فارسی پچژگک" {
		glyph, err := font.GlyphIndex(&buffer, char)
		if err != nil {
			t.Fatalf("look up Persian character %U: %v", char, err)
		}
		if glyph == 0 {
			t.Errorf("bundled font has no glyph for Persian character %U", char)
		}
	}
	for _, char := range "𐬀𐎠𐭀" {
		glyph, err := font.GlyphIndex(&buffer, char)
		if err != nil {
			t.Fatalf("look up ancient character %U: %v", char, err)
		}
		if glyph == 0 {
			t.Errorf("bundled font has no glyph for ancient character %U", char)
		}
	}
}
