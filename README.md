# Cuneiform

[فارسی](README.fa.md)

Cuneiform is a Unicode text editor for ancient scripts and other writing systems. Its software keyboard starts with Avesta, Pahlavi, Parthian, and Ancient Persian, followed by other historic scripts and the Unicode scripts supported by the Go toolchain.

It runs on Linux, Windows, and Android. You can open, edit, and save Unicode text documents in UTF-8 or UTF-16.

## Run and build

```sh
go run .
go test ./...
```

Build for desktop with `go build .`. Fyne mobile packaging and Android SDK/NDK setup are described in the [Fyne mobile documentation](https://docs.fyne.io/started/mobile/). Create distributable packages with `fyne package --target windows` or `fyne package --target android/arm64` (with the required SDKs installed).

## Keyboard and scripts

The language selector puts the four requested ancient Iranian keyboards first, followed by other ancient Iranian and historic scripts. It then adds the other scripts listed in the Go Unicode tables, plus general symbols and an **All printable Unicode** keyboard. The displayed Unicode version comes from the Go runtime. Rows are virtualized to support large ranges such as CJK. Keys show a character, one space, and its reading; combining marks are displayed with a dotted circle but insert only the mark.

The bundled Unicode Script Fallback font contains glyphs for the requested Iranian scripts and several other historic scripts. You can choose it or the system default with **Font**. **Settings** controls the theme and text size; changes are saved between runs. The right-to-left checkbox is next to the file controls, saves its value, and right-aligns the editor text without changing keyboard order. Cursor movement follows the selected text direction.

The keyboard occupies 70% of the window by default. Dragging the splitter changes the proportions, which are saved for the next run.

## Documents and fonts

Documents may be opened and saved as UTF-8 (with or without BOM) or UTF-16 LE/BE when marked with a byte-order mark. Existing files retain their encoding on Save. The bundled font is licensed under the SIL Open Font License; see `assets/fonts/LICENSE.txt`. The application is licensed under GPLv3 or later; see `LICENSE`.

Copyright (c) 2026 @ArdeshirV
