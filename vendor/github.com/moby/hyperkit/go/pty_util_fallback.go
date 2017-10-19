// +build !darwin

package hyperkit

import (
	"log"
	"os"
)

func saneTerminal(f *os.File) error {
	log.Fatal("Function not supported on your OS")
	return nil
}

func setRaw(f *os.File) error {
	log.Fatal("Function not supported on your OS")
	return nil
}

// isTerminal checks if the provided file is a terminal
func isTerminal(f *os.File) bool {
	log.Fatal("Function not supported on your OS")
	return false
}
