package main

import (
	_ "embed"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
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

//go:embed Icon.png
var appIconBytes []byte

//go:embed assets/fonts/UnicodeScriptFallback.ttf
var unicodeFontBytes []byte

//go:embed assets/ethereum-donation.png
var ethereumDonationImageBytes []byte

type editor struct {
	application fyne.App
	window      fyne.Window
	text        *widget.Entry
	status      *widget.Label
	editorHost  *fyne.Container
	keyboard    *characterKeyboard
	rightToLeft bool
	encoding    textEncoding
	uri         fyne.URI
	dirty       bool
	loading     bool
}

func keyboardColumns() int {
	return int(configNumber("keyboard_columns"))
}

func main() {
	application := app.NewWithID(appConfig.Application.ID)
	window := application.NewWindow(appConfig.Application.Name)
	window.Resize(fyne.NewSize(float32(configNumber("initial_window_width")), float32(configNumber("initial_window_height"))))
	appIcon := fyne.NewStaticResource(appConfig.Application.Icon, appIconBytes)
	window.SetIcon(appIcon)

	e := &editor{application: application, window: window, encoding: encodingUTF8}
	e.applySettings()
	rightToLeft := widget.NewCheck(configString("right_to_left"), nil)
	rightToLeft.SetChecked(application.Preferences().BoolWithFallback(preferenceKey("right_to_left"), appConfig.RTLDefault))
	e.rightToLeft = rightToLeft.Checked
	e.text = newEditorEntry(e.rightToLeft)
	e.text.SetPlaceHolder(configString("placeholder"))
	e.bindTextEntry()
	rightToLeft.OnChanged = func(enabled bool) {
		application.Preferences().SetBool(preferenceKey("right_to_left"), enabled)
		e.setTextDirection(enabled)
	}
	e.status = widget.NewLabel(configString("empty"))

	languageSelect := widget.NewSelect(languageNames(), nil)
	languageSelect.SetSelected(languageOptionName(languages[0]))
	e.keyboard = newKeyboard(languages[0], e.text)
	languageSelect.OnChanged = func(name string) {
		for _, lang := range languages {
			if languageOptionName(lang) == name {
				e.keyboard.SetList(lang, e.text)
				break
			}
		}
	}

	openButton := widget.NewButton(configString("open"), e.open)
	saveButton := widget.NewButton(configString("save"), e.save)
	saveAsButton := widget.NewButton(configString("save_as"), e.saveAs)
	settingsButton := widget.NewButton(configString("settings"), e.showSettingsDialog)
	aboutButton := widget.NewButton(configString("about"), e.showAbout)
	exitButton := widget.NewButton(configString("exit"), e.requestClose)
	top := container.NewVBox(
		container.NewBorder(nil, nil,
			container.NewHBox(openButton, saveButton, saveAsButton, settingsButton, aboutButton, rightToLeft),
			exitButton, nil),
		container.NewBorder(nil, nil, widget.NewLabel(configString("language")), nil, languageSelect),
	)
	e.editorHost = container.NewMax(e.text)
	content := container.NewVSplit(e.editorHost, e.keyboard)
	preferences := application.Preferences()
	content.SetOffset(preferences.FloatWithFallback(preferenceKey("split_offset"), configNumber("default_editor_fraction")))
	window.SetContent(container.NewBorder(top, e.status, nil, nil, content))
	window.Canvas().Focus(e.text)
	stopSavingSplit := make(chan struct{})
	window.SetOnClosed(func() {
		preferences.SetFloat(preferenceKey("split_offset"), content.Offset)
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
	entry        *widget.Entry
	languageName string
}

func newKeyboard(lang language, entry *widget.Entry) *characterKeyboard {
	k := &characterKeyboard{entry: entry}
	k.list = widget.NewList(
		func() int { return (len(k.chars) + keyboardColumns() - 1) / keyboardColumns() },
		func() fyne.CanvasObject {
			buttons := make([]fyne.CanvasObject, keyboardColumns())
			for i := range buttons {
				buttons[i] = widget.NewButton(configString("empty"), nil)
			}
			return container.NewGridWithColumns(keyboardColumns(), buttons...)
		},
		func(id widget.ListItemID, object fyne.CanvasObject) {
			grid := object.(*fyne.Container)
			for col, object := range grid.Objects {
				button := object.(*widget.Button)
				index := int(id)*keyboardColumns() + col
				if index >= len(k.chars) {
					button.Hide()
					button.OnTapped = nil
					continue
				}
				char := k.chars[index]
				button.Show()
				button.SetText(keyboardKeyText(char, phoneticFor(k.languageName, char)))
				button.OnTapped = func() {
					if canvasForEditor := fyne.CurrentApp().Driver().CanvasForObject(k.entry); canvasForEditor != nil {
						canvasForEditor.Focus(k.entry)
					}
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

func (k *characterKeyboard) SetList(lang language, entry *widget.Entry) {
	k.entry = entry
	k.languageName = lang.name
	k.chars = charactersFor(lang)
	k.list.Refresh()
}

func (k *characterKeyboard) SetEntry(entry *widget.Entry) {
	k.entry = entry
}

func keyLabel(char rune) string {
	if char == ' ' {
		return configString("space_key")
	}
	if unicode.Is(unicode.M, char) {
		return configString("combining_circle") + string(char)
	}
	return string(char)
}

func keyboardKeyText(char rune, pronunciation string) string {
	return configString("key_label_prefix") + keyLabel(char) +
		configString("key_label_separator") + strings.TrimSpace(pronunciation)
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
	saveDialog.SetFileName(configString("untitled_filename"))
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
	name := configString("untitled")
	if e.uri != nil {
		name = e.uri.Name()
	}
	marker := configString("empty")
	if e.dirty {
		marker = configString("dirty_marker")
	}
	e.status.SetText(formatConfigString("status_format",
		name, marker, e.encoding, unicode.Version, utf8RuneCount(e.text.Text)))
}

func (e *editor) bindTextEntry() {
	e.text.OnChanged = func(string) {
		if !e.loading {
			e.dirty = true
		}
		e.updateStatus()
	}
}

func (e *editor) setTextDirection(rightToLeft bool) {
	if e.rightToLeft == rightToLeft {
		return
	}

	text := e.text.Text
	row, column := e.text.CursorRow, e.text.CursorColumn
	hadDirectionMark := false
	if e.rightToLeft {
		lines := strings.Split(text, "\n")
		mark := configString("right_to_left_mark")
		if row >= 0 && row < len(lines) {
			hadDirectionMark = strings.HasPrefix(lines[row], mark)
		}
		text = stripEditorDirectionMarks(text)
	}
	if hadDirectionMark && column > 0 {
		column--
	}
	entry := newEditorEntry(rightToLeft)
	entry.SetPlaceHolder(configString("placeholder"))
	e.loading = true
	entry.SetText(text)
	e.loading = false
	entry.CursorRow, entry.CursorColumn = row, column
	e.text = entry
	e.rightToLeft = rightToLeft
	e.bindTextEntry()
	e.editorHost.Objects = []fyne.CanvasObject{entry}
	e.editorHost.Refresh()
	e.keyboard.SetEntry(entry)
	if canvas := e.window.Canvas(); canvas != nil {
		canvas.Focus(entry)
	}
}

func newEditorEntry(rightToLeft bool) *widget.Entry {
	style := widget.RichTextStyleInline
	style.Alignment = editorTextAlignment(rightToLeft)
	originalStyle := widget.RichTextStyleInline
	widget.RichTextStyleInline = style
	entry := widget.NewMultiLineEntry()
	entry.Wrapping = fyne.TextWrapWord
	entry.Scroll = fyne.ScrollVerticalOnly
	entry.SetText(configString("empty"))
	widget.RichTextStyleInline = originalStyle
	return entry
}

func editorTextAlignment(rightToLeft bool) fyne.TextAlign {
	if rightToLeft {
		return fyne.TextAlignTrailing
	}
	return fyne.TextAlignLeading
}

func stripEditorDirectionMarks(text string) string {
	mark := configString("right_to_left_mark")
	lines := strings.Split(text, "\n")
	for i := range lines {
		lines[i] = strings.TrimPrefix(lines[i], mark)
	}
	return strings.Join(lines, "\n")
}

func (e *editor) applySettings() {
	preferences := e.application.Preferences()
	scaleChoice := preferences.StringWithFallback(preferenceKey("font_size"), configString("font_size_medium"))
	scale := float32(configNumber("font_scale_medium"))
	switch scaleChoice {
	case configString("font_size_small"):
		scale = float32(configNumber("font_scale_small"))
	case configString("font_size_large"):
		scale = float32(configNumber("font_scale_large"))
	}
	font := fyne.Resource(fyne.NewStaticResource(configString("font_bundled_name"), unicodeFontBytes))
	if encodedFont := preferences.String(preferenceKey("custom_font_data")); encodedFont != "" {
		fontData, err := base64.StdEncoding.DecodeString(encodedFont)
		if err != nil {
			dialog.ShowError(fmt.Errorf(configString("font_load_error"), err), e.window)
			preferences.SetString(preferenceKey("custom_font_name"), "")
			preferences.SetString(preferenceKey("custom_font_data"), "")
		} else {
			font = fyne.NewStaticResource(preferences.StringWithFallback(preferenceKey("custom_font_name"), configString("font_file_default_name")), fontData)
		}
	}
	e.application.Settings().SetTheme(metallicTheme{
		mode:      preferences.StringWithFallback(preferenceKey("theme"), configString("theme_dark")),
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
		fontLimit := int64(configNumber("default_font_file_limit_mib") * configNumber("bytes_per_mib"))
		fontData, err := io.ReadAll(io.LimitReader(reader, fontLimit+1))
		if err != nil {
			dialog.ShowError(err, e.window)
			return
		}
		if int64(len(fontData)) > fontLimit {
			dialog.ShowError(fmt.Errorf(configString("font_file_too_large"), int(configNumber("default_font_file_limit_mib"))), e.window)
			return
		}
		if _, err := sfnt.Parse(fontData); err != nil {
			dialog.ShowError(fmt.Errorf(configString("font_file_invalid"), err), e.window)
			return
		}
		preferences := e.application.Preferences()
		preferences.SetString(preferenceKey("custom_font_name"), filepath.Base(reader.URI().Name()))
		preferences.SetString(preferenceKey("custom_font_data"), base64.StdEncoding.EncodeToString(fontData))
		e.applySettings()
	}, e.window)
	fontDialog.SetFilter(storage.NewExtensionFileFilter([]string{configString("font_extension_ttf"), configString("font_extension_otf")}))
	fontDialog.Show()
}

func (e *editor) showSettingsDialog() {
	preferences := e.application.Preferences()
	themeSelection := widget.NewSelect(appConfig.ThemeOptions, nil)
	themeSelection.SetSelected(preferences.StringWithFallback(preferenceKey("theme"), configString("theme_dark")))
	themeSelection.OnChanged = func(value string) {
		preferences.SetString(preferenceKey("theme"), value)
		e.applySettings()
	}
	sizeSelection := widget.NewSelect(appConfig.FontSizeOptions, nil)
	sizeSelection.SetSelected(preferences.StringWithFallback(preferenceKey("font_size"), configString("font_size_medium")))
	sizeSelection.OnChanged = func(value string) {
		preferences.SetString(preferenceKey("font_size"), value)
		e.applySettings()
	}
	content := container.NewVBox(
		widget.NewLabel(configString("theme_settings_label")),
		themeSelection,
		widget.NewLabel(configString("font_settings_label")),
		sizeSelection,
		widget.NewButton(configString("font"), e.showFontDialog),
		widget.NewButton(configString("font_bundled_button"), func() {
			preferences := e.application.Preferences()
			preferences.SetString(preferenceKey("custom_font_name"), "")
			preferences.SetString(preferenceKey("custom_font_data"), "")
			e.applySettings()
		}),
	)
	dialog.NewCustom(configString("settings"), configString("settings_close"), content, e.window).Show()
}

func (e *editor) showAbout() {
	icon := canvas.NewImageFromResource(fyne.NewStaticResource(appConfig.Application.Icon, appIconBytes))
	icon.FillMode = canvas.ImageFillContain
	icon.SetMinSize(fyne.NewSize(float32(configNumber("about_icon_size")), float32(configNumber("about_icon_size"))))
	appName := canvas.NewText(appConfig.Application.Name, theme.ForegroundColor())
	appName.TextSize = theme.Size(theme.SizeNameText) * float32(configNumber("about_name_font_multiplier"))
	version := canvas.NewText(configString("about_version")+appConfig.Application.Version, theme.ForegroundColor())
	version.TextSize = theme.Size(theme.SizeNameCaptionText)
	appDetails := container.NewVBox(appName, version)
	donationImage := canvas.NewImageFromResource(fyne.NewStaticResource(configString("donation_image_name"), ethereumDonationImageBytes))
	donationImage.FillMode = canvas.ImageFillContain
	donationImage.SetMinSize(fyne.NewSquareSize(float32(configNumber("donation_image_size"))))
	donationTitle := canvas.NewText(configString("donation_title"), theme.ForegroundColor())
	donationTitle.TextSize = theme.Size(theme.SizeNameText) * float32(configNumber("donation_title_font_multiplier"))
	donationTitle.TextStyle.Bold = true
	donationSection := container.NewVBox(
		donationTitle,
		widget.NewLabel(configString("donation_description")),
		container.NewCenter(donationImage),
		container.NewHBox(
			widget.NewLabel(configString("donation_address")),
			widget.NewButton(configString("clipboard_copy"), func() {
				e.window.Clipboard().SetContent(configString("donation_clipboard_address"))
			}),
		),
	)
	copyrightBefore, copyrightAfter, hasAuthorHandle := strings.Cut(appConfig.Application.Copyright, appConfig.Application.AuthorHandle)
	copyrightRow := fyne.CanvasObject(widget.NewLabel(appConfig.Application.Copyright))
	if hasAuthorHandle {
		copyrightRow = container.NewHBox(
			widget.NewLabel(copyrightBefore),
			widget.NewHyperlink(appConfig.Application.AuthorHandle, appConfig.Application.AuthorLink),
			widget.NewLabel(copyrightAfter),
		)
	}
	var aboutDialog dialog.Dialog
	closeButton := widget.NewButton(configString("close"), func() {
		aboutDialog.Hide()
	})
	closeButtonRow := container.NewCenter(container.NewGridWrap(
		fyne.NewSize(float32(configNumber("about_close_button_width")), closeButton.MinSize().Height),
		closeButton,
	))
	details := container.NewVBox(
		container.NewHBox(icon, appDetails),
		copyrightRow,
		widget.NewSeparator(),
		donationSection,
		widget.NewSeparator(),
		widget.NewLabel(configString("project_link_description")),
		widget.NewHyperlink(appConfig.Application.ProjectURL, appConfig.Application.ProjectLink),
		closeButtonRow,
	)
	aboutDialog = dialog.NewCustomWithoutButtons(configString("about_dialog_title"), details, e.window)
	aboutDialog.Show()
}

func (e *editor) requestClose() {
	if !e.dirty {
		e.quit()
		return
	}

	confirmation := dialog.NewCustomWithoutButtons(
		configString("save_changes_title"),
		widget.NewLabel(configString("save_changes_message")),
		e.window,
	)
	confirmation.SetButtons([]fyne.CanvasObject{
		widget.NewButton(configString("yes"), func() {
			confirmation.Hide()
			e.saveWithCompletion(func(saved bool) {
				if saved {
					e.quit()
				}
			})
		}),
		widget.NewButton(configString("no"), func() {
			confirmation.Hide()
			e.quit()
		}),
		widget.NewButton(configString("cancel"), confirmation.Hide),
	})
	confirmation.Show()
}

func (e *editor) quit() {
	e.window.SetCloseIntercept(nil)
	e.application.Quit()
}

func persistSplitOffset(split *container.Split, preferences fyne.Preferences, done <-chan struct{}) {
	ticker := time.NewTicker(time.Duration(configNumber("splitter_save_interval_millis")) * time.Millisecond)
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
					preferences.SetFloat(preferenceKey("split_offset"), last)
				}
			})
		}
	}
}

func utf8RuneCount(text string) int {
	return len([]rune(strings.ToValidUTF8(text, configString("invalid_utf8_replacement"))))
}
