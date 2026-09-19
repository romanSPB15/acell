package main

import (
	"fmt"
	"os"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/romanSPB15/acell"
)

var acellPalette = []string{
	"30", "31", "32", "33", "34", "35", "36", "37",
}

var redraws atomic.Int64

func envInt(name string, def int) int {
	if v := os.Getenv(name); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return def
}

func main() {
	cells := envInt("STRESS_CELLS", 300)
	paletteSz := envInt("STRESS_PALETTE", 8)
	if paletteSz > len(acellPalette) {
		paletteSz = len(acellPalette)
	}

	in, out := acell.Default()
	term := acell.New(in, out)
	defer term.Close()

	w, h := term.Size()
	term.Buf = makeBuf(w, h)

	logFile, _ := os.Create("acell_fps.log")
	defer logFile.Close()

	renderCh := make(chan struct{}, 1)
	stopCh := make(chan struct{})

	go func() {
		for {
			select {
			case <-stopCh:
				return
			default:
				select {
				case renderCh <- struct{}{}:
				case <-stopCh:
					return
				}
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		var last int64
		for range ticker.C {
			now := redraws.Load()
			fmt.Fprintf(logFile, "FPS: %d (cells=%d palette=%d)\n", now-last, cells, paletteSz)
			last = now
		}
	}()

	frame := 0
	total := 80 * 24
	if w*h < total {
		total = w * h
	}
	// Use a fixed 80x24 grid like tcell bench, if buffer is big enough
	benchW, benchH := 80, 24
	if w < benchW || h < benchH {
		benchW, benchH = w, h
	}

	for {
		select {
		case ev := <-term.Events():
			switch e := ev.(type) {
			case *acell.KeyboardEvent:
				if e.Rune == 'q' || e.Key == acell.KeyEsc || e.Key == acell.KeyCtrlC {
					close(stopCh)
					return
				}
			}
		case <-renderCh:
			frame++
			for k := 0; k < cells; k++ {
				idx := (frame*7919 + k*104729) % total
				y, x := idx/benchW, idx%benchW
				fg := acellPalette[(frame+k)%paletteSz]
				term.Buf[y][x] = acell.Cell{
					Char:  rune('A' + (frame+k)%26),
					Style: acell.Style{Fg: fg},
				}
			}
			term.Flush()
			redraws.Add(1)
		}
	}
}

func makeBuf(w, h int) [][]acell.Cell {
	buf := make([][]acell.Cell, h)
	for y := range buf {
		buf[y] = make([]acell.Cell, w)
		for x := range buf[y] {
			buf[y][x] = acell.Cell{Char: ' '}
		}
	}
	return buf
}
