package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"
	"unicode/utf8"
)

//go:embed assets/app.json assets/languages.json assets/readings.json
var configurationAssets embed.FS

type appConfiguration struct {
	Application struct {
		ID          string   `json:"id"`
		Name        string   `json:"name"`
		Version     string   `json:"version"`
		Icon        string   `json:"icon"`
		ProjectURL  string   `json:"project_url"`
		Copyright   string   `json:"copyright"`
		ProjectLink *url.URL `json:"-"`
	} `json:"application"`
	Strings                           map[string]string       `json:"strings"`
	Numbers                           map[string]float64      `json:"numbers"`
	Preferences                       map[string]string       `json:"preferences"`
	ThemeOptions                      []string                `json:"theme_options"`
	FontSizeOptions                   []string                `json:"font_size_options"`
	ThemePalettes                     map[string]themePalette `json:"theme_palettes"`
	RTLDefault                        bool                    `json:"right_to_left_default"`
	ExcludedScripts                   []string                `json:"excluded_unicode_scripts"`
	LanguageNameReplacements          [][2]string             `json:"language_name_replacements"`
	LanguageNormalizationReplacements [][2]string             `json:"language_normalization_replacements"`
	Languages                         languageConfiguration   `json:"-"`
	Readings                          map[string]readingSet   `json:"-"`
}

type themePalette struct {
	Variant    string `json:"variant"`
	Background string `json:"background"`
	Input      string `json:"input"`
	Button     string `json:"button"`
	Primary    string `json:"primary"`
	Foreground string `json:"foreground"`
	Hover      string `json:"hover"`
}

type languageConfiguration struct {
	Languages []languageDefinition `json:"languages"`
	Aliases   map[string]string    `json:"aliases"`
}

type languageDefinition struct {
	Name                string   `json:"name"`
	Ranges              [][2]int `json:"ranges"`
	Sample              string   `json:"sample"`
	AfterUnicodeScripts bool     `json:"after_unicode_scripts"`
}

type readingSet struct {
	Ranges        []readingRange    `json:"ranges"`
	Special       map[string]string `json:"special"`
	SpecialByRune map[rune]string
}

type readingRange struct {
	Start     string   `json:"start"`
	End       string   `json:"end"`
	Readings  []string `json:"readings"`
	FirstRune rune
	LastRune  rune
}

var appConfig = mustLoadAppConfiguration()

func mustLoadAppConfiguration() appConfiguration {
	config, err := loadAppConfiguration()
	if err != nil {
		panic(fmt.Errorf("load embedded application assets: %w", err))
	}
	return config
}

func loadAppConfiguration() (appConfiguration, error) {
	var config appConfiguration
	if err := decodeAsset("assets/app.json", &config); err != nil {
		return appConfiguration{}, err
	}
	if err := decodeAsset("assets/languages.json", &config.Languages); err != nil {
		return appConfiguration{}, err
	}
	if err := decodeAsset("assets/readings.json", &config.Readings); err != nil {
		return appConfiguration{}, err
	}
	if len(config.Strings) == 0 || len(config.Numbers) == 0 || len(config.Languages.Languages) == 0 {
		return appConfiguration{}, fmt.Errorf("application configuration is incomplete")
	}
	projectLink, err := url.Parse(config.Application.ProjectURL)
	if err != nil || projectLink.Scheme == "" || projectLink.Host == "" {
		return appConfiguration{}, fmt.Errorf("application project link is invalid")
	}
	config.Application.ProjectLink = projectLink
	number := func(key string) int {
		return int(config.Numbers[key])
	}
	for languageName, set := range config.Readings {
		set.SpecialByRune = make(map[rune]string, len(set.Special))
		for rawCodepoint, reading := range set.Special {
			codepoint, err := parseReadingCodepoint(rawCodepoint, config.Strings["unicode_codepoint_prefix"],
				number("numeric_base_hex"), number("unicode_scalar_bit_size"))
			if err != nil || !utf8.ValidRune(codepoint) {
				return appConfiguration{}, fmt.Errorf("%s special reading code point %q is invalid", languageName, rawCodepoint)
			}
			set.SpecialByRune[codepoint] = reading
		}
		for i := range set.Ranges {
			reading := &set.Ranges[i]
			first, err := strconv.ParseInt(reading.Start, number("numeric_base_hex"), number("unicode_scalar_bit_size"))
			if err != nil {
				return appConfiguration{}, fmt.Errorf("%s reading range start %q: %w", languageName, reading.Start, err)
			}
			last, err := strconv.ParseInt(reading.End, number("numeric_base_hex"), number("unicode_scalar_bit_size"))
			if err != nil || last < first || int64(len(reading.Readings)) != last-first+1 {
				return appConfiguration{}, fmt.Errorf("%s reading range %s-%s has invalid bounds or reading count", languageName, reading.Start, reading.End)
			}
			reading.FirstRune, reading.LastRune = rune(first), rune(last)
		}
		config.Readings[languageName] = set
	}
	return config, nil
}

func decodeAsset(name string, target any) error {
	file, err := configurationAssets.Open(name)
	if err != nil {
		return fmt.Errorf("open %s: %w", name, err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("decode %s: %w", name, err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return fmt.Errorf("decode %s: trailing JSON data", name)
		}
		return fmt.Errorf("decode %s: %w", name, err)
	}
	return nil
}

func configString(key string) string {
	value, exists := appConfig.Strings[key]
	if !exists {
		panic(fmt.Sprintf("missing string asset %q", key))
	}
	return value
}

func configNumber(key string) float64 {
	value, exists := appConfig.Numbers[key]
	if !exists {
		panic(fmt.Sprintf("missing number asset %q", key))
	}
	return value
}

func formatConfigString(key string, values ...any) string {
	return fmt.Sprintf(configString(key), values...)
}

func preferenceKey(key string) string {
	value, exists := appConfig.Preferences[key]
	if !exists {
		panic(fmt.Sprintf("missing preference asset %q", key))
	}
	return value
}

func parseReadingCodepoint(value, prefix string, base, bitSize int) (rune, error) {
	codepoint, err := strconv.ParseInt(strings.TrimPrefix(value, prefix), base, bitSize)
	return rune(codepoint), err
}
