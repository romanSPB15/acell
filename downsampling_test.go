package acell

import (
	"testing"

	"github.com/romanSPB15/acell/terminfo"
)

func TestNearest(t *testing.T) {
	if got := nearest(255, 0, 0, ansi16RGB[:]); got != 9 {
		t.Errorf("nearest red in 16 = %d, want 9", got)
	}
	if got := nearest(0, 0, 0, ansi16RGB[:]); got != 0 {
		t.Errorf("nearest black in 16 = %d, want 0", got)
	}
	if got := nearest(255, 0, 0, ansi16RGB[:8]); got != 1 {
		t.Errorf("nearest red in 8 = %d, want 1", got)
	}
	if got := nearest(255, 255, 255, ansi16RGB[:]); got != 15 {
		t.Errorf("nearest white in 16 = %d, want 15", got)
	}
}

func TestNearest256(t *testing.T) {
	if got := nearest256(255, 0, 0); got != 196 {
		t.Errorf("nearest256 red = %d, want 196", got)
	}
	if got := nearest256(255, 255, 255); got != 231 {
		t.Errorf("nearest256 white = %d, want 231", got)
	}
	if got := nearest256(0, 0, 0); got != 16 {
		t.Errorf("nearest256 black = %d, want 16", got)
	}
}

func TestParseColorCode(t *testing.T) {
	cases := []struct {
		in   string
		isBg bool
		kind int
	}{
		{"", false, 0},
		{"abc", false, 0},
		{"31", false, 1},
		{"91", false, 1},
		{"37", false, 1},
		{"97", false, 1},
		{"41", true, 1},
		{"101", true, 1},
		{"47", true, 1},
		{"107", true, 1},
		{"40", false, 0},
		{"31", true, 0},
		{"999", false, 0},
		{"38;2;10;20;30", false, 3},
		{"38;2;1;2", false, 0},
		{"38;2;1;2;3;4", false, 0},
		{"38;5;196", false, 2},
		{"38;5;", false, 0},
		{"48;2;10;20;30", true, 3},
		{"48;5;196", true, 2},
	}
	for i, tc := range cases {
		got := parseColorCode(tc.in, tc.isBg)
		if got.kind != tc.kind {
			t.Errorf("#%d: %q (isBg=%v) kind = %d, want %d",
				i, tc.in, tc.isBg, got.kind, tc.kind)
		}
	}
}

func TestParseColorCodeValues(t *testing.T) {
	pc := parseColorCode("38;2;10;20;30", false)
	if pc.r != 10 || pc.g != 20 || pc.b != 30 {
		t.Errorf("rgb = %d;%d;%d, want 10;20;30", pc.r, pc.g, pc.b)
	}

	pc = parseColorCode("38;5;196", false)
	if pc.idx != 196 {
		t.Errorf("idx = %d, want 196", pc.idx)
	}

	pc = parseColorCode("91", false)
	if pc.idx != 9 {
		t.Errorf("bright red idx = %d, want 9", pc.idx)
	}

	pc = parseColorCode("41", true)
	if pc.idx != 1 {
		t.Errorf("bg red idx = %d, want 1", pc.idx)
	}

	pc = parseColorCode("101", true)
	if pc.idx != 9 {
		t.Errorf("bg bright red idx = %d, want 9", pc.idx)
	}
}

func TestParsedColorToRGB(t *testing.T) {
	cases := []parsedColor{
		{kind: 1, idx: 1},
		{kind: 1, idx: 9},
		{kind: 2, idx: 196},
		{kind: 3, r: 10, g: 20, b: 30},
	}
	for i, pc := range cases {
		r, g, b := pc.toRGB()
		_ = r
		_ = g
		_ = b
		if pc.kind == 3 {
			if r != 10 || g != 20 || b != 30 {
				t.Errorf("#%d: rgb = %d;%d;%d, want 10;20;30", i, r, g, b)
			}
		}
	}
}

func TestANSI8Code(t *testing.T) {
	if got := ansi8Code(1, false); got != "31" {
		t.Errorf("fg 8 red = %q, want 31", got)
	}
	if got := ansi8Code(1, true); got != "41" {
		t.Errorf("bg 8 red = %q, want 41", got)
	}
	if got := ansi8Code(7, false); got != "37" {
		t.Errorf("fg 8 white = %q, want 37", got)
	}
}

func TestANSI16Code(t *testing.T) {
	if got := ansi16Code(1, false); got != "31" {
		t.Errorf("fg 16 red = %q, want 31", got)
	}
	if got := ansi16Code(9, false); got != "91" {
		t.Errorf("fg 16 bright red = %q, want 91", got)
	}
	if got := ansi16Code(1, true); got != "41" {
		t.Errorf("bg 16 red = %q, want 41", got)
	}
	if got := ansi16Code(9, true); got != "101" {
		t.Errorf("bg 16 bright red = %q, want 101", got)
	}
	if got := ansi16Code(0, false); got != "30" {
		t.Errorf("fg 16 black = %q, want 30", got)
	}
	if got := ansi16Code(15, false); got != "97" {
		t.Errorf("fg 16 bright white = %q, want 97", got)
	}
}

func TestANSI256Code(t *testing.T) {
	if got := ansi256Code(196, false); got != "38;5;196" {
		t.Errorf("fg 256 = %q, want 38;5;196", got)
	}
	if got := ansi256Code(196, true); got != "48;5;196" {
		t.Errorf("bg 256 = %q, want 48;5;196", got)
	}
}

func TestConvertColorEdges(t *testing.T) {
	cases := []struct {
		in     string
		target terminfo.ColorBits
		isBg   bool
		want   string
	}{
		{"31", terminfo.Color8, false, "31"},
		{"91", terminfo.Color16, false, "91"},
		{"38;5;196", terminfo.Color256, false, "38;5;196"},
		{"31", terminfo.ColorTrue, false, "31"},
		{"38;2;255;0;0", terminfo.ColorTrue, false, "38;2;255;0;0"},
		{"31", terminfo.ColorNone, false, "31"},
		{"48;2;255;0;0", terminfo.Color8, true, "41"},
		{"41", terminfo.Color16, true, "41"},
		{"48;2;255;0;0", terminfo.Color256, true, "48;5;196"},
		{"abc", terminfo.Color8, false, "abc"},
		{"", terminfo.Color8, false, ""},
		{"38;2;255;0;0", terminfo.Color256, false, "38;5;196"},
		{"38;2;255;0;0", terminfo.Color16, false, "91"},
		{"38;2;255;0;0", terminfo.Color8, false, "31"},
		{"38;5;196", terminfo.Color16, false, "91"},
		{"38;5;196", terminfo.Color8, false, "31"},
		{"91", terminfo.Color8, false, "31"},
		{"31", terminfo.Color16, false, "31"},
	}
	for i, tc := range cases {
		got := convertColor(tc.in, tc.target, tc.isBg)
		if got != tc.want {
			t.Errorf("#%d: %q (target=%v isBg=%v) = %q, want %q",
				i, tc.in, tc.target, tc.isBg, got, tc.want)
		}
	}
}

func TestConvertColorCache(t *testing.T) {
	tr := newFakeRawTerminal()
	tr.SetInfo(terminfo.Info{Colors: terminfo.Color8})
	term := NewWithTerm(tr)

	_ = term.maskStyle(Style{Fg: "38;2;255;0;0"})
	_ = term.maskStyle(Style{Fg: "38;2;255;0;0"})

	if len(term.fgCache) != 1 {
		t.Errorf("fgCache size = %d, want 1", len(term.fgCache))
	}
	if len(term.bgCache) != 0 {
		t.Errorf("bgCache size = %d, want 0", len(term.bgCache))
	}

	_ = term.maskStyle(Style{Bg: "48;2;0;255;0"})
	_ = term.maskStyle(Style{Bg: "48;2;0;255;0"})

	if len(term.bgCache) != 1 {
		t.Errorf("bgCache size = %d, want 1", len(term.bgCache))
	}
}

func TestMaskStyleTrueColor(t *testing.T) {
	tr := newFakeRawTerminal()
	tr.SetInfo(terminfo.Info{Colors: terminfo.ColorTrue})
	term := NewWithTerm(tr)

	s := term.maskStyle(Style{Fg: "38;2;255;0;0", Bg: "48;2;10;20;30"})
	if s.Fg != "38;2;255;0;0" || s.Bg != "48;2;10;20;30" {
		t.Errorf("TrueColor downsampled: %+v", s)
	}
}

func TestMaskStyleDownsample(t *testing.T) {
	tr := newFakeRawTerminal()
	tr.SetInfo(terminfo.Info{Colors: terminfo.Color8})
	term := NewWithTerm(tr)

	s := term.maskStyle(Style{Fg: "38;2;255;0;0", Bg: "48;2;0;0;255"})
	if s.Fg != "31" {
		t.Errorf("Fg = %q, want 31", s.Fg)
	}
	if s.Bg != "44" {
		t.Errorf("Bg = %q, want 44", s.Bg)
	}
}

func TestMaskStyleEmpty(t *testing.T) {
	tr := newFakeRawTerminal()
	tr.SetInfo(terminfo.Info{Colors: terminfo.Color8})
	term := NewWithTerm(tr)

	s := term.maskStyle(Style{})
	if s.Fg != "" || s.Bg != "" {
		t.Errorf("empty style mangled: %+v", s)
	}
}

func TestNearest256Fallback(t *testing.T) {
	got := nearest256(137, 89, 203)
	if got < 0 || got > 255 {
		t.Errorf("out of range: %d", got)
	}

	if got < 16 {
		t.Errorf("expected cube/gray index, got %d", got)
	}
}

func TestNearest256NoExactMatch(t *testing.T) {
	cases := [][3]uint8{
		{1, 1, 1},
		{254, 254, 254},
		{128, 64, 32},
		{200, 100, 50},
		{50, 100, 150},
	}
	for _, c := range cases {
		got := nearest256(c[0], c[1], c[2])
		if got < 0 || got > 255 {
			t.Errorf("nearest256(%v) = %d", c, got)
		}
	}
}
