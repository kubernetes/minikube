// +build !windows

package shell

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnknownShell(t *testing.T) {
	defer func(shell string) { os.Setenv("SHELL", shell) }(os.Getenv("SHELL"))
	os.Setenv("SHELL", "")

	shell, err := Detect()

	assert.Equal(t, err, ErrUnknownShell)
	assert.Empty(t, shell)
}
