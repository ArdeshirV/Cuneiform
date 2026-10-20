# Avestan Editor

[فارسی](README.fa.md)

Avestan Editor is a Unicode text editor for ancient scripts and other writing systems. Its software keyboard starts with Avesta, Pahlavi, Parthian, and Ancient Persian, followed by other historic scripts and the Unicode scripts supported by the Go toolchain.

It runs on Linux, Windows, and Android. You can open, edit, and save Unicode text documents in UTF-8 or UTF-16.

## Run and build

```sh
go run .
go test ./...
```

Build for desktop with `go build .`. Fyne mobile packaging and Android SDK/NDK setup are described in the [Fyne mobile documentation](https://docs.fyne.io/started/mobile/). Create distributable packages with `fyne package --target windows` or `fyne package --target android/arm64` (with the required SDKs installed).

## Keyboard and scripts

The language selector puts the four requested ancient Iranian keyboards first, followed by other ancient Iranian and historic scripts. It then adds the other scripts listed in the Go Unicode tables, plus general symbols and an **All printable Unicode** keyboard. The displayed Unicode version comes from the Go runtime. Rows are virtualized to support large ranges such as CJK. Keys display a space, the character, another space, and the reading; combining marks are displayed with a dotted circle but insert only the mark.

The bundled Unicode Script Fallback is a renamed, merged font containing Noto Avestan, Old Persian, Parthian, Pahlavi, and 40 other Noto script faces, plus Noto Naskh Arabic shaping for Persian/Arabic text. It is distributed under SIL OFL 1.1; face names and attribution are in `assets/fonts/FONT-SOURCES.txt`, and the license is in `assets/fonts/LICENSE.txt`. **Settings** contains font selection, bundled-font restore, theme, and text size controls. Changes are saved between runs. The right-to-left checkbox follows **About**, defaults to on, and is remembered. It changes the native editor's paragraph alignment to the right; when off, paragraphs align to the left. Long lines wrap to the available editor width.

The language selector previews each script with a sample word or representative characters. Keyboard keys show the character and its pronunciation separated by spaces. Verified readings are currently bundled for Avesta, Ancient Persian, Parthian, and Pahlavi; other characters fall back to their hexadecimal code point rather than an asserted pronunciation. Right-click or long-press in the editor for mouse-based cut/copy/paste/select-all; standard keyboard shortcuts are also available. Save As shows Fyne's overwrite confirmation before replacing an existing file. If the document has unsaved changes, Exit or closing the window asks whether to save, discard, or cancel.

The keyboard occupies 70% of the window by default. Dragging the splitter changes the proportions, which are saved for the next run.

## Documents and fonts

Documents may be opened and saved as UTF-8 (with or without BOM) or UTF-16 LE/BE when marked with a byte-order mark. Existing files retain their encoding on Save. The bundled font is licensed under the SIL Open Font License; see `assets/fonts/LICENSE.txt`. The application is licensed under GPLv3 or later; see `LICENSE`.

## Bundled configuration

Startup settings and labels are embedded from `assets/app.json`; the ordered script catalogue, code-point ranges, aliases, and previews are in `assets/languages.json`; transliterations are in `assets/readings.json`. They load directly from the application binary. Unicode classification and character enumeration remain algorithmic and use the Go Unicode tables.

Wistaspa, Avesta, and `ave_bl_l4` are not included because their redistribution terms could not be verified. The bundled, freely licensed Noto Sans Avestan face provides Avestan Unicode coverage.

Copyright (c) 2026 @ArdeshirV
