package main

import (
	"fmt"
	"image"
	"os"

	"github.com/romanSPB15/acell"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

func loadImage(filename string) (image.Image, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	return img, err
}

func fitImage(img image.Image, maxCols, maxRows, cellW, cellH int, pixelAspect float64) (cols, rows int) {
	if cellW <= 0 {
		cellW = 8
	}
	if cellH <= 0 {
		cellH = 16
	}
	if pixelAspect <= 0 {
		pixelAspect = 2.0
	}

	b := img.Bounds()
	iw := float64(b.Dx())
	ih := float64(b.Dy())
	if iw <= 0 || ih <= 0 {
		return 1, 1
	}

	naturalCols := iw / float64(cellW)
	naturalRows := ih * pixelAspect / float64(cellH)

	scale := 1.0
	if naturalCols > float64(maxCols) {
		scale = float64(maxCols) / naturalCols
	}
	if s := float64(maxRows) / naturalRows; s < scale {
		scale = s
	}

	cols = int(naturalCols * scale)
	rows = int(naturalRows * scale)
	if cols < 1 {
		cols = 1
	}
	if rows < 1 {
		rows = 1
	}
	return cols, rows
}

func main() {
	if len(os.Args) == 1 {
		fmt.Println("usage: images <file>")
		os.Exit(1)
	}

	img, err := loadImage(os.Args[1])
	if err != nil {
		fmt.Println("invalid file:", err)
		os.Exit(1)
	}

	in, out := acell.Default()
	t := acell.New(in, out)
	defer t.Close()

	info := t.Info()
	cellW := info.CellW
	cellH := info.CellH

	rebuild := func() {
		w, h := t.Size()

		cols, rows := fitImage(img, w, h-1, cellW, cellH, 2.0)

		px := (w - cols) / 2
		py := (h - rows) / 2
		if px < 0 {
			px = 0
		}
		if py < 0 {
			py = 0
		}

		t.Images = t.Images[:0]
		t.Images = append(t.Images,
			acell.ParseImage(img, acell.Point{X: px, Y: py}, cols, rows))
	}

	rebuild()
	t.Flush()

	for ev := range t.Events() {
		switch e := ev.(type) {
		case *acell.KeyboardEvent:
			if e.Rune == 'q' || e.Key == acell.KeyCtrlC {
				return
			}
		case *acell.ResizeEvent:
			t.Buf = acell.NewBuf(e.Width, e.Height)
			t.Invalidate()
			rebuild()
			t.Flush()
		}
	}
}
