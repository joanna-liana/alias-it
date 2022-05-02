package main

import (
	"alias_it/cli"
	"os"
)

func main() {
	cli := cli.New(os.Stdout, os.UserHomeDir)

	cli.Add()
}
