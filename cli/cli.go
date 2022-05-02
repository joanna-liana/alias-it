package cli

import (
	"fmt"
	"os"
	"path"
	"strings"
)

type HomeDirResolver = func() (string, error)
type AliasCLI struct {
	homeDirResolver HomeDirResolver
}

func (cli AliasCLI) Add() {
	command, aliasName := parseArgs()

	shellConfigPath := getShellConfigPath(cli.homeDirResolver)

	addAlias(aliasName, command, shellConfigPath)
}

func New(homeDirResolver HomeDirResolver) *AliasCLI {

	return &AliasCLI{
		homeDirResolver: homeDirResolver,
	}
}

func parseArgs() (string, string) {
	aliasName := os.Args[1]
	commandParts := os.Args[2:]
	command := strings.Join(commandParts, " ")

	fmt.Printf("🏡 Alias name:\t%s\n", aliasName)
	fmt.Printf("💻 Command:\t%s\n", command)

	return command, aliasName
}

func getShellConfigPath(homeDirResolver HomeDirResolver) string {
	homeDir, err := homeDirResolver()

	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	shellConfigPath := path.Join(homeDir, ".zshrc")

	return shellConfigPath
}

func appendToShellConfig(shellConfigPath string, toAppend string) {
	f, err := os.OpenFile(shellConfigPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		fmt.Println(err)
	}

	defer f.Close()

	if _, err := f.WriteString(toAppend); err != nil {
		fmt.Println(err)
	}
}

func addAlias(name, command, shellConfigPath string) {
	aliasString := "\nalias " + name + "=\"" + command + "\""

	appendToShellConfig(shellConfigPath, aliasString)

	fmt.Printf("\nAdded alias:%v\n\n", aliasString)
	fmt.Println("Hint: to prevent variable expansion, remember about prepending $ with a slash, e.g. $PWD -> \\$PWD")
	fmt.Println("\n👉 Remember to run \"source ~/.zshrc\" or open a new terminal tab to start using your alias")
}
