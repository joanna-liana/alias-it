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
var UNSUPPORTED_SHELL_ALIAS_FILE_PATH string

var homeDirResolver = func() (string, error) {
	return HOME_DIR, nil
}

var output bytes.Buffer

func TestCLI(t *testing.T) {
	supportedShellPaths := []string{
		"/bin/zsh",
		"/usr/bin/zsh",
	}

	for _, shellPath := range supportedShellPaths {
		t.Run("Appends an alias to the shell config file (ZSH)", func(t *testing.T) {
			// given
			os.Setenv("SHELL", shellPath)
			setUpPaths(t)
			setValidCliArgs()

			createNonEmptyConfig(t)

			aliasCli := cli.New(&output, homeDirResolver)

			want := "alias testAliasName=\"echo $PWD\""

			// when
			aliasCli.Add()

			// then
			assertAppendedAlias(want, t)
		})
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

	os.Setenv("SHELL", "/bin/zsh")
}

func setUpPaths(t *testing.T) {
	t.Helper()

	HOME_DIR = t.TempDir()
	ZSH_ALIAS_FILE_PATH = path.Join(HOME_DIR, ".zshrc")
	UNSUPPORTED_SHELL_ALIAS_FILE_PATH = path.Join(HOME_DIR, ".unsupportedrc")

	t.Log("ALIAS_FILE_PATH", ZSH_ALIAS_FILE_PATH)
}

func createNonEmptyConfig(t *testing.T) {
	t.Helper()

	f, err := os.Create(ZSH_ALIAS_FILE_PATH)

	if err != nil {
		t.Log("Could not create an empty temp config file")
		panic(err)
	}

	defer f.Close()

	if _, err := f.WriteString(ORIGINAL_LAST_LINE); err != nil {

		t.Log("Could not write to an empty temp config file")
		panic(err)
	}
}

func assertAppendedAlias(aliasLine string, t *testing.T) {
	t.Helper()

	assertAliasHasBeenSaved(aliasLine, t)
	assertFileSize(aliasLine, t)
}

func assertAliasHasBeenSaved(want string, t *testing.T) {
	t.Helper()

	got := getLastLine(ZSH_ALIAS_FILE_PATH, t)

	fmt.Println("shell", os.Getenv("SHELL"))
	fmt.Println("last line", got)
	fmt.Println("want line", want)

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

func assertFileSize(appendedLine string, t *testing.T) {
	t.Helper()

	want := getFileSize(ZSH_ALIAS_FILE_PATH, t)
	got := len(ORIGINAL_LAST_LINE) + len(appendedLine)

	fmt.Println("ZSH_ALIAS_FILE_PATH", ZSH_ALIAS_FILE_PATH)
	fmt.Println("len(ORIGINAL_LAST_LINE)", len(ORIGINAL_LAST_LINE))
	fmt.Println("want", want)
	fmt.Println("got", got)

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
