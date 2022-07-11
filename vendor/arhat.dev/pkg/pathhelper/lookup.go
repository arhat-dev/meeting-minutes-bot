package pathhelper

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func lookPathDir(cwd, target, pathListEnv, pathExtEnv string, find findAny) (string, error) {
	if find == nil {
		panic("no find function found")
	}

	chars := `/`
	if runtime.GOOS == "windows" {
		chars = `:\/`
	}
	exts := pathExts(pathExtEnv)

	if strings.ContainsAny(target, chars) {
		return find(cwd, target, exts)
	}

	pathList := splitPathList(cwd, target, pathListEnv, pathExtEnv)

	for _, elem := range pathList {
		var path string
		switch elem {
		case "", ".":
			// otherwise "foo" won't be "./foo"
			path = "." + string(filepath.Separator) + target
		default:
			path = filepath.Join(elem, target)
		}

		if f, err := find(cwd, path, exts); err == nil {
			return f, nil
		}
	}

	return "", fmt.Errorf("%q: executable file not found in $PATH", target)
}

// splitPathList normalize PATH list as absolute paths
// both ; and : are treated as path separator
func splitPathList(cwd, target, pathListEnv, pathExtEnv string) []string {
	isWindows := runtime.GOOS == "windows"

	isSlash := IsUnixSlash
	if isWindows {
		isSlash = IsWindowsSlash
	}

	// split semi-colon in addition to colon on windows
	list := SplitList(pathListEnv, true, isWindows, isSlash)

	for i, v := range list {
		if filepath.IsAbs(v) {
			continue
		}

		if !isWindows {
			list[i] = JoinUnixPath(cwd, v)
			continue
		}

		const (
			cygpathBin = "cygpath"
		)

		if target == cygpathBin {
			list[i], _ = AbsWindowsPath(cwd, v, func(path string) (string, error) {
				return "", nil
			})
			continue
		}

		var err error
		list[i], err = AbsWindowsPath(cwd, v, func(path string) (string, error) {
			// find root path of the fhs root using cygpath
			// but first lookup cygpath itself
			cygpath, err := lookPathDir(cwd, cygpathBin, pathListEnv, pathExtEnv, findExecutable)
			if err != nil {
				return "", err
			}

			// NOTE for some environments:
			// 	 github action windows:
			//		there is `cygpath` at C:\Program Files\Git\usr\bin\cygpath
			// 		when running inside gitbash (set `shell: bash`)

			output, err2 := exec.Command(cygpath, "-w", path).CombinedOutput()
			if err2 == nil {
				return strings.TrimSpace(string(output)), nil
			}

			switch {
			case os.Getenv("GITHUB_ACTIONS") == "true":
				// github action has msys2 installed without PATH added
				return `C:\msys64`, nil
			default:
				// TODO: other defaults?
				return "", err2
			}
		})

		// error can only happen when looking up fhs root
		_ = err
	}

	return list
}

// Following code copied from mvdan.cc/sh/v3

// findAny defines a function to pass to lookPathDir.
type findAny = func(dir string, file string, exts []string) (string, error)

func pathExts(pathExtEnv string) []string {
	if runtime.GOOS != "windows" {
		return nil
	}

	if len(pathExtEnv) == 0 {
		// include ""
		return []string{".com", ".exe", ".bat", ".cmd", ""}
	}

	var exts []string
	for _, e := range strings.Split(strings.ToLower(pathExtEnv), `;`) {
		if e == "" {
			continue
		}
		if e[0] != '.' {
			e = "." + e
		}
		exts = append(exts, e)
	}

	// allow no extension at last
	return append(exts, "")
}

func checkStat(dir, file string, checkExec bool) (string, error) {
	if !filepath.IsAbs(file) {
		file = filepath.Join(dir, file)
	}
	info, err := os.Stat(file)
	if err != nil {
		return "", err
	}
	m := info.Mode()
	if m.IsDir() {
		return "", fmt.Errorf("is a directory")
	}
	if checkExec && runtime.GOOS != "windows" && m&0o111 == 0 {
		return "", fmt.Errorf("permission denied")
	}
	return file, nil
}

func winHasExt(file string) bool {
	i := strings.LastIndex(file, ".")
	if i < 0 {
		return false
	}
	return strings.LastIndexAny(file, `:\/`) < i
}

// findExecutable returns the path to an existing executable file.
func findExecutable(dir, file string, exts []string) (string, error) {
	if len(exts) == 0 {
		// non-windows
		return checkStat(dir, file, true)
	}
	if winHasExt(file) {
		if file, err := checkStat(dir, file, true); err == nil {
			return file, nil
		}
	}
	for _, e := range exts {
		f := file + e
		if f, err := checkStat(dir, f, true); err == nil {
			return f, nil
		}
	}
	return "", fmt.Errorf("not found")
}

// findFile returns the path to an existing file.
func findFile(dir, file string, _ []string) (string, error) {
	return checkStat(dir, file, false)
}
