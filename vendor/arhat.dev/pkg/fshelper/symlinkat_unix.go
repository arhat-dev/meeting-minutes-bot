//go:build !(windows || solaris || js || aix || plan9)

package fshelper

import (
	"os"

	"golang.org/x/sys/unix"
)

// Symlinkat creates symlink relative to cwd rather actual working dir
// ref: https://man7.org/linux/man-pages/man2/symlink.2.html
func Symlinkat(cwd, file, linkTo string) error {
	dir, err := os.Open(cwd)
	if err != nil {
		return err
	}
	defer func() { _ = dir.Close() }()

	err = unix.Symlinkat(file, int(dir.Fd()), linkTo)
	if err != nil && err != unix.EINTR {
		return &os.LinkError{
			Op:  "symlinkat",
			Old: file,
			New: linkTo,
			Err: err,
		}
	}

	return nil
}
