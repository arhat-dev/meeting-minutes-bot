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

type Op uint32

const (
	Op_Unknown Op = iota

	Op_Abs // operation to get absolute path
	Op_Sub // operation to create sub fs

	Op_Create  // operation to create file
	Op_Symlink // operation to create symlink

	Op_Mkdir    // operation to create dir
	Op_MkdirAll // operation to create dir recursively

	Op_OpenFile // operation to open file with options
	Op_Open     // operation to open file for reading

	Op_Lstat // operation to lstat
	Op_Stat  // operation to stat

	Op_ReadFile // operation to read file content
	Op_ReadDir  // operation to read dir file list
	Op_Readlink // operation to read link destination

	Op_WriteFile // operation to write content to file

	Op_Chmod   // operation to change permission flags of file
	Op_Chown   // operation to change owner flags of file
	Op_Chtimes // operation to change time values of file

	Op_Rename // operation to rename file

	Op_Remove    // operation to remove file
	Op_RemoveAll // operation to remove files recursively
)

// op and name are operation and path value triggered this func,
// return value is the absolute path of current working dir
//
// when operation involves more than one path value (e.g. Op_Symlink),
// name is the old path value
type CwdGetFunc = func(op Op, name string) (string, error)

func sysCwd(Op, string) (string, error) {
	return os.Getwd()
}

// NewOSFS creates a new filesystem abstraction for real filesystem
// set strictIOFS to true to only allow fs path value
//
// getCwd is
func NewOSFS(strictIOFS bool, getCwd CwdGetFunc) *OSFS {
	if getCwd == nil {
		getCwd = sysCwd
	}

	return &OSFS{
		Strict: strictIOFS,
		GetCwd: getCwd,
	}
}

// OSFS is a context aware filesystem abstration for afero.FS and io/fs.FS
type OSFS struct {
	// Strict when set to true, adhere to io/fs.FS path value requirements
	// otherwise accepts all system path values
	Strict bool

	// used to determine current working dir, the string return value should be valid system
	// file path
	GetCwd CwdGetFunc

	// LookupFHS is the custom handler for unix style path on windows
	LookupFHS func(string) (string, error)
}

// getRealPath of name by joining current working dir when name is relative path
// name MUST be valid fs path value in strict mode
//
// the returned rpath value is always system file path
func (ofs *OSFS) getRealPath(op Op, name string) (cwd, rpath string, err error) {
	if (!fs.ValidPath(name) || runtime.GOOS == kernelconst.Windows && strings.ContainsAny(name, `\:`)) && ofs.Strict {
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

		cwd, err = ofs.GetCwd(op, name)
		if err != nil {
			return "", "", err
		}

		lookupFHS := ofs.LookupFHS
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

	cwd, err = ofs.GetCwd(op, name)
	if err != nil {
		return "", "", err
	}

	return cwd, filepath.Join(cwd, name), nil
}

// WriteFile is the os.WriteFile equivalent
func (ofs *OSFS) WriteFile(name string, data []byte, perm fs.FileMode) error {
	_, path, err := ofs.getRealPath(Op_WriteFile, name)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, perm)
}

// ReadDir is the os.ReadDir equivalent
// implements fs.ReadDirFS
func (ofs *OSFS) ReadDir(name string) ([]fs.DirEntry, error) {
	_, path, err := ofs.getRealPath(Op_ReadDir, name)
	if err != nil {
		return nil, err
	}

	return os.ReadDir(path)
}

// ReadFile is the os.ReadFile equivalent
// implements fs.ReadFileFS
func (ofs *OSFS) ReadFile(name string) ([]byte, error) {
	_, path, err := ofs.getRealPath(Op_ReadFile, name)
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
	_, path, err := ofs.getRealPath(Op_Sub, dir)
	if err != nil {
		return nil, err
	}

	return NewOSFS(ofs.Strict, func(op Op, name string) (string, error) { return path, nil }), nil
}

// Create is the os.Create equivalent
func (ofs *OSFS) Create(name string) (fs.File, error) {
	_, path, err := ofs.getRealPath(Op_Create, name)
	if err != nil {
		return nil, err
	}

	return os.Create(path)
}

// Mkdir is the os.Mkdir equivalent
func (ofs *OSFS) Mkdir(name string, perm fs.FileMode) error {
	_, path, err := ofs.getRealPath(Op_Mkdir, name)
	if err != nil {
		return err
	}

	return os.Mkdir(path, perm)
}

// MkdirAll is the os.MkdirAll equivalent
func (ofs *OSFS) MkdirAll(name string, perm fs.FileMode) error {
	_, path, err := ofs.getRealPath(Op_MkdirAll, name)
	if err != nil {
		return err
	}

	return os.MkdirAll(path, perm)
}

// Open is the os.Open equivalent
// implements fs.FS
func (ofs *OSFS) Open(name string) (fs.File, error) {
	_, path, err := ofs.getRealPath(Op_Open, name)
	if err != nil {
		return nil, err
	}

	return os.Open(path)
}

// OpenFile is the os.OpenFile equivalent
func (ofs *OSFS) OpenFile(name string, flag int, perm fs.FileMode) (fs.File, error) {
	_, path, err := ofs.getRealPath(Op_OpenFile, name)
	if err != nil {
		return nil, err
	}

	return os.OpenFile(path, flag, perm)
}

// Remove is the os.Remove equivalent
func (ofs *OSFS) Remove(name string) error {
	_, path, err := ofs.getRealPath(Op_Remove, name)
	if err != nil {
		return err
	}

	return os.Remove(path)
}

// RemoveAll is the os.RemoveAll equivalent
func (ofs *OSFS) RemoveAll(name string) error {
	_, path, err := ofs.getRealPath(Op_RemoveAll, name)
	if err != nil {
		return err
	}

	return os.RemoveAll(path)
}

// Rename is the os.Rename equivalent
func (ofs *OSFS) Rename(oldname, newname string) error {
	_, oldPath, err := ofs.getRealPath(Op_Rename, oldname)
	if err != nil {
		return err
	}

	_, newPath, err := ofs.getRealPath(Op_Rename, newname)
	if err != nil {
		return err
	}

	return os.Rename(oldPath, newPath)
}

// Stat is the os.Stat equivalent
func (ofs *OSFS) Stat(name string) (fs.FileInfo, error) {
	_, path, err := ofs.getRealPath(Op_Stat, name)
	if err != nil {
		return nil, err
	}

	return os.Stat(path)
}

// Chmod is the os.Chmod equivalent
func (ofs *OSFS) Chmod(name string, mode fs.FileMode) error {
	_, path, err := ofs.getRealPath(Op_Chmod, name)
	if err != nil {
		return err
	}

	return os.Chmod(path, mode)
}

// Chown is the os.Chown equivalent
func (ofs *OSFS) Chown(name string, uid, gid int) error {
	_, path, err := ofs.getRealPath(Op_Chown, name)
	if err != nil {
		return err
	}

	return os.Chown(path, uid, gid)
}

// Chtimes is the os.Chtimes equivalent
func (ofs *OSFS) Chtimes(name string, atime time.Time, mtime time.Time) error {
	_, path, err := ofs.getRealPath(Op_Chtimes, name)
	if err != nil {
		return err
	}

	return os.Chtimes(path, atime, mtime)
}

// Lstat is the os.Lstat equivalent
func (ofs *OSFS) Lstat(name string) (fs.FileInfo, error) {
	_, path, err := ofs.getRealPath(Op_Lstat, name)
	if err != nil {
		return nil, err
	}

	return os.Lstat(path)
}

// Symlink is the os.Symlink equivalent
func (ofs *OSFS) Symlink(oldname, newname string) error {
	if path.IsAbs(oldname) && path.IsAbs(newname) {
		if ofs.Strict {
			return &fs.PathError{
				Op:   "",
				Path: oldname,
				Err:  fs.ErrInvalid,
			}
		}

		// nothing to do with cwd
		return os.Symlink(oldname, newname)
	}

	if ofs.Strict {
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

	cwd, err := ofs.GetCwd(Op_Symlink, oldname)
	if err != nil {
		return err
	}

	// either oldname or newname is relative path
	// so we need to create symlink based on the current working dir
	return Symlinkat(cwd, oldname, newname)
}

// Readlink is the os.Readlink equivalent
func (ofs *OSFS) Readlink(name string) (string, error) {
	_, path, err := ofs.getRealPath(Op_Readlink, name)
	if err != nil {
		return "", err
	}

	return os.Readlink(path)
}

// Abs is the filepath.Abs equivalent
func (ofs *OSFS) Abs(name string) (path string, err error) {
	_, path, err = ofs.getRealPath(Op_Abs, name)
	if err != nil {
		return
	}

	return
}
