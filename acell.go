package acell

import (
	"io"
	"os"
	"sync"

	"github.com/romanSPB15/acell/builder"
	"github.com/romanSPB15/acell/term"
)

var (
	start = []byte("\033[?25l\033[?1003h\033[?1006h\033[?1049h\033[0m\033[H")
	end   = []byte("\033[0m\033[?25h\033[?1003l\033[?1000l\033[?1049l")
)

type pos struct {
	Line int
	Col  int
}

// Terminal управляет буфером, diff-рендером и терминалом.
type Terminal struct {
	raw    term.RawTerminal
	oldBuf [][]Cell
	Buf    [][]Cell

	cursorPos pos
	last      Style

	bb builder.Builder

	closeOnce sync.Once
	closeErr  error
}

func Default() (io.Reader, io.Writer) {
	return os.Stdin, os.Stdout
}

// New создаёт Terminal и переводит терминал в raw-режим, а также включает ENABLE_VIRTUAL_TERMINAL_PROCESSING для поддержки ANSI на Windows.
func New(in io.Reader, out io.Writer) *Terminal {
	return NewWithTerm(term.NewRawTerminal(in, out))
}

func NewWithTerm(t term.RawTerminal) *Terminal {
	t.MakeRaw()
	t.EnableANSI()
	t.Write(start)
	t.StartInput()

	w, h := t.Size()
	return &Terminal{
		raw:       t,
		Buf:       NewBuf(w, h),
		cursorPos: pos{-1, -1},
	}
}

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
	bb.WriteString("\033[?2026h")

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

			row[x].Style.WriteANSI(t.last, bb)
			t.last = row[x].Style

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
		bb.WriteString("\033[?2026l")

		t.bb.Copy(t.raw)
	}
}

func (t *Terminal) Size() (int, int) {
	return t.raw.Size()
}

// Events возвращает канал событий из терминала.
func (t *Terminal) Events() <-chan any {
	return t.raw.Events()
}

// Close останавливает терминал и восстанавливает режим.
func (t *Terminal) Close() error {
	t.closeOnce.Do(func() {
		t.raw.Write(end)
		if err := t.raw.Restore(); err != nil {
			t.closeErr = err
			return
		}
		t.closeErr = t.raw.Close()
	})
	return t.closeErr
}
