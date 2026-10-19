package main

import "fmt"

func phoneticFor(languageName string, char rune) string {
	if char == ' ' {
		return configString("phonetic_space")
	}
	if set, exists := appConfig.Readings[languageName]; exists {
		if reading, found := set.SpecialByRune[char]; found {
			return reading
		}
		for _, readingRange := range set.Ranges {
			if char >= readingRange.FirstRune && char <= readingRange.LastRune {
				return readingRange.Readings[char-readingRange.FirstRune]
			}
		}
	}
	return fmt.Sprintf(configString("phonetic_unknown_format"), char)
}
