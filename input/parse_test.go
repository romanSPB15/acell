package input

import (
	"reflect"
	"testing"
)

func TestParseAnsi(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		wantRune rune
		wantKey  Key
	}{
		{"enter", []byte{13}, 0, KeyEnter},
		{"tab", []byte{9}, 0, KeyTab},
		{"space", []byte{32}, ' ', KeySpace},
		{"slash", []byte{47}, '/', KeySlash},
		{"backslash", []byte{92}, '\\', KeyReverseSlash},
		{"backspace", []byte{127}, 0, KeyBackspace},
		{"normal 'a'", []byte{'a'}, 'a', KeyUnknown},
		{"normal 'A'", []byte{'A'}, 'A', KeyUnknown},
		{"ctrl+a", []byte{1}, 0, KeyCtrlA},
		{"ctrl+z", []byte{26}, 0, KeyCtrlZ},
		{"unknown", []byte{0}, 0, KeyUnknown},

		{"pgup", []byte{27, 91, 53, 126}, 0, KeyPgUp},
		{"pgdown", []byte{27, 91, 54, 126}, 0, KeyPgDown},
		{"shifttab", []byte{27, 91, 90}, 0, KeyShiftTab},
		{"delete", []byte{27, 91, 51, 126}, 0, KeyDelete},
		{"end", []byte{27, 91, 70}, 0, KeyEnd},
		{"home", []byte{27, 91, 72}, 0, KeyHome},
		{"insert", []byte{27, 91, 50, 126}, 0, KeyInsert},
		{"f1", []byte{27, 79, 80}, 0, KeyF1},
		{"f2", []byte{27, 79, 81}, 0, KeyF2},
		{"f3", []byte{27, 79, 82}, 0, KeyF3},
		{"f4", []byte{27, 79, 83}, 0, KeyF4},
		{"f5", []byte{27, 91, 49, 53, 126}, 0, KeyF5},
		{"f6", []byte{27, 91, 49, 55, 126}, 0, KeyF6},
		{"f7", []byte{27, 91, 49, 56, 126}, 0, KeyF7},
		{"f8", []byte{27, 91, 49, 57, 126}, 0, KeyF8},
		{"f9", []byte{27, 91, 50, 48, 126}, 0, KeyF9},
		{"f10", []byte{27, 91, 50, 49, 126}, 0, KeyF10},
		{"f11", []byte{27, 91, 50, 51, 126}, 0, KeyF11},
		{"f12", []byte{27, 91, 50, 52, 126}, 0, KeyF12},
		{"up", []byte{27, 91, 65}, 0, KeyArrowUp},
		{"right", []byte{27, 91, 67}, 0, KeyArrowRight},
		{"down", []byte{27, 91, 66}, 0, KeyArrowDown},
		{"left", []byte{27, 91, 68}, 0, KeyArrowLeft},
		{"unknown seq", []byte{27, 91, 1}, 0, KeyUnknown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, k := parseAnsi(tt.data)
			if r != tt.wantRune || k != tt.wantKey {
				t.Errorf("got (%q, %v), want (%q, %v)", r, k, tt.wantRune, tt.wantKey)
			}
		})
	}
}

func TestParseKeyboard(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want *KeyboardEvent
	}{
		{"empty", []byte{}, nil},
		{"normal 'a'", []byte{'a'}, &KeyboardEvent{Key: KeyUnknown, Rune: 'a', Alt: false}},
		{"enter", []byte{13}, &KeyboardEvent{Key: KeyEnter, Rune: 0, Alt: false}},
		{"ctrl+a", []byte{1}, &KeyboardEvent{Key: KeyCtrlA, Rune: 0, Alt: false}},
		{"alt+a", []byte{27, 'a'}, &KeyboardEvent{Key: KeyUnknown, Rune: 'a', Alt: true}},
		{"alt+enter", []byte{27, 13}, &KeyboardEvent{Key: KeyEnter, Rune: 0, Alt: true}},
		{"alt+up", []byte{27, 27, 91, 65}, &KeyboardEvent{Key: KeyArrowUp, Rune: 0, Alt: true}},
		{"unknown", []byte{0}, nil},

		// Mouse-последовательность не должна разбираться как клавиатура:
		// иначе ESC съестся и мы получим мусорное событие.
		{"mouse seq returns nil", []byte("\x1b[<0;10;20M"), nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseKeyboard(tt.data)
			if !reflect.DeepEqual(got, tt.want) {
				if got == nil {
					t.Errorf("got <nil>, want %v", *tt.want)
				} else {
					t.Errorf("got %v, want %v", got, *tt.want)
				}
			}
		})
	}
}

func TestParseMouse(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want *MouseEvent
	}{
		{"left press", []byte("\x1b[<0;10;20M"), &MouseEvent{Action: MousePress, Button: 0, Pos: Point{X: 9, Y: 19}}},
		{"middle press", []byte("\x1b[<1;5;5M"), &MouseEvent{Action: MousePress, Button: 1, Pos: Point{X: 4, Y: 4}}},
		{"right press", []byte("\x1b[<2;1;1M"), &MouseEvent{Action: MousePress, Button: 2, Pos: Point{X: 0, Y: 0}}},

		{"left release", []byte("\x1b[<0;10;20m"), &MouseEvent{Action: MouseRelease, Button: 0, Pos: Point{X: 9, Y: 19}}},
		{"right release", []byte("\x1b[<2;10;20m"), &MouseEvent{Action: MouseRelease, Button: 2, Pos: Point{X: 9, Y: 19}}},

		{"drag left", []byte("\x1b[<32;10;20M"), &MouseEvent{Action: MouseMove, Button: 0, Pos: Point{X: 9, Y: 19}}},
		{"drag middle", []byte("\x1b[<33;10;20M"), &MouseEvent{Action: MouseMove, Button: 1, Pos: Point{X: 9, Y: 19}}},
		{"pure motion", []byte("\x1b[<35;10;20M"), &MouseEvent{Action: MouseMove, Button: NoButton, Pos: Point{X: 9, Y: 19}}},

		{"wheel up", []byte("\x1b[<64;10;20M"), &MouseEvent{Action: MouseWheelUp, Button: NoButton, Pos: Point{X: 9, Y: 19}}},
		{"wheel down", []byte("\x1b[<65;10;20M"), &MouseEvent{Action: MouseWheelDown, Button: NoButton, Pos: Point{X: 9, Y: 19}}},
		{"wheel left", []byte("\x1b[<66;10;20M"), &MouseEvent{Action: MouseWheelLeft, Button: NoButton, Pos: Point{X: 9, Y: 19}}},
		{"wheel right", []byte("\x1b[<67;10;20M"), &MouseEvent{Action: MouseWheelRight, Button: NoButton, Pos: Point{X: 9, Y: 19}}},

		{"shift+click", []byte("\x1b[<4;10;20M"), &MouseEvent{Action: MousePress, Button: 0, Shift: true, Pos: Point{X: 9, Y: 19}}},
		{"alt+click", []byte("\x1b[<8;10;20M"), &MouseEvent{Action: MousePress, Button: 0, Alt: true, Pos: Point{X: 9, Y: 19}}},
		{"ctrl+click", []byte("\x1b[<16;10;20M"), &MouseEvent{Action: MousePress, Button: 0, Ctrl: true, Pos: Point{X: 9, Y: 19}}},
		{"ctrl+wheel", []byte("\x1b[<80;10;20M"), &MouseEvent{Action: MouseWheelUp, Button: NoButton, Ctrl: true, Pos: Point{X: 9, Y: 19}}},

		{"invalid prefix", []byte("abc"), nil},
		{"invalid parts", []byte("\x1b[<0;10m"), nil},
		{"invalid suffix", []byte("\x1b[<0;10;20X"), nil},
		{"invalid number 1", []byte("\x1b[<a;10;20M"), nil},
		{"invalid number 2", []byte("\x1b[<1;a;20M"), nil},
		{"invalid number 3", []byte("\x1b[<1;10;aM"), nil},
		{"negative cb", []byte("\x1b[<-1;10;20M"), nil},
		{"empty", []byte{}, nil},
		{"short", []byte("\x1b[<"), nil},

		// Keyboard-последовательность не должна разбираться как мышь.
		{"keyboard seq returns nil", []byte{'a'}, nil},
		{"arrow returns nil", []byte{27, 91, 65}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseMouse(tt.data)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParsePriority(t *testing.T) {
	cases := []struct {
		name string
		data []byte
		want any
	}{
		{"focus in", []byte("\x1b[I"), (*WindowFocusEvent)(nil)},
		{"focus out", []byte("\x1b[O"), (*WindowFocusEvent)(nil)},
		{"mouse", []byte("\x1b[<0;10;20M"), (*MouseEvent)(nil)},
		{"keyboard rune", []byte{'a'}, (*KeyboardEvent)(nil)},
		{"keyboard arrow", []byte{27, 91, 65}, (*KeyboardEvent)(nil)},
		{"garbage", []byte{0xff, 0xff}, nil},
		{"empty", []byte{}, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Parse(tc.data)
			if tc.want == nil {
				if got != nil {
					t.Errorf("got %T, want nil", got)
				}
				return
			}
			if reflect.TypeOf(got) != reflect.TypeOf(tc.want) {
				t.Errorf("got %T, want %T", got, tc.want)
			}
		})
	}
}

func TestParseFocus(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want *WindowFocusEvent
	}{
		{"focus in", []byte("\x1b[I"), &WindowFocusEvent{Focused: true}},
		{"focus out", []byte("\x1b[O"), &WindowFocusEvent{Focused: false}},
		{"garbage", []byte("abc"), nil},
		{"empty", nil, nil},
		{"keyboard", []byte{'a'}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseFocus(tt.data)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseOne(t *testing.T) {
	tests := []struct {
		name       string
		data       []byte
		wantEvType string
		wantN      int
	}{
		{"empty", []byte{}, "", 0},

		{"focus in", []byte("\x1b[I"), "focus", 3},
		{"focus out", []byte("\x1b[O"), "focus", 3},

		{"mouse press", []byte("\x1b[<0;10;20M"), "mouse", 11},
		{"mouse release", []byte("\x1b[<0;10;20m"), "mouse", 11},
		{"mouse incomplete", []byte("\x1b[<0;10;20"), "", 0},
		{"mouse invalid body", []byte("\x1b[<a;10;20M"), "", 11},

		{"esc alone", []byte{27}, "", 0},

		{"arrow up", []byte("\x1b[A"), "keyboard", 3},
		{"f1", []byte("\x1bOP"), "keyboard", 3},
		{"csi incomplete", []byte("\x1b["), "", 0},
		{"csi invalid", []byte("\x1b[\x01A"), "keyboard", 4},

		{"ss3 incomplete", []byte("\x1bO"), "", 0},
		{"ss3 f1", []byte("\x1bOP"), "keyboard", 3},
		{"ss3 invalid", []byte("\x1bOz"), "keyboard", 3},

		{"alt+a", []byte{27, 'a'}, "keyboard", 2},
		{"alt+enter", []byte{27, 13}, "keyboard", 2},

		{"ascii a", []byte{'a'}, "keyboard", 1},
		{"ascii 0", []byte{0}, "", 1},

		{"utf8 2byte", []byte("é"), "keyboard", 2},
		{"utf8 3byte", []byte("€"), "keyboard", 3},
		{"utf8 4byte", []byte("😀"), "keyboard", 4},
		{"utf8 2byte incomplete", []byte{0xC3}, "", 0},
		{"utf8 3byte incomplete", []byte{0xE2, 0x82}, "", 0},
		{"utf8 4byte incomplete", []byte{0xF0, 0x9F, 0x98}, "", 0},
		{"utf8 invalid lead", []byte{0xFF}, "", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ev, n := ParseOne(tt.data)
			if n != tt.wantN {
				t.Errorf("n = %d, want %d", n, tt.wantN)
			}
			gotType := ""
			switch ev.(type) {
			case *WindowFocusEvent:
				gotType = "focus"
			case *MouseEvent:
				gotType = "mouse"
			case *KeyboardEvent:
				gotType = "keyboard"
			case nil:
				gotType = ""
			}
			if gotType != tt.wantEvType {
				t.Errorf("event = %s, want %s", gotType, tt.wantEvType)
			}
		})
	}
}
