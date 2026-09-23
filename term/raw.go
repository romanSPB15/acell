package term

import (
	"io"
	"os"
	"sync"

	"github.com/romanSPB15/acell/terminfo"
	"golang.org/x/term"
)

// ResizeEvent отправляется в Events при изменении размера терминала.
type ResizeEvent struct {
	Width  int
	Height int
}

// rawTerminal — стандартная реализация RawTerminal.
type rawTerminal struct {
	in  io.Reader
	out io.Writer

	inFile  *os.File // если in удалось привести к *os.File
	outFile *os.File // если out удалось привести к *os.File

	events chan any
	stopCh chan struct{}
	once   sync.Once

	resizeHandlers []func(w, h int)
	resizeMu       sync.Mutex

	lastW, lastH int

	oldState *term.State
	oldFile  *os.File

	info terminfo.Info
}

func (t *rawTerminal) Info() terminfo.Info {
	return t.info
}

// NewRawTerminal оборачивает in/out в RawTerminal.
//
// Если in или out являются *os.File, операции с терминалом
// (MakeRaw, SizeFd, EnableANSI) используют их файловые дескрипторы.
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
		in:     in,
		out:    out,
		events: make(chan any, 64),
		stopCh: make(chan struct{}),
		info:   terminfo.Detect(),
	}
	if f, ok := in.(*os.File); ok {
		rt.inFile = f
	}
	if f, ok := out.(*os.File); ok {
		rt.outFile = f
	}

	return rt
}

// StartInput запускает чтение из in.
func (rt *rawTerminal) StartInput() {
	inputCh := onInput(nil)
	start(rt.in)

	go rt.inputLoop(inputCh)
	go rt.resizeLoop()
}

// Write реализует io.Writer.
func (rt *rawTerminal) Write(p []byte) (int, error) {
	return rt.out.Write(p)
}

// Read реализует io.Reader.
//
// Внимание: одновременный вызов Read и работа inputLoop могут
// конфликтовать за одни и те же байты. Обычно Read нужен только
// для тестов или для интеграции с внешним event loop.
func (rt *rawTerminal) Read(p []byte) (int, error) {
	return rt.in.Read(p)
}

// MakeRaw вводит входной файл в raw-режим.
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

// Events возвращает канал событий: *KeyboardEvent, *MouseEvent, *ResizeEvent.
// Канал буферизован; при переполнении события отбрасываются.
func (rt *rawTerminal) Events() <-chan any {
	return rt.events
}

// Close останавливает event loop и восстанавливает терминал.
// Идемпотентен.
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

func (rt *rawTerminal) inputLoop(inputCh <-chan []byte) {
	for {
		select {
		case <-rt.stopCh:
			return
		case data, ok := <-inputCh:
			if !ok {
				return
			}
			rt.dispatch(data)
		}
	}
}

func (rt *rawTerminal) dispatch(data []byte) {
	if ev := parseMouseEvent(data); ev != nil {
		rt.emit(ev)
		return
	}
	if ev := parseKeyboardInput(data); ev != nil {
		rt.emit(ev)
	}
}

func (rt *rawTerminal) emit(ev any) {
	select {
	case <-rt.stopCh:
	case rt.events <- ev:
	default:

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
