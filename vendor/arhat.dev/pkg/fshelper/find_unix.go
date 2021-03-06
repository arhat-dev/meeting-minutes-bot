//go:build darwin || freebsd || netbsd

package fshelper

import (
	"fmt"
	"syscall"
)

func (ofs *OSFS) matchFileSysinfo(opts *FindOptions, path string, f any) (bool, error) {
	const (
		needUserGroup = FindOp_CheckUser | FindOp_CheckGroup
		needTime      = FindOp_CheckCreationTime | FindOp_CheckLastAccessTime | FindOp_CheckLastMetadataChangeTime | FindOp_CheckLastContentUpdatedTime
	)

	info, ok := f.(*syscall.Stat_t)
	if !ok {
		return false, fmt.Errorf("unexpected not *syscall.Stat_t: %T", f)
	}

	ops := opts.Ops

	if ops&FindOp_CheckUserInvalid != 0 {
		// TODO: currently os/user.LookupID is based on reading /etc/passwd when bulit without cgo
		//       doesn't work properly nor efficiently
	}

	if ops&needUserGroup != 0 {
		if ops&FindOp_CheckUser != 0 && uint32(info.Uid) != opts.UnixUID {
			return false, nil
		}

		if ops&FindOp_CheckGroup != 0 && uint32(info.Gid) != opts.UnixGID {
			return false, nil
		}
	}

	if ops&needTime != 0 {
		if ops&FindOp_CheckCreationTime != 0 &&
			(int64(info.Birthtimespec.Sec) < opts.MinCreationTime ||
				int64(info.Birthtimespec.Sec) > opts.MaxCreationTime) {
			return false, nil
		}

		if ops&FindOp_CheckLastAccessTime != 0 &&
			(int64(info.Atimespec.Sec) < opts.MinAtime ||
				int64(info.Atimespec.Sec) > opts.MaxAtime) {
			return false, nil
		}

		if ops&FindOp_CheckLastMetadataChangeTime != 0 &&
			(int64(info.Ctimespec.Sec) < opts.MinCtime ||
				int64(info.Ctimespec.Sec) > opts.MaxCtime) {
			return false, nil
		}

		if ops&FindOp_CheckLastContentUpdatedTime != 0 &&
			(int64(info.Mtimespec.Sec) < opts.MinMtime ||
				int64(info.Mtimespec.Sec) > opts.MaxMtime) {
			return false, nil
		}
	}

	return true, nil
}
