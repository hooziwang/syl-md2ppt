package cmd

import (
	"fmt"
	"io"
)

var (
	Version   = "dev"
	Commit    = "none"
	BuildTime = "unknown"
)

func versionText() string {
	return fmt.Sprintf("syl-md2ppt 版本：%s（commit: %s，构建时间: %s）", Version, Commit, BuildTime)
}

func printVersion(w io.Writer) {
	fmt.Fprintln(w, versionText())
}
