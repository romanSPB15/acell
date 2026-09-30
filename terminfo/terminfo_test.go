package terminfo

import (
	"encoding/json"
	"testing"
)

func TestColorBitsString(t *testing.T) {
	for i, v := range []struct {
		Cb       ColorBits
		Expected string
	}{
		{ColorNone, "None"},
		{Color8, "8-colors"},
		{Color16, "16-colors"},
		{Color256, "256-colors"},
		{ColorTrue, "True Color"},
	} {
		got := v.Cb.String()
		if got != v.Expected {
			t.Fatalf("#%d: ColorBits.String: expected '%s', but got: '%s'", i, v.Expected, got)
		}
	}
}

func TestColorBitsMarshalJSON(t *testing.T) {
	b, err := json.Marshal(Color256)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != `"256-colors"` {
		t.Errorf("got %s, want %q", b, "256-color")
	}
}

func clearEnv(t *testing.T) {
	t.Helper()
	for _, k := range []string{
		"TERM", "TERM_PROGRAM", "COLORTERM", "WT_SESSION", "MINTTY", "OS",
	} {
		t.Setenv(k, "")
	}
}

func TestDetect(t *testing.T) {
	t.Run("defaults with no env", func(t *testing.T) {
		clearEnv(t)
		info := Detect()
		if info.Name != "xterm-256color" {
			t.Errorf("Name = %q", info.Name)
		}
		if info.Colors != Color256 {
			t.Errorf("Colors = %v, want 256", info.Colors)
		}
		if info.AltScreenOn == "" {
			t.Errorf("AltScreenOn must survive from Default()")
		}
	})

	t.Run("TERM sets name", func(t *testing.T) {
		clearEnv(t)
		t.Setenv("TERM", "screen")
		info := Detect()
		if info.Name != "screen" {
			t.Errorf("Name = %q, want screen", info.Name)
		}
	})

	t.Run("COLORTERM truecolor", func(t *testing.T) {
		clearEnv(t)
		t.Setenv("TERM", "xterm-256color")
		t.Setenv("COLORTERM", "truecolor")
		info := Detect()
		if info.Colors != ColorTrue {
			t.Errorf("Colors = %v, want True", info.Colors)
		}
	})

	t.Run("COLORTERM 24bit", func(t *testing.T) {
		clearEnv(t)
		t.Setenv("TERM", "xterm-256color")
		t.Setenv("COLORTERM", "24bit")
		info := Detect()
		if info.Colors != ColorTrue {
			t.Errorf("Colors = %v, want True", info.Colors)
		}
	})

	t.Run("vscode", func(t *testing.T) {
		clearEnv(t)
		t.Setenv("TERM_PROGRAM", "vscode")
		info := Detect()
		if info.Name != "vscode" {
			t.Errorf("Name = %q", info.Name)
		}
		if info.CursorHide != "" || info.CursorShow != "" {
			t.Errorf("cursor hide/show must be empty, got %q/%q", info.CursorHide, info.CursorShow)
		}
		if info.Blink {
			t.Errorf("Blink must be false")
		}
	})

	t.Run("windows terminal", func(t *testing.T) {
		clearEnv(t)
		t.Setenv("WT_SESSION", "some-guid")
		info := Detect()
		if info.Name != "windows-terminal" {
			t.Errorf("Name = %q", info.Name)
		}
		if info.Colors != ColorTrue {
			t.Errorf("Colors = %v, want True", info.Colors)
		}
		if !info.SynchronizedUpdate {
			t.Errorf("SynchronizedUpdate must be true")
		}
		if info.Blink {
			t.Errorf("Blink must be false")
		}
	})

	t.Run("mintty by MINTTY env", func(t *testing.T) {
		clearEnv(t)
		t.Setenv("MINTTY", "1")
		info := Detect()
		if info.Name != "mintty" {
			t.Errorf("Name = %q", info.Name)
		}
		if info.Colors != ColorTrue {
			t.Errorf("Colors = %v, want True", info.Colors)
		}
		if info.MouseAny || info.MouseSGR {
			t.Errorf("mouse must be off")
		}
		if info.CursorHide != "" || info.CursorShow != "" {
			t.Errorf("cursor must be off")
		}
		if info.SynchronizedUpdate {
			t.Errorf("sync update must be off")
		}
		if info.WindowFocusEvents {
			t.Errorf("focus events must be off")
		}
	})

	t.Run("mintty by TERM_PROGRAM", func(t *testing.T) {
		clearEnv(t)
		t.Setenv("TERM_PROGRAM", "mintty")
		info := Detect()
		if info.Name != "mintty" {
			t.Errorf("Name = %q", info.Name)
		}
	})

	t.Run("mintty by TERM substring", func(t *testing.T) {
		clearEnv(t)
		t.Setenv("TERM", "xterm-mintty")
		info := Detect()
		if info.Name != "mintty" {
			t.Errorf("Name = %q", info.Name)
		}
	})

	t.Run("cygwin", func(t *testing.T) {
		clearEnv(t)
		t.Setenv("TERM", "cygwin")
		info := Detect()
		if info.Name != "cygwin" {
			t.Errorf("Name = %q", info.Name)
		}
		if info.Colors != Color16 {
			t.Errorf("Colors = %v, want 16", info.Colors)
		}
		if info.Dim || info.Italic || info.Strike {
			t.Errorf("Dim/Italic/Strike must be off")
		}
		if info.SynchronizedUpdate {
			t.Errorf("sync update must be off")
		}
	})

	t.Run("msys", func(t *testing.T) {
		clearEnv(t)
		t.Setenv("TERM", "msys")
		info := Detect()
		if info.Name != "cygwin" {
			t.Errorf("Name = %q", info.Name)
		}
	})

	t.Run("linux console", func(t *testing.T) {
		clearEnv(t)
		t.Setenv("TERM", "linux")
		info := Detect()
		if info.Name != "linux" {
			t.Errorf("Name = %q", info.Name)
		}
		if info.Colors != Color8 {
			t.Errorf("Colors = %v, want 8", info.Colors)
		}
		if info.AltScreenOn != "" || info.AltScreenOff != "" {
			t.Errorf("alt-screen must be off")
		}
		if info.CursorHide != "" || info.CursorShow != "" {
			t.Errorf("cursor must be off")
		}
		if info.MouseAny || info.MouseSGR {
			t.Errorf("mouse must be off")
		}
		if info.SynchronizedUpdate {
			t.Errorf("sync update must be off")
		}
		if info.Italic || info.Strike {
			t.Errorf("Italic/Strike must be off")
		}
	})

	t.Run("dumb", func(t *testing.T) {
		clearEnv(t)
		t.Setenv("TERM", "dumb")
		info := Detect()
		if info.Name != "dumb" {
			t.Errorf("Name = %q", info.Name)
		}
		if info.Colors != ColorNone {
			t.Errorf("Colors = %v, want None", info.Colors)
		}
		if info.AltScreenOn != "" || info.AltScreenOff != "" {
			t.Errorf("alt-screen must be off")
		}
		if info.Sgr0 != "" || info.Op != "" {
			t.Errorf("sgr must be off")
		}
		if info.Bold || info.Dim || info.Italic || info.Underline ||
			info.Reverse || info.Blink || info.Hidden || info.Strike {
			t.Errorf("all attrs must be off")
		}
		if info.MouseAny || info.MouseSGR {
			t.Errorf("mouse must be off")
		}
		if info.SynchronizedUpdate || info.WindowFocusEvents {
			t.Errorf("sync/focus must be off")
		}
		if info.CursorHide != "" || info.CursorShow != "" {
			t.Errorf("cursor must be off")
		}
	})

	t.Run("conhost", func(t *testing.T) {
		clearEnv(t)
		t.Setenv("OS", "Windows_NT")
		info := Detect()
		if info.Name != "windows-conhost" {
			t.Errorf("Name = %q", info.Name)
		}
		if info.Colors != ColorTrue {
			t.Errorf("Colors = %v, want True", info.Colors)
		}
		if info.SynchronizedUpdate {
			t.Errorf("sync update must be off on conhost")
		}
	})

	t.Run("16color suffix", func(t *testing.T) {
		clearEnv(t)
		t.Setenv("TERM", "xterm-16color")
		info := Detect()
		if info.Colors != Color16 {
			t.Errorf("Colors = %v, want 16", info.Colors)
		}
	})

	t.Run("256color suffix", func(t *testing.T) {
		clearEnv(t)
		t.Setenv("TERM", "xterm-256color")
		info := Detect()
		if info.Colors != Color256 {
			t.Errorf("Colors = %v, want 256", info.Colors)
		}
	})

	t.Run("plain xterm -> 8 colors", func(t *testing.T) {
		clearEnv(t)
		t.Setenv("TERM", "xterm")
		info := Detect()
		if info.Colors != Color8 {
			t.Errorf("Colors = %v, want 8", info.Colors)
		}
	})

	t.Run("screen -> 8 colors", func(t *testing.T) {
		clearEnv(t)
		t.Setenv("TERM", "screen")
		info := Detect()
		if info.Colors != Color8 {
			t.Errorf("Colors = %v, want 8", info.Colors)
		}
	})

	t.Run("tmux -> 8 colors", func(t *testing.T) {
		clearEnv(t)
		t.Setenv("TERM", "tmux")
		info := Detect()
		if info.Colors != Color8 {
			t.Errorf("Colors = %v, want 8", info.Colors)
		}
	})
}
