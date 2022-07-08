package versionhelper

import (
	"runtime"
	"strings"
	"time"
)

// values should be set at build time using ldflags
// -X arhat.dev/pkg/versionhelper.{branch,commit, ...}=...
var (
	branch, commit, tag, arch string

	worktreeClean string
	buildTime     string

	goCompilerPlatform string
)

func Version() string {
	var sb strings.Builder
	sb.WriteString("branch: ")
	sb.WriteString(Branch())

	sb.WriteString("\ncommit: ")
	sb.WriteString(Commit())

	sb.WriteString("\ntag: ")
	sb.WriteString(Tag())

	sb.WriteString("\narch: ")
	sb.WriteString(Arch())

	sb.WriteString("\ngoVersion: ")
	sb.WriteString(GoVersion())

	sb.WriteString("\nbuildTime: ")
	sb.WriteString(buildTime)

	sb.WriteString("\nworkTreeClean: ")
	sb.WriteString(worktreeClean)

	sb.WriteString("\ngoCompilerPlatform: ")
	sb.WriteString(GoCompilerPlatform())
	sb.WriteString("\n")

	return sb.String()
}

// Branch name of the source code
func Branch() string {
	return branch
}

// Commit hash of the source code
func Commit() string {
	return commit
}

// Tag is the tag name from the VCS of source code
func Tag() string {
	return tag
}

// Arch returns cpu arch set at build time
// usually contains more detain than runtime.GOARCH, such as armv7 instead of arm
func Arch() string {
	return arch
}

func GoVersion() string {
	return runtime.Version()
}

func BuildTime() time.Time {
	ret, err := time.Parse(time.RFC3339, buildTime)
	if err != nil {
		return time.Time{}
	}

	return ret
}

func WorktreeClean() bool {
	return worktreeClean == "true"
}

func GoCompilerPlatform() string {
	return goCompilerPlatform
}
