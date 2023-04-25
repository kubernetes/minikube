// +build !windows

package panicwrap

import (
	"github.com/bugsnag/osext"
	"os"
	"os/exec"
)

func monitor(c *WrapConfig) (int, error) {

	// If we're the child process, absorb panics.
	if Wrapped(c) {
		panicCh := make(chan string)

		go trackPanic(os.Stdin, os.Stderr, c.DetectDuration, panicCh)

		// Wait on the panic data
		panicTxt := <-panicCh
		if panicTxt != "" {
			if !c.HidePanic {
				os.Stderr.Write([]byte(panicTxt))
			}

			c.Handler(panicTxt)
		}

		os.Exit(0)
	}

	exePath, err := osext.Executable()
	if err != nil {
		return -1, err
	}
	cmd := exec.Command(exePath, os.Args[1:]...)

	read, write, err := os.Pipe()
	if err != nil {
		return -1, err
	}

	cmd.Stdin = read
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), c.CookieKey+"="+c.CookieValue)

	if err != nil {
		return -1, err
	}
	err = cmd.Start()
	if err != nil {
		return -1, err
	}

	err = dup2(int(write.Fd()), int(os.Stderr.Fd()))
	if err != nil {
		return -1, err
	}

	return -1, nil
}
