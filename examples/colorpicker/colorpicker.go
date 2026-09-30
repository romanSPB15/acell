package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/romanSPB15/acell"
	"github.com/romanSPB15/acell/term"
)

const (
	cellW  = 3
	gap    = 1
	startX = 1
	startY = 1
	blockH = 16
	gapV   = 1
	fgY    = startY + blockH + gapV

	rgbBlockX0 = 68
	rgbBlockX1 = 79
	rgbBlockY0 = 1
	rgbBlockY1 = 5

	rgbLabelY = 1
	rgbValueY = 2
	rgbTextY  = 4
	rgbBackY  = 5

	rgbRX = 68
	rgbGX = 72
	rgbBX = 76

	minRows = 35
)

var fgConstantByIndex = map[int]string{
	0: "Fg16Black", 1: "Fg16Red", 2: "Fg16Green", 3: "Fg16Yellow",
	4: "Fg16Blue", 5: "Fg16Magenta", 6: "Fg16Cyan", 7: "Fg16White",
	8: "Fg16Grey", 9: "Fg16LightRed", 10: "Fg16LightGreen",
	11: "Fg16LightYellow", 12: "Fg16LightBlue", 13: "Fg16LightMagenta",
	14: "Fg16LightCyan", 15: "Fg16LightWhite",

	16: "FgBlack", 231: "FgWhite",
	17: "FgNavy", 18: "FgDarkBlue", 19: "FgMediumBlue", 21: "FgBlue",
	27: "FgDodgerBlue", 33: "FgAzure", 39: "FgDeepSkyBlue",
	117: "FgSkyBlue", 63: "FgRoyalBlue", 69: "FgCornflower",
	99: "FgSlateBlue", 153: "FgLightBlue", 152: "FgLightSteel",
	30: "FgTeal", 51: "FgCyan", 37: "FgLightSea", 44: "FgTurquoise",
	67: "FgSteelBlue", 66: "FgCadetBlue", 195: "FgPaleCyan",
	22: "FgDarkGreen", 28: "FgForestGreen", 46: "FgGreen",
	47: "FgSpringGreen", 48: "FgMediumSpring", 118: "FgGreenYellow",
	120: "FgLightGreen", 121: "FgPaleGreen", 29: "FgSeaGreen",
	35: "FgMediumSea", 40: "FgLimeGreen", 100: "FgOlive",
	143: "FgDarkKhaki", 222: "FgKhaki", 226: "FgYellow",
	229: "FgPaleYellow", 230: "FgLightYellow", 136: "FgDarkGoldenrod",
	178: "FgGoldenrod", 202: "FgOrangeRed", 208: "FgDarkOrange",
	214: "FgOrange", 220: "FgGold", 166: "FgChocolate", 130: "FgSaddle",
	173: "FgPeru", 179: "FgTan", 215: "FgSandy", 223: "FgPeach",
	88: "FgDarkRed", 124: "FgFirebrick", 196: "FgRed", 161: "FgCrimson",
	203: "FgTomato", 209: "FgCoral", 210: "FgSalmon", 198: "FgDeepPink",
	205: "FgHotPink", 212: "FgPink", 217: "FgLightPink",
	54: "FgIndigo", 56: "FgBlueViolet", 90: "FgPurple", 91: "FgDarkViolet",
	128: "FgDarkOrchid", 134: "FgMediumOrchid", 141: "FgMediumPurple",
	177: "FgViolet", 183: "FgPlum", 225: "FgThistle",
	213: "FgOrchid", 201: "FgMagenta",
}

func nameFor(idx int, bg bool) string {
	if idx >= 232 && idx <= 255 {
		prefix := "Fg"
		if bg {
			prefix = "Bg"
		}
		return fmt.Sprintf("acell.%sGrey%d", prefix, idx-231)
	}
	if n, ok := fgConstantByIndex[idx]; ok {
		if bg {
			n = "Bg" + strings.TrimPrefix(n, "Fg")
		}
		return "acell." + n
	}
	if bg {
		return fmt.Sprintf("acell.Bg256(%d)", idx)
	}
	return fmt.Sprintf("acell.Fg256(%d)", idx)
}

func cubeVal(v int) int {
	if v == 0 {
		return 0
	}
	return 55 + 40*v
}

func rgb256(idx int) (int, int, int) {
	switch {
	case idx < 16:
		tbl := [16][3]int{
			{0, 0, 0}, {205, 0, 0}, {0, 205, 0}, {205, 205, 0},
			{0, 0, 238}, {205, 0, 205}, {0, 205, 205}, {229, 229, 229},
			{127, 127, 127}, {255, 0, 0}, {0, 255, 0}, {255, 255, 0},
			{92, 92, 255}, {255, 0, 255}, {0, 255, 255}, {255, 255, 255},
		}
		return tbl[idx][0], tbl[idx][1], tbl[idx][2]
	case idx < 232:
		v := idx - 16
		return cubeVal(v / 36), cubeVal((v % 36) / 6), cubeVal(v % 6)
	default:
		v := 8 + 10*(idx-232)
		return v, v, v
	}
}

func luminance256(idx int) float64 {
	r, g, b := rgb256(idx)
	return 0.2126*float64(r) + 0.7152*float64(g) + 0.0722*float64(b)
}

func drawTextInto(t *acell.Terminal, x, y, w int, style acell.Style, s string) {
	runes := []rune(s)
	pad := (w - len(runes)) / 2
	if pad < 0 {
		pad = 0
	}
	for i, r := range runes {
		t.DrawRune(x+pad+i, y, style, r)
	}
}

func drawBgPalette(t *acell.Terminal, startRow, sw, sh int) {
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			idx := y*16 + x
			bg := acell.Bg256(idx)
			row := startRow + y
			baseCol := startX + x*(cellW+gap)

			for dx := 0; dx < cellW; dx++ {
				col := baseCol + dx
				if row >= sh || col >= sw {
					continue
				}
				t.Buf[row][col] = acell.Cell{Char: ' ', Style: acell.Style{Bg: bg}}
			}

			var fg string
			if luminance256(idx) > 128 {
				fg = acell.Fg16Black
			} else {
				fg = acell.Fg16White
			}
			drawTextInto(t, baseCol, row, cellW,
				acell.Style{Fg: fg, Bg: bg}, fmt.Sprintf("%d", idx))
		}
	}
}

func drawFgPalette(t *acell.Terminal, startRow, sw, sh int) {
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			idx := y*16 + x
			row := startRow + y
			baseCol := startX + x*(cellW+gap)
			if row >= sh {
				continue
			}
			drawTextInto(t, baseCol, row, cellW,
				acell.Style{Fg: acell.Fg256(idx)},
				fmt.Sprintf("%d", idx))
		}
	}
}

func hitTest(x, y int) (int, bool, bool) {
	if x < startX {
		return 0, false, false
	}
	relX := x - startX
	stride := cellW + gap
	col := relX / stride
	if col >= 16 || relX-col*stride >= cellW {
		return 0, false, false
	}
	if y >= startY && y < startY+blockH {
		return (y-startY)*16 + col, true, true
	}
	if y >= fgY && y < fgY+blockH {
		return (y-fgY)*16 + col, false, true
	}
	return 0, false, false
}

func adjust(cur *int, delta int) {
	*cur += delta
	if *cur < 0 {
		*cur = 0
	}
	if *cur > 255 {
		*cur = 255
	}
}

func wheelStep(alt, ctrl bool) int {
	switch {
	case ctrl:
		return 16
	case alt:
		return 1
	default:
		return 5
	}
}

func channelAt(x, y int) (byte, bool) {
	if y != rgbLabelY && y != rgbValueY {
		return 0, false
	}
	switch {
	case x >= rgbRX && x <= rgbRX+2:
		return 'r', true
	case x >= rgbGX && x <= rgbGX+2:
		return 'g', true
	case x >= rgbBX && x <= rgbBX+2:
		return 'b', true
	}
	return 0, false
}

func inRGBBlock(x, y int) bool {
	return x >= rgbBlockX0 && x <= rgbBlockX1 && y >= rgbBlockY0 && y <= rgbBlockY1
}

func rgbChannelText(r, g, b int, bg bool) string {
	fn := "FgRGB"
	if bg {
		fn = "BgRGB"
	}
	return fmt.Sprintf("acell.%s(%d, %d, %d)", fn, r, g, b)
}

func main() {
	in, out := acell.Default()
	t := acell.New(in, out)
	defer t.Close()

	sw, sh := t.Size()
	if sh < minRows {
		fmt.Fprintf(os.Stderr, "нужно минимум %d строк, у вас %d\n", minRows, sh)
		return
	}
	t.Buf = acell.NewBuf(sw, sh)

	r, g, b := 255, 128, 0
	hoverCh := byte(0) // 'r' | 'g' | 'b' | 0
	status := ""

	render := func() {
		drawBgPalette(t, startY, sw, sh)
		drawFgPalette(t, fgY, sw, sh)

		labelStyle := func(c byte) acell.Style {
			st := acell.Style{}
			switch c {
			case 'r':
				st.Fg = acell.Fg16Red
			case 'g':
				st.Fg = acell.Fg16Green
			case 'b':
				st.Fg = acell.Fg16Blue
			}
			if c == hoverCh {
				st.Args = acell.Bold | acell.Underline
			}
			return st
		}
		valueStyle := func(c byte) acell.Style {
			st := labelStyle(c)
			st.Args &^= acell.Underline
			return st
		}

		t.DrawRune(rgbRX+1, rgbLabelY, labelStyle('r'), 'R')
		t.DrawRune(rgbGX+1, rgbLabelY, labelStyle('g'), 'G')
		t.DrawRune(rgbBX+1, rgbLabelY, labelStyle('b'), 'B')

		t.DrawString(rgbRX, rgbValueY, valueStyle('r'), fmt.Sprintf("%3d", r))
		t.DrawString(rgbGX, rgbValueY, valueStyle('g'), fmt.Sprintf("%3d", g))
		t.DrawString(rgbBX, rgbValueY, valueStyle('b'), fmt.Sprintf("%3d", b))

		t.DrawString(rgbRX+2, rgbTextY,
			acell.Style{Fg: acell.FgRGB(r, g, b)}, " TEXT ")
		t.DrawString(rgbRX+2, rgbBackY,
			acell.Style{Bg: acell.BgRGB(r, g, b)}, " BACK ")

		clearRow(t, 1, sh-1, sw)
		if status != "" {
			t.DrawString(1, sh-1,
				acell.Style{Fg: acell.FgGreenYellow, Args: acell.Bold},
				"copied: "+status)
		}
		help := "wheel ±5  alt ±1  ctrl ±16  click copy  q quit"
		if sw > len(help)+2 {
			t.DrawString(sw-len(help)-1, sh-1,
				acell.Style{Fg: acell.FgGrey13}, help)
		}
	}

	redraw := func() {
		render()
		t.Flush()
	}

	t.SetTitle(fmt.Sprintf("acell palette — RGB(%d,%d,%d)", r, g, b))
	redraw()

	for ev := range t.Events() {
		switch e := ev.(type) {
		case *acell.KeyboardEvent:
			if e.Rune == 'q' || e.Key == acell.KeyEsc || e.Key == acell.KeyCtrlC {
				return
			}
			if e.Rune == 'c' {
				status = ""
				redraw()
			}

		case *acell.ResizeEvent:
			sw, sh = e.Width, e.Height
			if sh < minRows {
				continue
			}
			t.Buf = acell.NewBuf(sw, sh)
			t.Invalidate()
			redraw()

		case *acell.MouseEvent:
			switch e.Action {
			case acell.MouseMove:
				newHover := byte(0)
				if ch, ok := channelAt(e.Pos.X, e.Pos.Y); ok {
					newHover = ch
				}
				if newHover != hoverCh {
					hoverCh = newHover
					redraw()
				}

			case acell.MousePress:
				// Клик по R/G/B — копирует FgRGB/BgRGB целиком.
				if _, ok := channelAt(e.Pos.X, e.Pos.Y); ok {
					bg := e.Shift
					text := rgbChannelText(r, g, b, bg)
					term.CopyToClipboard(text)
					status = text
					redraw()
					continue
				}
				if e.Pos.Y == rgbTextY && inRGBBlock(e.Pos.X, e.Pos.Y) {
					text := rgbChannelText(r, g, b, false)
					term.CopyToClipboard(text)
					status = text
					redraw()
					continue
				}
				if e.Pos.Y == rgbBackY && inRGBBlock(e.Pos.X, e.Pos.Y) {
					text := rgbChannelText(r, g, b, true)
					term.CopyToClipboard(text)
					status = text
					redraw()
					continue
				}
				idx, isBg, ok := hitTest(e.Pos.X, e.Pos.Y)
				if !ok {
					continue
				}
				text := nameFor(idx, isBg)
				term.CopyToClipboard(text)
				status = text
				redraw()

			case acell.MouseWheelUp, acell.MouseWheelDown:
				ch, ok := channelAt(e.Pos.X, e.Pos.Y)
				if !ok {
					continue
				}
				step := wheelStep(e.Alt, e.Ctrl)
				if e.Action == acell.MouseWheelUp {
					step = -step
				}
				switch ch {
				case 'r':
					adjust(&r, step)
				case 'g':
					adjust(&g, step)
				case 'b':
					adjust(&b, step)
				}
				t.SetTitle(fmt.Sprintf("acell palette — RGB(%d,%d,%d)", r, g, b))
				redraw()
			}
		}
	}
}

// clearRow заполняет строку y от x0 до x1 пробелами с пустым стилем.
func clearRow(t *acell.Terminal, x0, y, sw int) {
	if y < 0 || y >= len(t.Buf) {
		return
	}
	for x := x0; x < sw; x++ {
		t.Buf[y][x] = acell.Cell{Char: ' '}
	}
}
