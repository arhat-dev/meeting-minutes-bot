//go:build !(windows || linux || openbsd || freebsd || darwin || dragonfly || solaris || aix || netbsd || plan9)

package fshelper

import "arhat.dev/pkg/wellknownerrors"

func (ofs *OSFS) matchFileSysinfo(opts *FindOptions, path string, f any) (ok bool, err error) {
	return false, wellknownerrors.ErrNotSupported
}
