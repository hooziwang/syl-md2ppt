package main

import (
	"fmt"
	"os"

	"syl-md2ppt/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		if msg := cmd.FriendlyError(err); msg != "" {
			fmt.Fprintln(os.Stderr, msg)
		}
		os.Exit(1)
	}
}
