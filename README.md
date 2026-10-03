<img src="docs/acell.gif" alt="acell demo" width="800"/>

# acell

[![Go Reference](https://pkg.go.dev/badge/github.com/romanSPB15/acell.svg)](https://pkg.go.dev/github.com/romanSPB15/acell)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Coverage](https://img.shields.io/badge/coverage-97%25-green)](./calc-coverage-noterm.ps1)
[![Test](https://github.com/romanSPB15/acell/actions/workflows/test.yaml/badge.svg)](https://github.com/romanSPB15/acell/actions/workflows/test.yaml)
[![Examples](https://img.shields.io/badge/View-examples-white?logo=github)](https://github.com/romanSPB15/acell/tree/main/examples)


[Russian version](README.ru.md)

**A low-level cell-based terminal layer for Go.**

- 🎯 Full mouse support: click, release, motion, wheel (SGR 1006)
- 🚀 5000 FPS in the benchmark — 2× faster than tcell v3
- 🎨 Automatic color downsampling — TrueColor → 256 → 16 → 8, matching terminal capabilities
- 🎁 Windows without WSL, without CGO
- 📦 Only two dependencies — `x/sys` and `x/term`
- ✅ 97% test coverage of the core

<h3 align="center"><pre>go get github.com/romanSPB15/acell</pre></h3>

## Quick start

```go
package main

import "github.com/romanSPB15/acell"

func main() {
	in, out := acell.Default()
	t := acell.New(in, out)
	defer t.Close()

	draw := func() {
		t.Clear()
		t.DrawString(0, 0, acell.Style{}, "Hello,")
		t.DrawString(7, 0, acell.Style{Fg: acell.FgRed, Args: acell.Bold}, "acell")
		t.DrawRune(12, 0, acell.Style{}, '!')
		t.DrawString(0, 2, acell.Style{Fg: acell.Fg16Grey}, "press q to quit")
		t.Flush()
	}

	draw()

	for ev := range t.Events() {
		switch e := ev.(type) {
		case *acell.KeyboardEvent:
			if e.Rune == 'q' || e.Key == acell.KeyCtrlC {
				return
			}
		case *acell.ResizeEvent:
			t.Buf = acell.NewBuf(e.Width, e.Height)
			t.Invalidate()
			draw()
		}
	}
}

```

## How it works

The rendering model is a `[][]Cell` buffer (`Terminal.Buf`) plus a shadow
copy (`oldBuf`). Each `Flush`:

1. Walks both buffers cell by cell.
2. For every changed cell, writes a `CUP` (only if the cursor isn't
   already there), then the style delta, then the rune.
3. Caches the previous style and cursor position to avoid emitting
   redundant sequences.
4. Wraps the frame in mode 2026 (`CSI ?2026h` … `CSI ?2026l`) for
   atomic updates.

All output is collected into a reusable `builder.Builder` — no intermediate strings.

## Performance

[Stress benchmark](https://github.com/romanSPB15/acell/blob/main/bench) — 120×30 terminal, 80×24 widget, 300 changing cells
per frame, 8 colors, Windows 10 x64, Windows Terminal, including I/O:

| Implementation                        | Raw     |
|---------------------------------------|---------|
| **acell**                             | ~5000   |
| tcell v3.5.0                          | ~2450   |

## Packages

| Package   | Purpose                                                    |
|-----------|------------------------------------------------------------|
| `acell`   | `Terminal`, `Cell`, `Style`, diff rendering                |
| `ansi`    | ANSI escape sequence parsing: `Strip`, `Find`              |
| `builder` | A `strings.Builder` analogue with an extended API          |
| `term`    | `RawTerminal` and terminal handling                        |
| `input`   | Mouse and keyboard parser                                  |
| `terminfo`| Extracting terminal information from environment variables |

## Test coverage

- Core: **97.0%**
- Including the terminal layer: 81.8%

```
./calc-coverage.ps1          # full coverage, including term
./calc-coverage-noterm.ps1   # core only
```

## License

[MIT](./LICENSE)
