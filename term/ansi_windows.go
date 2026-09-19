//go:build windows

package term

import (
	"os"

	"golang.org/x/sys/windows"
)

// enableANSIWindows включает ENABLE_VIRTUAL_TERMINAL_PROCESSING на Windows для поддержки ANSI в указанном файле.
// На других ОС заглушка.
func enableANSIWindowsFile(f *os.File) {
	if f == nil {
		return
	}
	h := windows.Handle(f.Fd())
	var mode uint32
	if err := windows.GetConsoleMode(h, &mode); err != nil {
		return
	}
	mode |= windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING
	_ = windows.SetConsoleMode(h, mode)
}
