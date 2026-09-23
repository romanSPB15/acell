package acell

import (
	"bytes"
	"io"
	"sync"

	"github.com/romanSPB15/acell/terminfo"
)

var _ RawTerminal = (*fakeRawTerminal)(nil)

type fakeRawTerminal struct {
	mu sync.Mutex

	written bytes.Buffer
	readBuf bytes.Buffer
	events  chan any

	raw          bool
	ansiEnabled  bool
	inputStarted bool
	closed       bool

	width, height int

	info terminfo.Info

	makeRawErr    error
	restoreErr    error
	enableANSIErr error
	closeErr      error
}

func newFakeRawTerminal() *fakeRawTerminal {
	return &fakeRawTerminal{
		events: make(chan any, 64),
		width:  80,
		height: 24,
		info: terminfo.Info{
			Name:   "fake",
			Colors: 24,

			CursorHide:   "\033[?25l",
			CursorShow:   "\033[?12l\033[?25h",
			AltScreenOn:  "\033[?1049h",
			AltScreenOff: "\033[?1049l",
			Home:         "\033[H",

			Bold:      true,
			Dim:       true,
			Italic:    true,
			Underline: true,
			Reverse:   true,
			Blink:     true,
			Hidden:    true,
			Strike:    true,

			MouseAny: true,
			MouseSGR: true,

			SynchronizedUpdate: true,
		},
	}
}

func (f *fakeRawTerminal) Write(p []byte) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.closed {
		return 0, io.ErrClosedPipe
	}
	return f.written.Write(p)
}

func (f *fakeRawTerminal) Read(p []byte) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.readBuf.Len() == 0 {
		return 0, io.EOF
	}
	return f.readBuf.Read(p)
}

func (f *fakeRawTerminal) MakeRaw() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.makeRawErr != nil {
		return f.makeRawErr
	}
	f.raw = true
	return nil
}

func (f *fakeRawTerminal) Restore() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.restoreErr != nil {
		return f.restoreErr
	}
	f.raw = false
	return nil
}

func (f *fakeRawTerminal) EnableANSI() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.enableANSIErr != nil {
		return f.enableANSIErr
	}
	f.ansiEnabled = true
	return nil
}

func (f *fakeRawTerminal) Events() <-chan any {
	return f.events
}

func (f *fakeRawTerminal) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.closed = true
	f.inputStarted = false
	return f.closeErr
}

func (f *fakeRawTerminal) Size() (int, int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.width, f.height
}

func (f *fakeRawTerminal) StartInput() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.inputStarted = true
}

func (f *fakeRawTerminal) WrittenBytes() []byte {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]byte, f.written.Len())
	copy(out, f.written.Bytes())
	return out
}

func (f *fakeRawTerminal) ANSIEnabled() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.ansiEnabled
}

func (f *fakeRawTerminal) Raw() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.raw
}

func (f *fakeRawTerminal) InputStarted() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.inputStarted
}

func (f *fakeRawTerminal) WrittenString() string {
	return string(f.WrittenBytes())
}

func (f *fakeRawTerminal) ResetWritten() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.written.Reset()
}

func (f *fakeRawTerminal) Closed() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.closed
}

func (f *fakeRawTerminal) FeedInput(p []byte) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.readBuf.Write(p)
}

func (f *fakeRawTerminal) Emit(ev any) {
	f.mu.Lock()
	closed := f.closed
	f.mu.Unlock()
	if closed {
		return
	}
	select {
	case f.events <- ev:
	default:
	}
}

func (f *fakeRawTerminal) SetSize(w, h int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.width, f.height = w, h
}

func (f *fakeRawTerminal) SetMakeRawErr(err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.makeRawErr = err
}

func (f *fakeRawTerminal) SetRestoreErr(err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.restoreErr = err
}

func (f *fakeRawTerminal) SetEnableANSIErr(err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.enableANSIErr = err
}

func (f *fakeRawTerminal) SetCloseErr(err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.closeErr = err
}

func (f *fakeRawTerminal) Info() terminfo.Info {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.info
}

func (f *fakeRawTerminal) SetInfo(info terminfo.Info) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.info = info
}
