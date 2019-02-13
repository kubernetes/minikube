// Package exit contains functions useful for exiting gracefully.
package exit

import (
	"os"

	"github.com/golang/glog"
	"k8s.io/minikube/pkg/minikube/console"
)

// Exit codes based on sysexits(3)
const (
	Failure     = 1  // Failure represents a general failure code
	BadUsage    = 64 // Usage represents an incorrect command line
	Data        = 65 // Data represents incorrect data supplied by the user
	NoInput     = 66 // NoInput represents that the input file did not exist or was not readable
	Unavailable = 69 // Unavailable represents when a service was unavailable
	Software    = 70 // Software represents an internal software error.
	IO          = 74 // IO represents an I/O error
	Config      = 78 // Config represents an unconfigured or miscon­figured state
	Permissions = 77 // Permissions represents a permissions error
)

// Usage outputs a usage error and exits with error code 64
func Usage(format string, a ...interface{}) {
	console.OutStyle("usage", format, a...)
	os.Exit(BadUsage)
}

// WithCode outputs a fatal error message and exits with a supplied error code.
func WithCode(code int, format string, a ...interface{}) {
	// use Warning because Error will display a duplicate message to stderr
	glog.Warningf(format, a...)
	console.Fatal(format, a...)
	os.Exit(code)
}

// WithError outputs an error and exits.
func WithError(msg string, err error) {
	console.Fatal(msg+": %v", err)
	console.Out("\n")
	console.OutStyle("sad", "Sorry that minikube crashed. If this was unexpected, we would love to hear from you:")
	console.OutStyle("url", "https://github.com/kubernetes/minikube/issues/new")
	// use Warning because Error will display a duplicate message to stderr
	glog.Warningf(msg)
	// Here is where we would insert code to optionally upload a stack trace.

	// We can be smarter about guessing exit codes, but EX_SOFTWARE should suffice.
	os.Exit(Software)
}
