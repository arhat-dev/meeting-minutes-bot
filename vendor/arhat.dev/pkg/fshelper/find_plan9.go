//go:build plan9

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

	info, ok := f.(*syscall.Dir)
	if !ok {
		return false, fmt.Errorf("unexpected not *syscall.Stat_t: %T", f)
	}

	ops := opts.Ops

	if ops&FindOp_CheckUserInvalid != 0 {
		// TODO: currently os/user.LookupID is based on reading /etc/passwd when bulit without cgo
		//       doesn't work properly nor efficiently
	}

	if ops&needUserGroup != 0 {
		if ops&FindOp_CheckUser != 0 && info.Uid != opts.WindowsOrPlan9User {
			return false, nil
		}

		if ops&FindOp_CheckGroup != 0 && info.Gid != opts.WindowsOrPlan9Group {
			return false, nil
		}
	}

	if ops&needTime != 0 {
		// NOTE: there is no btime
		// if ops&FindOp_CheckCreationTime != 0 &&
		// 	(info.Birthtimespec.Sec < opts.MinCreationTime ||
		// 		info.Birthtimespec.Sec > opts.MaxCreationTime) {
		// 	return false, nil
		// }

		if ops&FindOp_CheckLastAccessTime != 0 &&
			(int64(info.Atime) < opts.MinAtime || int64(info.Atime) > opts.MaxAtime) {
			return false, nil
		}

		// NOTE: there is no ctime
		// if ops&FindOp_CheckLastMetadataChangeTime != 0 &&
		// 	(info.Ctimespec.Sec < opts.MinCtime ||
		// 		info.Ctimespec.Sec > opts.MaxCtime) {
		// 	return false, nil
		// }

		if ops&FindOp_CheckLastContentUpdatedTime != 0 &&
			(int64(info.Mtime) < opts.MinMtime || int64(info.Mtime) > opts.MaxMtime) {
			return false, nil
		}
	}

	return true, nil
}
