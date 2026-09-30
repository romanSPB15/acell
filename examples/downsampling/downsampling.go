package main

import (
	"github.com/romanSPB15/acell"
)

func main() {
	in, out := acell.Default()
	t := acell.New(in, out)
	defer t.Close()

	colors := []string{
		"38;2;255;0;0",    // красный
		"38;2;255;128;0",  // оранжевый
		"38;2;255;255;0",  // жёлтый
		"38;2;0;255;0",    // зелёный
		"38;2;0;255;255",  // голубой
		"38;2;0;0;255",    // синий
		"38;2;128;0;255",  // фиолетовый
		"38;2;255;0;255",  // маджента
		"38;2;128;64;32",  // коричневый
		"38;2;64;128;192", // серо-голубой
	}

	render := func() {

		for i, fg := range colors {
			for x := 0; x < 20; x++ {
				t.Buf[i][x] = acell.Cell{Char: '█', Style: acell.Style{Fg: fg}}
			}
		}

		// Информация
		info := t.Info()
		t.DrawString(25, 0, acell.Style{Args: acell.Bold}, "> Info")
		t.DrawString(25, 1, acell.Style{}, "Name: "+info.Name)
		t.DrawString(25, 2, acell.Style{}, "Colors: "+info.Colors.String())

		t.Flush()
	}

	render()

	for ev := range t.Events() {
		switch e := ev.(type) {
		case *acell.KeyboardEvent:
			if e.Rune == 'q' || e.Key == acell.KeyCtrlC || e.Key == acell.KeyEsc {
				return
			}
		case *acell.ResizeEvent:
			t.Buf = acell.NewBuf(e.Width, e.Height)
			t.Invalidate()
			render()
		}
	}
}
