package pathhelper

import (
	"strings"
)

// IsReservedWindowsName returns truen if the name is reserved in windows
// ref: https://docs.microsoft.com/en-us/windows/win32/fileio/naming-a-file#naming-conventions
func IsReservedWindowsName(name string) bool {
	if len(name) == 0 {
		return false
	}

	for _, reserved := range []string{
		"CON", "PRN", "AUX", "NUL",
		"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
		"LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9",
	} {
		if strings.EqualFold(name, reserved) {
			return true
		}
	}

	return false
}

// IsReservedWindowsPathChar checks whether c is a reserved character for windows path
// ref: https://docs.microsoft.com/en-us/windows/win32/fileio/naming-a-file#naming-conventions
func IsReservedWindowsPathChar(c rune) bool {
	switch c {
	case '<', '>', ':', '"', '/', '\\', '|', '?', '*':
		return true
	}

	return false
}
