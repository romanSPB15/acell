package term

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"time"

	"github.com/romanSPB15/acell/terminfo"
	"golang.org/x/term"
)

// RawTerminal представляет собой терминал.
type RawTerminal interface {
	io.Writer
	io.Reader
	MakeRaw() error
	Restore() error
	EnableANSI() error
	Events() <-chan any
	Close() error
	Size() (int, int)
	StartInput()

	Info() terminfo.Info
}

var (
	chans  []chan []byte
	h      []func([]byte)
	mx     sync.Mutex
	stopCh chan struct{}

	started bool
)

// onInput подписывается на ввод.
// Если fn != nil, добавляет функцию-обработчик.
// Если fn == nil, создаёт и возвращает новый канал.
func onInput(fn func([]byte)) <-chan []byte {
	if fn == nil {
		ch := make(chan []byte, 16)
		mx.Lock()
		chans = append(chans, ch)
		mx.Unlock()
		return ch
	}

	mx.Lock()
	h = append(h, fn)
	mx.Unlock()

	return nil
}

// start запускает чтение из reader.
func start(input io.Reader) {
	if input == nil {
		input = os.Stdin
	}
	mx.Lock()
	defer mx.Unlock()
	if started {
		return
	}
	started = true
	stopCh = make(chan struct{})
	go readLoop(input)
}

// isStarted возвращает, запущено ли чтение.
func isStarted() bool {
	mx.Lock()
	defer mx.Unlock()
	v := started
	return v
}

// stop останавливает чтение из stdin.
func stop() {
	mx.Lock()
	defer mx.Unlock()

	if !started {
		return
	}
	started = false

	close(stopCh)

	for _, ch := range chans {
		func() {
			defer func() { recover() }()
			close(ch)
		}()
	}
	chans = nil
	h = nil
}

func readLoop(r io.Reader) {
	buf := make([]byte, 1024)

	type deadlineReader interface {
		SetReadDeadline(time.Time) error
	}
	dr, hasDeadline := r.(deadlineReader)

	for {
		if hasDeadline {
			dr.SetReadDeadline(time.Now().Add(50 * time.Millisecond))
		}

		n, err := r.Read(buf)
		if err != nil {
			if os.IsTimeout(err) {
				select {
				case <-stopCh:
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

		mx.Lock()

		select {
		case <-stopCh:
			mx.Unlock()
			return
		default:
		}

		for _, ch := range chans {
			select {
			case <-stopCh:
				mx.Unlock()
				return
			case ch <- data:
			default:
			}
		}

		for _, h := range h {
			func() {
				defer func() {
					if err := recover(); err != nil {
						fmt.Printf("term: panic: %v\r\n", err)
					}
				}()
				h(data)
			}()
		}

		mx.Unlock()
	}
}

var (
	ErrorNotRaw = errors.New("term: terminal not in raw mode")
)

var (
	oldState *term.State
	oldFile  *os.File
)

// makeRawFile вводит переданный файл в raw режим.
func makeRawFile(f *os.File) error {
	s, err := term.MakeRaw(int(f.Fd()))
	if err == nil {
		oldState = s
		oldFile = f
	}
	return err
}

// makeRaw вводит os.Stdin в raw режим (для совместимости).
func makeRaw() error {
	return makeRawFile(os.Stdin)
}

// restore выводит из raw тот файл, который был введён makeRawFile.
func restore() error {
	if oldState == nil || oldFile == nil {
		return ErrorNotRaw
	}
	return term.Restore(int(oldFile.Fd()), oldState)
}

// width возвращает ширину терминала в символах.
// В случае ошибки возвращает 0.
func width() int {
	width, _, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		return 0
	}
	return width
}

// height возвращает высоту терминала в строках.
// В случае ошибки возвращает 0.
func height() int {
	_, height, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Println(err)
		return 0
	}
	return height
}

// size возвращает ширину и высоту терминала.
// В случае ошибки возвращает (0, 0).
func size() (int, int) {
	w, h, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		return 0, 0
	}
	return w, h
}

// sizeFd возвращает размеры терминала по заданному дескриптору.
// В случае ошибки возвращает (0, 0).
func sizeFd(fd uintptr) (int, int) {
	w, h, err := term.GetSize(int(fd))
	if err != nil {
		return 0, 0
	}
	return w, h
}

// OpenURL открывает переданный URL в браузере пользователя.
// Поддерживает Windows, macOS и Linux.
func OpenURL(url string) error {
	switch runtime.GOOS {
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		return exec.Command("open", url).Start()
	default:
		return exec.Command("xdg-open", url).Start()
	}
}
