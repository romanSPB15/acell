package main

import (
	"math"
	"time"
	"unicode/utf8"

	"github.com/romanSPB15/acell"
	"github.com/romanSPB15/acell/builder"
)

const scale = 4

var font = map[rune][]string{
	'a': {
		".###.",
		"#...#",
		"#...#",
		"#####",
		"#...#",
		"#...#",
		"#...#",
	},
	'c': {
		".###.",
		"#...#",
		"#....",
		"#....",
		"#....",
		"#...#",
		".###.",
	},
	'e': {
		".###.",
		"#...#",
		"#...#",
		"#####",
		"#....",
		"#...#",
		".###.",
	},
	'l': {
		".##..",
		"..#..",
		"..#..",
		"..#..",
		"..#..",
		"..#..",
		".###.",
	},
}

var brailleBit = [2][4]uint8{
	{1, 2, 4, 64},
	{8, 16, 32, 128},
}

type bitmap struct {
	w, h int
	bits []bool
}

func newBitmap(w, h int) *bitmap {
	return &bitmap{w: w, h: h, bits: make([]bool, w*h)}
}

func (b *bitmap) set(x, y int) {
	if x < 0 || y < 0 || x >= b.w || y >= b.h {
		return
	}
	b.bits[y*b.w+x] = true
}

func (b *bitmap) get(x, y int) bool {
	if x < 0 || y < 0 || x >= b.w || y >= b.h {
		return false
	}
	return b.bits[y*b.w+x]
}

func buildWord(word string, s int) *bitmap {
	const cellW, cellH, gap = 5, 7, 1
	sw, sh, sg := cellW*s, cellH*s, gap*s
	w := len(word)*sw + (len(word)-1)*sg
	bm := newBitmap(w, sh)
	x := 0
	for _, r := range word {
		g := font[r]
		for yy, row := range g {
			for xx, ch := range row {
				if ch != '#' {
					continue
				}
				for dy := 0; dy < s; dy++ {
					for dx := 0; dx < s; dx++ {
						bm.set(x+xx*s+dx, yy*s+dy)
					}
				}
			}
		}
		x += sw + sg
	}
	return bm
}

func toBraille(bm *bitmap) [][]rune {
	bw := (bm.w + 1) / 2
	bh := (bm.h + 3) / 4
	grid := make([][]rune, bh)
	for y := 0; y < bh; y++ {
		grid[y] = make([]rune, bw)
		for x := 0; x < bw; x++ {
			var bits uint8
			for dy := 0; dy < 4; dy++ {
				for dx := 0; dx < 2; dx++ {
					if bm.get(x*2+dx, y*4+dy) {
						bits |= brailleBit[dx][dy]
					}
				}
			}
			grid[y][x] = rune(0x2800 + int(bits))
		}
	}
	return grid
}

var gradBuf builder.Builder

func main() {
	in, out := acell.Default()
	t := acell.New(in, out)
	defer t.Close()

	grid := toBraille(buildWord("acell", scale))
	bh := len(grid)
	bw := len(grid[0])

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

		pad := 3
		boxW := bw + pad*2
		boxH := bh + pad*2
		offX := (w - boxW) / 2
		offY := (h - boxH) / 2

		grad := func(x, y int) acell.Style {
			u := float64(x)/float64(boxW) + float64(y)/float64(boxH)*0.5
			r := uint8(127 + 128*math.Sin(u*6.28+phase))
			g := uint8(127 + 128*math.Sin(u*6.28+phase+2.09))
			b := uint8(127 + 128*math.Sin(u*6.28+phase+4.19))

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

		put := func(x, y int, ch rune, st acell.Style) {
			px, py := offX+x, offY+y
			if py < 0 || py >= h || px < 0 || px >= w {
				return
			}
			t.Buf[py][px] = acell.Cell{Char: ch, Style: st}
		}

		put(0, 0, '┌', grad(0, 0))
		put(boxW-1, 0, '┐', grad(boxW-1, 0))
		put(0, boxH-1, '└', grad(0, boxH-1))
		put(boxW-1, boxH-1, '┘', grad(boxW-1, boxH-1))

		for x := 1; x < boxW-1; x++ {
			put(x, 0, '─', grad(x, 0))
			put(x, boxH-1, '─', grad(x, boxH-1))
		}
		for y := 1; y < boxH-1; y++ {
			put(0, y, '│', grad(0, y))
			put(boxW-1, y, '│', grad(boxW-1, y))
		}

		for y := 0; y < bh; y++ {
			for x := 0; x < bw; x++ {
				if grid[y][x] == 0x2800 {
					continue
				}
				put(pad+x, pad+y, grid[y][x], grad(pad+x, pad+y))
			}
		}

		label := "Fast TUI engine by romanSPB15"
		labelY := offY + boxH - 1
		labelX := (w - utf8.RuneCountInString(label)) / 2
		for i, r := range label {
			put(labelX+i-offX, labelY-offY, r, acell.Style{Args: acell.Bold})
		}

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
			}
		case <-ticker.C:
			phase += 0.2
			render()
		}
	}
}
