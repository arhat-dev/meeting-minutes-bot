package versionhelper

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

type versionOptions struct {
	branch    bool
	commit    bool
	tag       bool
	arch      bool
	goVersion bool
	buildTime bool

	goCompilerPlatform bool
}

func NewVersionCmd(output io.Writer) *cobra.Command {
	opt := new(versionOptions)
	versionCmd := &cobra.Command{
		Use:           "version",
		Short:         "Print version info",
		SilenceErrors: true,
		SilenceUsage:  true,
		Run: func(cmd *cobra.Command, args []string) {
			show := func(s string) {
				_, _ = fmt.Fprint(output, s)
			}

			switch {
			case opt.branch:
				show(Branch())
			case opt.commit:
				show(Commit())
			case opt.tag:
				show(Tag())
			case opt.arch:
				show(Arch())
			case opt.goVersion:
				show(GoVersion())
			case opt.buildTime:
				show(buildTime)
			case opt.goCompilerPlatform:
				show(GoCompilerPlatform())
			default:
				show(Version())
			}
		},
	}

	versionFlags := versionCmd.Flags()
	versionFlags.BoolVar(&opt.branch, "branch", false, "get branch name")
	versionFlags.BoolVar(&opt.commit, "commit", false, "get commit hash")
	versionFlags.BoolVar(&opt.tag, "tag", false, "get tag name")
	versionFlags.BoolVar(&opt.arch, "arch", false, "get arch")
	versionFlags.BoolVar(&opt.buildTime, "build-time", false, "get build time (rfc3339 format)")
	versionFlags.BoolVar(&opt.goVersion, "go-version", false, "get go runtime/compiler version")
	versionFlags.BoolVar(&opt.goCompilerPlatform, "go-compilerPlatform", false, "get os/arch pair of the go compiler")

	return versionCmd
}
