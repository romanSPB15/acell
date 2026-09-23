package acell

import (
	"reflect"
	"slices"
	"testing"
)

func cells(chars string, styles ...Style) []Cell {
	runes := []rune(chars)

	res := make([]Cell, len(runes))

	var currentStyle Style
	if len(styles) > 0 {
		currentStyle = styles[0]
	}

	sIdx := 1

	for i, ch := range runes {
		res[i] = Cell{Char: ch, Style: currentStyle}
		if sIdx < len(styles) {
			currentStyle = styles[sIdx]
			sIdx++
		}
	}
	return res
}

func TestParse(t *testing.T) {
	tt := []struct {
		Input    string
		Expected []Cell
	}{
		{Input: "\033[31mHello\033[0m World",
			Expected: append(
				cells("Hello", Style{Fg: "31"}),
				cells(" World", Style{})...,
			),
		},
		{
			Input: "\033[1;32mBold green\033[0mNormal",
			Expected: append(
				cells("Bold green", Style{Fg: "32", Args: Bold}),
				cells("Normal", Style{})...,
			),
		},
		{
			Input: "\033[40;30m▀\033[0m▀",
			Expected: append(
				cells("▀", Style{Fg: "30", Bg: "40"}),
				cells("▀", Style{})...,
			),
		},
		{
			Input: "\033[40;30m▀\033[0m▀\033[43m",
			Expected: append(
				cells("▀", Style{Fg: "30", Bg: "40"}),
				cells("▀", Style{})...,
			),
		},
		{
			Input:    "\033[30m▀\033[40m▀\033[35m",
			Expected: cells("▀▀", Style{Fg: "30"}, Style{Fg: "30", Bg: "40"}),
		},
		{
			Input:    "\033[30m⣾\033[0m⢿",
			Expected: cells("⣾⢿", Style{Fg: "30"}, Style{}),
		},
		{
			Input:    "\033[30m⣾\033[0m⢿",
			Expected: cells("⣾⢿", Style{Fg: "30"}, Style{}),
		},
		{
			Input: "\033[2mH\033[38;2;200;100;50m\033[48;2;120;255;80mi",
			Expected: cells("Hi",
				Style{Args: Dim},
				Style{
					Fg:   "38;2;200;100;50",
					Bg:   "48;2;120;255;80",
					Args: Dim,
				}),
		},
		{
			Input: "\033[101;5;4;7m\033[95mH\033[39mi",
			Expected: cells("Hi", Style{
				Fg:   "95",
				Bg:   "101",
				Args: Blink | Underline | Reverse,
			}, Style{
				Bg:   "101",
				Args: Blink | Underline | Reverse,
			}),
		},
		{
			Input: "\x1b[1;3;4;5;7m1\x1b[22;23;24;25;27m2",
			Expected: cells("12", Style{
				Args: Blink | Underline | Reverse | Bold | Italic,
			}, Style{}),
		},
		{
			Input:    "\033[m123",
			Expected: cells("123"),
		},
		{
			Input:    "\033[41m1\033[49m23",
			Expected: cells("123", Style{Bg: "41"}, Style{}),
		},
		{
			Input: "\033[2mDim\033[22mNormal",
			Expected: append(
				cells("Dim", Style{Args: Dim}),
				cells("Normal", Style{})...,
			),
		},
		{
			Input: "\033[1;2mBoth\033[22mNone",
			Expected: append(
				cells("Both", Style{Args: Bold | Dim}),
				cells("None", Style{})...,
			),
		},
		{
			Input: "\033[1;2mBoth\033[22;1mBold",
			Expected: append(
				cells("Both", Style{Args: Bold | Dim}),
				cells("Bold", Style{Args: Bold})...,
			),
		},
		{
			Input: "\033[8mHidden\033[28mNormal",
			Expected: append(
				cells("Hidden", Style{Args: Hidden}),
				cells("Normal", Style{})...,
			),
		},
		{
			Input: "\033[9mStrike\033[29mNormal",
			Expected: append(
				cells("Strike", Style{Args: Strike}),
				cells("Normal", Style{})...,
			),
		},
		{
			Input: "\033[1;2;3;4;5;7;8;9mAll\033[22;23;24;25;27;28;29mNone",
			Expected: append(
				cells("All", Style{Args: Bold | Dim | Italic | Underline | Blink | Reverse | Hidden | Strike}),
				cells("None", Style{})...,
			),
		},
		{
			Input:    "\033[1mB\033[2mD\033[22mN",
			Expected: cells("BDN", Style{Args: Bold}, Style{Args: Bold | Dim}, Style{}),
		},
	}
	buf := make([]Cell, 0, 256)
	for i, test := range tt {
		got := Parse(test.Input, &buf)
		if !slices.Equal(got, test.Expected) {
			t.Errorf("#%d: expected %v, but got %v", i, test.Expected, got)
		}
	}
}

func TestParseFromTo(t *testing.T) {
	tt := []struct {
		Input    string
		Expected []Cell
		From, To Style
	}{
		{Input: "\033[31mHello\033[0m World",
			Expected: append(
				cells("Hello", Style{Fg: "31", Args: Bold}),
				cells(" World", Style{})...,
			),
			From: Style{Args: Bold},
			To:   Style{},
		},
		{Input: "1\033[33m2\033[3m3\033[5m4\033[0m",
			Expected: cells("1234", Style{Args: Bold, Fg: "90"}, Style{Fg: "33", Args: Bold}, Style{Fg: "33", Args: Bold | Italic}, Style{Fg: "33", Args: Bold | Italic | Blink}),
			From:     Style{Args: Bold, Fg: "90"},
			To:       Style{},
		},
		{Input: "",
			Expected: nil,
			From:     Style{Args: Underline, Fg: "90"},
			To:       Style{Args: Underline, Fg: "90"},
		},
		{Input: "Hello",
			Expected: cells("Hello", Style{Args: Underline, Fg: "90"}),
			From:     Style{Args: Underline, Fg: "90"},
			To:       Style{Args: Underline, Fg: "90"},
		},
	}
	buf := make([]Cell, 0, 128)
	for i, test := range tt {
		got, to := ParseFromTo(test.Input, &buf, test.From)
		if !slices.Equal(got, test.Expected) {
			t.Errorf("#%d: expected cells %v, but got %v", i, test.Expected, got)
		}
		if to != test.To {
			t.Errorf("#%d: expected to %v, but got %v", i, test.To, to)
		}
	}
}

func TestStyleANSI(t *testing.T) {
	tt := []struct {
		name     string
		last     Style
		new      Style
		expected string
	}{
		{
			name:     "both empty",
			last:     Style{},
			new:      Style{},
			expected: "",
		},
		{
			name:     "from empty to fg color",
			last:     Style{},
			new:      Style{Fg: "31"},
			expected: "\x1b[31m",
		},
		{
			name:     "from empty to bg color",
			last:     Style{},
			new:      Style{Bg: "44"},
			expected: "\x1b[44m",
		},
		{
			name:     "from empty to bold",
			last:     Style{},
			new:      Style{Args: Bold},
			expected: "\x1b[1m",
		},
		{
			name:     "from fg to empty -> reset",
			last:     Style{Fg: "31"},
			new:      Style{},
			expected: "\x1b[0m",
		},
		{
			name:     "from bg to empty -> reset",
			last:     Style{Bg: "44"},
			new:      Style{},
			expected: "\x1b[0m",
		},
		{
			name:     "from bold to empty -> reset",
			last:     Style{Args: Bold},
			new:      Style{},
			expected: "\x1b[0m",
		},
		{
			name:     "change fg color",
			last:     Style{Fg: "31"},
			new:      Style{Fg: "32"},
			expected: "\x1b[32m",
		},
		{
			name:     "change bg color",
			last:     Style{Bg: "44"},
			new:      Style{Bg: "45"},
			expected: "\x1b[45m",
		},
		{
			name:     "turn off bold, turn on italic",
			last:     Style{Args: Bold},
			new:      Style{Args: Italic},
			expected: "\x1b[22;3m",
		},
		{
			name:     "turn off italic only",
			last:     Style{Args: Bold | Italic},
			new:      Style{Args: Bold},
			expected: "\x1b[23m",
		},
		{
			name:     "multiple changes: fg and bold to empty",
			last:     Style{Fg: "31", Args: Bold | Underline},
			new:      Style{},
			expected: "\x1b[0m",
		},
		{
			name:     "change fg and bold -> only fg change (bold stays)",
			last:     Style{Fg: "31", Args: Bold},
			new:      Style{Fg: "32", Args: Bold},
			expected: "\x1b[32m",
		},
		{
			name:     "turn off underline, keep fg",
			last:     Style{Fg: "31", Args: Underline},
			new:      Style{Fg: "31"},
			expected: "\x1b[24m",
		},
		{
			name:     "complex: last with italic and bg, new with fg and bold",
			last:     Style{Bg: "44", Args: Italic},
			new:      Style{Fg: "31", Args: Bold},
			expected: "\x1b[1;23;31;49m",
		},
		{
			name:     "from non-empty to empty",
			last:     Style{Fg: "31", Args: Bold | Underline},
			new:      Style{},
			expected: "\x1b[0m",
		},
		{
			name:     "bright black",
			last:     Style{Fg: "90"},
			new:      Style{Fg: "90"},
			expected: "",
		},
		{
			name:     "reverse - set",
			last:     Style{Fg: "31", Args: Reverse},
			new:      Style{Fg: "31"},
			expected: "\033[27m",
		},
		{
			name:     "reverse - reset",
			last:     Style{},
			new:      Style{Args: Reverse},
			expected: "\033[7m",
		},
		{
			name:     "blink - set",
			last:     Style{Fg: "31", Args: Blink},
			new:      Style{Fg: "31"},
			expected: "\033[25m",
		},
		{
			name:     "blink - reset",
			last:     Style{},
			new:      Style{Args: Blink},
			expected: "\033[5m",
		},
		{
			name:     "italic off, bold on",
			last:     Style{Args: Italic},
			new:      Style{Args: Bold},
			expected: "\033[1;23m",
		},
		{
			name:     "dim on",
			last:     Style{},
			new:      Style{Args: Dim},
			expected: "\x1b[2m",
		},
		{
			name:     "dim off",
			last:     Style{Fg: "31", Args: Dim},
			new:      Style{Fg: "31"},
			expected: "\x1b[22m",
		},
		{
			name:     "bold to dim",
			last:     Style{Args: Bold},
			new:      Style{Args: Dim},
			expected: "\x1b[22;2m",
		},
		{
			name:     "dim to bold",
			last:     Style{Args: Dim},
			new:      Style{Args: Bold},
			expected: "\x1b[22;1m",
		},
		{
			name:     "bold+dim to bold",
			last:     Style{Args: Bold | Dim},
			new:      Style{Args: Bold},
			expected: "\x1b[22;1m",
		},
		{
			name:     "bold+dim to dim",
			last:     Style{Args: Bold | Dim},
			new:      Style{Args: Dim},
			expected: "\x1b[22;2m",
		},
		{
			name:     "bold+dim to nothing",
			last:     Style{Args: Bold | Dim, Fg: "31"},
			new:      Style{Fg: "31"},
			expected: "\x1b[22m",
		},
		{
			name:     "bold on, dim stays",
			last:     Style{Args: Dim},
			new:      Style{Args: Bold | Dim},
			expected: "\x1b[1m",
		},
		{
			name:     "hidden on",
			last:     Style{},
			new:      Style{Args: Hidden},
			expected: "\x1b[8m",
		},
		{
			name:     "hidden off",
			last:     Style{Fg: "31", Args: Hidden},
			new:      Style{Fg: "31"},
			expected: "\x1b[28m",
		},
		{
			name:     "strike on",
			last:     Style{},
			new:      Style{Args: Strike},
			expected: "\x1b[9m",
		},
		{
			name:     "strike off",
			last:     Style{Fg: "31", Args: Strike},
			new:      Style{Fg: "31"},
			expected: "\x1b[29m",
		},
		{
			name:     "all attributes on",
			last:     Style{},
			new:      Style{Args: Bold | Dim | Italic | Underline | Reverse | Blink | Hidden | Strike},
			expected: "\x1b[1;2;3;4;7;5;8;9m",
		},
		{
			name:     "all attributes off",
			last:     Style{Args: Bold | Dim | Italic | Underline | Reverse | Blink | Hidden | Strike, Fg: "31"},
			new:      Style{Fg: "31"},
			expected: "\x1b[22;23;24;27;25;28;29m",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.new.ANSI(tc.last)
			if got != tc.expected {
				t.Errorf("ANSI(%+v) = %q, want %q", tc.last, got, tc.expected)
			}
		})
	}
}

func TestToString(t *testing.T) {
	tt := []struct {
		input    [][]Cell
		expected string
	}{
		{
			input: [][]Cell{
				cells("123"),
			},
			expected: "123",
		},
		{
			input: [][]Cell{
				cells("456"),
				cells("123"),
			},
			expected: "456\n123",
		},
		{
			input: [][]Cell{
				cells("456", Style{Fg: "30", Args: Italic}),
				cells("123"),
			},
			expected: "\033[3;30m456\033[0m\n123",
		},
		{
			input: [][]Cell{
				cells("text", Style{Args: Bold | Italic | Underline}),
			},
			expected: "\x1b[1;3;4mtext\x1b[0m",
		},
		{
			input:    [][]Cell{},
			expected: "",
		},
		{
			input: [][]Cell{cells("123", Style{
				Fg: "38;2;200;100;50",
				Bg: "48;2;200;100;50",
			})},
			expected: "\033[38;2;200;100;50;48;2;200;100;50m123\033[0m",
		},
	}

	for i, tc := range tt {
		got := ToString(tc.input)
		if got != tc.expected {
			t.Errorf("#%d: expected %v, but got %v", i, []byte(tc.expected), []byte(got))
		}
	}
}

func TestParseMultiline(t *testing.T) {
	tt := []struct {
		Input    string
		Expected [][]Cell
	}{
		{Input: "",
			Expected: nil,
		},
		{Input: "Hello\nH",
			Expected: [][]Cell{
				cells("Hello"),
				cells("H    "),
			},
		},
	}
	for i, test := range tt {
		got := ParseMultiline(test.Input)
		if !reflect.DeepEqual(got, test.Expected) {
			t.Errorf("#%d: expected cells %v, but got %v", i, test.Expected, got)
		}
	}
}

func FuzzParseANSI(f *testing.F) {
	f.Add("\x1b[38;2;1;2;3m")
	f.Add("\x1b[38;2;1;2m")
	f.Add("\x1b[48;2;255;0m")
	f.Fuzz(func(t *testing.T, s string) {
		var buf []Cell
		_ = Parse(s, &buf)
	})
}
