package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type metallicTheme struct {
	mode      string
	font      fyne.Resource
	fontScale float32
}

func (t metallicTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if t.mode == "System" {
		return theme.DefaultTheme().Color(name, variant)
	}
	if t.mode == "Metallic Light" {
		switch name {
		case theme.ColorNameBackground:
			return color.NRGBA{R: 0xD8, G: 0xDC, B: 0xE0, A: 0xFF}
		case theme.ColorNameInputBackground:
			return color.NRGBA{R: 0xF1, G: 0xF3, B: 0xF5, A: 0xFF}
		case theme.ColorNameButton:
			return color.NRGBA{R: 0xC1, G: 0xC8, B: 0xCE, A: 0xFF}
		case theme.ColorNamePrimary:
			return color.NRGBA{R: 0x3F, G: 0x75, B: 0x91, A: 0xFF}
		case theme.ColorNameForeground:
			return color.NRGBA{R: 0x19, G: 0x22, B: 0x28, A: 0xFF}
		}
		return theme.DefaultTheme().Color(name, theme.VariantLight)
	}

	switch name {
	case theme.ColorNameBackground:
		return color.NRGBA{R: 0x20, G: 0x25, B: 0x2A, A: 0xFF}
	case theme.ColorNameInputBackground:
		return color.NRGBA{R: 0x15, G: 0x19, B: 0x1D, A: 0xFF}
	case theme.ColorNameButton:
		return color.NRGBA{R: 0x38, G: 0x40, B: 0x47, A: 0xFF}
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 0x75, G: 0xB5, B: 0xD1, A: 0xFF}
	case theme.ColorNameForeground:
		return color.NRGBA{R: 0xE5, G: 0xE9, B: 0xEC, A: 0xFF}
	case theme.ColorNameHover:
		return color.NRGBA{R: 0x4A, G: 0x55, B: 0x5E, A: 0xFF}
	}
	return theme.DefaultTheme().Color(name, theme.VariantDark)
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
