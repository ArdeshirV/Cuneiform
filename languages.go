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
	sample  string
	ranges  []runeRange
	scripts []*unicode.RangeTable
}

var languages = buildLanguages(appConfig)

func buildLanguages(config appConfiguration) []language {
	var entries []language
	var afterScripts []language
	known := make(map[string]int, len(config.Languages.Languages))
	for _, definition := range config.Languages.Languages {
		entry := language{
			name:   definition.Name,
			sample: definition.Sample,
		}
		for _, bounds := range definition.Ranges {
			entry.ranges = append(entry.ranges, runeRange{first: rune(bounds[0]), last: rune(bounds[1])})
		}
		if definition.AfterUnicodeScripts {
			afterScripts = append(afterScripts, entry)
			continue
		}
		known[normalizeLanguageName(entry.name)] = len(entries)
		entries = append(entries, entry)
	}

	scriptNames := make([]string, 0, len(unicode.Scripts))
	for scriptName := range unicode.Scripts {
		scriptNames = append(scriptNames, scriptName)
	}
	sort.Strings(scriptNames)
	for _, scriptName := range scriptNames {
		if containsString(config.ExcludedScripts, scriptName) {
			continue
		}
		normalized := normalizeLanguageName(scriptName)
		if alias, exists := config.Languages.Aliases[normalized]; exists {
			normalized = alias
		}
		if index, exists := known[normalized]; exists {
			entries[index].scripts = append(entries[index].scripts, unicode.Scripts[scriptName])
			continue
		}
		name := strings.NewReplacer(flattenReplacements(config.LanguageNameReplacements)...).Replace(scriptName)
		known[normalizeLanguageName(name)] = len(entries)
		entries = append(entries, language{
			name:    name,
			scripts: []*unicode.RangeTable{unicode.Scripts[scriptName]},
		})
	}
	return append(entries, afterScripts...)
}

func normalizeLanguageName(name string) string {
	name = strings.ToLower(name)
	return strings.NewReplacer(flattenReplacements(appConfig.LanguageNormalizationReplacements)...).Replace(name)
}

func flattenReplacements(replacements [][2]string) []string {
	values := make([]string, 0, len(replacements)*2)
	for _, pair := range replacements {
		values = append(values, pair[0], pair[1])
	}
	return values
}

func containsString(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}

func languageOptionName(lang language) string {
	sample := lang.sample
	if sample == "" {
		chars := charactersFor(lang)
		runes := make([]rune, 0, int(configNumber("language_preview_characters")))
		for _, char := range chars {
			if char == ' ' || unicode.Is(unicode.M, char) {
				continue
			}
			runes = append(runes, char)
			if len(runes) == cap(runes) {
				break
			}
		}
		sample = string(runes)
	}
	if sample == "" {
		return lang.name
	}
	return formatConfigString("language_sample_format", lang.name, sample)
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
