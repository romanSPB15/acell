package acell

import (
	"strconv"
	"strings"

	"github.com/romanSPB15/acell/terminfo"
)

var ansi16RGB = [16][3]uint8{
	{0, 0, 0}, {205, 0, 0}, {0, 205, 0}, {205, 205, 0},
	{0, 0, 238}, {205, 0, 205}, {0, 205, 205}, {229, 229, 229},
	{127, 127, 127}, {255, 0, 0}, {0, 255, 0}, {255, 255, 0},
	{92, 92, 255}, {255, 0, 255}, {0, 255, 255}, {255, 255, 255},
}

var ansi256RGB [256][3]uint8
var ansi256To16 [256]uint8
var ansi256To8 [256]uint8

func init() {
	copy(ansi256RGB[:16], ansi16RGB[:])
	levels := [6]uint8{0, 95, 135, 175, 215, 255}
	for r := 0; r < 6; r++ {
		for g := 0; g < 6; g++ {
			for b := 0; b < 6; b++ {
				ansi256RGB[16+36*r+6*g+b] = [3]uint8{levels[r], levels[g], levels[b]}
			}
		}
	}
	for i := 0; i < 24; i++ {
		v := uint8(8 + i*10)
		ansi256RGB[232+i] = [3]uint8{v, v, v}
	}
	for i := 0; i < 256; i++ {
		c := ansi256RGB[i]
		ansi256To16[i] = uint8(nearest(c[0], c[1], c[2], ansi16RGB[:]))
		ansi256To8[i] = uint8(nearest(c[0], c[1], c[2], ansi16RGB[:8]))
	}
}

func nearest(r, g, b uint8, palette [][3]uint8) int {
	best, bestD := 0, 1<<30
	for i, c := range palette {
		dr := int(c[0]) - int(r)
		dg := int(c[1]) - int(g)
		db := int(c[2]) - int(b)
		d := dr*dr + dg*dg + db*db
		if d < bestD {
			best, bestD = i, d
		}
	}
	return best
}

func nearest256(r, g, b uint8) int {
	best, bestD := 0, 1<<30
	for i := 0; i < 256; i++ {
		c := ansi256RGB[i]
		dr := int(c[0]) - int(r)
		dg := int(c[1]) - int(g)
		db := int(c[2]) - int(b)
		d := dr*dr + dg*dg + db*db
		if d < bestD {
			best, bestD = i, d
		}
	}
	return best
}

type parsedColor struct {
	kind    int // 0 = невалидный, 1 = 16-индекс, 2 = 256-индекс, 3 = RGB
	idx     int
	r, g, b uint8
}

func parseColorCode(s string, isBg bool) parsedColor {
	if s == "" {
		return parsedColor{}
	}
	if strings.HasPrefix(s, "38;2;") || strings.HasPrefix(s, "48;2;") {
		parts := strings.Split(s, ";")
		if len(parts) != 5 {
			return parsedColor{}
		}
		rv, _ := strconv.Atoi(parts[2])
		gv, _ := strconv.Atoi(parts[3])
		bv, _ := strconv.Atoi(parts[4])
		return parsedColor{kind: 3, r: uint8(rv), g: uint8(gv), b: uint8(bv)}
	}
	if strings.HasPrefix(s, "38;5;") || strings.HasPrefix(s, "48;5;") {
		parts := strings.Split(s, ";")
		if len(parts) != 3 {
			return parsedColor{}
		}
		n, _ := strconv.Atoi(parts[2])
		return parsedColor{kind: 2, idx: n}
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return parsedColor{}
	}
	if isBg {
		if n >= 40 && n <= 47 {
			return parsedColor{kind: 1, idx: n - 40}
		}
		if n >= 100 && n <= 107 {
			return parsedColor{kind: 1, idx: n - 100 + 8}
		}
	} else {
		if n >= 30 && n <= 37 {
			return parsedColor{kind: 1, idx: n - 30}
		}
		if n >= 90 && n <= 97 {
			return parsedColor{kind: 1, idx: n - 90 + 8}
		}
	}
	return parsedColor{}
}

func (pc parsedColor) toRGB() (uint8, uint8, uint8) {
	switch pc.kind {
	case 1:
		c := ansi16RGB[pc.idx]
		return c[0], c[1], c[2]
	case 2:
		c := ansi256RGB[pc.idx]
		return c[0], c[1], c[2]
	case 3:
		return pc.r, pc.g, pc.b
	}
	return 0, 0, 0
}

func ansi8Code(idx int, isBg bool) string {
	if isBg {
		return strconv.Itoa(40 + idx)
	}
	return strconv.Itoa(30 + idx)
}

func ansi16Code(idx int, isBg bool) string {
	if idx < 8 {
		if isBg {
			return strconv.Itoa(40 + idx)
		}
		return strconv.Itoa(30 + idx)
	}
	if isBg {
		return strconv.Itoa(100 + idx - 8)
	}
	return strconv.Itoa(90 + idx - 8)
}

func ansi256Code(idx int, isBg bool) string {
	if isBg {
		return "48;5;" + strconv.Itoa(idx)
	}
	return "38;5;" + strconv.Itoa(idx)
}

func convertColor(s string, target terminfo.ColorBits, isBg bool) string {
	if s == "" || target >= terminfo.ColorTrue {
		return s
	}
	pc := parseColorCode(s, isBg)
	if pc.kind == 0 {
		return s
	}
	switch target {
	case terminfo.Color256:
		if pc.kind <= 2 {
			return s
		}
	case terminfo.Color16:
		if pc.kind == 1 || (pc.kind == 2 && pc.idx < 16) {
			return s
		}
	case terminfo.Color8:
		if pc.kind == 1 && pc.idx < 8 {
			return s
		}
	}

	r, g, b := pc.toRGB()

	switch target {
	case terminfo.Color8:
		return ansi8Code(nearest(r, g, b, ansi16RGB[:8]), isBg)
	case terminfo.Color16:
		if pc.kind == 2 {
			return ansi16Code(int(ansi256To16[pc.idx]), isBg)
		}
		return ansi16Code(nearest(r, g, b, ansi16RGB[:]), isBg)
	case terminfo.Color256:
		return ansi256Code(nearest256(r, g, b), isBg)
	}
	return s
}
