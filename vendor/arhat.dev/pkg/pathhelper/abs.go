package pathhelper

import (
	"strings"
)

// IsWindowsAbs reports whether the path is absolute on windows
// the path is considered absolute as following cases
// - reserved windows names (e.g. COM1)
// - starts with `\\<some-host>\<share>` (UNC path)
// - starts with `C:\` (driver absolute path)
// - starts with `\` or `/`
func IsWindowsAbs(path string) (b bool) {
	if IsReservedWindowsName(path) {
		return true
	}
	l := volumeNameLen(path)
	if l == 0 {
		return false
	}
	path = path[l:]
	if path == "" {
		return false
	}

	return IsWindowsSlash(path[0])
}

// AbsWindowsPath returns absolute path of path with custom cwd
// cwd SHOULD be a windows absolute path
// path SHOULD be a relative path (not starting with `[a-zA-Z]:` or `\\`)
//
// It tries to handle three different styles all at once:
// 	- windows native style:
// 		- `foo\bar` and `foo/bar` as path relative
// 		- `\foo` and `/foo` as driver relative
// 	- cygpath style absolute path (`/cygdrive/c`)
// 	- golang io/fs style absolute path for windows (`/[a-zA-Z]/`, e.g. /c/foo)
func AbsWindowsPath(
	cwd, path string,
	getFHSPath func(p string) (string, error),
) (string, error) {
	if len(path) == 0 {
		return CleanWindowsPath(cwd), nil
	}

	if path[0] != '/' {
		if IsWindowsAbs(path) {
			return path, nil
		}

		// starts with:
		// 	`\` (ONLY can be driver relative)
		// 	other (ONLY can be relative path)
		return JoinWindowsPath(cwd, path), nil
	}

	// starts with `/`

	// cygpath or driver relative path

	if strings.HasPrefix(path, "/cygdrive/") {
		// provide empty default volume name since /cygdrive/ MUST be followed
		// by the driver desigantor
		return ConvertFSPathToWindowsPath("", strings.TrimPrefix(path, "/cygdrive")), nil
	}

	return getFHSPath(path)
}
