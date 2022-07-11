package fshelper

import (
	"syscall"
	"unsafe"
)

type statx_t struct {
	stx_mask            uint32
	stx_blksize         uint32
	stx_attributes      uint64
	stx_nlink           uint32
	stx_uid             uint32
	stx_gid             uint32
	stx_mode            uint32
	stx_ino             uint64
	stx_size            uint64
	stx_blocks          uint64
	stx_attributes_mask uint64
	stx_atime           syscall.Timespec
	stx_btime           syscall.Timespec // creation time
	stx_ctime           syscall.Timespec
	stx_mtime           syscall.Timespec
	stx_rdev            uint64
	stx_dev             uint64
	spare               [14]uint64
}

func fstat_statx(fd int, path string, stat *statx_t, flags int) (err error) {
	var _p0 *byte
	_p0, err = syscall.BytePtrFromString(path)
	if err != nil {
		return
	}

	_, _, e1 := syscall.Syscall6(SYS_Statx, uintptr(fd), uintptr(unsafe.Pointer(_p0)), uintptr(unsafe.Pointer(stat)), uintptr(flags), 0, 0)
	if e1 != 0 {
		err = e1
	}
	return
}
