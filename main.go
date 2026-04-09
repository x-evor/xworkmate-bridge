package main

import (
	"fmt"
	"os"

	"xworkmate-bridge/internal/acp"
	"xworkmate-bridge/internal/toolbridge"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "serve" {
		if err := acp.Serve(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "acp-stdio" {
		acp.RunStdio(os.Stdin, os.Stdout)
		return
	}

	toolbridge.Run(os.Stdin, os.Stdout)
}
