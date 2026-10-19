package main

import (
	"sort"
	"strings"
	"unicode"
)

type runeRange struct {
	first rune
	last  rune
}

type language struct {
	name    string
	ranges  []runeRange
	scripts []*unicode.RangeTable
}

func block(first, last rune) []runeRange {
	return []runeRange{{first: first, last: last}}
}

var languages = []language{
	{name: "Avesta", ranges: block(0x10B00, 0x10B3F)},
	{name: "Pahlavi", ranges: []runeRange{{0x10B60, 0x10B7F}, {0x10B80, 0x10BAF}}},
	{name: "Parthian", ranges: block(0x10B40, 0x10B5F)},
	{name: "Ancient Persian", ranges: block(0x103A0, 0x103DF)},
	{name: "Sogdian", ranges: block(0x10F30, 0x10F6F)},
	{name: "Old Sogdian", ranges: block(0x10F00, 0x10F2F)},
	{name: "Manichaean", ranges: block(0x10AC0, 0x10AFF)},
	{name: "Khwarezmian", ranges: block(0x10FB0, 0x10FDF)},
	{name: "Old Uyghur", ranges: block(0x10F70, 0x10FAF)},
	{name: "Cuneiform", ranges: block(0x12000, 0x123FF)},
	{name: "Egyptian Hieroglyphs", ranges: block(0x13000, 0x1345F)},
	{name: "Phoenician", ranges: block(0x10900, 0x1091F)},
	{name: "Ugaritic", ranges: block(0x10380, 0x1039F)},
	{name: "Linear B", ranges: block(0x10000, 0x1007F)},
	{name: "Cypriot", ranges: block(0x10800, 0x1083F)},
	{name: "Old Italic", ranges: block(0x10300, 0x1032F)},
	{name: "Runic", ranges: block(0x16A0, 0x16FF)},
	{name: "Gothic", ranges: block(0x10330, 0x1034F)},
	{name: "Glagolitic", ranges: block(0x2C00, 0x2C5F)},
	{name: "Coptic", ranges: block(0x2C80, 0x2CFF)},
	{name: "Greek", ranges: []runeRange{{0x0370, 0x03FF}, {0x1F00, 0x1FFF}}},
	{name: "Latin", ranges: []runeRange{{0x0020, 0x024F}, {0x1E00, 0x1EFF}, {0xA720, 0xA7FF}, {0xAB30, 0xAB6F}}},
	{name: "Armenian", ranges: block(0x0530, 0x058F)},
	{name: "Hebrew", ranges: block(0x0590, 0x05FF)},
	{name: "Arabic", ranges: []runeRange{{0x0600, 0x06FF}, {0x0750, 0x077F}, {0x08A0, 0x08FF}}},
	{name: "Syriac", ranges: block(0x0700, 0x074F)},
	{name: "Thaana", ranges: block(0x0780, 0x07BF)},
	{name: "Devanagari", ranges: block(0x0900, 0x097F)},
	{name: "Bengali", ranges: block(0x0980, 0x09FF)},
	{name: "Gurmukhi", ranges: block(0x0A00, 0x0A7F)},
	{name: "Gujarati", ranges: block(0x0A80, 0x0AFF)},
	{name: "Odia", ranges: block(0x0B00, 0x0B7F)},
	{name: "Tamil", ranges: block(0x0B80, 0x0BFF)},
	{name: "Telugu", ranges: block(0x0C00, 0x0C7F)},
	{name: "Kannada", ranges: block(0x0C80, 0x0CFF)},
	{name: "Malayalam", ranges: block(0x0D00, 0x0D7F)},
	{name: "Sinhala", ranges: block(0x0D80, 0x0DFF)},
	{name: "Thai", ranges: block(0x0E00, 0x0E7F)},
	{name: "Lao", ranges: block(0x0E80, 0x0EFF)},
	{name: "Tibetan", ranges: block(0x0F00, 0x0FFF)},
	{name: "Myanmar", ranges: block(0x1000, 0x109F)},
	{name: "Georgian", ranges: block(0x10A0, 0x10FF)},
	{name: "Ethiopic", ranges: block(0x1200, 0x137F)},
	{name: "Cherokee", ranges: block(0x13A0, 0x13FF)},
	{name: "Canadian Aboriginal", ranges: block(0x1400, 0x167F)},
	{name: "Mongolian", ranges: block(0x1800, 0x18AF)},
	{name: "Ogham", ranges: block(0x1680, 0x169F)},
	{name: "Bopomofo", ranges: block(0x3100, 0x312F)},
	{name: "Hangul", ranges: []runeRange{{0x1100, 0x11FF}, {0x3130, 0x318F}, {0xAC00, 0xD7AF}}},
	{name: "Japanese Hiragana", ranges: block(0x3040, 0x309F)},
	{name: "Japanese Katakana", ranges: []runeRange{{0x30A0, 0x30FF}, {0x31F0, 0x31FF}}},
	{name: "Chinese Han (CJK)", ranges: []runeRange{
		{0x3400, 0x4DBF}, {0x4E00, 0x9FFF}, {0xF900, 0xFAFF},
		{0x20000, 0x2A6DF}, {0x2A700, 0x2B73F}, {0x2B740, 0x2B81F},
		{0x2B820, 0x2CEAF}, {0x2CEB0, 0x2EBEF}, {0x2EBF0, 0x2EE5F},
		{0x30000, 0x3134F}, {0x31350, 0x323AF},
	}},
	{name: "Yi", ranges: block(0xA000, 0xA4CF)},
	{name: "Vai", ranges: block(0xA500, 0xA63F)},
	{name: "Adlam", ranges: block(0x1E900, 0x1E95F)},
	{name: "Tifinagh", ranges: block(0x2D30, 0x2D7F)},
	{name: "N'Ko", ranges: block(0x07C0, 0x07FF)},
	{name: "Etruscan", ranges: block(0x10300, 0x1032F)},
	{name: "General Unicode Symbols", ranges: []runeRange{{0x2000, 0x2BFF}, {0x1F300, 0x1FAFF}}},
	{name: "All printable Unicode", ranges: block(0, unicode.MaxRune)},
}

func init() {
	special := languages[len(languages)-2:]
	all := append([]language(nil), languages[:len(languages)-2]...)
	known := make(map[string]int, len(all))
	for i, lang := range all {
		known[normalizeLanguageName(lang.name)] = i
	}
	aliases := map[string]string{
		"avestan":               "avesta",
		"oldpersian":            "ancientpersian",
		"inscriptionalparthian": "parthian",
		"inscriptionalpahlavi":  "pahlavi",
		"psalterpahlavi":        "pahlavi",
		"hiragana":              "japanesehiragana",
		"katakana":              "japanesekatakana",
		"han":                   "chinesehancjk",
	}
	scriptNames := make([]string, 0, len(unicode.Scripts))
	for scriptName := range unicode.Scripts {
		scriptNames = append(scriptNames, scriptName)
	}
	sort.Strings(scriptNames)
	for _, scriptName := range scriptNames {
		if scriptName == "Common" || scriptName == "Inherited" {
			continue
		}
		normalized := normalizeLanguageName(scriptName)
		if alias, ok := aliases[normalized]; ok {
			normalized = alias
		}
		if index, exists := known[normalized]; exists {
			all[index].scripts = append(all[index].scripts, unicode.Scripts[scriptName])
			continue
		}
		name := strings.ReplaceAll(scriptName, "_", " ")
		all = append(all, language{name: name, scripts: []*unicode.RangeTable{unicode.Scripts[scriptName]}})
		known[normalizeLanguageName(name)] = len(all) - 1
	}
	languages = append(all, special...)
}

func normalizeLanguageName(name string) string {
	name = strings.ToLower(name)
	name = strings.NewReplacer(" ", "", "-", "", "_", "", "'", "").Replace(name)
	return name
}

func charactersFor(lang language) []rune {
	var chars []rune
	hasSpace := false
	seen := make(map[rune]struct{})
	appendCharacter := func(codepoint rune) {
		if !unicode.IsPrint(codepoint) {
			return
		}
		if _, exists := seen[codepoint]; exists {
			return
		}
		seen[codepoint] = struct{}{}
		chars = append(chars, codepoint)
		hasSpace = hasSpace || codepoint == ' '
	}
	for _, r := range lang.ranges {
		for codepoint := r.first; codepoint <= r.last; codepoint++ {
			appendCharacter(codepoint)
		}
	}
	for _, script := range lang.scripts {
		for _, r := range script.R16 {
			for codepoint := rune(r.Lo); codepoint <= rune(r.Hi); codepoint += rune(r.Stride) {
				appendCharacter(codepoint)
			}
		}
		for _, r := range script.R32 {
			for codepoint := rune(r.Lo); codepoint <= rune(r.Hi); codepoint += rune(r.Stride) {
				appendCharacter(codepoint)
			}
		}
	}
	sort.Slice(chars, func(i, j int) bool { return chars[i] < chars[j] })
	if !hasSpace {
		chars = append(chars, ' ')
	}
	return chars
}
