package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"math"
	"os"
	"sync/atomic"
	"time"

	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/tty"
	"github.com/gdamore/tcell/v3/vt"
	"github.com/romanSPB15/acell"
	"github.com/romanSPB15/acell/terminfo"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

const (
	benchW = 80
	benchH = 24
	frames = 1000
)

// ---------------------------------------------------------------------------
// палитры
// ---------------------------------------------------------------------------

type paletteType int

const (
	pal256 paletteType = iota
	palTrueColor
)

var acellPalette256 = []string{
	acell.FgBlack, acell.FgRed, acell.FgGreen, acell.FgYellow,
	acell.FgBlue, acell.FgPurple, acell.FgTeal, acell.FgWhite,
}

var tcellPalette256 = []tcell.Color{
	tcell.ColorBlack, tcell.ColorRed, tcell.ColorGreen, tcell.ColorYellow,
	tcell.ColorBlue, tcell.ColorPurple, tcell.ColorTeal, tcell.ColorWhite,
}

type rgbColor struct{ r, g, b int }

var trueColorPalette = []rgbColor{
	{0, 0, 0}, {255, 0, 0}, {0, 255, 0}, {255, 255, 0},
	{0, 0, 255}, {255, 0, 255}, {0, 255, 255}, {255, 255, 255},
}

// ---------------------------------------------------------------------------
// fill modes
// ---------------------------------------------------------------------------

type fillMode int

const (
	modeBoth fillMode = iota
	modeCharOnly
	modeColorOnly
	modeOneCell
	modeNoChanges
)

// ---------------------------------------------------------------------------
// fake terminals
// ---------------------------------------------------------------------------

type acellFake struct {
	writes, bytes atomic.Int64
	events        chan any
	trueColor     bool
}

func (f *acellFake) Write(p []byte) (int, error) {
	f.writes.Add(1)
	f.bytes.Add(int64(len(p)))
	return len(p), nil
}
func (f *acellFake) Read(p []byte) (int, error) { return 0, io.EOF }
func (f *acellFake) MakeRaw() error             { return nil }
func (f *acellFake) Restore() error             { return nil }
func (f *acellFake) EnableANSI() error          { return nil }
func (f *acellFake) Events() <-chan any         { return f.events }
func (f *acellFake) Close() error               { return nil }
func (f *acellFake) Size() (int, int)           { return benchW, benchH }
func (f *acellFake) StartInput()                {}
func (f *acellFake) Info() terminfo.Info {
	info := terminfo.Default()
	if f.trueColor {
		info.Colors = terminfo.ColorTrue
	}
	return info
}

type countingTty struct {
	tty.Tty
	writes atomic.Int64
	bytes  atomic.Int64
}

func newCountingTty(w, h int) *countingTty {
	mt := vt.NewMockTerm(vt.MockOptSize{X: vt.Col(w), Y: vt.Row(h)})
	return &countingTty{Tty: mt}
}

func (c *countingTty) Write(p []byte) (int, error) {
	c.writes.Add(1)
	c.bytes.Add(int64(len(p)))
	return len(p), nil
}

// ---------------------------------------------------------------------------
// заполнение кадров
// ---------------------------------------------------------------------------

func fillAc(term *acell.Terminal, frame, n int, mode fillMode, pal paletteType) {
	// noChanges — вообще ничего не трогаем.
	if mode == modeNoChanges {
		return
	}

	// oneCell — меняем ровно одну ячейку; она «путешествует» по экрану,
	// чтобы заодно проверить cursor tracking.
	if mode == modeOneCell {
		idx := frame % (benchW * benchH)
		x, y := idx%benchW, idx/benchW

		var fg string
		switch pal {
		case palTrueColor:
			c := trueColorPalette[frame%len(trueColorPalette)]
			fg = acell.FgRGB(c.r, c.g, c.b)
		default:
			fg = acellPalette256[frame%len(acellPalette256)]
		}

		term.Buf[y][x] = acell.Cell{
			Char:  rune('A' + frame%26),
			Style: acell.Style{Fg: fg},
		}
		return
	}

	// остальные режимы — стандартный сценарий
	total := benchW * benchH
	for k := 0; k < n; k++ {
		var idx int
		if mode == modeBoth {
			idx = (frame*7919 + k*104729) % total
		} else {
			idx = (k * 104729) % total
		}
		x, y := idx%benchW, idx/benchW

		ch := rune('A' + (frame+k)%26)

		var fg string
		switch pal {
		case palTrueColor:
			c := trueColorPalette[(frame+k)%len(trueColorPalette)]
			if mode == modeCharOnly {
				c = trueColorPalette[0]
			}
			fg = acell.FgRGB(c.r, c.g, c.b)
		default:
			fg = acellPalette256[(frame+k)%len(acellPalette256)]
			if mode == modeCharOnly {
				fg = acellPalette256[0]
			}
		}

		if mode == modeColorOnly {
			ch = 'X'
		}

		term.Buf[y][x] = acell.Cell{Char: ch, Style: acell.Style{Fg: fg}}
	}
}

func fillTc(s tcell.Screen, frame, n int, mode fillMode, pal paletteType) {
	if mode == modeNoChanges {
		return
	}

	if mode == modeOneCell {
		idx := frame % (benchW * benchH)
		x, y := idx%benchW, idx/benchW

		var fg tcell.Color
		switch pal {
		case palTrueColor:
			c := trueColorPalette[frame%len(trueColorPalette)]
			fg = tcell.NewRGBColor(int32(c.r), int32(c.g), int32(c.b))
		default:
			fg = tcellPalette256[frame%len(tcellPalette256)]
		}

		s.SetContent(x, y, rune('A'+frame%26), nil, tcell.StyleDefault.Foreground(fg))
		return
	}

	total := benchW * benchH
	for k := 0; k < n; k++ {
		var idx int
		if mode == modeBoth {
			idx = (frame*7919 + k*104729) % total
		} else {
			idx = (k * 104729) % total
		}
		x, y := idx%benchW, idx/benchW

		ch := rune('A' + (frame+k)%26)

		var fg tcell.Color
		switch pal {
		case palTrueColor:
			c := trueColorPalette[(frame+k)%len(trueColorPalette)]
			if mode == modeCharOnly {
				c = trueColorPalette[0]
			}
			fg = tcell.NewRGBColor(int32(c.r), int32(c.g), int32(c.b))
		default:
			fg = tcellPalette256[(frame+k)%len(tcellPalette256)]
			if mode == modeCharOnly {
				fg = tcellPalette256[0]
			}
		}

		if mode == modeColorOnly {
			ch = 'X'
		}

		s.SetContent(x, y, ch, nil, tcell.StyleDefault.Foreground(fg))
	}
}

// ---------------------------------------------------------------------------
// измерения
// ---------------------------------------------------------------------------

type measurement struct {
	fps    float64
	writes int64
	bytes  int64
}

type benchResult struct {
	scenario string
	acell    measurement
	tcell    measurement
}

func runAcell(n int, mode fillMode, pal paletteType) measurement {
	f := &acellFake{
		events:    make(chan any),
		trueColor: pal == palTrueColor,
	}
	term := acell.NewWithTerm(f)
	defer term.Close()

	// seed: для noChanges заполняем экран целиком, чтобы diff что-то сравнивал.
	if mode == modeNoChanges {
		fillAc(term, 0, benchW*benchH, modeBoth, pal)
	} else {
		fillAc(term, 0, n, mode, pal)
	}
	term.Flush()

	f.writes.Store(0)
	f.bytes.Store(0)

	start := time.Now()
	for i := 0; i < frames; i++ {
		fillAc(term, i+1, n, mode, pal)
		term.Flush()
	}
	elapsed := time.Since(start)

	return measurement{
		fps:    float64(frames) / elapsed.Seconds(),
		writes: f.writes.Load() / frames,
		bytes:  f.bytes.Load() / frames,
	}
}

func runTcell(n int, mode fillMode, pal paletteType) (measurement, error) {
	mt := newCountingTty(benchW, benchH)
	s, err := tcell.NewTerminfoScreenFromTty(mt)
	if err != nil {
		return measurement{}, err
	}
	if err := s.Init(); err != nil {
		return measurement{}, err
	}
	defer s.Fini()

	if mode == modeNoChanges {
		fillTc(s, 0, benchW*benchH, modeBoth, pal)
	} else {
		fillTc(s, 0, n, mode, pal)
	}
	s.Show()

	mt.writes.Store(0)
	mt.bytes.Store(0)

	start := time.Now()
	for i := 0; i < frames; i++ {
		fillTc(s, i+1, n, mode, pal)
		s.Show()
	}
	elapsed := time.Since(start)

	return measurement{
		fps:    float64(frames) / elapsed.Seconds(),
		writes: mt.writes.Load() / frames,
		bytes:  mt.bytes.Load() / frames,
	}, nil
}

// ---------------------------------------------------------------------------
// график
// ---------------------------------------------------------------------------

var (
	colBg     = color.RGBA{26, 26, 26, 255}
	colFg     = color.RGBA{230, 230, 230, 255}
	colSub    = color.RGBA{140, 140, 140, 255}
	colGrid   = color.RGBA{60, 60, 60, 255}
	colAcell  = color.RGBA{220, 60, 60, 255}
	colTcell  = color.RGBA{70, 130, 220, 255}
	colBorder = color.RGBA{90, 90, 90, 255}
	colDelta  = color.RGBA{180, 220, 130, 255}
)

type metricKind int

const (
	metricFPS metricKind = iota
	metricBytes
)

func (m metricKind) value(x measurement) float64 {
	if m == metricFPS {
		return x.fps
	}
	return float64(x.bytes)
}

func (m metricKind) title() string {
	if m == metricFPS {
		return "FPS: acell vs tcell"
	}
	return "Bytes/flush: acell vs tcell"
}

func (m metricKind) axisLabel() string {
	if m == metricFPS {
		return "frames per second"
	}
	return "bytes per flush"
}

func formatDelta(d float64) string {
	if math.Abs(d) < 0.5 {
		return fmt.Sprintf("%+.1f%%", d)
	}
	return fmt.Sprintf("%+.0f%%", d)
}

func drawChart(results []benchResult, m metricKind, path string) error {
	const (
		W       = 1200
		marginL = 260
		marginR = 280
		marginT = 110
		marginB = 70
		rowH    = 62
		barH    = 20
		barGap  = 4
		barFill = 0.80
	)

	H := marginT + marginB + rowH*len(results)
	plotL := marginL
	plotR := W - marginR
	plotT := marginT
	plotB := H - marginB
	plotW := plotR - plotL

	img := image.NewRGBA(image.Rect(0, 0, W, H))
	draw.Draw(img, img.Bounds(), &image.Uniform{colBg}, image.Point{}, draw.Src)

	maxV := 1.0
	for _, r := range results {
		if v := m.value(r.acell); v > maxV {
			maxV = v
		}
		if v := m.value(r.tcell); v > maxV {
			maxV = v
		}
	}

	var step float64
	if m == metricFPS {
		step = 5000
	} else {
		step = 5000
	}
	maxV = float64(int(maxV/step)+1) * step

	steps := 5
	for i := 0; i <= steps; i++ {
		x := plotL + i*plotW/steps
		for y := plotT; y <= plotB; y++ {
			img.Set(x, y, colGrid)
		}
		val := maxV * float64(i) / float64(steps)
		label := fmt.Sprintf("%.0f", val)
		w := font.MeasureString(basicfont.Face7x13, label).Ceil()
		drawText(img, x-w/2, plotB+25, label, colSub)
	}

	axisLbl := m.axisLabel()
	aw := font.MeasureString(basicfont.Face7x13, axisLbl).Ceil()
	drawText(img, plotL+(plotW-aw)/2, plotB+50, axisLbl, colSub)

	drawRect(img, plotL, plotT, plotR, plotB, colBorder)

	usableW := float64(plotW) * barFill
	deltaRightX := W - 20

	for i, r := range results {
		rowTop := plotT + i*rowH + (rowH-(2*barH+barGap))/2

		sw := font.MeasureString(basicfont.Face7x13, r.scenario).Ceil()
		drawText(img, plotL-15-sw, rowTop+barH-2, r.scenario, colFg)

		av := m.value(r.acell)
		tv := m.value(r.tcell)

		aw := int(usableW * av / maxV)
		fillRect(img, plotL+1, rowTop, plotL+1+aw, rowTop+barH, colAcell)
		drawText(img, plotL+aw+8, rowTop+barH-4, fmt.Sprintf("%.0f", av), colFg)

		tw := int(usableW * tv / maxV)
		fillRect(img, plotL+1, rowTop+barH+barGap, plotL+1+tw, rowTop+2*barH+barGap, colTcell)
		drawText(img, plotL+tw+8, rowTop+2*barH+barGap-4, fmt.Sprintf("%.0f", tv), colFg)

		var deltaStr string
		if m == metricFPS {
			if r.tcell.fps > 0 {
				deltaStr = "acell " + formatDelta((r.acell.fps-r.tcell.fps)/r.tcell.fps*100)
			}
		} else {
			if r.acell.bytes > 0 {
				deltaStr = "tcell " + formatDelta(float64(r.tcell.bytes-r.acell.bytes)/float64(r.acell.bytes)*100)
			} else {
				deltaStr = "both 0 bytes"
			}
		}
		dw := font.MeasureString(basicfont.Face7x13, deltaStr).Ceil()
		drawText(img, deltaRightX-dw, rowTop+barH-2, deltaStr, colDelta)
	}

	legendY := marginT - 55
	legendX := plotL
	fillRect(img, legendX, legendY, legendX+18, legendY+18, colAcell)
	drawText(img, legendX+26, legendY+14, "acell", colFg)
	legendX += 110
	fillRect(img, legendX, legendY, legendX+18, legendY+18, colTcell)
	drawText(img, legendX+26, legendY+14, "tcell", colFg)

	title := m.title()
	tw := font.MeasureString(basicfont.Face7x13, title).Ceil()
	drawText(img, (W-tw)/2, 40, title, colFg)

	sub := fmt.Sprintf("%dx%d, %d frames/scenario",
		benchW, benchH, frames)
	sw := font.MeasureString(basicfont.Face7x13, sub).Ceil()
	drawText(img, (W-sw)/2, 62, sub, colSub)

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func drawText(img *image.RGBA, x, y int, s string, c color.Color) {
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(c),
		Face: basicfont.Face7x13,
		Dot:  fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y)},
	}
	d.DrawString(s)
}

func fillRect(img *image.RGBA, x0, y0, x1, y1 int, c color.Color) {
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			img.Set(x, y, c)
		}
	}
}

func drawRect(img *image.RGBA, x0, y0, x1, y1 int, c color.Color) {
	for x := x0; x <= x1; x++ {
		img.Set(x, y0, c)
		img.Set(x, y1, c)
	}
	for y := y0; y <= y1; y++ {
		img.Set(x0, y, c)
		img.Set(x1, y, c)
	}
}

// ---------------------------------------------------------------------------
// main
// ---------------------------------------------------------------------------

type scenario struct {
	name        string
	n           int
	mode        fillMode
	pal         paletteType
	skipInChart bool
}

var scenarios = []scenario{
	{"oneCell (1)", 0, modeOneCell, pal256, false},
	{"tiny (10 cells)", 10, modeBoth, pal256, false},
	{"light (100 cells)", 100, modeBoth, pal256, false},
	{"medium (300 cells)", 300, modeBoth, pal256, false},
	{"heavy (500 cells)", 500, modeBoth, pal256, false},
	{"dense (1000 cells)", 1000, modeBoth, pal256, false},
	{"full (1920 cells)", benchW * benchH, modeBoth, pal256, false},
	{"textOnly (300)", 300, modeCharOnly, pal256, false},
	{"colorOnly (300)", 300, modeColorOnly, pal256, false},
	{"trueColor (300)", 300, modeBoth, palTrueColor, false},
	{"trueColor full (1920)", benchW * benchH, modeBoth, palTrueColor, false},
	// noChanges — «пол» overhead'а. Исключаем из графиков:
	// FPS в сотни тысяч искажает шкалу и делает остальные бары нечитаемыми.
	{"noChanges (0)", 0, modeNoChanges, pal256, true},
}

func main() {
	var results []benchResult
	var chartResults []benchResult

	for _, s := range scenarios {
		fmt.Printf("=== %s ===\n", s.name)

		ac := runAcell(s.n, s.mode, s.pal)
		fmt.Printf("acell:  %.0f FPS, %d writes/flush, %d bytes/flush\n",
			ac.fps, ac.writes, ac.bytes)

		tc, err := runTcell(s.n, s.mode, s.pal)
		if err != nil {
			fmt.Fprintln(os.Stderr, "tcell:", err)
			continue
		}
		fmt.Printf("tcell:  %.0f FPS, %d writes/flush, %d bytes/flush\n",
			tc.fps, tc.writes, tc.bytes)

		var fpsDeltaStr string
		if tc.fps > 0 {
			fpsDeltaStr = formatDelta((ac.fps - tc.fps) / tc.fps * 100)
		} else {
			fpsDeltaStr = "n/a"
		}

		var bDeltaStr string
		if ac.bytes > 0 {
			bDeltaStr = formatDelta(float64(tc.bytes-ac.bytes) / float64(ac.bytes) * 100)
		} else if tc.bytes == 0 {
			bDeltaStr = "0 (both)"
		} else {
			bDeltaStr = "n/a"
		}

		fmt.Printf("delta:  acell %s faster (FPS), tcell %s more bytes\n\n",
			fpsDeltaStr, bDeltaStr)

		r := benchResult{scenario: s.name, acell: ac, tcell: tc}
		results = append(results, r)
		if !s.skipInChart {
			chartResults = append(chartResults, r)
		}
	}

	if err := drawChart(chartResults, metricFPS, "bench_fps.png"); err != nil {
		fmt.Fprintln(os.Stderr, "chart fps:", err)
		os.Exit(1)
	}
	fmt.Println("chart: bench_fps.png")

	if err := drawChart(chartResults, metricBytes, "bench_bytes.png"); err != nil {
		fmt.Fprintln(os.Stderr, "chart bytes:", err)
		os.Exit(1)
	}
	fmt.Println("chart: bench_bytes.png")
}
