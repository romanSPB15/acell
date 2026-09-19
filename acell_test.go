package acell

import (
	"os"
	"slices"
	"testing"
)

func TestDefault(t *testing.T) {
	in, out := Default()
	if in != os.Stdin {
		t.Fatalf("invalid Default result - in!=stdin")
	}
	if out != os.Stdout {
		t.Fatalf("invalid Default result - out!=stdout")
	}
}

func TestNewWithTerm(t *testing.T) {
	tr := newFakeRawTerminal()
	NewWithTerm(tr)
	if !tr.Raw() {
		t.Fatal("terminal is not raw")
	}
	if !tr.ANSIEnabled() {
		t.Fatal("ansi is not enabled")
	}
	if !slices.Equal(tr.WrittenBytes(), start) {
		t.Fatalf("invalid bytes written: %v", tr.WrittenBytes())
	}
	if !tr.InputStarted() {
		t.Fatal("input is not started")
	}
}

func TestClose(t *testing.T) {
	tr := newFakeRawTerminal()
	NewWithTerm(tr).Close()
	if tr.Raw() {
		t.Fatal("terminal is raw")
	}

	expected := append(start, end...)
	if !slices.Equal(tr.WrittenBytes(), expected) {
		t.Fatalf("invalid bytes written: expected: %v, but got: %v", expected, tr.WrittenBytes())
	}
	if tr.InputStarted() {
		t.Fatal("input is started")
	}
	if !tr.Closed() {
		t.Fatal("terminal is not closed")
	}
}

func buildBuf(h, w int, ch rune, st Style) [][]Cell {
	buf := make([][]Cell, h)
	for y := range buf {
		buf[y] = make([]Cell, w)
		for x := range buf[y] {
			buf[y][x] = Cell{Char: ch, Style: st}
		}
	}
	return buf
}

func TestFlush(t *testing.T) {
	t.Run("empty buffer writes nothing", func(t *testing.T) {
		tr := newFakeRawTerminal()
		term := NewWithTerm(tr)
		tr.ResetWritten()

		term.Flush()

		if got := tr.WrittenString(); got != "" {
			t.Fatalf("want no writes, got %q", got)
		}
	})

	t.Run("zero-width buffer writes nothing", func(t *testing.T) {
		tr := newFakeRawTerminal()
		term := NewWithTerm(tr)
		term.Buf = [][]Cell{{}}
		tr.ResetWritten()

		term.Flush()

		if got := tr.WrittenString(); got != "" {
			t.Fatalf("want no writes, got %q", got)
		}
	})

	t.Run("single cell", func(t *testing.T) {
		tr := newFakeRawTerminal()
		term := NewWithTerm(tr)
		term.Buf = [][]Cell{{{Char: 'A', Style: Style{Fg: "31"}}}}
		tr.ResetWritten()

		term.Flush()

		want := "\033[?2026h\033[1;1H\033[31mA\033[?2026l"
		if got := tr.WrittenString(); got != want {
			t.Fatalf("want %q, got %q", want, got)
		}
	})

	t.Run("null rune becomes space", func(t *testing.T) {
		tr := newFakeRawTerminal()
		term := NewWithTerm(tr)
		term.Buf = [][]Cell{{{Char: 0}}}
		tr.ResetWritten()

		term.Flush()

		want := "\033[?2026h\033[1;1H \033[?2026l"
		if got := tr.WrittenString(); got != want {
			t.Fatalf("want %q, got %q", want, got)
		}
	})

	t.Run("adjacent cells don't emit redundant CUP", func(t *testing.T) {
		tr := newFakeRawTerminal()
		term := NewWithTerm(tr)
		term.Buf = [][]Cell{{
			{Char: 'A', Style: Style{Fg: "31"}},
			{Char: 'B', Style: Style{Fg: "32"}},
		}}
		tr.ResetWritten()

		term.Flush()

		// Второй CUP не должен появиться: после 'A' курсор уже на [0,1].
		want := "\033[?2026h\033[1;1H\033[31mA\033[32mB\033[?2026l"
		if got := tr.WrittenString(); got != want {
			t.Fatalf("want %q, got %q", want, got)
		}
	})

	t.Run("row rollover doesn't emit CUP", func(t *testing.T) {
		tr := newFakeRawTerminal()
		term := NewWithTerm(tr)
		st := Style{Fg: "31"}
		term.Buf = [][]Cell{
			{{Char: 'A', Style: st}, {Char: 'B', Style: st}},
			{{Char: 'C', Style: st}, {Char: 'D', Style: st}},
		}
		tr.ResetWritten()

		term.Flush()

		// После 'B' курсор сам переходит на [1,0], CUP не нужен.
		want := "\033[?2026h\033[1;1H\033[31mABCD\033[?2026l"
		if got := tr.WrittenString(); got != want {
			t.Fatalf("want %q, got %q", want, got)
		}
	})

	t.Run("style reset to empty emits \\033[0m", func(t *testing.T) {
		tr := newFakeRawTerminal()
		term := NewWithTerm(tr)
		term.Buf = [][]Cell{{
			{Char: 'A', Style: Style{Fg: "31"}},
			{Char: 'B'}, // пустой Style
		}}
		tr.ResetWritten()

		term.Flush()

		want := "\033[?2026h\033[1;1H\033[31mA\033[0mB\033[?2026l"
		if got := tr.WrittenString(); got != want {
			t.Fatalf("want %q, got %q", want, got)
		}
	})

	t.Run("unchanged buffer emits writes nothing", func(t *testing.T) {
		tr := newFakeRawTerminal()
		term := NewWithTerm(tr)
		term.Buf = [][]Cell{{{Char: 'A', Style: Style{Fg: "31"}}}}
		term.Flush()
		tr.ResetWritten()

		term.Flush()

		if got := tr.WrittenString(); got != "" {
			t.Fatalf("want no writes, got %q", got)
		}
	})

	t.Run("only changed cells are rewritten", func(t *testing.T) {
		tr := newFakeRawTerminal()
		term := NewWithTerm(tr)
		term.Buf = buildBuf(1, 4, ' ', Style{})
		term.Buf[0][0] = Cell{Char: 'A'}
		term.Buf[0][2] = Cell{Char: 'C'}
		term.Flush()
		tr.ResetWritten()

		// Меняем только [0][2]: должен уйти CUP именно в эту клетку.
		term.Buf[0][2] = Cell{Char: 'X'}
		term.Flush()

		want := "\033[?2026h\033[1;3HX\033[?2026l"
		if got := tr.WrittenString(); got != want {
			t.Fatalf("want %q, got %q", want, got)
		}
	})

	t.Run("grow buffer resets oldBuf", func(t *testing.T) {
		tr := newFakeRawTerminal()
		term := NewWithTerm(tr)
		st := Style{Fg: "31"}
		term.Buf = [][]Cell{{{Char: 'A', Style: st}}}
		term.Flush()
		tr.ResetWritten()

		term.Buf = [][]Cell{{
			{Char: 'A', Style: st},
			{Char: 'B', Style: st},
		}}
		term.Flush()

		// Обе клетки переписаны, так как oldBuf сброшен.
		want := "\033[?2026h\033[1;1H\033[31mAB\033[?2026l"
		if got := tr.WrittenString(); got != want {
			t.Fatalf("want %q, got %q", want, got)
		}
	})

	t.Run("shrink buffer resets oldBuf", func(t *testing.T) {
		tr := newFakeRawTerminal()
		term := NewWithTerm(tr)
		st := Style{Fg: "31"}
		term.Buf = buildBuf(2, 2, 'A', st)
		term.Flush()
		tr.ResetWritten()

		term.Buf = [][]Cell{{{Char: 'A', Style: st}}}
		term.Flush()

		want := "\033[?2026h\033[1;1H\033[31mA\033[?2026l"
		if got := tr.WrittenString(); got != want {
			t.Fatalf("want %q, got %q", want, got)
		}
	})

	t.Run("style persists between flushes", func(t *testing.T) {
		tr := newFakeRawTerminal()
		term := NewWithTerm(tr)
		st := Style{Fg: "31"}

		term.Buf = [][]Cell{{{Char: 'A', Style: st}}}
		term.Flush()
		tr.ResetWritten()

		// Меняем только символ, стиль тот же —
		// повторного \033[31m быть не должно.
		term.Buf[0][0] = Cell{Char: 'B', Style: st}
		term.Flush()

		want := "\033[?2026h\033[1;1HB\033[?2026l"
		if got := tr.WrittenString(); got != want {
			t.Fatalf("want %q, got %q", want, got)
		}
	})
}

func TestCloseIdempotent(t *testing.T) {
	tr := newFakeRawTerminal()
	term := NewWithTerm(tr)
	tr.ResetWritten()

	if err := term.Close(); err != nil {
		t.Fatalf("first Close: %v", err)
	}
	afterFirst := len(tr.WrittenBytes())

	if err := term.Close(); err != nil {
		t.Fatalf("second Close: %v", err)
	}
	if len(tr.WrittenBytes()) != afterFirst {
		t.Fatalf("second Close wrote extra bytes: %q", tr.WrittenString())
	}
}

func TestNewBuf(t *testing.T) {
	const width, height = 20, 10
	buf := NewBuf(20, 10)
	if len(buf) != height {
		t.Fatalf("invalid buf height: expected: %d, but got: %v", height, len(buf))
	}
	if len(buf[0]) != width {
		t.Fatalf("invalid buf width: expected: %d, but got: %v", width, len(buf[0]))
	}
	expected := Cell{Char: ' '}
	for y := range height {
		for x := range width {
			if buf[y][x] != expected {
				t.Fatalf("invalid cell at [%d, %d]: expected %v, but got: %v", x, y, expected, buf[y][x])
			}
		}
	}
}
