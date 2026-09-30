package main

import (
	"fmt"
	"os"
	"sort"
	"sync/atomic"

	"github.com/romanSPB15/acell"
	"github.com/romanSPB15/acell/terminfo"
)

// relevantEnvKeys — переменные, важные для диагностики терминала.
var relevantEnvKeys = []string{
	"TERM",
	"TERM_PROGRAM",
	"TERM_PROGRAM_VERSION",
	"COLORTERM",
	"WT_SESSION",
	"WT_PROFILE_ID",
	"MINTTY",
	"MSYSTEM",
	"LANG",
	"LC_ALL",
	"LC_CTYPE",
	"SSH_TTY",
	"SSH_CLIENT",
	"TMUX",
	"KITTY_WINDOW_ID",
	"WEZTERM_EXECUTABLE",
	"ALACRITTY_LOG",
	"VTE_VERSION",
	"GOOS",
	"GOARCH",
}

func collectEnv() []string {
	out := make([]string, 0, len(relevantEnvKeys))
	for _, k := range relevantEnvKeys {
		if v, ok := os.LookupEnv(k); ok {
			out = append(out, k+"="+v)
		}
	}
	sort.Strings(out)
	return out
}

func main() {
	in, out := acell.Default()
	t := acell.NewWithInfo(in, out, terminfo.All)
	defer t.Close()

	w, h := t.Size()

	var focus atomic.Bool
	focus.Store(true)

	mouseX, mouseY := -1, -1
	var lastMouse atomic.Value

	env := collectEnv()

	render := func() {
		t.Clear()

		// env vars на правой половине
		const envX = 30
		for i := 0; i < h; i++ {
			idx := i
			if idx >= len(env) {
				break
			}
			line := env[idx]
			maxW := w - envX
			if maxW < 1 {
				maxW = 1
			}
			if len(line) > maxW {
				line = line[:maxW]
			}
			t.DrawString(envX, i, acell.Style{Fg: "38;5;250"}, line)
		}

		t.DrawString(0, 0, acell.Style{Args: acell.Bold}, "bold")
		t.DrawString(0, 1, acell.Style{Args: acell.Dim}, "dim")
		t.DrawString(0, 2, acell.Style{Args: acell.Underline}, "underline")
		t.DrawString(0, 3, acell.Style{Args: acell.Italic}, "italic")
		t.DrawString(0, 4, acell.Style{Args: acell.Reverse}, "reverse")
		t.DrawString(0, 5, acell.Style{Args: acell.Blink}, "blink")
		t.DrawString(0, 6, acell.Style{Args: acell.Hidden}, "hidden")
		t.DrawString(0, 7, acell.Style{Args: acell.Strike}, "strike")
		t.DrawString(0, 8, acell.Style{Fg: acell.FgRGB(255, 128, 0)}, "rgb orange")
		t.DrawString(0, 9, acell.Style{Fg: acell.FgRed}, "256 red")

		if t.Info().WindowFocusEvents {
			if focus.Load() {
				t.DrawString(0, 11, acell.Style{Fg: acell.FgGreenYellow}, "Window in focus")
			} else {
				t.DrawString(0, 11, acell.Style{Fg: acell.FgOrangeRed}, "Focus lost")
			}
		}

		mouseLabel := "Mouse: no events yet"
		if s, ok := lastMouse.Load().(string); ok && s != "" {
			mouseLabel = "Mouse: " + s
		}
		t.DrawString(0, 12, acell.Style{Fg: acell.FgRoyalBlue}, mouseLabel)
		if mouseX >= 0 {
			t.DrawString(0, 13, acell.Style{Fg: acell.FgYellow},
				fmt.Sprintf("Last position: X=%d Y=%d", mouseX, mouseY))
		}

		t.DrawString(0, 14, acell.Style{Fg: acell.FgRoyalBlue},
			fmt.Sprintf("Window size: %dx%d", w, h))

		t.Flush()
	}

	render()

	for ev := range t.Events() {
		switch e := ev.(type) {
		case *acell.KeyboardEvent:
			if e.Key == acell.KeyCtrlC || e.Rune == 'q' || e.Key == acell.KeyEsc {
				return
			}
		case *acell.ResizeEvent:
			w, h = e.Width, e.Height
			t.Buf = acell.NewBuf(e.Width, e.Height)
			t.Invalidate()
			render()
		case *acell.WindowFocusEvent:
			focus.Store(e.Focused)
			render()
		case *acell.MouseEvent:
			mouseX, mouseY = e.Pos.X, e.Pos.Y
			action := "?"
			switch e.Action {
			case acell.MousePress:
				action = fmt.Sprintf("press btn=%d", e.Button)
			case acell.MouseRelease:
				action = fmt.Sprintf("release btn=%d", e.Button)
			case acell.MouseMove:
				action = fmt.Sprintf("move btn=%d", e.Button)
			case acell.MouseWheelUp:
				action = "wheel up"
			case acell.MouseWheelDown:
				action = "wheel down"
			case acell.MouseWheelLeft:
				action = "wheel left"
			case acell.MouseWheelRight:
				action = "wheel right"
			}
			mods := ""
			if e.Shift {
				mods += " shift"
			}
			if e.Alt {
				mods += " alt"
			}
			if e.Ctrl {
				mods += " ctrl"
			}
			lastMouse.Store(fmt.Sprintf("%s%s @ (%d, %d)", action, mods, e.Pos.X, e.Pos.Y))
			render()
		}
	}
}
