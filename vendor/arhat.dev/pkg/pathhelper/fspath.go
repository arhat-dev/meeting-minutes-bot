package pathhelper

// ConvertFSPathToWindowsPath converts paths like `/c/foo` to absolute windows path `c:/foo`
// defaultVolumeName is used to resolve path as driver relative path, it's usually obtained by
// calling filepath.VolumeName, and its value is like `c:`, `\\some-host\path`
func ConvertFSPathToWindowsPath(defaultVolumeName, path string) string {
	if len(path) == 0 || path[0] != '/' {
		// path relative path (e.g. foo)
		// or absolute path (e.g. \\foo, c:/foo)
		return CleanWindowsPath(path)
	}

	// absolute fs path
	// now we get path starts with '/'

	if len(path) == 1 {
		// is `/`, drive relative path, windows can handle this
		if len(defaultVolumeName) == 0 {
			return CleanWindowsPath(path)
		}

		if isWindowsDiskVolumeName(defaultVolumeName) && !IsWindowsSlash(defaultVolumeName[len(defaultVolumeName)-1]) {
			defaultVolumeName += string(WindowsSlash)
		}

		return JoinWindowsPath(defaultVolumeName, path)
	}

	// when second character is [a-zA-Z], the path is a absolute
	// windows path when the third character is '/'

	if !IsWindowsDiskDesignatorChar(path[1]) || len(path) == 2 || path[2] != '/' {
		if len(defaultVolumeName) == 0 {
			return CleanWindowsPath(path)
		}

		if isWindowsDiskVolumeName(defaultVolumeName) && !IsWindowsSlash(defaultVolumeName[len(defaultVolumeName)-1]) {
			defaultVolumeName += string(WindowsSlash)
		}

		return JoinWindowsPath(defaultVolumeName, path[1:])
	}

	return JoinWindowsPath(string([]byte{path[1], ':', '\\'}), path[3:])
}

// isWindowsDiskVolumeName returns true when s is `[a-zA-Z]:`
func isWindowsDiskVolumeName(s string) bool {
	if len(s) != 2 {
		return false
	}

	return s[1] == ':' && IsWindowsDiskDesignatorChar(s[0])
}
