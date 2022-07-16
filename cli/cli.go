package cli

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"
)

type HomeDirResolver = func() (string, error)
type AliasCLI struct {
	homeDirResolver HomeDirResolver
	printer         io.Writer
}

var ErrUnsupportedShell = errors.New("unsupported shell, only ZSH is supported for now")

type ShellName int64

const (
	Zsh ShellName = iota + 1
	Bash
)

var shellConfigFile = map[ShellName]string{
	Bash: ".bashrc",
	Zsh:  ".zshrc",
}

var shellSuffixes = map[ShellName]string{
	Bash: "/bash",
	Zsh:  "/zsh",
}

func (cli AliasCLI) Add() {
	command, aliasName, err := cli.parseArgs()

	if err != nil {
		if err.Error() == "insufficient args" {
			return
		}

		panic(err)
	}

	shellName, err := cli.ensureSupportedShell()

	if err != nil {
		if errors.Is(err, ErrUnsupportedShell) {
			return
		}

		panic(err)
	}

	shellConfigPath := cli.getShellConfigPath(cli.homeDirResolver, shellName)

	cli.addAlias(aliasName, command, shellConfigPath)
}

func (cli AliasCLI) ensureSupportedShell() (ShellName, error) {
	shellPath, _ := exec.Command("echo", os.ExpandEnv("$SHELL")).Output()

	var usedShellName ShellName

	for shellName, suffix := range shellSuffixes {
		isUsingSupportedShell := bytes.HasSuffix(shellPath, []byte(suffix+"\n"))

		if !isUsingSupportedShell {
			continue
		}

		usedShellName = shellName
	}

	if usedShellName == 0 {
		err := ErrUnsupportedShell

		cli.println("Error:\t", err.Error())

		return usedShellName, err
	}

	return usedShellName, nil
}

func New(printer io.Writer, homeDirResolver HomeDirResolver) *AliasCLI {

	return &AliasCLI{
		homeDirResolver: homeDirResolver,
		printer:         printer,
	}
}

func (cli AliasCLI) parseArgs() (string, string, error) {
	if len(os.Args) < 3 {
		var errorString string

		if len(os.Args) < 2 {
			errorString = "missing alias name"
		} else {
			errorString = "missing command"
		}

		cli.println("Error:\t", errorString, "\nUsage:\t alias-it <alias_name> <command_name>")

		return "", "", errors.New("insufficient args")
	}

	aliasName := os.Args[1]
	commandParts := os.Args[2:]
	command := strings.Join(commandParts, " ")

	cli.printf("🏡 Alias name:\t%s\n", aliasName)
	cli.printf("💻 Command:\t%s\n", command)

	return command, aliasName, nil
}

func (cli AliasCLI) getShellConfigPath(homeDirResolver HomeDirResolver, shell ShellName) string {
	homeDir, err := homeDirResolver()

	if err != nil {
		cli.println(err)
		panic(err)
	}

	shellConfigPath := path.Join(homeDir, shellConfigFile[shell])

	return shellConfigPath
}

func (cli AliasCLI) appendToShellConfig(shellConfigPath string, toAppend string) {
	f, err := os.OpenFile(shellConfigPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		cli.println(err)
	}

	defer f.Close()

	if _, err := f.WriteString(toAppend); err != nil {
		cli.println(err)
	}
}

func (cli AliasCLI) addAlias(name, command, shellConfigPath string) {
	aliasString := "\nalias " + name + "=\"" + command + "\""

	cli.appendToShellConfig(shellConfigPath, aliasString)

	cli.printf("\nAdded alias:%v\n\n", aliasString)
	cli.println("Hint: to prevent variable expansion, remember about prepending $ with a slash, e.g. $PWD -> \\$PWD")
	cli.println("\n👉 Remember to run \"source ~/.zshrc\" or open a new terminal tab to start using your alias")
}

func (cli AliasCLI) printf(format string, toPrint ...any) {
	fmt.Fprintf(cli.printer, format, toPrint...)
}

func (cli AliasCLI) println(toPrint ...any) {
	fmt.Fprintln(cli.printer, toPrint...)
}
