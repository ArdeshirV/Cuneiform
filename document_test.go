package main

import (
	"bytes"
	"testing"
)

func TestDecodeAndEncodeDocument(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		text     string
		encoding textEncoding
	}{
		{"UTF-8", []byte("Avesta: 𐬀𐬁"), "Avesta: 𐬀𐬁", encodingUTF8},
		{"UTF-8 BOM", []byte{0xEF, 0xBB, 0xBF, 'A'}, "A", encodingUTF8BOM},
		{"UTF-16 LE", []byte{0xFF, 0xFE, 0x41, 0, 0x01, 0xD8, 0, 0xDC}, "A𐐀", encodingUTF16LE},
		{"UTF-16 BE", []byte{0xFE, 0xFF, 0, 0x41, 0xD8, 0x01, 0xDC, 0}, "A𐐀", encodingUTF16BE},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			text, encoding, err := decodeDocument(test.data)
			if err != nil {
				t.Fatal(err)
			}
			if text != test.text || encoding != test.encoding {
				t.Fatalf("decodeDocument() = (%q, %v), want (%q, %v)", text, encoding, test.text, test.encoding)
			}
			if encoded := encodeDocument(text, encoding); !bytes.Equal(encoded, test.data) {
				t.Fatalf("round-trip = %v, want %v", encoded, test.data)
			}
		})
	}
}

func TestDecodeDocumentRejectsInvalidText(t *testing.T) {
	for _, data := range [][]byte{{0xFF}, {0xFF, 0xFE, 0x00}, {0xFF, 0xFE, 0x00, 0xD8}} {
		if _, _, err := decodeDocument(data); err == nil {
			t.Errorf("decodeDocument(%v) unexpectedly succeeded", data)
		}
	}
}

func TestAncientIranianLanguagesAppearFirst(t *testing.T) {
	want := []string{"Avesta", "Pahlavi", "Parthian", "Ancient Persian"}
	for i, name := range want {
		if languages[i].name != name {
			t.Fatalf("languages[%d] = %q, want %q", i, languages[i].name, name)
		}
		if len(charactersFor(languages[i])) == 0 {
			t.Errorf("%s keyboard has no printable characters", name)
		}
	}
}
