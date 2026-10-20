package main

import "testing"

func TestEmbeddedConfigurationLoadsLanguageAndReadingAssets(t *testing.T) {
	config, err := loadAppConfiguration()
	if err != nil {
		t.Fatalf("load embedded configuration: %v", err)
	}
	if got, want := config.Languages.Languages[0].Name, "Avesta"; got != want {
		t.Fatalf("first language = %q, want %q", got, want)
	}
	if got := config.Readings["Avesta"].Ranges[0].Readings[0]; got != "a" {
		t.Fatalf("first Avesta reading = %q, want %q", got, "a")
	}
	if got, want := config.Application.AuthorLink.String(), "https://ardeshirv.github.io/"; got != want {
		t.Fatalf("author link = %q, want %q", got, want)
	}
	if got, want := config.Application.AuthorHandle, "@ArdeshirV"; got != want {
		t.Fatalf("author handle = %q, want %q", got, want)
	}
	if len(config.ThemePalettes) == 0 || len(config.Strings) == 0 || len(config.Numbers) == 0 {
		t.Fatal("embedded app settings are incomplete")
	}
}

func TestConfiguredThemeColorsAreValid(t *testing.T) {
	if _, err := parseThemeColor(configString("selection_highlight")); err != nil {
		t.Fatalf("parse selection color: %v", err)
	}
	for themeName, palette := range appConfig.ThemePalettes {
		for _, value := range []string{palette.Background, palette.Input, palette.Button, palette.Primary, palette.Foreground, palette.Hover} {
			if value == "" {
				continue
			}
			if _, err := parseThemeColor(value); err != nil {
				t.Errorf("theme %q has invalid color %q: %v", themeName, value, err)
			}
		}
	}
}
