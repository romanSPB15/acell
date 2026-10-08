package acell

import (
	"image"
	"image/color"

	"github.com/romanSPB15/acell/builder"
)

func encodeSixel(img image.Image, cols, rows, cellW, cellH int) []byte {
	if cols <= 0 || rows <= 0 {
		return nil
	}
	if cellW <= 0 {
		cellW = 8
	}
	if cellH <= 0 {
		cellH = 16
	}
	pixelAspect := float64(cellH) / float64(cellW)

	targetW := cols * cellW
	targetH := rows * cellH

	sixelRows := int(float64(targetH) / pixelAspect)
	if sixelRows < 1 {
		sixelRows = 1
	}

	scaled := scaleNearest(img, targetW, sixelRows)

	var sb builder.Builder
	sb.WriteString("\x1bPq")

	bounds := scaled.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	// Квантование 6×6×6 = 216 цветов
	palette := make(map[color.RGBA]int, 256)
	var paletteList []color.RGBA

	getColor := func(c color.RGBA) int {
		q := color.RGBA{
			R: (c.R / 51) * 51,
			G: (c.G / 51) * 51,
			B: (c.B / 51) * 51,
			A: 255,
		}
		if idx, ok := palette[q]; ok {
			return idx
		}
		idx := len(paletteList)
		palette[q] = idx
		paletteList = append(paletteList, q)
		return idx
	}

	colors := make([][]int, h)
	for y := 0; y < h; y++ {
		colors[y] = make([]int, w)
		for x := 0; x < w; x++ {
			r, g, b, _ := scaled.At(x+bounds.Min.X, y+bounds.Min.Y).RGBA()
			c := color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), 255}
			colors[y][x] = getColor(c)
		}
	}

	for i, c := range paletteList {
		sb.WriteString("#")
		sb.WriteInt((i))
		sb.WriteString(";2;")
		sb.WriteInt(int(c.R))
		sb.WriteString(";")
		sb.WriteInt(int(c.G))
		sb.WriteString(";")
		sb.WriteInt(int(c.B))
	}

	for bandStart := 0; bandStart < h; bandStart += 6 {
		used := make(map[int]bool)
		for y := bandStart; y < bandStart+6 && y < h; y++ {
			for x := 0; x < w; x++ {
				used[colors[y][x]] = true
			}
		}

		order := make([]int, 0, len(used))
		for c := range used {
			order = append(order, c)
		}
		// Сортировка вставками — маленький массив, stdlib sort не тащим
		for i := 1; i < len(order); i++ {
			for j := i; j > 0 && order[j-1] > order[j]; j-- {
				order[j-1], order[j] = order[j], order[j-1]
			}
		}

		first := true
		for _, c := range order {
			if !first {
				sb.WriteByte('$')
			}
			first = false

			sb.WriteByte('#')
			sb.WriteInt(c)

			lastX := -1
			for x := 0; x < w; x++ {
				for dy := 0; dy < 6; dy++ {
					y := bandStart + dy
					if y >= h {
						break
					}
					if colors[y][x] == c {
						lastX = x
						break
					}
				}
			}
			if lastX < 0 {
				continue
			}

			for x := 0; x <= lastX; x++ {
				var bits byte
				for dy := 0; dy < 6; dy++ {
					y := bandStart + dy
					if y >= h {
						break
					}
					if colors[y][x] == c {
						bits |= 1 << dy
					}
				}
				sb.WriteByte(0x3F + bits)
			}
		}
		sb.WriteByte('-')
	}

	sb.WriteString("\x1b\\")
	return []byte(sb.String())
}

// scaleNearest — простой nearest-neighbor ресайз без зависимостей.
func scaleNearest(src image.Image, w, h int) image.Image {
	if w <= 0 || h <= 0 {
		return src
	}
	b := src.Bounds()
	sw, sh := b.Dx(), b.Dy()
	if sw == w && sh == h {
		return src
	}

	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		sy := b.Min.Y + y*sh/h
		for x := 0; x < w; x++ {
			sx := b.Min.X + x*sw/w
			dst.Set(x, y, src.At(sx, sy))
		}
	}
	return dst
}
