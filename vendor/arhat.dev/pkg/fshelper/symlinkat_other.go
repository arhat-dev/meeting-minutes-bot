//go:build windows || solaris || js || aix || plan9

package fshelper

import "os"

// Symlinkat creates symlink relative to cwd rather actual working dir
func Symlinkat(cwd, oldname, newname string) error {
	actualCwd, err := os.Getwd()
	if err != nil {
		return err
	}

	defer func() {
		err = os.Chdir(actualCwd)
		if err != nil {
			// TODO: better error handling
			panic(err)
		}
	}()

	err = os.Chdir(cwd)
	if err != nil {
		return err
	}

	return os.Symlink(oldname, newname)
}
