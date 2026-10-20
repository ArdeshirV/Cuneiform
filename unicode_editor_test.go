package main

import (
	"testing"

	"fyne.io/fyne/v2"
	fynetest "fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

func TestUnicodeEditorSupportsSuccessiveTypingAndInsertion(t *testing.T) {
	testApp := fynetest.NewApp()
	defer testApp.Quit()

	editor := widget.NewMultiLineEntry()
	if !editor.MultiLine {
		t.Fatal("editor is not multiline")
	}
	if editor.Wrapping != fyne.TextWrapWord {
		t.Fatalf("editor wrapping = %v, want word wrapping", editor.Wrapping)
	}

	editor.TypedRune('a')
	editor.TypedRune('b')
	if editor.Text != "ab" {
		t.Fatalf("text after successive typing = %q, want %q", editor.Text, "ab")
	}

	editor.TypedKey(&fyne.KeyEvent{Name: fyne.KeyLeft})
	editor.TypedRune('𐬀')
	if editor.Text != "a𐬀b" {
		t.Fatalf("text after inserting at the cursor = %q, want %q", editor.Text, "a𐬀b")
	}
}

func TestKeyboardKeySeparatesSymbolAndPronunciation(t *testing.T) {
	got := keyboardKeyText('𐬀', "a")
	want := configString("left_to_right_mark") + "𐬀/a"
	if got != want {
		t.Fatalf("keyboard label = %q, want %q", got, want)
	}
}
