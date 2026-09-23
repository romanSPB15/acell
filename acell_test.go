package acell

import (
	"os"
	"strings"
	"testing"

	"github.com/romanSPB15/acell/terminfo"
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
	if !tr.InputStarted() {
		t.Fatal("input is not started")
	}

	out := tr.WrittenString()
	for _, want := range []string{
		"\033[?1049h", // alt-screen on
		"\033[?1003h", // mouse any
		"\033[?1006h", // mouse sgr
		"\033[H",      // home
		"\033[?25l",   // cursor hide
	} {
		if !strings.Contains(out, want) {
			t.Errorf("start: expected %q in %q", want, out)
		}
	}
}

func TestNewWithTermNoCursorHide(t *testing.T) {
	tr := newFakeRawTerminal()
	info := terminfo.Default()
	info.CursorHide = ""
	tr.SetInfo(info)

	NewWithTerm(tr)

	if out := tr.WrittenString(); strings.Contains(out, "\033[?25l") {
		t.Errorf("did not expect cursor-hide, got %q", out)
	}
}

func TestNewWithTermNoMouse(t *testing.T) {
	tr := newFakeRawTerminal()
	info := terminfo.Default()
	info.MouseAny = false
	info.MouseSGR = false
	tr.SetInfo(info)

	NewWithTerm(tr)

	out := tr.WrittenString()
	if strings.Contains(out, "\033[?1003h") {
		t.Errorf("did not expect mouse-any, got %q", out)
	}
	if strings.Contains(out, "\033[?1006h") {
		t.Errorf("did not expect mouse-sgr, got %q", out)
	}
}

func TestNewWithTermNoAltScreen(t *testing.T) {
	tr := newFakeRawTerminal()
	info := terminfo.Default()
	info.AltScreenOn = ""
	info.AltScreenOff = ""
	tr.SetInfo(info)

	NewWithTerm(tr)

	if out := tr.WrittenString(); strings.Contains(out, "\033[?1049h") {
		t.Errorf("did not expect alt-screen, got %q", out)
	}
}

func TestClose(t *testing.T) {
	tr := newFakeRawTerminal()
	term := NewWithTerm(tr)
	tr.ResetWritten()

	if err := term.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	if tr.Raw() {
		t.Fatal("terminal is raw")
	}
	if tr.InputStarted() {
		t.Fatal("input is started")
	}
	if !tr.Closed() {
		t.Fatal("terminal is not closed")
	}

	out := tr.WrittenString()
	for _, want := range []string{
		"\033[?1003l", // mouse any off
		"\033[?1006l", // mouse sgr off
		"\033[?25h",   // cursor show
		"\033[?1049l", // alt-screen off
	} {
		if !strings.Contains(out, want) {
			t.Errorf("end: expected %q in %q", want, out)
		}
	}
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
		term.Buf = nil
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
			{Char: 'B'},
		}}
		tr.ResetWritten()

		term.Flush()

		want := "\033[?2026h\033[1;1H\033[31mA\033[0mB\033[?2026l"
		if got := tr.WrittenString(); got != want {
			t.Fatalf("want %q, got %q", want, got)
		}
	})

	t.Run("unchanged buffer writes nothing", func(t *testing.T) {
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

		term.Buf[0][0] = Cell{Char: 'B', Style: st}
		term.Flush()

		want := "\033[?2026h\033[1;1HB\033[?2026l"
		if got := tr.WrittenString(); got != want {
			t.Fatalf("want %q, got %q", want, got)
		}
	})

	t.Run("no sync update if unsupported", func(t *testing.T) {
		tr := newFakeRawTerminal()
		info := terminfo.Default()
		info.SynchronizedUpdate = false
		tr.SetInfo(info)

		term := NewWithTerm(tr)
		term.Buf = [][]Cell{{{Char: 'A', Style: Style{Fg: "31"}}}}
		tr.ResetWritten()

		term.Flush()

		want := "\033[1;1H\033[31mA"
		if got := tr.WrittenString(); got != want {
			t.Fatalf("want %q, got %q", want, got)
		}
	})

	t.Run("mask blink if unsupported", func(t *testing.T) {
		tr := newFakeRawTerminal()
		info := terminfo.Default()
		info.Blink = false
		tr.SetInfo(info)

		term := NewWithTerm(tr)
		term.Buf = [][]Cell{{{Char: 'A', Style: Style{Args: Blink}}}}
		tr.ResetWritten()

		term.Flush()

		out := tr.WrittenString()
		if strings.Contains(out, "\033[5m") {
			t.Errorf("blink should be masked, got %q", out)
		}
	})

	t.Run("mask dim if unsupported", func(t *testing.T) {
		tr := newFakeRawTerminal()
		info := terminfo.Default()
		info.Dim = false
		tr.SetInfo(info)

		term := NewWithTerm(tr)
		term.Buf = [][]Cell{{{Char: 'A', Style: Style{Args: Dim}}}}
		tr.ResetWritten()

		term.Flush()

		out := tr.WrittenString()
		if strings.Contains(out, "\033[2m") {
			t.Errorf("dim should be masked, got %q", out)
		}
	})
}

func TestNewBuf(t *testing.T) {
	const width, height = 20, 10
	buf := NewBuf(width, height)
	if len(buf) != height {
		t.Fatalf("invalid buf height: expected %d, got %d", height, len(buf))
	}
	if len(buf[0]) != width {
		t.Fatalf("invalid buf width: expected %d, got %d", width, len(buf[0]))
	}
	expected := Cell{Char: ' '}
	for y := range height {
		for x := range width {
			if buf[y][x] != expected {
				t.Fatalf("invalid cell at [%d, %d]: expected %v, got %v", x, y, expected, buf[y][x])
			}
		}
	}
}

func TestVSCodeNoCursorHide(t *testing.T) {
	ft := newFakeRawTerminal()
	ft.SetInfo(terminfo.Info{
		Name:               "vscode",
		Colors:             256,
		AltScreenOn:        "\033[?1049h",
		AltScreenOff:       "\033[?1049l",
		Home:               "\033[H",
		CursorHide:         "",
		CursorShow:         "",
		SynchronizedUpdate: false,
		Blink:              false,
		MouseAny:           false,
		MouseSGR:           false,
	})

	term := NewWithTerm(ft)
	defer term.Close()

	out := ft.WrittenString()
	if strings.Contains(out, "\033[?25l") {
		t.Errorf("vscode: не должно быть \\033[?25l в выводе")
	}
	if strings.Contains(out, "\033[?2026h") {
		t.Errorf("vscode: не должно быть \\033[?2026h в выводе")
	}
}

func TestDumbNoAltScreen(t *testing.T) {
	ft := newFakeRawTerminal()
	ft.SetInfo(terminfo.Info{
		Name:         "dumb",
		Colors:       8,
		AltScreenOn:  "",
		AltScreenOff: "",
		CursorHide:   "",
		CursorShow:   "",
		MouseAny:     false,
		MouseSGR:     false,
	})

	term := NewWithTerm(ft)
	defer term.Close()

	out := ft.WrittenString()
	if strings.Contains(out, "\033[?1049h") {
		t.Errorf("dumb: не должно быть alt-screen")
	}
	if strings.Contains(out, "\033[?1003h") {
		t.Errorf("dumb: не должно быть mouse")
	}
}

func TestTerminalInfo(t *testing.T) {
	tr := newFakeRawTerminal()
	info := terminfo.Default()
	info.Name = "custom"
	tr.SetInfo(info)

	term := NewWithTerm(tr)

	if term.info.Name != "custom" {
		t.Errorf("expected info.Name=%q, got %q", "custom", term.info.Name)
	}
}

func TestSize(t *testing.T) {
	tr := newFakeRawTerminal()
	tr.SetSize(120, 40)
	term := NewWithTerm(tr)

	w, h := term.Size()
	if w != 120 || h != 40 {
		t.Errorf("expected (120, 40), got (%d, %d)", w, h)
	}
}

func TestEvents(t *testing.T) {
	tr := newFakeRawTerminal()
	term := NewWithTerm(tr)

	if term.Events() != tr.Events() {
		t.Error("Events() should return underlying channel")
	}
}

func TestDrawString(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		tr := newFakeRawTerminal()
		term := NewWithTerm(tr)

		term.DrawString(2, 1, Style{Fg: "31"}, "abc")

		for i, ch := range "abc" {
			cell := term.Buf[1][2+i]
			if cell.Char != ch {
				t.Errorf("[%d] = %q, want %q", i, cell.Char, ch)
			}
			if cell.Style.Fg != "31" {
				t.Errorf("[%d].Style.Fg = %q, want %q", i, cell.Style.Fg, "31")
			}
		}
	})

	t.Run("clip right", func(t *testing.T) {
		tr := newFakeRawTerminal()
		term := NewWithTerm(tr)
		term.Buf = NewBuf(5, 3)

		term.DrawString(3, 0, Style{}, "abcdef")

		// влезли только 'a', 'b' на [0][3] и [0][4]
		if term.Buf[0][3].Char != 'a' {
			t.Errorf("[3] = %q, want 'a'", term.Buf[0][3].Char)
		}
		if term.Buf[0][4].Char != 'b' {
			t.Errorf("[4] = %q, want 'b'", term.Buf[0][4].Char)
		}
	})

	t.Run("negative x clips left", func(t *testing.T) {
		tr := newFakeRawTerminal()
		term := NewWithTerm(tr)
		term.Buf = NewBuf(10, 3)

		term.DrawString(-2, 0, Style{}, "abcdef")

		// 'a','b' отброшены, 'c' попал в [0][0]
		if term.Buf[0][0].Char != 'c' {
			t.Errorf("[0] = %q, want 'c'", term.Buf[0][0].Char)
		}
		if term.Buf[0][3].Char != 'f' {
			t.Errorf("[3] = %q, want 'f'", term.Buf[0][3].Char)
		}
	})

	t.Run("y out of bounds", func(t *testing.T) {
		tr := newFakeRawTerminal()
		term := NewWithTerm(tr)

		// не должно паниковать
		term.DrawString(0, 1000, Style{}, "abc")
		term.DrawString(0, -1, Style{}, "abc")
	})

	t.Run("empty string", func(t *testing.T) {
		tr := newFakeRawTerminal()
		term := NewWithTerm(tr)

		before := term.Buf[0][0]
		term.DrawString(0, 0, Style{}, "")
		if term.Buf[0][0] != before {
			t.Errorf("buffer changed")
		}
	})

	t.Run("null rune becomes space", func(t *testing.T) {
		tr := newFakeRawTerminal()
		term := NewWithTerm(tr)

		term.DrawString(0, 0, Style{}, string(rune(0)))
		if term.Buf[0][0].Char != ' ' {
			t.Errorf("[0] = %q, want ' '", term.Buf[0][0].Char)
		}
	})

	t.Run("multiline LF", func(t *testing.T) {
		tr := newFakeRawTerminal()
		term := NewWithTerm(tr)

		term.DrawString(0, 0, Style{}, "ab\ncd")

		if term.Buf[0][0].Char != 'a' || term.Buf[0][1].Char != 'b' {
			t.Errorf("line 0 wrong: %q %q", term.Buf[0][0].Char, term.Buf[0][1].Char)
		}
		if term.Buf[1][0].Char != 'c' || term.Buf[1][1].Char != 'd' {
			t.Errorf("line 1 wrong: %q %q", term.Buf[1][0].Char, term.Buf[1][1].Char)
		}
	})

	t.Run("multiline CRLF", func(t *testing.T) {
		tr := newFakeRawTerminal()
		term := NewWithTerm(tr)

		term.DrawString(0, 0, Style{}, "ab\r\ncd")

		if term.Buf[0][0].Char != 'a' || term.Buf[0][1].Char != 'b' {
			t.Errorf("line 0 wrong")
		}
		if term.Buf[1][0].Char != 'c' || term.Buf[1][1].Char != 'd' {
			t.Errorf("line 1 wrong")
		}
	})

	t.Run("empty buffer", func(t *testing.T) {
		tr := newFakeRawTerminal()
		term := NewWithTerm(tr)
		term.Buf = nil

		term.DrawString(0, 0, Style{}, "abc") // не должно паниковать
	})

	t.Run("zero width", func(t *testing.T) {
		tr := newFakeRawTerminal()
		term := NewWithTerm(tr)
		term.Buf = [][]Cell{{}}

		term.DrawString(0, 0, Style{}, "abc") // не должно паниковать
	})
}
