package main

import (
	"unicode"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/driver/mobile"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type unicodeEditor struct {
	widget.BaseWidget
	Text        string
	PlaceHolder string
	OnChanged   func(string)

	rightToLeft bool
	cursor      int
	anchor      int
	shiftDown   bool
	focused     bool
	undoStack   []editAction
	redoStack   []editAction
	richText    *widget.RichText
	scroll      *container.Scroll
	caret       *canvas.Rectangle
	highlight   *canvas.Rectangle
}

func (e *unicodeEditor) scrollOffsetX() float32 {
	if e.scroll == nil {
		return 0
	}
	return e.scroll.Offset.X
}

type editAction struct {
	position int
	removed  string
	inserted string
}

func newUnicodeEditor() *unicodeEditor {
	editor := &unicodeEditor{anchor: -1}
	editor.ExtendBaseWidget(editor)
	return editor
}

func (e *unicodeEditor) CreateRenderer() fyne.WidgetRenderer {
	e.ExtendBaseWidget(e)
	e.richText = widget.NewRichText()
	e.richText.Wrapping = fyne.TextWrapWord
	e.scroll = container.NewVScroll(e.richText)
	background := canvas.NewRectangle(theme.Color(theme.ColorNameInputBackground))
	background.CornerRadius = theme.Size(theme.SizeNameInputRadius)
	e.caret = canvas.NewRectangle(theme.Color(theme.ColorNamePrimary))
	highlightColor, err := parseThemeColor(configString("selection_highlight"))
	if err != nil {
		panic(err)
	}
	e.highlight = canvas.NewRectangle(highlightColor)
	e.refreshText()
	return &unicodeEditorRenderer{background: background, scroll: e.scroll, caret: e.caret, highlight: e.highlight, editor: e}
}

func (e *unicodeEditor) MinSize() fyne.Size {
	e.ExtendBaseWidget(e)
	return fyne.NewSize(float32(configNumber("editor_minimum_width")),
		theme.Size(theme.SizeNameText)*float32(configNumber("editor_minimum_height_lines")))
}

func (e *unicodeEditor) Refresh() {
	e.refreshText()
	e.BaseWidget.Refresh()
}

func (e *unicodeEditor) SetText(text string) {
	e.Text = text
	e.cursor = len([]rune(text))
	e.anchor = -1
	e.undoStack = nil
	e.redoStack = nil
	e.Refresh()
}

func (e *unicodeEditor) SetPlaceHolder(text string) {
	e.PlaceHolder = text
	e.Refresh()
}

func (e *unicodeEditor) SetRightToLeft(enabled bool) {
	if e.rightToLeft == enabled {
		return
	}
	e.rightToLeft = enabled
	e.Refresh()
}

func (e *unicodeEditor) refreshText() {
	if e.richText == nil {
		return
	}
	alignment := e.textAlignment()
	if e.Text == "" && !e.focused {
		e.richText.Segments = []widget.RichTextSegment{
			&widget.TextSegment{
				Text: e.PlaceHolder,
				Style: widget.RichTextStyle{
					Alignment: alignment,
					ColorName: theme.ColorNamePlaceHolder,
				},
			},
		}
		e.richText.Refresh()
		return
	}

	runes := []rune(e.Text)
	e.cursor = e.clamp(e.cursor, len(runes))
	e.anchor = e.clamp(e.anchor, len(runes))
	e.richText.Segments = []widget.RichTextSegment{
		&widget.TextSegment{
			Text: e.Text,
			Style: widget.RichTextStyle{
				Alignment: alignment,
				ColorName: theme.ColorNameForeground,
			},
		},
	}
	e.richText.Refresh()
	e.updateOverlays()
}

func (e *unicodeEditor) textAlignment() fyne.TextAlign {
	if e.rightToLeft {
		return fyne.TextAlignTrailing
	}
	return fyne.TextAlignLeading
}

func (e *unicodeEditor) FocusGained() {
	e.focused = true
	e.Refresh()
}

func (e *unicodeEditor) FocusLost() {
	e.focused = false
	e.shiftDown = false
	e.anchor = -1
	e.Refresh()
}

func (e *unicodeEditor) TypedRune(char rune) {
	if unicode.IsControl(char) && char != '\n' && char != '\t' {
		return
	}
	e.replaceSelection(string(char))
}

func (e *unicodeEditor) TypedKey(key *fyne.KeyEvent) {
	switch key.Name {
	case fyne.KeyBackspace:
		start, end := e.selection()
		if start == end && start > 0 {
			start--
		}
		e.replaceRange(start, end, "")
	case fyne.KeyDelete:
		start, end := e.selection()
		if start == end && end < len([]rune(e.Text)) {
			end++
		}
		e.replaceRange(start, end, "")
	case fyne.KeyReturn, fyne.KeyEnter:
		e.replaceSelection("\n")
	case fyne.KeyTab:
		e.replaceSelection("\t")
	case fyne.KeyLeft, fyne.KeyRight:
		delta := 1
		if key.Name == fyne.KeyLeft {
			delta = -1
		}
		row, _ := e.rowColumn(e.cursor)
		start, end := e.lineBoundsForRow(row)
		if isRTLText(string([]rune(e.Text)[start:end])) {
			delta *= -1
		}
		e.moveCursor(e.cursor+delta, e.shiftDown)
	case fyne.KeyHome:
		e.moveCursor(e.lineStart(e.cursor), e.shiftDown)
	case fyne.KeyEnd:
		e.moveCursor(e.lineEnd(e.cursor), e.shiftDown)
	case fyne.KeyUp:
		e.moveLine(-1)
	case fyne.KeyDown:
		e.moveLine(1)
	}
}

func (e *unicodeEditor) KeyDown(key *fyne.KeyEvent) {
	if key.Name == desktop.KeyShiftLeft || key.Name == desktop.KeyShiftRight {
		if !e.shiftDown {
			e.anchor = e.cursor
		}
		e.shiftDown = true
	}
}

func (e *unicodeEditor) KeyUp(key *fyne.KeyEvent) {
	if key.Name == desktop.KeyShiftLeft || key.Name == desktop.KeyShiftRight {
		e.shiftDown = false
		if e.anchor == e.cursor {
			e.anchor = -1
		}
	}
}

func (e *unicodeEditor) Keyboard() mobile.KeyboardType {
	return mobile.DefaultKeyboard
}

func (e *unicodeEditor) TypedShortcut(shortcut fyne.Shortcut) {
	switch shortcut.(type) {
	case *fyne.ShortcutCopy:
		clipboard := fyne.CurrentApp().Clipboard()
		if typed, ok := shortcut.(*fyne.ShortcutCopy); ok && typed.Clipboard != nil {
			clipboard = typed.Clipboard
		}
		e.copySelectionTo(clipboard)
	case *fyne.ShortcutCut:
		clipboard := fyne.CurrentApp().Clipboard()
		if typed, ok := shortcut.(*fyne.ShortcutCut); ok && typed.Clipboard != nil {
			clipboard = typed.Clipboard
		}
		e.copySelectionTo(clipboard)
		start, end := e.selection()
		if start != end {
			e.replaceRange(start, end, "")
		}
	case *fyne.ShortcutPaste:
		clipboard := fyne.CurrentApp().Clipboard()
		if typed, ok := shortcut.(*fyne.ShortcutPaste); ok && typed.Clipboard != nil {
			clipboard = typed.Clipboard
		}
		e.replaceSelection(clipboard.Content())
	case *fyne.ShortcutSelectAll:
		e.anchor = 0
		e.cursor = len([]rune(e.Text))
		e.Refresh()
	case *fyne.ShortcutUndo:
		e.undo()
	case *fyne.ShortcutRedo:
		e.redo()
	}
}

func (e *unicodeEditor) MouseDown(event *desktop.MouseEvent) {
	if event.Button != desktop.MouseButtonPrimary {
		return
	}
	e.focus()
	e.moveCursor(e.cursorAt(event.Position), false)
}

func (e *unicodeEditor) MouseUp(*desktop.MouseEvent) {}

func (e *unicodeEditor) Dragged(event *fyne.DragEvent) {
	e.moveCursor(e.cursorAt(event.Position), true)
}

func (e *unicodeEditor) DragEnd() {}

func (e *unicodeEditor) TouchDown(event *mobile.TouchEvent) {
	e.focus()
	e.moveCursor(e.cursorAt(event.Position), false)
}

func (e *unicodeEditor) TouchUp(*mobile.TouchEvent)     {}
func (e *unicodeEditor) TouchCancel(*mobile.TouchEvent) {}

func (e *unicodeEditor) TappedSecondary(event *fyne.PointEvent) {
	canvas := fyne.CurrentApp().Driver().CanvasForObject(e)
	if canvas == nil {
		return
	}
	e.focus()
	start, end := e.selection()
	clicked := e.cursorAt(event.Position)
	if start == end || clicked < start || clicked > end {
		e.moveCursor(clicked, false)
	}
	menu := fyne.NewMenu(configString("empty"),
		fyne.NewMenuItem(configString("clipboard_cut"), func() {
			e.copySelectionTo(fyne.CurrentApp().Clipboard())
			start, end := e.selection()
			if start != end {
				e.replaceRange(start, end, "")
			}
		}),
		fyne.NewMenuItem(configString("clipboard_copy"), func() {
			e.copySelectionTo(fyne.CurrentApp().Clipboard())
		}),
		fyne.NewMenuItem(configString("clipboard_paste"), func() {
			e.replaceSelection(fyne.CurrentApp().Clipboard().Content())
		}),
		fyne.NewMenuItem(configString("clipboard_select_all"), func() {
			e.anchor = 0
			e.cursor = len([]rune(e.Text))
			e.Refresh()
		}),
	)
	widget.NewPopUpMenu(menu, canvas).ShowAtPosition(event.AbsolutePosition)
}

func (e *unicodeEditor) focus() {
	canvas := fyne.CurrentApp().Driver().CanvasForObject(e)
	if canvas != nil {
		canvas.Focus(e)
	}
}

func (e *unicodeEditor) cursorAt(position fyne.Position) int {
	runes := []rune(e.Text)
	rowHeight := theme.Size(theme.SizeNameText) * theme.Size(theme.SizeNameLineSpacing)
	if rowHeight <= 0 {
		rowHeight = theme.Size(theme.SizeNameText) * float32(configNumber("line_height_fallback_factor"))
	}
	row := int((position.Y + e.scrollOffsetY()) / rowHeight)
	lineStart, lineEnd := e.lineBoundsForRow(row)
	if lineStart == lineEnd {
		return lineStart
	}
	line := string(runes[lineStart:lineEnd])
	x := position.X
	if e.scroll != nil {
		x += e.scroll.Offset.X
	}
	return cursorIndexAtX(line, lineStart, e.scrollWidth()+e.scrollOffsetX(), x, e.rightToLeft, isRTLText(line))
}

func cursorIndexAtX(line string, lineStart int, viewportWidth, x float32, rightAligned, rightToLeft bool) int {
	return cursorIndexAtXMeasured(line, lineStart, viewportWidth, x, rightAligned, rightToLeft, func(text string) float32 {
		return canvas.NewText(text, theme.Color(theme.ColorNameForeground)).MinSize().Width
	})
}

func cursorIndexAtXMeasured(line string, lineStart int, viewportWidth, x float32, rightAligned, rightToLeft bool, measure func(string) float32) int {
	runes := []rune(line)
	lineWidth := measure(line)
	startX := float32(0)
	if rightAligned {
		startX = viewportWidth - lineWidth
	}
	best := 0
	bestDistance := float32(-1)
	for col := 0; col <= len(runes); col++ {
		prefixWidth := measure(string(runes[:col]))
		point := startX + prefixWidth
		if rightToLeft {
			point = startX + lineWidth - prefixWidth
		}
		distance := absFloat32(point - x)
		if bestDistance < 0 || distance < bestDistance {
			bestDistance = distance
			best = col
		}
	}
	return lineStart + best
}

func isRTLText(text string) bool {
	for _, char := range text {
		if unicode.Is(unicode.M, char) || unicode.Is(unicode.N, char) {
			continue
		}
		if unicode.IsLetter(char) {
			for _, scriptName := range appConfig.RTLScripts {
				if unicode.Is(unicode.Scripts[scriptName], char) {
					return true
				}
			}
			return false
		}
	}
	return false
}

func (e *unicodeEditor) scrollOffsetY() float32 {
	if e.scroll == nil {
		return 0
	}
	return e.scroll.Offset.Y
}

func (e *unicodeEditor) scrollWidth() float32 {
	if e.scroll == nil {
		return e.Size().Width
	}
	return e.scroll.Size().Width
}

func (e *unicodeEditor) updateOverlays() {
	if e.caret == nil || e.highlight == nil {
		return
	}
	start, end := e.selection()
	if !e.focused {
		e.caret.Hide()
		e.highlight.Hide()
		return
	}
	if start == end {
		e.highlight.Hide()
		row, _ := e.rowColumn(e.cursor)
		lineStart, lineEnd := e.lineBoundsForRow(row)
		line := string([]rune(e.Text)[lineStart:lineEnd])
		x := textPositionX(line, e.cursor-lineStart, e.scrollWidth(), e.rightToLeft, isRTLText(line))
		y := float32(row)*theme.Size(theme.SizeNameText)*theme.Size(theme.SizeNameLineSpacing) - e.scrollOffsetY()
		e.caret.Move(fyne.NewPos(x-e.scrollOffsetX(), y))
		e.caret.Resize(fyne.NewSize(theme.Size(theme.SizeNameInputBorder)*float32(configNumber("caret_width_multiplier")),
			theme.Size(theme.SizeNameText)*theme.Size(theme.SizeNameLineSpacing)))
		e.caret.Show()
		return
	}
	e.caret.Hide()
	startRow, startCol := e.rowColumn(start)
	endRow, endCol := e.rowColumn(end)
	if startRow != endRow {
		e.highlight.Hide()
		return
	}
	lineStart, lineEnd := e.lineBoundsForRow(startRow)
	line := string([]rune(e.Text)[lineStart:lineEnd])
	left := textPositionX(line, startCol, e.scrollWidth(), e.rightToLeft, isRTLText(line))
	right := textPositionX(line, endCol, e.scrollWidth(), e.rightToLeft, isRTLText(line))
	if right < left {
		left, right = right, left
	}
	y := float32(startRow)*theme.Size(theme.SizeNameText)*theme.Size(theme.SizeNameLineSpacing) - e.scrollOffsetY()
	e.highlight.Move(fyne.NewPos(left-e.scrollOffsetX(), y))
	e.highlight.Resize(fyne.NewSize(right-left, theme.Size(theme.SizeNameText)*theme.Size(theme.SizeNameLineSpacing)))
	e.highlight.Show()
}

func textPositionX(line string, column int, viewportWidth float32, rightAligned, rightToLeft bool) float32 {
	runes := []rune(line)
	column = min(max(column, 0), len(runes))
	lineWidth := canvas.NewText(line, theme.Color(theme.ColorNameForeground)).MinSize().Width
	prefixWidth := canvas.NewText(string(runes[:column]), theme.Color(theme.ColorNameForeground)).MinSize().Width
	startX := float32(0)
	if rightAligned {
		startX = viewportWidth - lineWidth
	}
	if rightToLeft {
		return startX + lineWidth - prefixWidth
	}
	return startX + prefixWidth
}

func (e *unicodeEditor) moveLine(delta int) {
	row, col := e.rowColumn(e.cursor)
	targetRow := row + delta
	start, end := e.lineBoundsForRow(targetRow)
	e.moveCursor(e.clamp(start+col, end), e.shiftDown)
}

func (e *unicodeEditor) moveCursor(position int, selecting bool) {
	runes := []rune(e.Text)
	position = e.clamp(position, len(runes))
	if selecting {
		if e.anchor < 0 {
			e.anchor = e.cursor
		}
	} else {
		e.anchor = -1
	}
	e.cursor = position
	e.Refresh()
}

func (e *unicodeEditor) replaceSelection(value string) {
	start, end := e.selection()
	e.replaceRange(start, end, value)
}

func (e *unicodeEditor) replaceRange(start, end int, value string) {
	runes := []rune(e.Text)
	start = e.clamp(start, len(runes))
	end = e.clamp(end, len(runes))
	if start > end {
		start, end = end, start
	}
	removed := string(runes[start:end])
	if removed == value {
		return
	}
	action := editAction{position: start, removed: removed, inserted: value}
	e.undoStack = append(e.undoStack, action)
	e.redoStack = nil
	e.applyTextChange(start, end, value, start+len([]rune(value)))
}

func (e *unicodeEditor) applyTextChange(start, end int, value string, cursor int) {
	runes := []rune(e.Text)
	insert := []rune(value)
	updated := make([]rune, 0, len(runes)-(end-start)+len(insert))
	updated = append(updated, runes[:start]...)
	updated = append(updated, insert...)
	updated = append(updated, runes[end:]...)
	e.Text = string(updated)
	e.cursor = cursor
	e.anchor = -1
	e.Refresh()
	if e.OnChanged != nil {
		e.OnChanged(e.Text)
	}
}

func (e *unicodeEditor) undo() {
	if len(e.undoStack) == 0 {
		return
	}
	action := e.undoStack[len(e.undoStack)-1]
	e.undoStack = e.undoStack[:len(e.undoStack)-1]
	e.redoStack = append(e.redoStack, action)
	insertedLen := len([]rune(action.inserted))
	e.applyTextChange(action.position, action.position+insertedLen, action.removed, action.position+len([]rune(action.removed)))
}

func (e *unicodeEditor) redo() {
	if len(e.redoStack) == 0 {
		return
	}
	action := e.redoStack[len(e.redoStack)-1]
	e.redoStack = e.redoStack[:len(e.redoStack)-1]
	e.undoStack = append(e.undoStack, action)
	removedLen := len([]rune(action.removed))
	e.applyTextChange(action.position, action.position+removedLen, action.inserted, action.position+len([]rune(action.inserted)))
}

func (e *unicodeEditor) copySelectionTo(clipboard fyne.Clipboard) {
	start, end := e.selection()
	if start != end {
		clipboard.SetContent(string([]rune(e.Text)[start:end]))
	}
}

func (e *unicodeEditor) selection() (int, int) {
	if e.anchor < 0 || e.anchor == e.cursor {
		return e.cursor, e.cursor
	}
	if e.anchor < e.cursor {
		return e.anchor, e.cursor
	}
	return e.cursor, e.anchor
}

func (e *unicodeEditor) rowColumn(position int) (int, int) {
	runes := []rune(e.Text)
	position = e.clamp(position, len(runes))
	row, start := 0, 0
	for i, char := range runes[:position] {
		if char == '\n' {
			row++
			start = i + 1
		}
	}
	return row, position - start
}

func (e *unicodeEditor) lineStart(position int) int {
	runes := []rune(e.Text)
	position = e.clamp(position, len(runes))
	for i := position - 1; i >= 0; i-- {
		if runes[i] == '\n' {
			return i + 1
		}
	}
	return 0
}

func (e *unicodeEditor) lineEnd(position int) int {
	runes := []rune(e.Text)
	position = e.clamp(position, len(runes))
	for i := position; i < len(runes); i++ {
		if runes[i] == '\n' {
			return i
		}
	}
	return len(runes)
}

func (e *unicodeEditor) lineBoundsForRow(row int) (int, int) {
	if row < 0 {
		return 0, 0
	}
	runes := []rune(e.Text)
	start, current := 0, 0
	for i, char := range runes {
		if char == '\n' {
			if current == row {
				return start, i
			}
			current++
			start = i + 1
		}
	}
	if current < row {
		return len(runes), len(runes)
	}
	return start, len(runes)
}

func (e *unicodeEditor) clamp(value, maximum int) int {
	if value < 0 {
		return 0
	}
	if value > maximum {
		return maximum
	}
	return value
}

type unicodeEditorRenderer struct {
	background *canvas.Rectangle
	scroll     *container.Scroll
	caret      *canvas.Rectangle
	highlight  *canvas.Rectangle
	editor     *unicodeEditor
}

func (r *unicodeEditorRenderer) Destroy() {}

func (r *unicodeEditorRenderer) Layout(size fyne.Size) {
	r.background.Resize(size)
	r.scroll.Resize(size)
	r.editor.updateOverlays()
}

func (r *unicodeEditorRenderer) MinSize() fyne.Size {
	return fyne.NewSize(float32(configNumber("editor_minimum_width")),
		theme.Size(theme.SizeNameText)*float32(configNumber("editor_minimum_height_lines")))
}

func (r *unicodeEditorRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.background, r.scroll, r.highlight, r.caret}
}

func (r *unicodeEditorRenderer) Refresh() {
	r.background.FillColor = theme.Color(theme.ColorNameInputBackground)
	r.background.Refresh()
	r.scroll.Refresh()
	r.editor.updateOverlays()
}

func absFloat32(value float32) float32 {
	if value < 0 {
		return -value
	}
	return value
}

var _ fyne.Widget = (*unicodeEditor)(nil)
var _ fyne.Focusable = (*unicodeEditor)(nil)
var _ fyne.Shortcutable = (*unicodeEditor)(nil)
var _ desktop.Mouseable = (*unicodeEditor)(nil)
var _ fyne.Draggable = (*unicodeEditor)(nil)
var _ mobile.Keyboardable = (*unicodeEditor)(nil)
var _ mobile.Touchable = (*unicodeEditor)(nil)
var _ fyne.SecondaryTappable = (*unicodeEditor)(nil)
