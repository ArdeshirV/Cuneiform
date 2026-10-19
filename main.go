package main

import (
	_ "embed"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"golang.org/x/image/font/sfnt"
)

const (
	keyboardColumns = 5
	appVersion      = "1.0"
	bundledFontName = "Unicode Script Fallback (bundled)"
	systemFontName  = "System default"
)

//go:embed Icon.png
var appIconBytes []byte

//go:embed assets/fonts/UnicodeScriptFallback.ttf
var unicodeFontBytes []byte

type editor struct {
	application fyne.App
	window      fyne.Window
	text        *unicodeEditor
	status      *widget.Label
	encoding    textEncoding
	uri         fyne.URI
	dirty       bool
	loading     bool
}

func main() {
	application := app.NewWithID("org.ardeshirv.cuneiform")
	window := application.NewWindow("Cuneiform")
	window.Resize(fyne.NewSize(1000, 760))
	appIcon := fyne.NewStaticResource("Icon.png", appIconBytes)
	window.SetIcon(appIcon)

	e := &editor{application: application, window: window, encoding: encodingUTF8}
	e.applySettings()
	e.text = newUnicodeEditor()
	e.text.SetPlaceHolder("Open or start a Unicode text document…")
	e.status = widget.NewLabel("")
	e.text.OnChanged = func(string) {
		if !e.loading {
			e.dirty = true
		}
		e.updateStatus()
	}

	languageSelect := widget.NewSelect(languageNames(), nil)
	languageSelect.SetSelected(languageOptionName(languages[0]))
	keyboard := newKeyboard(languages[0], e.text)
	rightToLeft := widget.NewCheck("Right to left", nil)
	rightToLeft.SetChecked(application.Preferences().BoolWithFallback("right_to_left", true))
	e.text.SetRightToLeft(rightToLeft.Checked)
	rightToLeft.OnChanged = func(enabled bool) {
		application.Preferences().SetBool("right_to_left", enabled)
		e.text.SetRightToLeft(enabled)
	}
	languageSelect.OnChanged = func(name string) {
		for _, lang := range languages {
			if languageOptionName(lang) == name {
				keyboard.SetList(lang, e.text)
				break
			}
		}
	}

	openButton := widget.NewButton("Open", e.open)
	saveButton := widget.NewButton("Save", e.save)
	saveAsButton := widget.NewButton("Save As", e.saveAs)
	fontButton := widget.NewButton("Font", e.showFontDialog)
	settingsButton := widget.NewButton("Settings", e.showSettingsDialog)
	aboutButton := widget.NewButton("About", e.showAbout)
	exitButton := widget.NewButton("Exit", e.requestClose)
	top := container.NewVBox(
		container.NewBorder(nil, nil,
			container.NewHBox(openButton, saveButton, saveAsButton, fontButton, settingsButton, aboutButton, rightToLeft),
			exitButton, nil),
		container.NewBorder(nil, nil, widget.NewLabel("Language:"), nil, languageSelect),
	)
	content := container.NewVSplit(e.text, keyboard)
	preferences := application.Preferences()
	content.SetOffset(preferences.FloatWithFallback("split_offset", 0.3))
	window.SetContent(container.NewBorder(top, e.status, nil, nil, content))
	stopSavingSplit := make(chan struct{})
	window.SetOnClosed(func() {
		preferences.SetFloat("split_offset", content.Offset)
		close(stopSavingSplit)
	})
	window.SetCloseIntercept(e.requestClose)
	go persistSplitOffset(content, preferences, stopSavingSplit)
	e.updateStatus()
	window.ShowAndRun()
}

func languageNames() []string {
	names := make([]string, len(languages))
	for i, lang := range languages {
		names[i] = languageOptionName(lang)
	}
	return names
}

type characterKeyboard struct {
	widget.BaseWidget
	list         *widget.List
	chars        []rune
	entry        *unicodeEditor
	languageName string
}

func newKeyboard(lang language, entry *unicodeEditor) *characterKeyboard {
	k := &characterKeyboard{entry: entry}
	k.list = widget.NewList(
		func() int { return (len(k.chars) + keyboardColumns - 1) / keyboardColumns },
		func() fyne.CanvasObject {
			buttons := make([]fyne.CanvasObject, keyboardColumns)
			for i := range buttons {
				buttons[i] = widget.NewButton("", nil)
			}
			return container.NewGridWithColumns(keyboardColumns, buttons...)
		},
		func(id widget.ListItemID, object fyne.CanvasObject) {
			grid := object.(*fyne.Container)
			for col, object := range grid.Objects {
				button := object.(*widget.Button)
				index := int(id)*keyboardColumns + col
				if index >= len(k.chars) {
					button.Hide()
					button.OnTapped = nil
					continue
				}
				char := k.chars[index]
				button.Show()
				button.SetText(keyLabel(char) + " " + phoneticFor(k.languageName, char))
				button.OnTapped = func() {
					k.entry.TypedRune(char)
				}
			}
		},
	)
	k.ExtendBaseWidget(k)
	k.SetList(lang, entry)
	return k
}

func (k *characterKeyboard) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(k.list)
}

func (k *characterKeyboard) SetList(lang language, entry *unicodeEditor) {
	k.entry = entry
	k.languageName = lang.name
	k.chars = charactersFor(lang)
	k.list.Refresh()
}

func keyLabel(char rune) string {
	if char == ' ' {
		return "␠"
	}
	if unicode.Is(unicode.M, char) {
		return "◌" + string(char)
	}
	return string(char)
}

func (e *editor) open() {
	openDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, e.window)
			return
		}
		if reader == nil {
			return
		}
		defer reader.Close()
		data, err := io.ReadAll(reader)
		if err != nil {
			dialog.ShowError(err, e.window)
			return
		}
		text, encoding, err := decodeDocument(data)
		if err != nil {
			dialog.ShowError(err, e.window)
			return
		}
		e.loading = true
		e.text.SetText(text)
		e.loading = false
		e.uri = reader.URI()
		e.encoding = encoding
		e.dirty = false
		e.updateStatus()
	}, e.window)
	openDialog.Show()
}

func (e *editor) save() {
	e.saveWithCompletion(nil)
}

func (e *editor) saveWithCompletion(completion func(bool)) {
	if e.uri == nil {
		e.saveAsWithCompletion(completion)
		return
	}
	writer, err := storage.SaveFileToURI(e.uri)
	if err != nil {
		dialog.ShowError(err, e.window)
		if completion != nil {
			completion(false)
		}
		return
	}
	if err := writeDocument(writer, e.text.Text, e.encoding); err != nil {
		dialog.ShowError(err, e.window)
		if completion != nil {
			completion(false)
		}
		return
	}
	e.dirty = false
	e.updateStatus()
	if completion != nil {
		completion(true)
	}
}

func (e *editor) saveAs() {
	e.saveAsWithCompletion(nil)
}

func (e *editor) saveAsWithCompletion(completion func(bool)) {
	saveDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, e.window)
			if completion != nil {
				completion(false)
			}
			return
		}
		if writer == nil {
			if completion != nil {
				completion(false)
			}
			return
		}
		if err := writeDocument(writer, e.text.Text, e.encoding); err != nil {
			dialog.ShowError(err, e.window)
			if completion != nil {
				completion(false)
			}
			return
		}
		e.uri = writer.URI()
		e.dirty = false
		e.updateStatus()
		if completion != nil {
			completion(true)
		}
	}, e.window)
	saveDialog.SetFileName("untitled.txt")
	saveDialog.Show()
}

func writeDocument(writer fyne.URIWriteCloser, text string, encoding textEncoding) error {
	data := encodeDocument(text, encoding)
	written, writeErr := writer.Write(data)
	closeErr := writer.Close()
	if writeErr != nil {
		return errors.Join(writeErr, closeErr)
	}
	if written != len(data) {
		return errors.Join(io.ErrShortWrite, closeErr)
	}
	return closeErr
}

func (e *editor) updateStatus() {
	name := "Untitled"
	if e.uri != nil {
		name = e.uri.Name()
	}
	marker := ""
	if e.dirty {
		marker = " •"
	}
	e.status.SetText(fmt.Sprintf("%s%s   |   %s   |   Unicode data: %s   |   %d code points",
		name, marker, e.encoding, unicode.Version, utf8RuneCount(e.text.Text)))
}

func (e *editor) applySettings() {
	preferences := e.application.Preferences()
	scaleChoice := preferences.StringWithFallback("font_size", "Medium")
	scale := float32(1)
	switch scaleChoice {
	case "Small":
		scale = 0.85
	case "Large":
		scale = 1.15
	}
	font := fyne.Resource(fyne.NewStaticResource("UnicodeScriptFallback.ttf", unicodeFontBytes))
	if encodedFont := preferences.String("custom_font_data"); encodedFont != "" {
		fontData, err := base64.StdEncoding.DecodeString(encodedFont)
		if err != nil {
			dialog.ShowError(fmt.Errorf("load saved font: %w", err), e.window)
			preferences.SetString("custom_font_name", "")
			preferences.SetString("custom_font_data", "")
		} else {
			font = fyne.NewStaticResource(preferences.StringWithFallback("custom_font_name", "custom-font.ttf"), fontData)
		}
	}
	e.application.Settings().SetTheme(metallicTheme{
		mode:      preferences.StringWithFallback("theme", "Metallic Dark"),
		font:      font,
		fontScale: scale,
	})
}

func (e *editor) showFontDialog() {
	fontDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, e.window)
			return
		}
		if reader == nil {
			return
		}
		defer reader.Close()
		fontData, err := io.ReadAll(io.LimitReader(reader, (20<<20)+1))
		if err != nil {
			dialog.ShowError(err, e.window)
			return
		}
		if len(fontData) > 20<<20 {
			dialog.ShowError(errors.New("font file exceeds the 20 MiB limit"), e.window)
			return
		}
		if _, err := sfnt.Parse(fontData); err != nil {
			dialog.ShowError(fmt.Errorf("invalid TrueType/OpenType font: %w", err), e.window)
			return
		}
		preferences := e.application.Preferences()
		preferences.SetString("custom_font_name", filepath.Base(reader.URI().Name()))
		preferences.SetString("custom_font_data", base64.StdEncoding.EncodeToString(fontData))
		e.applySettings()
	}, e.window)
	fontDialog.SetFilter(storage.NewExtensionFileFilter([]string{".ttf", ".otf"}))
	fontDialog.Show()
}

func (e *editor) showSettingsDialog() {
	preferences := e.application.Preferences()
	themeSelection := widget.NewSelect([]string{"Metallic Dark", "Metallic Light", "System"}, nil)
	themeSelection.SetSelected(preferences.StringWithFallback("theme", "Metallic Dark"))
	themeSelection.OnChanged = func(value string) {
		preferences.SetString("theme", value)
		e.applySettings()
	}
	sizeSelection := widget.NewSelect([]string{"Small", "Medium", "Large"}, nil)
	sizeSelection.SetSelected(preferences.StringWithFallback("font_size", "Medium"))
	sizeSelection.OnChanged = func(value string) {
		preferences.SetString("font_size", value)
		e.applySettings()
	}
	content := container.NewVBox(
		widget.NewLabel("Theme"),
		themeSelection,
		widget.NewLabel("Text size"),
		sizeSelection,
		widget.NewButton("Use bundled font", func() {
			preferences := e.application.Preferences()
			preferences.SetString("custom_font_name", "")
			preferences.SetString("custom_font_data", "")
			e.applySettings()
		}),
	)
	dialog.NewCustom("Settings", "Close", content, e.window).Show()
}

func (e *editor) showAbout() {
	icon := canvas.NewImageFromResource(fyne.NewStaticResource("Icon.png", appIconBytes))
	icon.FillMode = canvas.ImageFillContain
	icon.SetMinSize(fyne.NewSize(72, 72))
	projectURL := &url.URL{Scheme: "https", Host: "github.com", Path: "/ArdeshirV/cuneiform"}
	appName := canvas.NewText("Cuneiform", theme.ForegroundColor())
	appName.TextSize = theme.Size(theme.SizeNameText) * 3
	version := canvas.NewText("Version: "+appVersion, theme.ForegroundColor())
	version.TextSize = theme.Size(theme.SizeNameCaptionText)
	version.Move(fyne.NewPos(0, -theme.Size(theme.SizeNameCaptionText)/2))
	versionHolder := container.NewWithoutLayout(version)
	details := container.NewVBox(
		container.NewHBox(icon, appName, versionHolder),
		widget.NewLabel("Copyright (c) 2026 @ArdeshirV, Licenced under GPLv3+"),
		widget.NewHyperlink("https://github.com/ArdeshirV/cuneiform", projectURL),
	)
	dialog.NewCustom("About", "Close", details, e.window).Show()
}

func (e *editor) requestClose() {
	if !e.dirty {
		e.quit()
		return
	}

	confirmation := dialog.NewCustomWithoutButtons(
		"Unsaved changes",
		widget.NewLabel("Do you want to save changes?"),
		e.window,
	)
	confirmation.SetButtons([]fyne.CanvasObject{
		widget.NewButton("Yes", func() {
			confirmation.Hide()
			e.saveWithCompletion(func(saved bool) {
				if saved {
					e.quit()
				}
			})
		}),
		widget.NewButton("No", func() {
			confirmation.Hide()
			e.quit()
		}),
		widget.NewButton("Cancel", confirmation.Hide),
	})
	confirmation.Show()
}

func (e *editor) quit() {
	e.window.SetCloseIntercept(nil)
	e.application.Quit()
}

func persistSplitOffset(split *container.Split, preferences fyne.Preferences, done <-chan struct{}) {
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()
	last := split.Offset
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			fyne.Do(func() {
				if split.Offset != last {
					last = split.Offset
					preferences.SetFloat("split_offset", last)
				}
			})
		}
	}
}

func utf8RuneCount(text string) int {
	return len([]rune(strings.ToValidUTF8(text, "\uFFFD")))
}
