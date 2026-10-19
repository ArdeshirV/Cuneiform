package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"unicode/utf16"
	"unicode/utf8"
)

type textEncoding uint8

const (
	encodingUTF8 textEncoding = iota
	encodingUTF8BOM
	encodingUTF16LE
	encodingUTF16BE
)

func decodeDocument(data []byte) (string, textEncoding, error) {
	switch {
	case len(data) >= 2 && data[0] == 0xFF && data[1] == 0xFE:
		return decodeUTF16(data[2:], encodingUTF16LE, binary.LittleEndian)
	case len(data) >= 2 && data[0] == 0xFE && data[1] == 0xFF:
		return decodeUTF16(data[2:], encodingUTF16BE, binary.BigEndian)
	case len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF:
		if !utf8.Valid(data[3:]) {
			return "", encodingUTF8BOM, errors.New(configString("document_invalid_utf8_bom"))
		}
		return string(data[3:]), encodingUTF8BOM, nil
	default:
		if !utf8.Valid(data) {
			return "", encodingUTF8, errors.New(configString("document_invalid_utf8"))
		}
		return string(data), encodingUTF8, nil
	}
}

func decodeUTF16(data []byte, encoding textEncoding, order binary.ByteOrder) (string, textEncoding, error) {
	if len(data)%2 != 0 {
		return "", encoding, errors.New(configString("document_incomplete_utf16_unit"))
	}
	units := make([]uint16, len(data)/2)
	for i := range units {
		units[i] = order.Uint16(data[i*2:])
	}
	for i := 0; i < len(units); i++ {
		if 0xD800 <= units[i] && units[i] <= 0xDBFF {
			if i+1 >= len(units) || units[i+1] < 0xDC00 || units[i+1] > 0xDFFF {
				return "", encoding, fmt.Errorf(configString("document_invalid_utf16_surrogate"), i)
			}
			i++
		} else if 0xDC00 <= units[i] && units[i] <= 0xDFFF {
			return "", encoding, fmt.Errorf(configString("document_unpaired_utf16_surrogate"), i)
		}
	}
	return string(utf16.Decode(units)), encoding, nil
}

func encodeDocument(text string, encoding textEncoding) []byte {
	switch encoding {
	case encodingUTF16LE, encodingUTF16BE:
		order := binary.ByteOrder(binary.LittleEndian)
		bom := []byte{0xFF, 0xFE}
		if encoding == encodingUTF16BE {
			order = binary.BigEndian
			bom = []byte{0xFE, 0xFF}
		}
		units := utf16.Encode([]rune(text))
		data := make([]byte, 2*len(units))
		for i, unit := range units {
			order.PutUint16(data[i*2:], unit)
		}
		return append(bom, data...)
	case encodingUTF8BOM:
		return append([]byte{0xEF, 0xBB, 0xBF}, []byte(text)...)
	default:
		return []byte(text)
	}
}

func (encoding textEncoding) String() string {
	switch encoding {
	case encodingUTF8BOM:
		return configString("encoding_utf8_bom")
	case encodingUTF16LE:
		return configString("encoding_utf16_le")
	case encodingUTF16BE:
		return configString("encoding_utf16_be")
	default:
		return configString("encoding_utf8")
	}
}
