package acell

import (
	"errors"
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

		want := "\033[?2026h\033[H\033[31mA\033[?2026l"
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

		want := "\033[?2026h\033[H \033[?2026l"
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

		want := "\033[?2026h\033[H\033[31mA\033[32mB\033[?2026l"
		if got := tr.WrittenString(); got != want {
			t.Fatalf("want %q, got %q", want, got)
		}
	})

	t.Run("row rollover emits CUP after line boundary", func(t *testing.T) {
		tr := newFakeRawTerminal()
		term := NewWithTerm(tr)
		st := Style{Fg: "31"}
		term.Buf = [][]Cell{
			{{Char: 'A', Style: st}, {Char: 'B', Style: st}},
			{{Char: 'C', Style: st}, {Char: 'D', Style: st}},
		}
		tr.ResetWritten()

		term.Flush()

		want := "\033[?2026h\033[H\033[31mAB\033[2;1HCD\033[?2026l"
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

		want := "\033[?2026h\033[H\033[31mA\033[0mB\033[?2026l"
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

		want := "\033[?2026h\033[H\033[31mAB\033[?2026l"
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

		want := "\033[?2026h\033[H\033[31mA\033[?2026l"
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

		want := "\033[?2026h\033[HB\033[?2026l"
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

		want := "\033[H\033[31mA"
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

	t.Run("only CR", func(t *testing.T) {
		tr := newFakeRawTerminal()
		term := NewWithTerm(tr)

		term.DrawString(0, 0, Style{}, "a\rb")
		if term.Buf[1][0].Char != 'b' {
			t.Errorf("[1][0] = %q, want 'b'", term.Buf[1][0].Char)
		}
	})

	t.Run("only LF", func(t *testing.T) {
		tr := newFakeRawTerminal()
		term := NewWithTerm(tr)

		term.DrawString(0, 0, Style{}, "a\nb")
		if term.Buf[1][0].Char != 'b' {
			t.Errorf("[1][0] = %q, want 'b'", term.Buf[1][0].Char)
		}
	})
}

func TestRuneWidth(t *testing.T) {
	cases := []struct {
		r    rune
		want int
	}{
		{0, 0},
		{'\n', 0},
		{'\t', 0},
		{0x7F, 0},
		{'a', 1},
		{'─', 1},
		{'█', 1},
		{'⠋', 1},
		{0x0301, 0}, // combining
		{0x200B, 0}, // zero-width space
		{0xFE00, 0}, // variation selector
		{0xFE20, 0}, // combining half marks
		{'猫', 2},
		{'中', 2},
		{'あ', 2},
		{'한', 2},
		{'Ａ', 2}, // fullwidth A
		{'😀', 2},
		{'🚀', 2},
		{'✅', 2},
		{'⭐', 2},
		{0x20000, 2}, // CJK Ext B
		{0x30000, 2}, // CJK Ext G
	}
	for _, tc := range cases {
		if got := RuneWidth(tc.r); got != tc.want {
			t.Errorf("RuneWidth(%U) = %d, want %d", tc.r, got, tc.want)
		}
	}
}

func TestDrawRuneWide(t *testing.T) {
	t.Run("wide rune occupies two cells with placeholder", func(t *testing.T) {
		tr := newFakeRawTerminal()
		term := NewWithTerm(tr)
		term.Buf = NewBuf(10, 3)

		w, drawn := term.DrawRune(0, 0, Style{Fg: "31"}, '猫')
		if !drawn {
			t.Fatalf("drawn = false")
		}
		if w != 2 {
			t.Fatalf("width = %d, want 2", w)
		}
		if term.Buf[0][0].Char != '猫' {
			t.Errorf("[0][0] = %q, want '猫'", term.Buf[0][0].Char)
		}
		if term.Buf[0][1].Char != ' ' {
			t.Errorf("[0][1] = %q, want ' '", term.Buf[0][1].Char)
		}
		if term.Buf[0][0].Style.Fg != "31" || term.Buf[0][1].Style.Fg != "31" {
			t.Errorf("placeholder style mismatch")
		}
	})

	t.Run("wide rune doesn't fit on right edge", func(t *testing.T) {
		tr := newFakeRawTerminal()
		term := NewWithTerm(tr)
		term.Buf = NewBuf(5, 1)

		w, drawn := term.DrawRune(4, 0, Style{}, '猫')
		if drawn {
			t.Fatalf("drawn = true, want false")
		}
		if w != 2 {
			t.Fatalf("width = %d, want 2", w)
		}
		for x := 0; x < 5; x++ {
			if term.Buf[0][x].Char != ' ' {
				t.Errorf("[%d] = %q, want ' '", x, term.Buf[0][x].Char)
			}
		}
	})

	t.Run("wide rune at last valid position", func(t *testing.T) {
		tr := newFakeRawTerminal()
		term := NewWithTerm(tr)
		term.Buf = NewBuf(5, 1)

		w, drawn := term.DrawRune(3, 0, Style{}, '猫')
		if !drawn || w != 2 {
			t.Fatalf("drawn=%v width=%d, want true/2", drawn, w)
		}
		if term.Buf[0][3].Char != '猫' || term.Buf[0][4].Char != ' ' {
			t.Errorf("wrong cells: [3]=%q [4]=%q", term.Buf[0][3].Char, term.Buf[0][4].Char)
		}
	})

	t.Run("wide rune overwrites previous narrow", func(t *testing.T) {
		tr := newFakeRawTerminal()
		term := NewWithTerm(tr)
		term.Buf = NewBuf(10, 1)
		term.Buf[0][0] = Cell{Char: 'A'}
		term.Buf[0][1] = Cell{Char: 'B'}

		term.DrawRune(0, 0, Style{}, '猫')

		if term.Buf[0][0].Char != '猫' {
			t.Errorf("[0] = %q, want '猫'", term.Buf[0][0].Char)
		}
		if term.Buf[0][1].Char != ' ' {
			t.Errorf("[1] = %q, want ' '", term.Buf[0][1].Char)
		}
	})

	t.Run("wide rune at y out of bounds", func(t *testing.T) {
		tr := newFakeRawTerminal()
		term := NewWithTerm(tr)
		term.Buf = NewBuf(10, 3)

		if w, drawn := term.DrawRune(0, 100, Style{}, '猫'); drawn || w != 2 {
			t.Errorf("y=100: drawn=%v w=%d, want false/2", drawn, w)
		}
		if w, drawn := term.DrawRune(0, -1, Style{}, '猫'); drawn || w != 2 {
			t.Errorf("y=-1: drawn=%v w=%d, want false/2", drawn, w)
		}
	})

	t.Run("wide rune with negative x", func(t *testing.T) {
		tr := newFakeRawTerminal()
		term := NewWithTerm(tr)
		term.Buf = NewBuf(10, 1)

		if w, drawn := term.DrawRune(-1, 0, Style{}, '猫'); drawn || w != 2 {
			t.Errorf("x=-1: drawn=%v w=%d, want false/2", drawn, w)
		}
	})

	t.Run("empty buffer", func(t *testing.T) {
		tr := newFakeRawTerminal()
		term := NewWithTerm(tr)
		term.Buf = nil

		if w, drawn := term.DrawRune(0, 0, Style{}, '猫'); drawn || w != 2 {
			t.Errorf("nil buf: drawn=%v w=%d, want false/2", drawn, w)
		}
	})

	t.Run("zero-width row", func(t *testing.T) {
		tr := newFakeRawTerminal()
		term := NewWithTerm(tr)
		term.Buf = [][]Cell{{}}

		if w, drawn := term.DrawRune(0, 0, Style{}, '猫'); drawn || w != 2 {
			t.Errorf("zero-width: drawn=%v w=%d, want false/2", drawn, w)
		}
	})

	t.Run("two wide runes side by side", func(t *testing.T) {
		tr := newFakeRawTerminal()
		term := NewWithTerm(tr)
		term.Buf = NewBuf(10, 1)

		term.DrawRune(0, 0, Style{}, '猫')
		term.DrawRune(2, 0, Style{}, '中')

		want := []rune{'猫', ' ', '中', ' '}
		for i, ch := range want {
			if term.Buf[0][i].Char != ch {
				t.Errorf("[%d] = %q, want %q", i, term.Buf[0][i].Char, ch)
			}
		}
	})

	t.Run("wide rune after wide rune without gap", func(t *testing.T) {
		tr := newFakeRawTerminal()
		term := NewWithTerm(tr)
		term.Buf = NewBuf(10, 1)

		term.DrawRune(0, 0, Style{}, '猫')
		// перезаписываем placeholder вторым широким
		term.DrawRune(1, 0, Style{}, '中')

		if term.Buf[0][0].Char != '猫' {
			t.Errorf("[0] = %q, want '猫'", term.Buf[0][0].Char)
		}
		if term.Buf[0][1].Char != '中' {
			t.Errorf("[1] = %q, want '中'", term.Buf[0][1].Char)
		}
		if term.Buf[0][2].Char != ' ' {
			t.Errorf("[2] = %q, want ' '", term.Buf[0][2].Char)
		}
	})

	t.Run("narrow rune one cell", func(t *testing.T) {
		tr := newFakeRawTerminal()
		term := NewWithTerm(tr)
		term.Buf = NewBuf(5, 1)

		w, drawn := term.DrawRune(1, 0, Style{Fg: "32"}, 'a')
		if !drawn || w != 1 {
			t.Fatalf("drawn=%v w=%d, want true/1", drawn, w)
		}
		if term.Buf[0][1].Char != 'a' {
			t.Errorf("[1] = %q, want 'a'", term.Buf[0][1].Char)
		}
		// соседняя клетка не тронута
		if term.Buf[0][2].Char != ' ' {
			t.Errorf("[2] = %q, want ' ' (not touched)", term.Buf[0][2].Char)
		}
	})

	t.Run("zero rune becomes space", func(t *testing.T) {
		tr := newFakeRawTerminal()
		term := NewWithTerm(tr)
		term.Buf = NewBuf(5, 1)

		w, drawn := term.DrawRune(0, 0, Style{}, 0)
		if drawn || w != 0 {
			t.Fatalf("drawn=%v w=%d, want false/0", drawn, w)
		}
	})

	t.Run("combining rune returns zero width not drawn", func(t *testing.T) {
		tr := newFakeRawTerminal()
		term := NewWithTerm(tr)
		term.Buf = NewBuf(5, 1)

		w, drawn := term.DrawRune(0, 0, Style{}, 0x0301)
		if drawn || w != 0 {
			t.Fatalf("combining: drawn=%v w=%d, want false/0", drawn, w)
		}
	})
}

func TestFlushCursorAndWide(t *testing.T) {
	t.Run("CUF when dx == 2", func(t *testing.T) {
		tr := newFakeRawTerminal()
		term := NewWithTerm(tr)
		term.Buf = buildBuf(1, 10, ' ', Style{})
		term.Buf[0][0] = Cell{Char: 'A'}
		term.Flush()
		tr.ResetWritten()

		term.Buf[0][3] = Cell{Char: 'B'}
		term.Flush()

		want := "\033[?2026h\033[2CB\033[?2026l"
		if got := tr.WrittenString(); got != want {
			t.Fatalf("want %q, got %q", want, got)
		}
	})

	t.Run("CUF when dx == 4", func(t *testing.T) {
		tr := newFakeRawTerminal()
		term := NewWithTerm(tr)
		term.Buf = buildBuf(1, 12, ' ', Style{})
		term.Buf[0][0] = Cell{Char: 'A'}
		term.Flush()
		tr.ResetWritten()

		term.Buf[0][5] = Cell{Char: 'B'}
		term.Flush()

		want := "\033[?2026h\033[4CB\033[?2026l"
		if got := tr.WrittenString(); got != want {
			t.Fatalf("want %q, got %q", want, got)
		}
	})

	t.Run("no CUF when dx > 4 — fallback to CUP", func(t *testing.T) {
		tr := newFakeRawTerminal()
		term := NewWithTerm(tr)
		term.Buf = buildBuf(1, 15, ' ', Style{})
		term.Buf[0][0] = Cell{Char: 'A'}
		term.Flush()
		tr.ResetWritten()

		term.Buf[0][10] = Cell{Char: 'B'}
		term.Flush()

		want := "\033[?2026h\033[1;11HB\033[?2026l"
		if got := tr.WrittenString(); got != want {
			t.Fatalf("want %q, got %q", want, got)
		}
	})

	t.Run("CUD when dy == 1", func(t *testing.T) {
		tr := newFakeRawTerminal()
		term := NewWithTerm(tr)
		term.Buf = buildBuf(5, 5, ' ', Style{})
		term.Buf[0][2] = Cell{Char: 'A'}
		term.Flush()
		tr.ResetWritten()

		term.Buf[1][3] = Cell{Char: 'B'}
		term.Flush()

		want := "\033[?2026h\033[BB\033[?2026l"
		if got := tr.WrittenString(); got != want {
			t.Fatalf("want %q, got %q", want, got)
		}
	})

	t.Run("CUD when dy == 3", func(t *testing.T) {
		tr := newFakeRawTerminal()
		term := NewWithTerm(tr)
		term.Buf = buildBuf(6, 5, ' ', Style{})
		term.Buf[0][2] = Cell{Char: 'A'}
		term.Flush()
		tr.ResetWritten()

		term.Buf[3][3] = Cell{Char: 'B'}
		term.Flush()

		want := "\033[?2026h\033[3BB\033[?2026l"
		if got := tr.WrittenString(); got != want {
			t.Fatalf("want %q, got %q", want, got)
		}
	})
}

func TestInfoMethod(t *testing.T) {
	tr := newFakeRawTerminal()
	info := terminfo.Default()
	info.Name = "custom-info"
	tr.SetInfo(info)

	term := NewWithTerm(tr)
	if term.Info().Name != "custom-info" {
		t.Errorf("Info().Name = %q", term.Info().Name)
	}
}

func TestCloseRestoreError(t *testing.T) {
	tr := newFakeRawTerminal()
	want := errors.New("restore failed")
	tr.SetRestoreErr(want)

	term := NewWithTerm(tr)
	if err := term.Close(); !errors.Is(err, want) {
		t.Errorf("got %v, want %v", err, want)
	}
}

func TestCloseError(t *testing.T) {
	tr := newFakeRawTerminal()
	want := errors.New("close failed")
	tr.SetCloseErr(want)

	term := NewWithTerm(tr)
	if err := term.Close(); !errors.Is(err, want) {
		t.Errorf("got %v, want %v", err, want)
	}
}
