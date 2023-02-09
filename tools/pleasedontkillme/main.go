package main

import (
	"os"
	"os/signal"
	"syscall"
)

// This is used to unittest functions that kill processes,
// in a cross-platform way.
func main() {
	ch := make(chan os.Signal)
	done := make(chan struct{})
	defer close(ch)

	signal.Notify(ch, syscall.SIGKILL)
	defer signal.Stop(ch)

	go func() {
		<-ch
		close(done)
	}()

	<-done
}
