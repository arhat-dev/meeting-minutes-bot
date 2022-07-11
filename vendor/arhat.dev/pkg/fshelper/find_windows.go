//go:build windows

package fshelper

import (
	"fmt"
	"path/filepath"
	"syscall"
	"time"

	"golang.org/x/sys/windows"
)

func (ofs *OSFS) matchFileSysinfo(opts *FindOptions, path string, f any) (ok bool, err error) {
	const (
		needUserGroup = FindOp_CheckUser | FindOp_CheckGroup | FindOp_CheckUserInvalid
	)

	info, ok := f.(*syscall.Win32FileAttributeData)
	if !ok {
		err = fmt.Errorf("unexpected not *syscall.Win32FileAttributeData: %T", f)
		return
	}

	ops := opts.Ops

	if ops&needUserGroup != 0 {
		var (
			sd  *windows.SECURITY_DESCRIPTOR
			sid *windows.SID
		)

		sd, err = windows.GetNamedSecurityInfo(
			filepath.FromSlash(path),
			windows.SE_FILE_OBJECT,
			windows.OWNER_SECURITY_INFORMATION,
		)
		if err != nil {
			return false, err
		}

		if ops&FindOp_CheckUserInvalid != 0 && sd.IsValid() {
			return false, nil
		}

		if ops&FindOp_CheckUser != 0 {
			sid, _, err = sd.Owner()
			if err != nil {
				return
			}

			if sid.String() != opts.WindowsOrPlan9User {
				return false, nil
			}
		}

		if ops&FindOp_CheckGroup != 0 {
			sid, _, err = sd.Group()
			if err != nil {
				return
			}

			if sid.String() != opts.WindowsOrPlan9Group {
				return false, nil
			}
		}
	}

	if ops&FindOp_CheckCreationTime != 0 {
		birthTime := info.CreationTime.Nanoseconds() / int64(time.Second)
		if birthTime < opts.MinCreationTime || birthTime > opts.MaxCreationTime {
			return false, nil
		}
	}

	if ops&FindOp_CheckLastAccessTime != 0 {
		atime := info.LastAccessTime.Nanoseconds() / int64(time.Second)
		if atime < opts.MinAtime || atime > opts.MaxAtime {
			return false, nil
		}
	}

	if ops&FindOp_CheckLastMetadataChangeTime != 0 {
		// TODO: there is no ctime support on windows
	}

	if ops&FindOp_CheckLastContentUpdatedTime != 0 {
		mtime := info.LastWriteTime.Nanoseconds() / int64(time.Second)
		if mtime < opts.MinCtime || mtime > opts.MaxCtime {
			return false, nil
		}
	}

	return true, nil
}
