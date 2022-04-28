package cli_test

import (
	"alias_it/cli"
	"fmt"
	"io"
	"os"
	"path"
	"testing"
)

var HOME_DIR = "test_home"

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

		aliasFilePath := path.Join(HOME_DIR, ".zshrc")

		f, err := os.Create(aliasFilePath)

		if err != nil {
			panic(err)
		}

		defer f.Close()

		if _, err := f.WriteString("lorem\nipsum"); err != nil {
			panic(err)
		}

		homeDirResolver := func() (string, error) {
			return HOME_DIR, nil
		}

		aliasCli := cli.New(homeDirResolver)

		want := "alias testAliasName=\"echo $PWD\""

		// when
		aliasCli.Add()

		// then
		got := getLastLine(aliasFilePath, t)

		if want != got {
			t.Errorf("Want %q, got %q", want, got)
		}
	})
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
