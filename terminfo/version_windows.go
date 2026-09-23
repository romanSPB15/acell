//go:build windows

package terminfo

import (
	"strconv"

	"golang.org/x/sys/windows/registry"
)

// windowsBuild возвращает номер сборки Windows (например, 19045 для Win10 22H2).
// 0 если не удалось определить.
func windowsBuild() uint32 {
	k, err := registry.OpenKey(
		registry.LOCAL_MACHINE,
		`SOFTWARE\Microsoft\Windows NT\CurrentVersion`,
		registry.QUERY_VALUE,
	)
	if err != nil {
		return 0
	}
	defer k.Close()

	s, _, err := k.GetStringValue("CurrentBuildNumber")
	if err != nil {
		return 0
	}

	n, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0
	}
	return uint32(n)
}
