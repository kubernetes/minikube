package constants

import (
	"os"
	"path/filepath"
)

const MachineName = "minikubeVM"

// Fix for windows
var Minipath = filepath.Join(os.Getenv("HOME"), "minikube")

func MakeMiniPath(fileName string) string {
	return filepath.Join(Minipath, fileName)
}
