package main

import (
	"encoding/json"

	"github.com/romanSPB15/acell"
)

func main() {
	in, out := acell.Default()
	t := acell.New(in, out)
	defer t.Close()

	render := func() {
		t.DrawString(0, 0, acell.Style{Args: acell.Bold}, "bold")
		t.DrawString(0, 1, acell.Style{Args: acell.Dim}, "dim")
		t.DrawString(0, 2, acell.Style{Args: acell.Underline}, "underline")
		t.DrawString(0, 3, acell.Style{Args: acell.Italic}, "italic")
		t.DrawString(0, 4, acell.Style{Args: acell.Reverse}, "reverse")
		t.DrawString(0, 5, acell.Style{Args: acell.Blink}, "blink")
		t.DrawString(0, 6, acell.Style{Args: acell.Hidden}, "hidden")
		t.DrawString(0, 7, acell.Style{Args: acell.Strike}, "strike")
		t.DrawString(0, 8, acell.Style{Fg: "38;2;255;128;0"}, "rgb orange")
		t.DrawString(0, 9, acell.Style{Fg: "38;5;196"}, "256 red")

		d, _ := json.MarshalIndent(t.Info(), "", "  ")
		t.DrawString(15, 0, acell.Style{}, string(d))

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
			t.Buf = acell.NewBuf(e.Width, e.Height)
			t.Invalidate()
			render()
		}
	}
}
