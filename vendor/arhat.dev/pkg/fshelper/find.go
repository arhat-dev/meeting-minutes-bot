package fshelper

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unsafe"

	"arhat.dev/pkg/pathhelper"
	"github.com/bmatcuk/doublestar/v4"
)

type FindOptions struct {
	Ops FindOp

	FileType fs.FileMode
	Perm     fs.FileMode // TODO: currently not used

	MinDepth, MaxDepth int32
	MinSize, MaxSize   int64

	// unix timestamps (seconds since unix epoch)
	MinCreationTime, MaxCreationTime int64
	MinAtime, MaxAtime               int64
	MinCtime, MaxCtime               int64
	MinMtime, MaxMtime               int64

	Regexpr, RegexprIgnoreCase *regexp.Regexp

	UnixUID, UnixGID uint32
	// windows SID of user & group or plan9 user & group name
	WindowsOrPlan9User, WindowsOrPlan9Group string

	PathPattern,
	// lower case pattern to match name ignore case (MUST be in lower case to be effective)
	PathPatternLower string

	NamePattern,
	NamePatternFollowSymlink,
	// lower case pattern to match name ignore case (MUST be in lower case to be effective)
	NamePatternLower,
	// lower case pattern to match name ignore case (MUST be in lower case to be effective)
	NamePatternLowerFollowSymlink string

	// Run on each matched path
	OnMatch func(path string, d fs.DirEntry)
}

// FindOp is the bitset of find operations
type FindOp uint32

const (
	FindOp_CheckDepth FindOp = 1 << iota

	// set FindOptions.FileType to filter
	FindOp_CheckTypeNotFile // not regular file (e.g. dir, symlink)

	// FindOptions.FileType is ignored, and file should report is regular file
	FindOp_CheckTypeIsFile // is regular file

	FindOp_CheckPerm // TODO: currently not implementd

	// set both FindOptions.{Min, Max}Size to filter
	FindOp_CheckSize

	FindOp_CheckUserInvalid // no unix implementation

	// set FindOptions.UnixUID or FindOptions.WindowsOrPlan9User to filter
	FindOp_CheckUser

	// set FindOptions.UnixGID or FindOptions.WindowsOrPlan9Group to filter
	FindOp_CheckGroup

	// set both FindOptions.{Min, Max}CreationTime to filter
	FindOp_CheckCreationTime // btime (birth time)

	// set both FindOptions.{Min, Max}Ctime to filter
	FindOp_CheckLastMetadataChangeTime // ctime

	// set both FindOptions.{Min, Max}Atime to filter
	FindOp_CheckLastAccessTime // atime

	// set both FindOptions.{Min, Max}Mtime to filter
	FindOp_CheckLastContentUpdatedTime // mtime

	// set FindOptions.NamePattern to filter
	FindOp_CheckName

	// set FindOptions.NamePatternFollowSymlink to filter
	FindOp_CheckNameFollowSymlink

	// set FindOptions.NamePatternLower to filter
	FindOp_CheckNameIgnoreCase

	// set FindOptions.NamePatternLowerFollowSymlink to filter
	FindOp_CheckNameIgnoreCaseFollowSymlink

	// path matching

	// set FindOptions.PathPattern to filter
	FindOp_CheckPath

	// set FindOptions.PathPatternLower to filter
	FindOp_CheckPathIgnoreCase

	// set FindOptions.Regexpr to filter
	FindOp_CheckRegex

	// set FindOptions.RegexprIgnoreCase to filter
	FindOp_CheckRegexIgnoreCase
)

// pathDepth return depth of s for find
// it assumes there is no relative path element in middle (. or ..)
func pathDepth(s string) (n int32) {
	sz := len(s)
	if sz < 2 {
		return 0
	}

	lastIsPathSep := false
	for _, c := range s {
		if os.IsPathSeparator(byte(c)) {
			if lastIsPathSep {
				continue
			} else {
				lastIsPathSep = true
				n++
			}
		} else {
			lastIsPathSep = false
		}
	}

	return
}

// Find all matched files by walking from startpath
func (ofs *OSFS) Find(fopts *FindOptions, startpath string) (ret []string, err error) {
	var (
		_buffer   [256]byte
		buf       []byte
		slashPath string
	)

	buf = _buffer[:]
	checkDepth := fopts.Ops&FindOp_CheckDepth != 0
	minDepth, maxDepth := fopts.MinDepth, fopts.MaxDepth

	startpath = filepath.Clean(startpath)

	// depth note for fs.WalkDir, depth need to be adjusted according to the starting dir
	//
	// case 1:
	// . => all child entries needs depth + 1
	//
	// case 2:
	// foo (including ./foo) will form path like foo/x => depth + 0
	// foo/bar (including ./foo/bar) will form path like foo/bar/x => depth - 1
	// foo/bar/woo
	// ...... => depth - pathDepth(dir)
	//
	// case 3:
	// .. => depth + 0
	// ../foo => depth - 1
	// ...... => same as case 2

	depthAdd := -pathDepth(startpath)
	if startpath == "." {
		depthAdd = 1
	}

	onMatch := fopts.OnMatch
	if onMatch == nil {
		onMatch = func(path string, d fs.DirEntry) {}
	}

	notFirst := false

	err = fs.WalkDir(ofs, startpath, func(path string, ent fs.DirEntry, dirErr error) error {
		var (
			ok    bool
			depth int32
		)

		if dirErr != nil {
			// ignore
			if ent == nil {
				// errored at root, end this walk
				//
				// ref: comments of fs.WalkDirFunc
				return dirErr
			}

			return fs.SkipDir
		}

		// ent is not nil when dirErr is nil
		slashPath, buf = pathhelper.ToSlash(buf, path)

		if checkDepth {
			if notFirst {
				depth = pathDepth(slashPath) + depthAdd
			} else {
				depth = 0
				notFirst = true
			}

			if depth < minDepth || depth > maxDepth {
				return nil
			}
		}

		ok, dirErr = ofs.TryMatch(fopts, slashPath, ent)
		if ok {
			// matched, handle path buffering
			if slashPath != path {
				// slashPath inside buffer, make a copy
				pathBuf := make([]byte, len(slashPath))
				copy(pathBuf, slashPath)
				path = *(*string)(unsafe.Pointer(&pathBuf))
			}
			ret = append(ret, path)

			onMatch(path, ent)
		}

		// TODO: handle err?
		return nil
	})

	return
}

// TryMatch checks whether DirEntry d satisfies FindOptions
//
// path and d should match fs.WalkFunc definition
func (ofs *OSFS) TryMatch(opts *FindOptions, path string, d fs.DirEntry) (matched bool, err error) {
	const (
		mayNeedReadlink = FindOp_CheckNameFollowSymlink | FindOp_CheckNameIgnoreCaseFollowSymlink

		needType    = FindOp_CheckTypeNotFile | FindOp_CheckTypeIsFile | FindOp_CheckSize /* only check size when it's a regular file */ | mayNeedReadlink
		needSysinfo = FindOp_CheckUserInvalid | FindOp_CheckUser | FindOp_CheckGroup |
			FindOp_CheckCreationTime | FindOp_CheckLastMetadataChangeTime |
			FindOp_CheckLastAccessTime | FindOp_CheckLastContentUpdatedTime
		needInfo = FindOp_CheckSize | needSysinfo
		needName = FindOp_CheckName | FindOp_CheckNameIgnoreCase | FindOp_CheckNameFollowSymlink | FindOp_CheckNameIgnoreCaseFollowSymlink
		needPath = FindOp_CheckPath | FindOp_CheckPathIgnoreCase | FindOp_CheckRegex | FindOp_CheckRegexIgnoreCase
	)

	var (
		ops  = opts.Ops
		typ  fs.FileMode
		name string
		sz   int64
	)

	if ops&needType != 0 {
		typ = d.Type()

		if ops&FindOp_CheckTypeNotFile != 0 && typ&opts.FileType == 0 {
			return false, nil
		}

		if ops&FindOp_CheckTypeIsFile != 0 && !typ.IsRegular() {
			return false, nil
		}
	}

	if ops&needName != 0 {
		name = d.Name()

		if ops&FindOp_CheckName != 0 {
			matched, err = doublestar.Match(opts.NamePattern, name)
			if err != nil || !matched {
				return
			}
		}

		if ops&FindOp_CheckNameIgnoreCase != 0 {
			matched, err = doublestar.Match(opts.NamePatternLower, strings.ToLower(name))
			if err != nil || !matched {
				return
			}
		}
	}

	if ops&needPath != 0 {
		path = filepath.ToSlash(path)

		if ops&FindOp_CheckRegex != 0 {
			matched = opts.Regexpr.MatchString(path)
			if !matched {
				return
			}
		}

		if ops&FindOp_CheckRegexIgnoreCase != 0 {
			matched = opts.RegexprIgnoreCase.MatchString(path)
			if !matched {
				return
			}
		}

		if ops&FindOp_CheckPath != 0 {
			matched, err = doublestar.Match(opts.PathPattern, path)
			if err != nil || !matched {
				return
			}
		}

		if ops&FindOp_CheckPathIgnoreCase != 0 {
			matched, err = doublestar.Match(opts.PathPatternLower, strings.ToLower(path))
			if err != nil || !matched {
				return
			}
		}
	}

	if ops&mayNeedReadlink != 0 {
		if typ&fs.ModeSymlink != 0 {
			name, err = ofs.Readlink(path)
			if err != nil {
				return false, err
			}

			name = filepath.Base(name)
		}

		if ops&FindOp_CheckNameFollowSymlink != 0 {
			matched, err = doublestar.Match(opts.NamePatternFollowSymlink, name)
			if err != nil || !matched {
				return
			}
		}

		if ops&FindOp_CheckNameIgnoreCaseFollowSymlink != 0 {
			matched, err = doublestar.Match(opts.NamePatternLowerFollowSymlink, strings.ToLower(name))
			if err != nil || !matched {
				return
			}
		}
	}

	if ops&needInfo != 0 {
		var info fs.FileInfo

		info, err = d.Info()
		if err != nil {
			return false, err
		}

		if ops&FindOp_CheckSize != 0 {
			if !typ.IsRegular() {
				// not a regular file
				return false, nil
			}

			sz = info.Size()
			if sz < opts.MinSize || sz > opts.MaxSize {
				return false, nil
			}
		}

		if ops&needSysinfo != 0 {
			matched, err = ofs.matchFileSysinfo(opts, path, info.Sys())
			if err != nil || !matched {
				return
			}
		}
	}

	return true, nil
}
