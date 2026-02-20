package main

import (
	"fmt"
	"os"

	"syl-md2ppt/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, cmd.FriendlyError(err))
		os.Exit(1)
	}
}
