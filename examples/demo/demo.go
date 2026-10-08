package main

import (
	"github.com/romanSPB15/acell"
	"github.com/romanSPB15/acell/builder"
)

const scale = 2

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

func lerpStyle(fromR, fromG, fromB, toR, toG, toB uint8, t float64) acell.Style {
	r := uint8(float64(fromR) + (float64(toR)-float64(fromR))*t)
	g := uint8(float64(fromG) + (float64(toG)-float64(fromG))*t)
	b := uint8(float64(fromB) + (float64(toB)-float64(fromB))*t)

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

// Розовый -> фиолетовый -> синий
func logoGrad(t float64) acell.Style {
	if t < 0.5 {
		return lerpStyle(255, 105, 180, 138, 43, 226, t*2)
	}
	return lerpStyle(138, 43, 226, 70, 70, 230, (t-0.5)*2)
}

var (
	lineStyle = acell.Style{Fg: "38;2;150;70;220"}         // фиолетовый
	signStyle = acell.Style{Fg: acell.FgRGB(200, 20, 100)} // розовый
)

func main() {
	in, out := acell.Default()
	t := acell.New(in, out)
	defer t.Close()

	grid := toBraille(buildWord("acell", scale))
	bh := len(grid)
	bw := len(grid[0])

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

		put := func(x, y int, ch rune, st acell.Style) {
			if y < 0 || y >= h || x < 0 || x >= w {
				return
			}
			t.Buf[y][x] = acell.Cell{Char: ch, Style: st}
		}

		const lineThick = 2
		const sigRow = 1
		blockH := 2*lineThick + sigRow + bh
		offY := (h - blockH) / 2
		offX := (w - bw) / 2

		// Верхняя линия
		for i := 0; i < lineThick; i++ {
			for x := 0; x < bw; x++ {
				put(offX+x, offY+i, '╱', lineStyle)
			}
		}

		// Подпись в щели между линией и логотипом, слева
		sign := "romanSPB15"
		signY := offY + lineThick
		for i, r := range sign {
			put(offX+i, signY, r, signStyle)
		}

		// Логотип
		logoY := offY + lineThick + sigRow
		for y := 0; y < bh; y++ {
			for x := 0; x < bw; x++ {
				if grid[y][x] == 0x2800 {
					continue
				}
				tpos := 0.0
				if bw > 1 {
					tpos = float64(x) / float64(bw-1)
				}
				put(offX+x, logoY+y, grid[y][x], logoGrad(tpos))
			}
		}

		// Нижняя линия
		botY := logoY + bh
		for i := 0; i < lineThick; i++ {
			for x := 0; x < bw; x++ {
				put(offX+x, botY+i, '╱', lineStyle)
			}
		}

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
