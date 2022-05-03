package cli_test

import (
	"alias_it/cli"
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"testing"
)

const HOME_DIR = "test_home"
const ORIGINAL_LAST_LINE = "<add alias after this line>"

var homeDirResolver = func() (string, error) {
	return HOME_DIR, nil
}

var ALIAS_FILE_PATH = path.Join(HOME_DIR, ".zshrc")

var output bytes.Buffer

func TestMain(m *testing.M) {
	setUp()

	// exec test and this returns an exit code to pass to os
	exitCode := m.Run()

	os.Exit(exitCode)
}

func setUp() {
	os.RemoveAll(HOME_DIR)
	os.Mkdir(HOME_DIR, 0777)
}

func TestCLI(t *testing.T) {
	t.Run("Appends an alias to the shell config file", func(t *testing.T) {
		// given
		os.Args = []string{"TEST_CMD", "testAliasName", "echo $PWD"}

		createNonEmptyConfig(t)

		aliasCli := cli.New(&output, homeDirResolver)

		want := "alias testAliasName=\"echo $PWD\""

		// when
		aliasCli.Add()

		// then
		assertAppendedAlias(want, t)
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
			createNonEmptyConfig(t)

			os.Args = []string{"TEST_CMD"}
			os.Args = append(os.Args, tt.cmdArgs...)

			aliasCli := cli.New(&output, homeDirResolver)

			want := ORIGINAL_LAST_LINE

			// when
			aliasCli.Add()

			// then
			got := getLastLine(ALIAS_FILE_PATH, t)

			if want != got {
				t.Errorf("Want %q, got %q", want, got)
			}
		})
	}
}

func createNonEmptyConfig(t *testing.T) {
	t.Helper()

	f, err := os.Create(ALIAS_FILE_PATH)

	if err != nil {
		panic(err)
	}

	defer f.Close()

	if _, err := f.WriteString(ORIGINAL_LAST_LINE); err != nil {
		panic(err)
	}
}

func assertAppendedAlias(aliasLine string, t *testing.T) {
	t.Helper()

	assertFileSize(aliasLine, t)
	assertAliasHasBeenSaved(aliasLine, t)

}

func assertAliasHasBeenSaved(want string, t *testing.T) {
	got := getLastLine(ALIAS_FILE_PATH, t)

	if want != got {
		t.Errorf("Want %q, got %q", want, got)
	}
}

func assertFileSize(appendedLine string, t *testing.T) {
	t.Helper()

	want := getFileSize(ALIAS_FILE_PATH, t)
	got := len(ORIGINAL_LAST_LINE) + len(appendedLine)

	if want <= got {
		t.Fatal("Wanted the content of the config file to grow, not get replaced")
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
