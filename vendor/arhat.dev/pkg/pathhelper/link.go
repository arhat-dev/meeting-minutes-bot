package pathhelper

import (
	"path"
	"strings"
)

// EvalLink returns lintTo's actual path from the point view of linkFile
// usually used in archive file or non actual os filesystem
//
// For unix style path only
func EvalLink(linkFile, linkTo string) string {
	linkTo = CleanUnixPath(linkTo)
	if path.IsAbs(linkTo) {
		return linkTo
	}

	var (
		upperDir  string
		remainder = linkTo
		actualDir = path.Dir(linkFile)
	)

	idx := strings.IndexByte(remainder, UnixSlash)
	for idx != -1 && actualDir != "." {
		upperDir, remainder = remainder[:idx], remainder[idx+1:]

		if upperDir != ".." {
			break
		}

		idx = strings.IndexByte(remainder, UnixSlash)
		actualDir = path.Dir(actualDir)
	}

	if actualDir == "." {
		return remainder
	}

	if strings.HasPrefix(linkTo, "..") {
		return path.Join(linkFile, linkTo)
	}

	return path.Join(path.Dir(linkFile), linkTo)
}
