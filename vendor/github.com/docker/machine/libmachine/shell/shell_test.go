package shell

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectBash(t *testing.T) {
	defer func(shell string) { os.Setenv("SHELL", shell) }(os.Getenv("SHELL"))
	os.Setenv("SHELL", "/bin/bash")

	shell, err := Detect()

	assert.Equal(t, "bash", shell)
	assert.NoError(t, err)
}

func TestDetectFish(t *testing.T) {
	defer func(shell string) { os.Setenv("SHELL", shell) }(os.Getenv("SHELL"))
	os.Setenv("SHELL", "/bin/bash")

	defer func(fishDir string) { os.Setenv("__fish_bin_dir", fishDir) }(os.Getenv("__fish_bin_dir"))
	os.Setenv("__fish_bin_dir", "/usr/local/Cellar/fish/2.2.0/bin")

	shell, err := Detect()

	assert.Equal(t, "fish", shell)
	assert.NoError(t, err)
}
