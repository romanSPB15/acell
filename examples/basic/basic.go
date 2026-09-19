package main

import (
	"time"

	"github.com/romanSPB15/acell"
)

func main() {
	in, out := acell.Default()
	term := acell.New(in, out)

	w, h := term.Size()
	term.Buf = newBuf(w, h)

	draw(term.Buf, time.Now())
	term.Flush()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case ev := <-term.Events():
			switch e := ev.(type) {
			case *acell.KeyboardEvent:
				if e.Rune == 'q' || e.Key == acell.KeyCtrlC {
					term.Close()
					return
				}
			case *acell.MouseEvent:
				if e.Action == acell.MousePress {
					drawAt(term.Buf, e.Pos.X, e.Pos.Y, 'X')
					term.Flush()
				}
			}
		case now := <-ticker.C:
			draw(term.Buf, now)
			term.Flush()
		}
	}
}

// newBuf создаёт пустой буфер заданного размера.
func newBuf(w, h int) [][]acell.Cell {
	buf := make([][]acell.Cell, h)
	for y := range buf {
		buf[y] = make([]acell.Cell, w)
		for x := range buf[y] {
			buf[y][x] = acell.Cell{Char: ' '}
		}
	}
	return buf
}

// draw рисует рамку и текущее время.
func draw(buf [][]acell.Cell, now time.Time) {
	h := len(buf)
	if h == 0 {
		return
	}
	w := len(buf[0])
	if w == 0 {
		return
	}

	// Очистка
	for y := range buf {
		for x := range buf[y] {
			buf[y][x] = acell.Cell{Char: ' '}
		}
	}

	style := acell.Style{Fg: "36"} // cyan

	// Рамка
	for x := 0; x < w; x++ {
		buf[0][x] = acell.Cell{Char: '─', Style: style}
		buf[h-1][x] = acell.Cell{Char: '─', Style: style}
	}
	for y := 0; y < h; y++ {
		buf[y][0] = acell.Cell{Char: '│', Style: style}
		buf[y][w-1] = acell.Cell{Char: '│', Style: style}
	}
	buf[0][0] = acell.Cell{Char: '┌', Style: style}
	buf[0][w-1] = acell.Cell{Char: '┐', Style: style}
	buf[h-1][0] = acell.Cell{Char: '└', Style: style}
	buf[h-1][w-1] = acell.Cell{Char: '┘', Style: style}

	// Текст по центру
	line := now.Format("15:04:05")
	startX := (w - len(line)) / 2
	midY := h / 2
	bold := acell.Style{Fg: "97", Args: acell.Bold}
	for i, r := range line {
		buf[midY][startX+i] = acell.Cell{Char: r, Style: bold}
	}

	// Подсказка внизу
	hint := "press q to quit, click to place X"
	startX = (w - len(hint)) / 2
	hintStyle := acell.Style{Fg: "90"}
	for i, r := range hint {
		buf[h-2][startX+i] = acell.Cell{Char: r, Style: hintStyle}
	}
}

// drawAt ставит символ в указанную позицию, если она внутри буфера.
func drawAt(buf [][]acell.Cell, x, y int, ch rune) {
	if y < 0 || y >= len(buf) {
		return
	}
	if x < 0 || x >= len(buf[y]) {
		return
	}
	buf[y][x] = acell.Cell{Char: ch, Style: acell.Style{Fg: "33"}}
}
