package constants

import (
	"os"
	"path/filepath"
)

// MachineName is the name to use for the VM.
const MachineName = "minikubeVM"

// Fix for windows
var Minipath = filepath.Join(os.Getenv("HOME"), ".minikube")

// MakeMiniPath is a utility to calculate a relative path to our directory.
func MakeMiniPath(fileName string) string {
	return filepath.Join(Minipath, fileName)
}
