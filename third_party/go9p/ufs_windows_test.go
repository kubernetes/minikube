package go9p

import (
	"fmt"
	"os"
	"testing"
)

func assert_NotNil(t *testing.T, i interface {}) {
  if (i == nil) {
    t.Error("Value should not be nil")
  }
}

func assert_NotEqual(t *testing.T, lhs uint32, rhs uint32, message string) {
  if (lhs == rhs) {
    t.Errorf("Value %d should not be %d. %s", lhs, rhs, message)
  }
}

func TestDir2DirTimestamp(t *testing.T) {
	fi, err := os.Stat(".")
	if err != nil {
		t.Error(err)
	}
	var st *Dir
	st, _ = dir2Dir(".", fi, false, nil)
	assert_NotNil(t, st)

	if testing.Verbose() {
		fmt.Printf("%s %d %d\n", st.Name, st.Mtime, st.Atime)
	}

	assert_NotEqual(t, st.Mtime, uint32(0), "Mtime should be set")
	assert_NotEqual(t, st.Atime, uint32(0), "Atime should be set")
}

func TestDir2DirTimestampDotu(t *testing.T) {
	fi, err := os.Stat(".")
	if err != nil {
		t.Error(err)
	}
	var st *Dir
	st, _ = dir2Dir(".", fi, true, nil)
	assert_NotNil(t, st)

	if testing.Verbose() {
		fmt.Printf("%s %d %d\n", st.Name, st.Mtime, st.Atime)
	}

	assert_NotEqual(t, st.Mtime, uint32(0), "Mtime should be set")
	assert_NotEqual(t, st.Atime, uint32(0), "Atime should be set")
}
