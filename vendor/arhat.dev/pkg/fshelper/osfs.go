package fshelper

import (
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"arhat.dev/pkg/kernelconst"
	"arhat.dev/pkg/pathhelper"
	"github.com/bmatcuk/doublestar/v4"
)

// NewOSFS creates a new filesystem abstraction for real filesystem
// set strictIOFS to true to only allow fs path value
// getCwd is used to determine current working dir, the string return value should be valid system
// file path
func NewOSFS(
	strictIOFS bool,
	getCwd func() (string, error),
) *OSFS {
	if getCwd == nil {
		getCwd = os.Getwd
	}

	return &OSFS{
		strict: strictIOFS,
		getCwd: getCwd,
	}
}

// OSFS is a context aware filesystem abstration for afero.FS and io/fs.FS
type OSFS struct {
	strict    bool
	getCwd    func() (string, error)
	lookupFHS func(string) (string, error)
}

// SetStrict sets strict mode to require io/fs.FS path value
func (ofs *OSFS) SetStrict(s bool) *OSFS {
	ofs.strict = s
	return ofs
}

// SetWindowsFHSLookup sets custom handler for unix style path
func (ofs *OSFS) SetWindowsFHSLookup(lookup func(path string) (string, error)) *OSFS {
	ofs.lookupFHS = lookup
	return ofs
}

// getRealPath of name by joining current working dir when name is relative path
// name MUST be valid fs path value in strict mode
//
// the returned rpath value is always system file path
func (ofs *OSFS) getRealPath(name string) (cwd, rpath string, err error) {
	if (!fs.ValidPath(name) || runtime.GOOS == kernelconst.Windows && strings.ContainsAny(name, `\:`)) && ofs.strict {
		return "", "", &fs.PathError{
			Op:   "",
			Err:  fs.ErrInvalid,
			Path: name,
		}
	}

	if path.IsAbs(name) {
		// path starts with `/`
		if runtime.GOOS != kernelconst.Windows {
			// on unix platforms, it means absolute path
			// no more action required
			return "", name, nil
		}

		// on windows it has different meaning
		// - /foo is driver relative, should be relative to current working dir's driver
		// - /c/foo is absolute path in driver c
		// - /cygdrive/c/foo is the same as /c/foo, but with extra /cygdrive prefix
		// - /usr/foo is the path inside msys2/mingw64/cygwin root

		cwd, err = ofs.getCwd()
		if err != nil {
			return "", "", err
		}

		lookupFHS := ofs.lookupFHS
		if lookupFHS == nil {
			lookupFHS = func(path string) (string, error) {
				ret, err := exec.Command("cygpath", "-w", path).CombinedOutput()
				if err == nil {
					return strings.TrimSpace(string(ret)), nil
				}

				ret, err = exec.Command("winepath", "-w", path).CombinedOutput()
				if err == nil {
					return strings.TrimSpace(string(ret)), nil
				}

				return pathhelper.ConvertFSPathToWindowsPath(filepath.VolumeName(cwd), path), nil
			}
		}

		rpath, err = pathhelper.AbsWindowsPath(cwd, name, lookupFHS)

		return "", rpath, err
	}

	if runtime.GOOS == kernelconst.Windows && pathhelper.IsWindowsAbs(name) {
		// handle paths like c:\\, \\host\share
		return "", name, nil
	}

	// is relative path for both windows and unix

	cwd, err = ofs.getCwd()
	if err != nil {
		return "", "", err
	}

	return cwd, filepath.Join(cwd, name), nil
}

// WriteFile is the os.WriteFile equivalent
func (ofs *OSFS) WriteFile(name string, data []byte, perm fs.FileMode) error {
	_, path, err := ofs.getRealPath(name)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, perm)
}

// ReadDir is the os.ReadDir equivalent
// implements fs.ReadDirFS
func (ofs *OSFS) ReadDir(name string) ([]fs.DirEntry, error) {
	_, path, err := ofs.getRealPath(name)
	if err != nil {
		return nil, err
	}

	return os.ReadDir(path)
}

// ReadFile is the os.ReadFile equivalent
// implements fs.ReadFileFS
func (ofs *OSFS) ReadFile(name string) ([]byte, error) {
	_, path, err := ofs.getRealPath(name)
	if err != nil {
		return nil, err
	}

	return os.ReadFile(path)
}

// Glob implements fs.GlobFS with `**` support
func (ofs *OSFS) Glob(pattern string) ([]string, error) {
	return doublestar.Glob(ofs, pattern)
}

// Sub implements fs.SubFS
func (ofs *OSFS) Sub(dir string) (fs.FS, error) {
	_, path, err := ofs.getRealPath(dir)
	if err != nil {
		return nil, err
	}

	return NewOSFS(ofs.strict, func() (string, error) { return path, nil }), nil
}

// Create is the os.Create equivalent
func (ofs *OSFS) Create(name string) (fs.File, error) {
	_, path, err := ofs.getRealPath(name)
	if err != nil {
		return nil, err
	}

	return os.Create(path)
}

// Mkdir is the os.Mkdir equivalent
func (ofs *OSFS) Mkdir(name string, perm fs.FileMode) error {
	_, path, err := ofs.getRealPath(name)
	if err != nil {
		return err
	}

	return os.Mkdir(path, perm)
}

// MkdirAll is the os.MkdirAll equivalent
func (ofs *OSFS) MkdirAll(name string, perm fs.FileMode) error {
	_, path, err := ofs.getRealPath(name)
	if err != nil {
		return err
	}

	return os.MkdirAll(path, perm)
}

// Open is the os.Open equivalent
// implements fs.FS
func (ofs *OSFS) Open(name string) (fs.File, error) {
	_, path, err := ofs.getRealPath(name)
	if err != nil {
		return nil, err
	}

	return os.Open(path)
}

// OpenFile is the os.OpenFile equivalent
func (ofs *OSFS) OpenFile(name string, flag int, perm fs.FileMode) (fs.File, error) {
	_, path, err := ofs.getRealPath(name)
	if err != nil {
		return nil, err
	}

	return os.OpenFile(path, flag, perm)
}

// Remove is the os.Remove equivalent
func (ofs *OSFS) Remove(name string) error {
	_, path, err := ofs.getRealPath(name)
	if err != nil {
		return err
	}

	return os.Remove(path)
}

// RemoveAll is the os.RemoveAll equivalent
func (ofs *OSFS) RemoveAll(name string) error {
	_, path, err := ofs.getRealPath(name)
	if err != nil {
		return err
	}

	return os.RemoveAll(path)
}

// Rename is the os.Rename equivalent
func (ofs *OSFS) Rename(oldname, newname string) error {
	_, oldPath, err := ofs.getRealPath(oldname)
	if err != nil {
		return err
	}

	_, newPath, err := ofs.getRealPath(newname)
	if err != nil {
		return err
	}

	return os.Rename(oldPath, newPath)
}

// Stat is the os.Stat equivalent
func (ofs *OSFS) Stat(name string) (fs.FileInfo, error) {
	_, path, err := ofs.getRealPath(name)
	if err != nil {
		return nil, err
	}

	return os.Stat(path)
}

// Chmod is the os.Chmod equivalent
func (ofs *OSFS) Chmod(name string, mode fs.FileMode) error {
	_, path, err := ofs.getRealPath(name)
	if err != nil {
		return err
	}

	return os.Chmod(path, mode)
}

// Chown is the os.Chown equivalent
func (ofs *OSFS) Chown(name string, uid, gid int) error {
	_, path, err := ofs.getRealPath(name)
	if err != nil {
		return err
	}

	return os.Chown(path, uid, gid)
}

// Chtimes is the os.Chtimes equivalent
func (ofs *OSFS) Chtimes(name string, atime time.Time, mtime time.Time) error {
	_, path, err := ofs.getRealPath(name)
	if err != nil {
		return err
	}

	return os.Chtimes(path, atime, mtime)
}

// Lstat is the os.Lstat equivalent
func (ofs *OSFS) Lstat(name string) (fs.FileInfo, error) {
	_, path, err := ofs.getRealPath(name)
	if err != nil {
		return nil, err
	}

	return os.Lstat(path)
}

// Symlink is the os.Symlink equivalent
func (ofs *OSFS) Symlink(oldname, newname string) error {
	if path.IsAbs(oldname) && path.IsAbs(newname) {
		if ofs.strict {
			return &fs.PathError{
				Op:   "",
				Path: oldname,
				Err:  fs.ErrInvalid,
			}
		}

		// nothing to do with cwd
		return os.Symlink(oldname, newname)
	}

	if ofs.strict {
		if !fs.ValidPath(oldname) {
			return &fs.PathError{
				Op:   "",
				Path: oldname,
				Err:  fs.ErrInvalid,
			}
		}

		if !fs.ValidPath(newname) {
			return &fs.PathError{
				Op:   "",
				Path: newname,
				Err:  fs.ErrInvalid,
			}
		}
	}

	cwd, err := ofs.getCwd()
	if err != nil {
		return err
	}

	// either oldname or newname is relative path
	// so we need to create symlink based on the current working dir
	return Symlinkat(cwd, oldname, newname)
}

// Readlink is the os.Readlink equivalent
func (ofs *OSFS) Readlink(name string) (string, error) {
	_, path, err := ofs.getRealPath(name)
	if err != nil {
		return "", err
	}

	return os.Readlink(path)
}

// Abs is the filepath.Abs equivalent
func (ofs *OSFS) Abs(name string) (path string, err error) {
	_, path, err = ofs.getRealPath(name)
	if err != nil {
		return
	}

	return
}
