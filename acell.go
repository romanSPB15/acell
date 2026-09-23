package acell

import (
	"io"
	"os"
	"strings"
	"sync"

	"github.com/romanSPB15/acell/builder"
	"github.com/romanSPB15/acell/term"
	"github.com/romanSPB15/acell/terminfo"
)

type pos struct {
	Line int
	Col  int
}

// Terminal управляет буфером, diff-рендером и терминалом.
type Terminal struct {
	raw  term.RawTerminal
	info terminfo.Info

	oldBuf [][]Cell
	Buf    [][]Cell

	cursorPos pos
	last      Style

	fgCache map[string]string
	bgCache map[string]string

	bb builder.Builder

	closeOnce sync.Once
	closeErr  error
}

// Default возвращает стандартные потоки ввода-вывода.
func Default() (io.Reader, io.Writer) {
	return os.Stdin, os.Stdout
}

// New создаёт Terminal и переводит терминал в raw-режим,
// а также включает ENABLE_VIRTUAL_TERMINAL_PROCESSING для ANSI на Windows.
func New(in io.Reader, out io.Writer) *Terminal {
	return NewWithTerm(term.NewRawTerminal(in, out))
}

// NewWithTerm создаёт Terminal на основе переданного RawTerminal.
// Возможности терминала берутся из t.Info().
func NewWithTerm(t term.RawTerminal) *Terminal {
	info := t.Info()

	t.MakeRaw()
	t.EnableANSI()

	var start []byte
	start = append(start, info.AltScreenOn...)
	if info.MouseAny {
		start = append(start, "\033[?1003h"...)
	}
	if info.MouseSGR {
		start = append(start, "\033[?1006h"...)
	}
	start = append(start, info.Sgr0...)
	start = append(start, info.Home...)
	start = append(start, info.CursorHide...)

	t.Write(start)
	t.StartInput()

	w, h := t.Size()
	return &Terminal{
		info:      info,
		raw:       t,
		Buf:       NewBuf(w, h),
		cursorPos: pos{-1, -1},
		fgCache:   make(map[string]string, 32),
		bgCache:   make(map[string]string, 32),
	}
}

// maskStyle снимает атрибуты, которые терминал не поддерживает.
func (t *Terminal) maskStyle(s Style) Style {
	if !t.info.Blink {
		s.Args &^= Blink
	}
	if !t.info.Dim {
		s.Args &^= Dim
	}
	if !t.info.Italic {
		s.Args &^= Italic
	}
	if !t.info.Strike {
		s.Args &^= Strike
	}
	if !t.info.Hidden {
		s.Args &^= Hidden
	}

	if t.info.Colors < terminfo.ColorTrue {
		s.Fg = t.convertFg(s.Fg)
		s.Bg = t.convertBg(s.Bg)
	}
	return s
}

func (t *Terminal) convertFg(s string) string {
	if s == "" {
		return s
	}
	if v, ok := t.fgCache[s]; ok {
		return v
	}
	v := convertColor(s, t.info.Colors, false)
	t.fgCache[s] = v
	return v
}

func (t *Terminal) convertBg(s string) string {
	if s == "" {
		return s
	}
	if v, ok := t.bgCache[s]; ok {
		return v
	}
	v := convertColor(s, t.info.Colors, true)
	t.bgCache[s] = v
	return v
}

// NewBuf создаёт пустой буфер заданного размера.
func NewBuf(w, h int) [][]Cell {
	buf := make([][]Cell, h)
	for y := range buf {
		buf[y] = make([]Cell, w)
		for x := range buf[y] {
			buf[y][x] = Cell{Char: ' '}
		}
	}
	return buf
}

// Flush сравнивает Buf с oldBuf, пишет разницу в терминал
// и обновляет oldBuf.
func (t *Terminal) Flush() {
	h := len(t.Buf)
	if h == 0 {
		return
	}
	w := len(t.Buf[0])
	if w == 0 {
		return
	}

	if len(t.oldBuf) != h || len(t.oldBuf[0]) != w {
		t.oldBuf = NewBuf(w, h)
		t.cursorPos = pos{-1, -1}
		t.last = Style{}
	}

	bb := &t.bb
	bb.Reset()

	if t.info.SynchronizedUpdate {
		bb.WriteString("\033[?2026h")
	}

	changed := false

	for y := range h {
		row := t.Buf[y]
		oldRow := t.oldBuf[y]
		for x := range w {
			if row[x] == oldRow[x] {
				continue
			}

			changed = true

			if t.cursorPos.Line != y || t.cursorPos.Col != x {
				bb.WriteString("\033[")
				bb.WriteInt(y + 1)
				bb.WriteByte(';')
				bb.WriteInt(x + 1)
				bb.WriteByte('H')
			}

			// if t.cursorPos.Line != y || t.cursorPos.Col != x {
			// 	moved := false

			// 	if t.cursorPos.Line == y {
			// 		dx := x - t.cursorPos.Col
			// 		if dx > 0 && dx <= 4 {
			// 			if dx == 1 {
			// 				bb.WriteString("\033[C")
			// 			} else {
			// 				bb.WriteString("\033[")
			// 				bb.WriteInt(dx)
			// 				bb.WriteByte('C')
			// 			}
			// 			moved = true
			// 		}
			// 	}

			// 	if !moved && x == 0 && y == 0 {
			// 		bb.WriteString("\033[H")
			// 		moved = true
			// 	}

			// 	if t.cursorPos.Col == x && t.cursorPos.Line < y {
			// 		dy := y - t.cursorPos.Line
			// 		if dy <= 4 {
			// 			if dy == 1 {
			// 				bb.WriteString("\033[B")
			// 			} else {
			// 				bb.WriteString("\033[")
			// 				bb.WriteInt(dy)
			// 				bb.WriteByte('B')
			// 			}
			// 			moved = true
			// 		}
			// 	}

			// 	if !moved {
			// 		bb.WriteString("\033[")
			// 		bb.WriteInt(y + 1)
			// 		bb.WriteByte(';')
			// 		bb.WriteInt(x + 1)
			// 		bb.WriteByte('H')
			// 	}
			// }

			st := t.maskStyle(row[x].Style)
			st.WriteANSI(t.last, bb)
			t.last = st

			ch := row[x].Char
			if ch == 0 {
				ch = ' '
			}
			bb.WriteRune(ch)

			oldRow[x] = row[x]

			x2, y2 := x+1, y
			if x2 == w {
				x2 = 0
				y2++
			}
			t.cursorPos = pos{Line: y2, Col: x2}
		}
	}

	if changed {
		if t.info.SynchronizedUpdate {
			bb.WriteString("\033[?2026l")
		}
		t.bb.Copy(t.raw)
	}
}

// Size возвращает размер терминала.
func (t *Terminal) Size() (int, int) {
	return t.raw.Size()
}

// Events возвращает канал событий из терминала.
func (t *Terminal) Events() <-chan any {
	return t.raw.Events()
}

// Info получает информацию об терминале из переменных среды.
func (t *Terminal) Info() terminfo.Info {
	return t.raw.Info()
}

// Close останавливает терминал и восстанавливает режим.
func (t *Terminal) Close() error {
	t.closeOnce.Do(func() {
		var end []byte
		if t.info.MouseAny {
			end = append(end, "\033[?1003l"...)
		}
		if t.info.MouseSGR {
			end = append(end, "\033[?1006l"...)
		}
		end = append(end, t.info.Sgr0...)
		end = append(end, t.info.CursorShow...)
		end = append(end, t.info.AltScreenOff...)

		t.raw.Write(end)

		if err := t.raw.Restore(); err != nil {
			t.closeErr = err
			return
		}
		t.closeErr = t.raw.Close()
	})
	return t.closeErr
}

// DrawString рисует строку str начиная с позиции (x, y) с заданным стилем.
// Возвращает количество нарисованных ячеек.
// Символы, вышедшие за границы буфера, отбрасываются.
// Поддерживает переносы в строке — \n или \r\n.
func (t *Terminal) DrawString(x, y int, style Style, str string) {
	h := len(t.Buf)
	if h == 0 {
		return
	}
	if y < 0 || y >= h {
		return
	}
	w := len(t.Buf[y])
	if w == 0 {
		return
	}

	if strings.Contains(str, "\n") {
		strs := strings.Split(strings.ReplaceAll(str, "\r\n", "\n"), "\n")
		for i, v := range strs {
			t.DrawString(x, y+i, style, v)
		}
		return
	}

	for _, r := range str {
		if x >= w {
			break
		}
		if x >= 0 {
			if r == 0 {
				r = ' '
			}
			t.Buf[y][x] = Cell{Char: r, Style: style}
		}
		x++
	}
}
