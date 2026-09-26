package acell

import (
	"io"
	"os"
	"strings"
	"sync"
	"unicode"

	"github.com/romanSPB15/acell/builder"
	"github.com/romanSPB15/acell/term"
	"github.com/romanSPB15/acell/terminfo"
)

type pos struct {
	Line int
	Col  int
}

// Terminal управляет буфером, diff-рендером и терминалом.
type Terminal struct {
	raw  term.RawTerminal
	info terminfo.Info

	oldBuf [][]Cell
	Buf    [][]Cell

	cursorPos pos
	last      Style

	fgCache map[string]string
	bgCache map[string]string

	bb builder.Builder

	closeOnce sync.Once
	closeErr  error

	forceRedraw bool
}

// Default возвращает стандартные потоки ввода-вывода.
func Default() (io.Reader, io.Writer) {
	return os.Stdin, os.Stdout
}

// New создаёт Terminal и переводит терминал в raw-режим,
// а также включает ENABLE_VIRTUAL_TERMINAL_PROCESSING для ANSI на Windows.
func New(in io.Reader, out io.Writer) *Terminal {
	return NewWithTerm(term.NewRawTerminal(in, out))
}

// NewWithTerm создаёт Terminal на основе переданного RawTerminal.
// Возможности терминала берутся из t.Info().
func NewWithTerm(t term.RawTerminal) *Terminal {
	info := t.Info()

	t.MakeRaw()
	t.EnableANSI()

	var start []byte
	start = append(start, info.AltScreenOn...)
	if info.MouseAny {
		start = append(start, "\033[?1003h"...)
	}
	if info.MouseSGR {
		start = append(start, "\033[?1006h"...)
	}
	start = append(start, info.Sgr0...)
	start = append(start, info.Home...)
	start = append(start, info.CursorHide...)

	t.Write(start)
	t.StartInput()

	w, h := t.Size()
	return &Terminal{
		info:      info,
		raw:       t,
		Buf:       NewBuf(w, h),
		cursorPos: pos{-1, -1},
		fgCache:   make(map[string]string, 32),
		bgCache:   make(map[string]string, 32),
	}
}

// maskStyle снимает атрибуты, которые терминал не поддерживает.
func (t *Terminal) maskStyle(s Style) Style {
	if !t.info.Blink {
		s.Args &^= Blink
	}
	if !t.info.Dim {
		s.Args &^= Dim
	}
	if !t.info.Italic {
		s.Args &^= Italic
	}
	if !t.info.Strike {
		s.Args &^= Strike
	}
	if !t.info.Hidden {
		s.Args &^= Hidden
	}

	if t.info.Colors < terminfo.ColorTrue {
		s.Fg = t.convertFg(s.Fg)
		s.Bg = t.convertBg(s.Bg)
	}
	return s
}

func (t *Terminal) convertFg(s string) string {
	if s == "" {
		return s
	}
	if v, ok := t.fgCache[s]; ok {
		return v
	}
	v := convertColor(s, t.info.Colors, false)
	t.fgCache[s] = v
	return v
}

func (t *Terminal) convertBg(s string) string {
	if s == "" {
		return s
	}
	if v, ok := t.bgCache[s]; ok {
		return v
	}
	v := convertColor(s, t.info.Colors, true)
	t.bgCache[s] = v
	return v
}

// NewBuf создаёт пустой буфер заданного размера.
func NewBuf(w, h int) [][]Cell {
	buf := make([][]Cell, h)
	for y := range buf {
		buf[y] = make([]Cell, w)
		for x := range buf[y] {
			buf[y][x] = Cell{Char: ' '}
		}
	}
	return buf
}

// Flush сравнивает Buf с oldBuf, пишет разницу в терминал
// и обновляет oldBuf.
func (t *Terminal) Flush() {
	h := len(t.Buf)
	if h == 0 {
		return
	}
	w := len(t.Buf[0])
	if w == 0 {
		return
	}

	force := t.forceRedraw
	t.forceRedraw = false

	if force || len(t.oldBuf) != h || len(t.oldBuf[0]) != w {
		t.oldBuf = NewBuf(w, h)
		t.cursorPos = pos{-1, -1}
		t.last = Style{}
	}

	bb := &t.bb
	bb.Reset()

	if t.info.SynchronizedUpdate {
		bb.WriteString("\033[?2026h")
	}

	changed := false

	for y := range h {
		row := t.Buf[y]
		oldRow := t.oldBuf[y]
		for x := range w {
			if !force && row[x] == oldRow[x] {
				continue
			}

			if x > 0 && RuneWidth(row[x-1].Char) == 2 {
				oldRow[x] = row[x]
				continue
			}

			changed = true

			if t.cursorPos.Line != y || t.cursorPos.Col != x {
				moved := false

				if t.cursorPos.Line == y {
					dx := x - t.cursorPos.Col
					if dx > 0 && dx <= 4 {
						if dx == 1 {
							bb.WriteString("\033[C")
						} else {
							bb.WriteString("\033[")
							bb.WriteInt(dx)
							bb.WriteByte('C')
						}
						moved = true
					}
				}

				if !moved && x == 0 && y == 0 {
					bb.WriteString("\033[H")
					moved = true
				}

				if t.cursorPos.Col == x && t.cursorPos.Line < y {
					dy := y - t.cursorPos.Line
					if dy <= 4 {
						if dy == 1 {
							bb.WriteString("\033[B")
						} else {
							bb.WriteString("\033[")
							bb.WriteInt(dy)
							bb.WriteByte('B')
						}
						moved = true
					}
				}

				if !moved {
					bb.WriteString("\033[")
					bb.WriteInt(y + 1)
					bb.WriteByte(';')
					bb.WriteInt(x + 1)
					bb.WriteByte('H')
				}
			}

			st := t.maskStyle(row[x].Style)
			st.WriteANSI(t.last, bb)
			t.last = st

			ch := row[x].Char
			if ch == 0 {
				ch = ' '
			}
			rw := RuneWidth(ch)

			if rw > 1 {
				bb.WriteString("\033[2X")
			}
			bb.WriteRune(ch)

			oldRow[x] = row[x]

			if rw > 1 {
				t.cursorPos = pos{-1, -1}
			} else {
				x2 := x + rw
				y2 := y
				if x2 >= w {
					x2 = 0
					y2++
					t.cursorPos = pos{-1, -1}
				} else {
					t.cursorPos = pos{Line: y2, Col: x2}
				}
			}
		}
	}

	if changed {
		if t.info.SynchronizedUpdate {
			bb.WriteString("\033[?2026l")
		}
		t.bb.Copy(t.raw)
	}
}

// Size возвращает размер терминала.
func (t *Terminal) Size() (int, int) {
	return t.raw.Size()
}

// Events возвращает канал событий из терминала.
func (t *Terminal) Events() <-chan any {
	return t.raw.Events()
}

// Info получает информацию об терминале из переменных среды.
func (t *Terminal) Info() terminfo.Info {
	return t.info
}

// Close останавливает терминал и восстанавливает режим.
func (t *Terminal) Close() error {
	t.closeOnce.Do(func() {
		var end []byte
		if t.info.MouseAny {
			end = append(end, "\033[?1003l"...)
		}
		if t.info.MouseSGR {
			end = append(end, "\033[?1006l"...)
		}
		end = append(end, t.info.Sgr0...)
		end = append(end, t.info.CursorShow...)
		end = append(end, t.info.AltScreenOff...)

		t.raw.Write(end)

		if err := t.raw.Restore(); err != nil {
			t.closeErr = err
			return
		}
		t.closeErr = t.raw.Close()
	})
	return t.closeErr
}

// DrawRune рисует одну руну в позиции (x, y) с заданным стилем.
// Возвращает ширину руны (0, 1 или 2) и признак того, что руна нарисована.
// Ширина 0 означает невидимую руну (комбинирующую) — писать её не нужно.
// drawn == false означает, что руна не влезла.
func (t *Terminal) DrawRune(x, y int, style Style, r rune) (width int, drawn bool) {
	rw := RuneWidth(r)
	if rw == 0 {
		return 0, false
	}

	h := len(t.Buf)
	if y < 0 || y >= h {
		return rw, false
	}
	w := len(t.Buf[y])
	if w == 0 {
		return rw, false
	}
	if x < 0 || x+rw > w {
		return rw, false
	}

	if r == 0 {
		r = ' '
	}
	t.Buf[y][x] = Cell{Char: r, Style: style}
	for i := 1; i < rw; i++ {
		t.Buf[y][x+i] = Cell{Char: ' ', Style: style}
	}
	return rw, true
}

// DrawString рисует строку str начиная с позиции (x, y) с заданным стилем с учётом ширины рун.
// Возвращает количество нарисованных ячеек.
// Символы, вышедшие за границы буфера, отбрасываются.
// Поддерживает переносы в строке — \n или \r\n.
func (t *Terminal) DrawString(x, y int, style Style, str string) int {
	if strings.ContainsAny(str, "\r\n") {
		str = strings.ReplaceAll(str, "\r\n", "\n")
		str = strings.ReplaceAll(str, "\r", "\n")
		lines := strings.Split(str, "\n")
		total := 0
		for i, line := range lines {
			total += t.DrawString(x, y+i, style, line)
		}
		return total
	}

	total := 0
	for _, r := range str {
		rw := RuneWidth(r)
		if rw == 0 {
			continue
		}
		if x >= 0 {
			_, drawn := t.DrawRune(x, y, style, r)
			if !drawn {
				break
			}
			total += rw
		}
		x += rw
	}
	return total
}

// RuneWidth возвращает ширину руны в ячейках терминала:
// 0 — невидимая (комбинирующие, zero-width, управляющие),
// 1 — обычная,
// 2 — широкая (CJK, Hangul, Kana, emoji, fullwidth).
func RuneWidth(r rune) int {
	if r == 0 {
		return 0
	}
	// Управляющие и DEL — невидимы
	if r < 0x20 || r == 0x7F {
		return 0
	}

	if r < 0x7F {
		return 1
	}

	// Комбинирующие и zero-width — 0
	if unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Me, r) {
		return 0
	}
	if r >= 0x200B && r <= 0x200F {
		return 0
	}
	if r >= 0xFE00 && r <= 0xFE0F {
		return 0
	}
	if r >= 0xFE20 && r <= 0xFE2F {
		return 0
	}

	// Широкие — 2
	switch {
	case r >= 0x1100 && r <= 0x115F: // Hangul Jamo
	case r >= 0x2600 && r <= 0x27BF: // Misc symbols, Dingbats
	case r >= 0x2B00 && r <= 0x2BFF: // Misc symbols and arrows
	case r >= 0x2E80 && r <= 0x303E: // CJK Radicals, Kangxi
	case r >= 0x3041 && r <= 0x33FF: // Hiragana, Katakana
	case r >= 0x3400 && r <= 0x4DBF: // CJK Ext A
	case r >= 0x4E00 && r <= 0x9FFF: // CJK Unified
	case r >= 0xA000 && r <= 0xA4CF: // Yi
	case r >= 0xAC00 && r <= 0xD7A3: // Hangul Syllables
	case r >= 0xF900 && r <= 0xFAFF: // CJK Compatibility Ideographs
	case r >= 0xFE10 && r <= 0xFE19: // Vertical Forms
	case r >= 0xFE30 && r <= 0xFE6F: // CJK Compatibility Forms
	case r >= 0xFF00 && r <= 0xFF60: // Fullwidth Forms
	case r >= 0xFFE0 && r <= 0xFFE6: // Fullwidth Signs
	case r >= 0x1F000 && r <= 0x1F02F: // Mahjong
	case r >= 0x1F0A0 && r <= 0x1F0FF: // Playing cards
	case r >= 0x1F100 && r <= 0x1F1FF: // Enclosed alphanumerics
	case r >= 0x1F200 && r <= 0x1F2FF: // Enclosed ideographic
	case r >= 0x1F300 && r <= 0x1F64F: // Emoji, Misc Symbols
	case r >= 0x1F680 && r <= 0x1F6FF: // Transport and Map
	case r >= 0x1F700 && r <= 0x1F77F: // Alchemical
	case r >= 0x1F780 && r <= 0x1F7FF: // Geometric Shapes Extended
	case r >= 0x1F800 && r <= 0x1F8FF: // Supplemental Arrows-C
	case r >= 0x1F900 && r <= 0x1F9FF: // Supplemental Symbols
	case r >= 0x1FA00 && r <= 0x1FAFF: // Symbols and Pictographs Ext
	case r >= 0x20000 && r <= 0x2FFFD: // CJK Ext B+
	case r >= 0x30000 && r <= 0x3FFFD: // CJK Ext G+
		return 2
	default:
		return 1
	}
	return 2
}

// Invalidate заставляет следующий Flush перерисовать весь экран
// без сравнения с oldBuf. Нужен при ресайзе: терминал делает reflow
// содержимого, и diff-логика не знает, что реально на экране.
func (t *Terminal) Invalidate() {
	t.forceRedraw = true
}
