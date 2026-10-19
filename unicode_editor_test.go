package main

import (
	"testing"

	"fyne.io/fyne/v2"
)

func TestUnicodeEditorMaintainsLogicalTextAndHistory(t *testing.T) {
	editor := &unicodeEditor{Text: "𐬀𐬁", cursor: 1, anchor: -1}
	editor.TypedRune('𐬂')
	if want := "𐬀𐬂𐬁"; editor.Text != want {
		t.Fatalf("inserted text = %q, want %q", editor.Text, want)
	}

	editor.undo()
	if want := "𐬀𐬁"; editor.Text != want {
		t.Fatalf("text after undo = %q, want %q", editor.Text, want)
	}
	editor.redo()
	if want := "𐬀𐬂𐬁"; editor.Text != want {
		t.Fatalf("text after redo = %q, want %q", editor.Text, want)
	}
}

func TestUnicodeEditorRTLChangesAlignmentAndArrowDirection(t *testing.T) {
	editor := &unicodeEditor{Text: "abc", cursor: 1, anchor: -1}
	editor.SetRightToLeft(true)
	if !editor.rightToLeft {
		t.Fatal("RTL mode was not enabled")
	}
	if got := editor.textAlignment(); got != fyne.TextAlignTrailing {
		t.Fatalf("RTL text alignment = %v, want trailing", got)
	}
	editor.TypedKey(&fyne.KeyEvent{Name: fyne.KeyLeft})
	if editor.cursor != 2 {
		t.Fatalf("cursor after left arrow in RTL mode = %d, want 2", editor.cursor)
	}
	editor.TypedKey(&fyne.KeyEvent{Name: fyne.KeyRight})
	if editor.cursor != 1 {
		t.Fatalf("cursor after right arrow in RTL mode = %d, want 1", editor.cursor)
	}
}

func TestUnicodeEditorDoesNotDirtyOnNoOpDelete(t *testing.T) {
	editor := &unicodeEditor{Text: "x", cursor: 0, anchor: -1}
	changed := false
	editor.OnChanged = func(string) { changed = true }
	editor.TypedKey(&fyne.KeyEvent{Name: fyne.KeyBackspace})
	if changed {
		t.Fatal("backspace at the beginning reported a text change")
	}
}
