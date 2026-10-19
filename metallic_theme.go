package main

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type metallicTheme struct {
	mode      string
	font      fyne.Resource
	fontScale float32
}

func (t metallicTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if t.mode == configString("theme_system") {
		return theme.DefaultTheme().Color(name, variant)
	}
	palette, exists := appConfig.ThemePalettes[t.mode]
	if !exists {
		return theme.DefaultTheme().Color(name, variant)
	}
	paletteVariant := theme.VariantDark
	if palette.Variant == configString("theme_variant_light") {
		paletteVariant = theme.VariantLight
	}
	var value string
	switch name {
	case theme.ColorNameBackground:
		value = palette.Background
	case theme.ColorNameInputBackground:
		value = palette.Input
	case theme.ColorNameButton:
		value = palette.Button
	case theme.ColorNamePrimary:
		value = palette.Primary
	case theme.ColorNameForeground:
		value = palette.Foreground
	case theme.ColorNameHover:
		value = palette.Hover
	}
	if value == "" {
		return theme.DefaultTheme().Color(name, paletteVariant)
	}
	parsed, err := parseThemeColor(value)
	if err != nil {
		panic(fmt.Errorf(configString("theme_color_invalid"), value, err))
	}
	return parsed
}

func parseThemeColor(value string) (color.NRGBA, error) {
	raw := strings.TrimPrefix(value, configString("color_hex_prefix"))
	if len(raw) != int(configNumber("color_hex_digits")) {
		return color.NRGBA{}, fmt.Errorf(configString("theme_color_hex_digits"), int(configNumber("color_hex_digits")))
	}
	parsed, err := strconv.ParseUint(raw, int(configNumber("color_hex_base")), int(configNumber("color_bits")))
	if err != nil {
		return color.NRGBA{}, err
	}
	return color.NRGBA{
		R: uint8(parsed >> uint(configNumber("color_red_shift"))),
		G: uint8(parsed >> uint(configNumber("color_green_shift"))),
		B: uint8(parsed >> uint(configNumber("color_blue_shift"))),
		A: uint8(parsed),
	}, nil
}

func (t metallicTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t metallicTheme) Font(style fyne.TextStyle) fyne.Resource {
	if t.font != nil {
		return t.font
	}
	return theme.DefaultTheme().Font(style)
}

func (t metallicTheme) Size(name fyne.ThemeSizeName) float32 {
	size := theme.DefaultTheme().Size(name)
	if name == theme.SizeNameText || name == theme.SizeNameHeadingText || name == theme.SizeNameCaptionText {
		return size * t.fontScale
	}
	return size
}
