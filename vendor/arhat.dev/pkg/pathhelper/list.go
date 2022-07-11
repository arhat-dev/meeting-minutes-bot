package pathhelper

import (
	"strings"
)

type CharMatchFunc func(byte) bool

// IsWindowsPathListSeparator returns true when c is colon or semi-colon
func IsWindowsPathListSeparator(c byte) bool {
	return c == UnixPathListSeparator || c == WindowsPathListSeparator
}

// IsUnixPathListSeparator returns true when c is colon
func IsUnixPathListSeparator(c byte) bool {
	return c == UnixPathListSeparator
}

// IsWindowsSlash returns true when c is slash or back slash
func IsWindowsSlash(c byte) bool {
	return c == WindowsSlash || c == UnixSlash
}

// IsUnixSlash returns true when c is slash
func IsUnixSlash(c byte) bool {
	return c == UnixSlash
}

// IsWindowsDiskDesignatorChar returns true when c matches [a-zA-Z]
func IsWindowsDiskDesignatorChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

// SplitList is like filepath.SplitList, but caller controls
// whether take colon and/or semi-colon as path separators
//
// it also handles double-quoted path segments
//
// NOTE for windows path lists:
// 	behavior differences from win32:
//	- when colonIsPathSeparator is true, "C:tmp.txt" will be splited into `C` and `tmp.txt`
// 	  instead of keeping it as like win32's behavior: a file named "tmp.txt" in the current directory on drive C
//
// NOTE for unix style path list:
// 	when colonIsPathSeparator is true (most of the case)
// 	`[a-zA-Z]:/` in middle of the path list followd by an absolute path will be joined with next path segment
// 		e.g. `/root/bin:c:/foo` becomes `/root/bin`, `c:/foo`
// 	so it's preferable to avoid single character relative path in the path list
func SplitList(
	path string,
	colonSep bool,
	semiColonSep bool,
	isSlash CharMatchFunc,
) []string {
	if path == "" {
		return []string{""}
	}

	// split path, respecting but preserving quotes

	list := []string{}
	start := 0
	quo := false
	for i := 0; i < len(path); i++ {
		switch c := path[i]; {
		case c == '"':
			quo = !quo
		case c == WindowsPathListSeparator && !quo && semiColonSep:
			list = append(list, path[start:i])
			start = i + 1
		case c == UnixPathListSeparator && !quo && colonSep:
			if i == 0 {
				// `:`
				start = i + 1
				continue
			}

			// eliminate corner case
			if len(path) == i+1 {
				// foo:<empty>, last segment is empty
				list = append(list, path[start:i])
				start = i + 1
				break
			}

			// look back, handle windows disk designator

			// last char can only be [a-zA-Z] if it's windows absolute path
			if !IsWindowsDiskDesignatorChar(path[i-1]) || i-start != 1 {
				// `[^a-zA-Z]:` or .+[a-zA-Z]
				list = append(list, path[start:i])
				start = i + 1
				continue
			}

			// look forward, continue handling windows disk designator

			// next char can only be path slash if it's windows absolute path
			if !isSlash(path[i+1]) {
				// `[a-zA-Z]:[^/\]` (e.g. `c:b``)
				list = append(list, path[start:i])
				start = i + 1
				continue
			}
		}
	}
	list = append(list, path[start:])

	// Remove quotes.
	for i, s := range list {
		list[i] = strings.ReplaceAll(s, `"`, ``)
	}

	return list
}
