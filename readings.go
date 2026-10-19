package main

import (
	"fmt"
	"strings"
)

var avestanReadings = stringsAsReadings(
	"a aa ao aao an aan ae aee e ee o oo i ii u uu ke xe xye xve ge gge ghe ce je te the de dhe tte pe fe be bhe nge ngye ngve ne nye nne me hme yye ye ve re le se ze she zhe shye sshe he",
)

var oldPersianReadings = stringsAsReadings(
	"a i u ka ku ga gu xa ca ja ji ta tu da di du tha pa ba fa na nu ma mi mu ya va vi ra ru la sa za sha ssa ha",
)

var parthianReadings = stringsAsReadings(
	"a b g d h w z h t y k l m n s ʿ p ṣ q r š t",
)

var pahlaviReadings = stringsAsReadings(
	"a b g d h w-ʿ-r z h t y k l m-q n s p ṣ š t",
)

func stringsAsReadings(values string) []string {
	return strings.Fields(values)
}

func phoneticFor(languageName string, char rune) string {
	if char == ' ' {
		return "space"
	}
	var readings []string
	var start rune
	switch languageName {
	case "Avesta":
		switch {
		case 0x10B00 <= char && char <= 0x10B35:
			readings, start = avestanReadings, 0x10B00
		case char == 0x10B39:
			return "abbreviation"
		case 0x10B3A <= char && char <= 0x10B3F:
			return "punctuation"
		}
	case "Ancient Persian":
		switch {
		case 0x103A0 <= char && char <= 0x103C3:
			readings, start = oldPersianReadings, 0x103A0
		case 0x103C8 <= char && char <= 0x103CF:
			return []string{"auramazdaa", "auramazdaa-2", "auramazdaaha", "xshaayathiya", "dahyaaush", "dahyaaush-2", "baga", "buumish"}[char-0x103C8]
		case char == 0x103D0:
			return "word divider"
		case 0x103D1 <= char && char <= 0x103D5:
			return []string{"one", "two", "ten", "twenty", "hundred"}[char-0x103D1]
		}
	case "Parthian":
		switch {
		case 0x10B40 <= char && char <= 0x10B55:
			readings, start = parthianReadings, 0x10B40
		case 0x10B58 <= char && char <= 0x10B5F:
			return []string{"one", "two", "three", "four", "ten", "twenty", "one hundred", "one thousand"}[char-0x10B58]
		}
	case "Pahlavi":
		switch {
		case 0x10B60 <= char && char <= 0x10B72:
			readings, start = pahlaviReadings, 0x10B60
		case 0x10B78 <= char && char <= 0x10B7F:
			return []string{"one", "two", "three", "four", "ten", "twenty", "one hundred", "one thousand"}[char-0x10B78]
		}
	}
	if readings != nil {
		return readings[char-start]
	}
	return fmt.Sprintf("%X", char)
}
