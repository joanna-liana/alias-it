package main

import (
	"alias_it/cli"
	"os"
)

func main() {
	cli := cli.New(os.UserHomeDir)

	cli.Add()
}
