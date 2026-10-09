package term

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/romanSPB15/acell/input"
	"github.com/romanSPB15/acell/terminfo"
	"golang.org/x/term"
)

// ResizeEvent отправляется в Events при изменении размера терминала.
type ResizeEvent struct {
	Width  int
	Height int
}

type rawTerminal struct {
	in  io.Reader
	out io.Writer

	inFile  *os.File
	outFile *os.File

	events  chan any
	inputCh chan []byte
	stopCh  chan struct{}
	once    sync.Once

	lastW, lastH int

	oldState *term.State
	oldFile  *os.File

	infoMu sync.RWMutex
	info   terminfo.Info
}

// Info возвращает текущие характеристики терминала.
func (t *rawTerminal) Info() terminfo.Info {
	t.infoMu.RLock()
	defer t.infoMu.RUnlock()
	return t.info
}

// NewRawTerminal оборачивает in/out в RawTerminal.
//
// Если in или out являются *os.File, операции с терминалом
// (MakeRaw, Size, EnableANSI) используют их файловые дескрипторы.
// В противном случае эти операции возвращают ошибку или ничего не делают.
//
// in/out можно передать nil — тогда используются os.Stdin / os.Stdout.
func NewRawTerminal(in io.Reader, out io.Writer) RawTerminal {
	if in == nil {
		in = os.Stdin
	}
	if out == nil {
		out = os.Stdout
	}

	rt := &rawTerminal{
		in:      in,
		out:     out,
		events:  make(chan any, 64),
		inputCh: make(chan []byte, 16),
		stopCh:  make(chan struct{}),
		info:    terminfo.Detect(),
	}
	if f, ok := in.(*os.File); ok {
		rt.inFile = f
	}
	if f, ok := out.(*os.File); ok {
		rt.outFile = f
	}

	return rt
}

// StartInput запускает чтение из in и запрашивает размер ячейки терминала.
func (rt *rawTerminal) StartInput() {
	rt.Write([]byte("\033[16t"))
	go rt.readLoop()
	go rt.inputLoop()
	go rt.resizeLoop()
}

// Write реализует io.Writer.
func (rt *rawTerminal) Write(p []byte) (int, error) {
	return rt.out.Write(p)
}

// Read реализует io.Reader.
func (rt *rawTerminal) Read(p []byte) (int, error) {
	return rt.in.Read(p)
}

// MakeRaw переводит входной файл в raw-режим.
func (rt *rawTerminal) MakeRaw() error {
	if rt.inFile == nil {
		return ErrorNotRaw
	}
	s, err := term.MakeRaw(int(rt.inFile.Fd()))
	if err != nil {
		return err
	}
	rt.oldState = s
	rt.oldFile = rt.inFile
	return nil
}

// Restore возвращает терминал в исходный режим.
func (rt *rawTerminal) Restore() error {
	if rt.oldState == nil || rt.oldFile == nil {
		return ErrorNotRaw
	}
	err := term.Restore(int(rt.oldFile.Fd()), rt.oldState)
	rt.oldState = nil
	rt.oldFile = nil
	return err
}

// EnableANSI включает ANSI на Windows (на других ОС — no-op).
func (rt *rawTerminal) EnableANSI() error {
	if rt.outFile != nil {
		enableANSIWindowsFile(rt.outFile)
	}
	return nil
}

// Events возвращает канал событий.
func (rt *rawTerminal) Events() <-chan any {
	return rt.events
}

// Close останавливает чтение и восстанавливает терминал. Идемпотентен.
func (rt *rawTerminal) Close() error {
	rt.once.Do(func() {
		close(rt.stopCh)
		if rt.oldState != nil && rt.oldFile != nil {
			_ = term.Restore(int(rt.oldFile.Fd()), rt.oldState)
			rt.oldState = nil
			rt.oldFile = nil
		}
	})
	return nil
}

func (rt *rawTerminal) readLoop() {
	buf := make([]byte, 1024)

	type deadlineReader interface {
		SetReadDeadline(time.Time) error
	}
	dr, hasDeadline := rt.in.(deadlineReader)

	for {
		if hasDeadline {
			dr.SetReadDeadline(time.Now().Add(50 * time.Millisecond))
		}

		n, err := rt.in.Read(buf)
		if err != nil {
			if os.IsTimeout(err) {
				select {
				case <-rt.stopCh:
					return
				default:
					continue
				}
			}
			return
		}
		if n == 0 {
			continue
		}

		data := make([]byte, n)
		copy(data, buf[:n])

		select {
		case <-rt.stopCh:
			return
		case rt.inputCh <- data:
		}
	}
}

func (rt *rawTerminal) inputLoop() {
	var pending []byte

	for {
		select {
		case <-rt.stopCh:
			return
		case data, ok := <-rt.inputCh:
			if !ok {
				return
			}
			pending = append(pending, data...)

			for len(pending) > 0 {
				if n := rt.tryCellSizeReply(pending); n > 0 {
					pending = pending[n:]
					continue
				}
				ev, n := input.ParseOne(pending)
				if n == 0 {
					break
				}
				pending = pending[n:]
				if ev != nil {
					rt.emit(ev)
				}
			}
		}
	}
}

// tryCellSizeReply разбирает ответ \033[6;H;Wt. Возвращает число
// потреблённых байт или 0, если это не ответ на запрос размера ячейки.
func (rt *rawTerminal) tryCellSizeReply(p []byte) int {
	if len(p) < 4 {
		return 0
	}
	if p[0] != 0x1b || p[1] != '[' || p[2] != '6' || p[3] != ';' {
		return 0
	}
	for i := 4; i < len(p); i++ {
		if p[i] == 0x1b {
			return 0
		}
		if p[i] == 't' {
			var h, w int
			if _, err := fmt.Sscanf(string(p[:i+1]), "\033[6;%d;%dt", &h, &w); err == nil && h > 0 && w > 0 {
				rt.infoMu.Lock()
				rt.info.CellH = h
				rt.info.CellW = w
				rt.infoMu.Unlock()
			}
			return i + 1
		}
	}
	return 0
}

func (rt *rawTerminal) emit(ev any) {
	select {
	case <-rt.stopCh:
	case rt.events <- ev:
	}
}

func (rt *rawTerminal) Size() (int, int) {
	if rt.outFile != nil {
		return sizeFd(rt.outFile.Fd())
	}
	if rt.inFile != nil {
		return sizeFd(rt.inFile.Fd())
	}
	return 0, 0
}

var _ RawTerminal = (*rawTerminal)(nil)
