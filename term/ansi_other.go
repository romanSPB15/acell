//go:build !windows

package term

import "os"

func enableANSIWindowsFile(f *os.File) {
}
