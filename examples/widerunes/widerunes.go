package main

import (
	"math"
	"time"

	"github.com/romanSPB15/acell"
	"github.com/romanSPB15/acell/builder"
)

var gradBuf builder.Builder

func main() {
	in, out := acell.Default()
	t := acell.New(in, out)
	defer t.Close()

	phase := 0.0

	render := func() {
		w, h := t.Size()
		if w == 0 || h == 0 {
			return
		}

		for y := range t.Buf {
			for x := range t.Buf[y] {
				t.Buf[y][x] = acell.Cell{Char: ' '}
			}
		}

		boxW := 50
		boxH := 10
		offX := (w - boxW) / 2
		offY := (h - boxH) / 2

		/************************Градиент*********************************/

		grad := func(x, y int) acell.Style {
			u := float64(x)/float64(boxW) + float64(y)/float64(boxH)*0.5
			r := uint8(127 + 128*math.Sin(u+phase))
			g := uint8(127 + 128*math.Sin(u+phase+1))
			b := uint8(127 + 128*math.Sin(u+phase+3))

			gradBuf.Reset()
			gradBuf.Grow(16)
			gradBuf.WriteString("38;2;")
			gradBuf.WriteUint(uint(r))
			gradBuf.WriteByte(';')
			gradBuf.WriteUint(uint(g))
			gradBuf.WriteByte(';')
			gradBuf.WriteUint(uint(b))
			return acell.Style{Fg: gradBuf.StringCopy()}
		}

		/***************************Фон***********************************/

		for y := range h {
			for x := 0; x < w; x += 2 {
				if x%4 == 0 {
					t.DrawRune(x, y, acell.Style{Fg: "38;2;50;50;50"}, '猫')
				} else {
					t.DrawRune(x, y, acell.Style{Fg: "38;2;50;50;50"}, '咪')
				}
			}
		}

		/**************************Рамка**********************************/

		t.DrawRune(offX, offY, grad(0, 0), '┌')
		t.DrawRune(offX+boxW-1, offY, grad(boxW-1, 0), '┐')
		t.DrawRune(offX, offY+boxH-1, grad(0, boxH-1), '└')
		t.DrawRune(offX+boxW-1, offY+boxH-1, grad(boxW-1, boxH-1), '┘')

		for x := 1; x < boxW-1; x++ {
			t.DrawRune(offX+x, offY, grad(x, 0), '─')
			t.DrawRune(offX+x, offY+boxH-1, grad(x, boxH-1), '─')
		}
		for y := 1; y < boxH-1; y++ {
			t.DrawRune(offX, offY+y, grad(0, y), '│')
			t.DrawRune(offX+boxW-1, offY+y, grad(boxW-1, y), '│')
		}

		for y := offY + 1; y < offY+boxH-1; y++ {
			for x := offX + 1; x < offX+boxW-1; x++ {
				t.DrawRune(x, y, acell.Style{}, ' ')
			}
		}

		/************************Содержимое********************************/

		_ = grad

		t.Flush()
	}

	render()

	ticker := time.NewTicker(time.Second / 30)
	defer ticker.Stop()

	for {
		select {
		case ev := <-t.Events():
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
		case <-ticker.C:
			phase += 0.02
			render()
		}
	}
}
