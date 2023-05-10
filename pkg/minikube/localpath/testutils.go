package localpath

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// FakeFile satisfies fdWriter
type FakeFile struct {
	b bytes.Buffer
}

// NewFakeFile creates a FakeFile
func NewFakeFile() *FakeFile {
	return &FakeFile{}
}

// Fd returns the file descriptor
func (f *FakeFile) Fd() uintptr {
	return uintptr(0)
}

func (f *FakeFile) Write(p []byte) (int, error) {
	return f.b.Write(p)
}
func (f *FakeFile) String() string {
	return f.b.String()
}

// MakeTempDir creates the temp dir and returns the path
func MakeTempDir(t *testing.T) string {
	tempDir := t.TempDir()
	tempDir = filepath.Join(tempDir, ".minikube")
	if err := os.MkdirAll(filepath.Join(tempDir, "addons"), 0777); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(tempDir, "cache"), 0777); err != nil {
		t.Fatal(err)
	}
	os.Setenv(MinikubeHome, tempDir)
	return MiniPath()
}
