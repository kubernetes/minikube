package getter

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestTar(t *testing.T) {
	mtime := time.Unix(0, 0)
	cases := []TestDecompressCase{
		{
			"extended_header.tar",
			true,
			false,
			[]string{"directory/", "directory/a", "directory/b"},
			"",
			nil,
		},
		{
			"implied_dir.tar",
			true,
			false,
			[]string{"directory/", "directory/sub/", "directory/sub/a", "directory/sub/b"},
			"",
			nil,
		},
		{
			"unix_time_0.tar",
			true,
			false,
			[]string{"directory/", "directory/sub/", "directory/sub/a", "directory/sub/b"},
			"",
			&mtime,
		},
	}

	for i, tc := range cases {
		cases[i].Input = filepath.Join("./testdata", "decompress-tar", tc.Input)
	}

	TestDecompressor(t, new(TarDecompressor), cases)
}

// testDecompressPermissions decompresses a directory and checks the permissions of the expanded files
func testDecompressorPermissions(t *testing.T, d Decompressor, input string, expected map[string]int, umask os.FileMode) {
	td, err := ioutil.TempDir("", "getter")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Destination is always joining result so that we have a new path
	dst := filepath.Join(td, "subdir", "result")

	err = d.Decompress(dst, input, true, umask)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	defer os.RemoveAll(dst)

	for name, mode := range expected {
		fi, err := os.Stat(filepath.Join(dst, name))
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		real := fi.Mode()
		if real != os.FileMode(mode) {
			t.Fatalf("err: %s expected mode %o got %o", name, mode, real)
		}
	}
}

func TestDecompressTarPermissions(t *testing.T) {
	d := new(TarDecompressor)
	input := "./test-fixtures/decompress-tar/permissions.tar"

	var expected map[string]int
	var masked int

	if runtime.GOOS == "windows" {
		expected = map[string]int{
			"directory/public":  0666,
			"directory/private": 0666,
			"directory/exec":    0666,
			"directory/setuid":  0666,
		}
		masked = 0666
	} else {
		expected = map[string]int{
			"directory/public":  0666,
			"directory/private": 0600,
			"directory/exec":    0755,
			"directory/setuid":  040000755,
		}
		masked = 0755
	}

	testDecompressorPermissions(t, d, input, expected, os.FileMode(0))

	expected["directory/setuid"] = masked
	testDecompressorPermissions(t, d, input, expected, os.FileMode(060000000))
}
