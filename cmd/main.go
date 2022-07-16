package main

import (
	"alias_it/cli"
	"fmt"
	"os"
	"runtime"
	"strings"
)

func main() {
	if strings.HasPrefix(runtime.GOOS, "windows") {
		fmt.Println("Apologies, you cannot use this app on Windows. Try WSL!")

		os.Exit(1)
	}

	cli := cli.New(os.Stdout, os.UserHomeDir)

	cli.Add()
}
