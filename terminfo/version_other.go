//go:build !windows

package terminfo

func windowsBuild() uint32 { return 0 }
