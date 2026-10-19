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
	undoStack   []string
	redoStack   []string
	richText    *widget.RichText
	scroll      *container.Scroll
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
	e.refreshText()
	return &unicodeEditorRenderer{background: background, scroll: e.scroll, editor: e}
}

func (e *unicodeEditor) MinSize() fyne.Size {
	e.ExtendBaseWidget(e)
	return fyne.NewSize(100, theme.Size(theme.SizeNameText)*3)
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
	cursor := e.clamp(e.cursor, len(runes))
	start, end := e.selection()
	segments := make([]widget.RichTextSegment, 0, 5)
	appendSegment := func(text string, selected bool) {
		if text == "" {
			return
		}
		colorName := theme.ColorNameForeground
		if selected {
			colorName = theme.ColorNamePrimary
		}
		segments = append(segments, &widget.TextSegment{
			Text: text,
			Style: widget.RichTextStyle{
				Alignment: alignment,
				ColorName: colorName,
				Inline:    true,
			},
		})
	}
	appendCursor := func() {
		if e.focused && start == end {
			appendSegment("|", false)
		}
	}
	if start == end {
		appendSegment(string(runes[:cursor]), false)
		appendCursor()
		appendSegment(string(runes[cursor:]), false)
	} else {
		appendSegment(string(runes[:start]), false)
		appendSegment(string(runes[start:end]), true)
		if cursor == end {
			appendCursor()
		}
		appendSegment(string(runes[end:]), false)
		if cursor == start {
			appendCursor()
		}
	}
	if len(segments) == 0 {
		segments = append(segments, &widget.TextSegment{
			Text: "",
			Style: widget.RichTextStyle{
				Alignment: alignment,
				ColorName: theme.ColorNameForeground,
			},
		})
	}
	e.richText.Segments = segments
	e.richText.Refresh()
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
		if e.rightToLeft {
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
		e.copySelection()
	case *fyne.ShortcutCut:
		e.copySelection()
		start, end := e.selection()
		if start != end {
			e.replaceRange(start, end, "")
		}
	case *fyne.ShortcutPaste:
		e.replaceSelection(fyne.CurrentApp().Clipboard().Content())
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
		rowHeight = theme.Size(theme.SizeNameText)
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
	best := lineStart
	bestDistance := float32(-1)
	for col := 0; col <= len([]rune(line)); col++ {
		prefix := string([]rune(line)[:col])
		width := canvas.NewText(prefix, theme.Color(theme.ColorNameForeground)).MinSize().Width
		point := width
		if e.rightToLeft {
			point = e.scrollWidth() - width
		}
		distance := absFloat32(point - x)
		if bestDistance < 0 || distance < bestDistance {
			bestDistance = distance
			best = lineStart + col
		}
	}
	return e.clamp(best, len(runes))
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
	insert := []rune(value)
	updated := make([]rune, 0, len(runes)-(end-start)+len(insert))
	updated = append(updated, runes[:start]...)
	updated = append(updated, insert...)
	updated = append(updated, runes[end:]...)
	nextText := string(updated)
	if nextText == e.Text {
		return
	}
	e.undoStack = append(e.undoStack, e.Text)
	e.redoStack = nil
	e.Text = nextText
	e.cursor = start + len(insert)
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
	e.redoStack = append(e.redoStack, e.Text)
	e.Text = e.undoStack[len(e.undoStack)-1]
	e.undoStack = e.undoStack[:len(e.undoStack)-1]
	e.cursor = len([]rune(e.Text))
	e.anchor = -1
	e.Refresh()
	if e.OnChanged != nil {
		e.OnChanged(e.Text)
	}
}

func (e *unicodeEditor) redo() {
	if len(e.redoStack) == 0 {
		return
	}
	e.undoStack = append(e.undoStack, e.Text)
	e.Text = e.redoStack[len(e.redoStack)-1]
	e.redoStack = e.redoStack[:len(e.redoStack)-1]
	e.cursor = len([]rune(e.Text))
	e.anchor = -1
	e.Refresh()
	if e.OnChanged != nil {
		e.OnChanged(e.Text)
	}
}

func (e *unicodeEditor) copySelection() {
	start, end := e.selection()
	if start != end {
		fyne.CurrentApp().Clipboard().SetContent(string([]rune(e.Text)[start:end]))
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
	editor     *unicodeEditor
}

func (r *unicodeEditorRenderer) Destroy() {}

func (r *unicodeEditorRenderer) Layout(size fyne.Size) {
	r.background.Resize(size)
	r.scroll.Resize(size)
}

func (r *unicodeEditorRenderer) MinSize() fyne.Size {
	return fyne.NewSize(100, theme.Size(theme.SizeNameText)*3)
}

func (r *unicodeEditorRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.background, r.scroll}
}

func (r *unicodeEditorRenderer) Refresh() {
	r.background.FillColor = theme.Color(theme.ColorNameInputBackground)
	r.background.Refresh()
	r.scroll.Refresh()
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
