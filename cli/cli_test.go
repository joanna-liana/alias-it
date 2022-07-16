package cli_test

import (
	"alias_it/cli"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"
)

const ORIGINAL_LAST_LINE = "<add alias after this line>"

var HOME_DIR string
var ZSH_ALIAS_FILE_PATH string
var BASH_ALIAS_FILE_PATH string
var UNSUPPORTED_SHELL_ALIAS_FILE_PATH string

var TEST_SHELL_PATH = "/bin/zsh"
var TEST_SHELL_CONFIG_FILE = ".zshrc"

var homeDirResolver = func() (string, error) {
	return HOME_DIR, nil
}

var output bytes.Buffer

func TestCLI(t *testing.T) {
	supportedShellPathCases := []struct {
		shellName      string
		aliasPath      *string
		supportedPaths []string
	}{
		{
			shellName:      "ZSH",
			aliasPath:      &ZSH_ALIAS_FILE_PATH,
			supportedPaths: []string{"/bin/zsh", "/usr/bin/zsh"},
		},
		{
			shellName:      "bash",
			aliasPath:      &BASH_ALIAS_FILE_PATH,
			supportedPaths: []string{"/bin/bash", "/usr/bin/bash"},
		},
	}

	for _, tt := range supportedShellPathCases {
		for _, shellPath := range tt.supportedPaths {
			t.Run("Appends an alias to the shell config file - "+tt.shellName+"("+shellPath+")", func(t *testing.T) {

				// given
				TEST_SHELL_PATH = shellPath

				setUpEnv(t)
				setUpPaths(t)
				setValidCliArgs()

				createNonEmptyConfig(t)

				aliasCli := cli.New(&output, homeDirResolver)

				want := "alias testAliasName=\"echo $PWD\""

				// when
				aliasCli.Add()

				// then
				assertAppendedAlias(t, *tt.aliasPath, want)
			})
		}
	}

	t.Run("Does not add any alias if the shell is not ZSH", func(t *testing.T) {
		// given
		setValidCliArgs()

		setUpEnv(t)
		setUpPaths(t)
		createNonEmptyConfig(t)

		os.Setenv("SHELL", "")

		out, _ := exec.Command("echo", os.ExpandEnv("$SHELL")).Output()

		fmt.Println(string(out))

		aliasCli := cli.New(&output, homeDirResolver)

		want := "unsupported shell"

		// when
		aliasCli.Add()

		// then
		got := output.String()

		if !strings.Contains(got, want) {
			t.Errorf("Wanted\n%q\nto contain\n%q", got, want)
		}

		assertNoAliasAdded(t)
	})

	missingArgsCases := []struct {
		warning string
		cmdArgs []string
	}{
		{
			warning: "missing alias name",
			cmdArgs: []string{},
		},
		{
			warning: "missing command",
			cmdArgs: []string{"testAliasName"},
		},
	}

	for _, tt := range missingArgsCases {
		t.Run("Warns about missing args - "+tt.warning, func(t *testing.T) {
			// given
			os.Args = []string{"TEST_CMD"}
			os.Args = append(os.Args, tt.cmdArgs...)

			aliasCli := cli.New(&output, homeDirResolver)

			want := tt.warning

			// when
			aliasCli.Add()

			// then
			got := output.String()

			if !strings.Contains(got, want) {
				t.Errorf("Wanted\n%q\nto contain\n%q", got, want)
			}
		})

		t.Run("Does not add any alias if args are missing - "+tt.warning, func(t *testing.T) {
			// given
			setUpPaths(t)
			createNonEmptyConfig(t)

			os.Args = []string{"TEST_CMD"}
			os.Args = append(os.Args, tt.cmdArgs...)

			aliasCli := cli.New(&output, homeDirResolver)

			// when
			aliasCli.Add()

			// then
			assertNoAliasAdded(t)
		})
	}
}

func setValidCliArgs() {
	os.Args = []string{"TEST_CMD", "testAliasName", "echo $PWD"}
}

func setUpEnv(t *testing.T) {
	t.Helper()

	os.Setenv("SHELL", TEST_SHELL_PATH)
}

func setUpPaths(t *testing.T) {
	t.Helper()

	HOME_DIR = t.TempDir()
	ZSH_ALIAS_FILE_PATH = path.Join(HOME_DIR, ".zshrc")
	BASH_ALIAS_FILE_PATH = path.Join(HOME_DIR, ".bashrc")
	UNSUPPORTED_SHELL_ALIAS_FILE_PATH = path.Join(HOME_DIR, ".unsupportedrc")
}

func createNonEmptyConfig(t *testing.T) {
	t.Helper()

	for _, path := range []string{ZSH_ALIAS_FILE_PATH, BASH_ALIAS_FILE_PATH} {
		f, err := os.Create(path)

		if err != nil {
			t.Log("Could not create an empty temp config file: " + path)
			panic(err)
		}

		defer f.Close()

		if _, err := f.WriteString(ORIGINAL_LAST_LINE); err != nil {

			t.Log("Could not write to an empty temp config file: " + path)
			panic(err)
		}
	}
}

func assertAppendedAlias(t *testing.T, aliasPath string, aliasLine string) {
	t.Helper()

	assertAliasHasBeenSaved(t, aliasPath, aliasLine)
	assertFileSize(t, aliasPath, aliasLine)
}

func assertAliasHasBeenSaved(t *testing.T, aliasPath string, want string) {
	t.Helper()

	got := getLastLine(aliasPath, t)

	if want != got {
		t.Fatalf("Want %q, got %q", want, got)
	}
}

func assertNoAliasAdded(t *testing.T) {
	t.Helper()

	want := ORIGINAL_LAST_LINE
	got := getLastLine(ZSH_ALIAS_FILE_PATH, t)

	if want != got {
		t.Fatalf("Want %q, got %q", want, got)
	}
}

func assertFileSize(t *testing.T, aliasPath string, appendedLine string) {
	t.Helper()

	want := getFileSize(aliasPath, t)
	got := len(ORIGINAL_LAST_LINE) + len(appendedLine)

	if want <= got {
		t.Fatalf("Wanted the content of the config file to grow, not get replaced. Want: %v, got: %v", want, got)
	}
}

// based on https://stackoverflow.com/a/51328256/12938569
func getLastLine(filepath string, t *testing.T) string {
	t.Helper()

	fileHandle, err := os.Open(filepath)

	if err != nil {
		panic("Cannot open file " + filepath)
	}

	defer fileHandle.Close()

	line := ""
	var offset int64 = 0

	stat, err := fileHandle.Stat()

	if err != nil {
		panic(err)
	}

	filesize := stat.Size()

	for {
		offset -= 1
		fileHandle.Seek(offset, io.SeekEnd)

		char := make([]byte, 1)
		fileHandle.Read(char)

		if offset != -1 && (char[0] == 10 || char[0] == 13) { // stop if we find a line
			break
		}

		line = fmt.Sprintf("%s%s", string(char), line)

		if offset == -filesize { // stop if we are at the begining
			break
		}
	}

	return line
}

func getFileSize(filepath string, t *testing.T) int {
	t.Helper()

	stat, err := os.Stat(filepath)

	if err != nil {
		panic(err)
	}

	return int(stat.Size())
}
