package main

import "github.com/romanSPB15/acell"

func main() {
	in, out := acell.Default()
	t := acell.New(in, out)
	defer t.Close()

	draw := func() {
		t.Clear()
		t.DrawString(0, 0, acell.Style{}, "Hello,")
		t.DrawString(7, 0, acell.Style{Fg: acell.FgRed, Args: acell.Bold}, "acell")
		t.DrawRune(12, 0, acell.Style{}, '!')
		t.DrawString(0, 2, acell.Style{Fg: acell.Fg16Grey}, "press q to quit")
		t.Flush()
	}

	draw()

	for ev := range t.Events() {
		switch e := ev.(type) {
		case *acell.KeyboardEvent:
			if e.Rune == 'q' || e.Key == acell.KeyCtrlC {
				return
			}
		case *acell.ResizeEvent:
			t.Buf = acell.NewBuf(e.Width, e.Height)
			t.Invalidate()
			draw()
		}
	}
}
