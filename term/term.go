// Пакет term предоставляет RawTerminal и его реализацию.
package term

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"runtime"

	"github.com/romanSPB15/acell/terminfo"
	"golang.org/x/term"
)

// RawTerminal представляет собой терминал.
type RawTerminal interface {
	io.ReadWriteCloser
	MakeRaw() error
	Restore() error
	EnableANSI() error
	Events() <-chan any
	Size() (int, int)
	StartInput()

	Info() terminfo.Info
}

// ErrorNotRaw возвращается, если терминал не был переведён в raw-режим.
var ErrorNotRaw = errors.New("term: terminal not in raw mode")

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

// TTY возвращает потоки ввода-вывода, привязанные к управляющему терминалу.
// На Windows — CONIN$/CONOUT$, на Unix — /dev/tty.
// Если управляющий терминал недоступен, возвращает os.Stdin/os.Stdout.
func TTY() (io.Reader, io.Writer) {
	if runtime.GOOS == "windows" {
		in, err := os.OpenFile("CONIN$", os.O_RDWR, 0)
		if err != nil {
			return os.Stdin, os.Stdout
		}
		out, err := os.OpenFile("CONOUT$", os.O_RDWR, 0)
		if err != nil {
			in.Close()
			return os.Stdin, os.Stdout
		}
		return in, out
	}

	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return os.Stdin, os.Stdout
	}
	return tty, tty
}
