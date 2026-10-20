package main

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	fynetest "fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"
)

func TestUnicodeEditorSupportsSuccessiveTypingAndInsertion(t *testing.T) {
	testApp := fynetest.NewApp()
	defer testApp.Quit()

	editor := newEditorEntry(false)
	if !editor.MultiLine {
		t.Fatal("editor is not multiline")
	}
	if editor.Wrapping != fyne.TextWrapWord {
		t.Fatalf("editor wrapping = %v, want word wrapping", editor.Wrapping)
	}
	if editor.Scroll != fyne.ScrollVerticalOnly {
		t.Fatalf("editor scroll direction = %v, want vertical-only scrolling", editor.Scroll)
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

func TestEditorTextAlignmentFollowsCheckbox(t *testing.T) {
	testApp := fynetest.NewApp()
	defer testApp.Quit()

	for _, test := range []struct {
		rightToLeft bool
		want        fyne.TextAlign
	}{
		{rightToLeft: true, want: fyne.TextAlignTrailing},
		{rightToLeft: false, want: fyne.TextAlignLeading},
	} {
		editor := newEditorEntry(test.rightToLeft)
		if got := editorTextAlignment(test.rightToLeft); got != test.want {
			t.Errorf("editor alignment (RTL %v) = %v, want %v", test.rightToLeft, got, test.want)
		}
		if got := entryTextAlignment(t, editor); got != test.want {
			t.Errorf("entry text style alignment (RTL %v) = %v, want %v", test.rightToLeft, got, test.want)
		}
		if got := editor.Text; got != "" {
			t.Errorf("new editor text = %q, want empty", got)
		}
	}
}

func TestEditorFillsAvailableWidthAndWrapsLongLines(t *testing.T) {
	testApp := fynetest.NewApp()
	defer testApp.Quit()

	editor := newEditorEntry(false)
	editor.SetText("a very long line that should wrap instead of forcing the editor wider than its window")
	window := fynetest.NewWindow(container.NewMax(editor))
	window.Resize(fyne.NewSize(640, 360))
	window.Show()

	if got := editor.Size().Width; got < 600 {
		t.Fatalf("editor width = %v, want it to fill the 640px window", got)
	}
	if got := editor.MinSize().Width; got >= 640 {
		t.Fatalf("editor minimum width = %v, want it below the window width", got)
	}
	window.Close()
}

func TestChangingTextDirectionKeepsDocumentAndSwitchesAlignment(t *testing.T) {
	testApp := fynetest.NewApp()
	defer testApp.Quit()

	entry := newEditorEntry(false)
	entry.SetText("first line\nsecond line")
	editorHost := container.NewMax(entry)
	window := fynetest.NewWindow(editorHost)
	window.Show()
	editor := &editor{
		window:     window,
		text:       entry,
		editorHost: editorHost,
		keyboard:   newKeyboard(languages[0], entry),
		status:     widget.NewLabel(""),
		encoding:   encodingUTF8,
	}
	editor.bindTextEntry()
	editor.setTextDirection(true)
	if got, want := editor.text.Text, "first line\nsecond line"; got != want {
		t.Fatalf("document after enabling RTL = %q, want %q", got, want)
	}
	if got := entryTextAlignment(t, editor.text); got != fyne.TextAlignTrailing {
		t.Fatalf("alignment after enabling RTL = %v, want right aligned", got)
	}

	editor.setTextDirection(false)
	if got, want := editor.text.Text, "first line\nsecond line"; got != want {
		t.Fatalf("document after disabling RTL = %q, want %q", got, want)
	}
	if got := entryTextAlignment(t, editor.text); got != fyne.TextAlignLeading {
		t.Fatalf("alignment after disabling RTL = %v, want left aligned", got)
	}
	window.Close()
}

func entryTextAlignment(t *testing.T, entry *widget.Entry) fyne.TextAlign {
	t.Helper()
	visited := make(map[fyne.Widget]bool)
	var find func([]fyne.CanvasObject) (fyne.TextAlign, bool)
	find = func(objects []fyne.CanvasObject) (fyne.TextAlign, bool) {
		for _, object := range objects {
			if richText, ok := object.(*widget.RichText); ok && len(richText.Segments) > 0 {
				if segment, ok := richText.Segments[0].(*widget.TextSegment); ok {
					return segment.Style.Alignment, true
				}
			}
			if objectContainer, ok := object.(*fyne.Container); ok {
				if alignment, found := find(objectContainer.Objects); found {
					return alignment, true
				}
			}
			if widgetObject, ok := object.(fyne.Widget); ok && !visited[widgetObject] {
				visited[widgetObject] = true
				renderer := fynetest.TempWidgetRenderer(t, widgetObject)
				if alignment, found := find(renderer.Objects()); found {
					return alignment, true
				}
			}
		}
		return fyne.TextAlignLeading, false
	}

	renderer := fynetest.TempWidgetRenderer(t, entry)
	if alignment, found := find(renderer.Objects()); found {
		return alignment
	}
	t.Fatal("could not locate the entry's rich-text segment")
	return fyne.TextAlignLeading
}

func TestKeyboardKeyShowsSymbolAndPronunciationSeparatedBySpaces(t *testing.T) {
	got := keyboardKeyText('𐬀', "a")
	want := " 𐬀 a"
	if got != want {
		t.Fatalf("keyboard label = %q, want %q", got, want)
	}
}
