// Package terminfo определяет возможности текущего терминала.
package terminfo

import (
	"encoding/json"
	"os"
	"strings"
)

// ColorBits описывает глубину цвета терминала.
type ColorBits int

const (
	ColorNone ColorBits = 0  // без цвета
	Color8    ColorBits = 3  // 8 цветов
	Color16   ColorBits = 4  // 16 цветов
	Color256  ColorBits = 8  // 256 цветов
	ColorTrue ColorBits = 24 // 16M цветов (True Color)
)

func (cb ColorBits) String() string {
	switch cb {
	case Color8:
		return "8-colors"
	case Color16:
		return "16-colors"
	case Color256:
		return "256-color"
	case ColorTrue:
		return "True Color"
	}
	return "None"
}

func (c ColorBits) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.String())
}

// Info описывает возможности терминала.
type Info struct {
	Name string

	// Colors — глубина цвета.
	Colors ColorBits

	CursorHide string
	CursorShow string

	AltScreenOn  string
	AltScreenOff string

	Clear string
	Home  string

	Sgr0 string
	Op   string

	Bold      bool
	Dim       bool
	Italic    bool
	Underline bool
	Reverse   bool
	Blink     bool
	Hidden    bool
	Strike    bool

	MouseAny bool
	MouseSGR bool

	SynchronizedUpdate bool
}

// Default возвращает Info с полным набором возможностей (xterm-256color).
func Default() Info {
	return Info{
		Name:   "xterm-256color",
		Colors: Color256,

		CursorHide:   "\033[?25l",
		CursorShow:   "\033[?12l\033[?25h",
		AltScreenOn:  "\033[?1049h",
		AltScreenOff: "\033[?1049l",
		Clear:        "\033[2J\033[H",
		Home:         "\033[H",

		Sgr0: "\033(B\033[m",
		Op:   "\033[39;49m",

		Bold:      true,
		Dim:       true,
		Italic:    true,
		Underline: true,
		Reverse:   true,
		Blink:     true,
		Hidden:    true,
		Strike:    true,

		MouseAny: true,
		MouseSGR: true,

		SynchronizedUpdate: true,
	}
}

// Detect возвращает Info для текущего окружения.
func Detect() Info {
	info := Default()

	term := os.Getenv("TERM")
	termProgram := os.Getenv("TERM_PROGRAM")
	colorTerm := os.Getenv("COLORTERM")
	wtSession := os.Getenv("WT_SESSION")

	if term != "" {
		info.Name = term
	}

	// True Color из $COLORTERM
	if colorTerm == "truecolor" || colorTerm == "24bit" {
		info.Colors = ColorTrue
	}

	// Windows Terminal
	if wtSession != "" {
		info.Name = "windows-terminal"
		info.Colors = ColorTrue
		info.SynchronizedUpdate = true
		info.Blink = false
		return info
	}

	// VS Code Terminal
	if termProgram == "vscode" {
		info.Name = "vscode"
		info.Colors = ColorTrue
		info.CursorHide = ""
		info.CursorShow = ""
		info.SynchronizedUpdate = false
		info.Blink = false
		return info
	}

	// mintty (Git Bash)
	if os.Getenv("MINTTY") != "" ||
		termProgram == "mintty" ||
		strings.Contains(term, "mintty") {
		info.Name = "mintty"
		info.Colors = ColorTrue
		info.SynchronizedUpdate = false

		info.Blink = false
		info.MouseAny = false
		info.MouseSGR = false
		info.AltScreenOn = ""
		info.AltScreenOff = ""
		info.CursorHide = ""
		info.CursorShow = ""

		return info
	}

	// cygwin / msys
	if strings.HasPrefix(term, "cygwin") || strings.HasPrefix(term, "msys") {
		info.Name = "cygwin"
		info.Colors = Color16
		info.Dim = false
		info.Italic = false
		info.Strike = false
		info.SynchronizedUpdate = false
		return info
	}

	// Linux-консоль (без X)
	if term == "linux" {
		info.Name = "linux"
		info.Colors = Color8
		info.CursorHide = ""
		info.CursorShow = ""
		info.AltScreenOn = ""
		info.AltScreenOff = ""
		info.Italic = false
		info.Strike = false
		info.MouseAny = false
		info.MouseSGR = false
		info.SynchronizedUpdate = false
		return info
	}

	// dumb
	if term == "dumb" {
		info.Name = "dumb"
		info.Colors = ColorNone
		info.CursorHide = ""
		info.CursorShow = ""
		info.AltScreenOn = ""
		info.AltScreenOff = ""
		info.Sgr0 = ""
		info.Op = ""
		info.Bold = false
		info.Dim = false
		info.Italic = false
		info.Underline = false
		info.Reverse = false
		info.Blink = false
		info.Hidden = false
		info.Strike = false
		info.MouseAny = false
		info.MouseSGR = false
		info.SynchronizedUpdate = false
		return info
	}

	// ConHost (cmd.exe, PowerShell, Windows PowerShell) — не WT, не mintty
	if term == "" && os.Getenv("OS") == "Windows_NT" {
		info.Name = "windows-conhost"
		info.Colors = ColorTrue
		info.SynchronizedUpdate = false

		if windowsBuild() < 22000 {
			info.Dim = false
			info.Italic = false
			info.Blink = false
			info.Hidden = false
			info.Strike = false
			info.CursorHide = ""
			info.CursorShow = ""
		}

		return info
	}

	if info.Colors != ColorTrue {
		switch {
		case strings.Contains(term, "256color"):
			info.Colors = Color256
		case strings.Contains(term, "16color"):
			info.Colors = Color16
		case term == "xterm" || term == "screen" || term == "tmux":
			info.Colors = Color8
		}
	}

	return info
}
