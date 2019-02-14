/*
Copyright 2019 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package exit contains functions useful for exiting gracefully.
package exit

import (
	"fmt"
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

	// MaxProblems controls the number of problems to show for each source
	MaxProblems = 3
)

// Usage outputs a usage error and exits with error code 64
func Usage(format string, a ...interface{}) {
	console.ErrStyle("usage", format, a...)
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
	displayError(msg, err)
	// Here is where we would insert code to optionally upload a stack trace.

	// We can be smarter about guessing exit codes, but EX_SOFTWARE should suffice.
	os.Exit(Software)
}

// WithProblems outputs an error along with any autodetected problems, and exits.
func WithProblems(msg string, err error, problems map[string][]string) {
	displayError(msg, err)

	for name, lines := range problems {
		console.OutStyle("failure", "Problems detected in %q:", name)
		if len(lines) > MaxProblems {
			lines = lines[:MaxProblems]
		}
		for _, l := range lines {
			console.OutStyle("log-entry", l)
		}
	}
	os.Exit(Software)
}

func displayError(msg string, err error) {
	// use Warning because Error will display a duplicate message to stderr
	glog.Warningf(fmt.Sprintf("%s: %v", msg, err))
	console.Fatal(msg+": %v", err)
	console.Err("\n")
	console.ErrStyle("sad", "Sorry that minikube crashed. If this was unexpected, we would love to hear from you:")
	console.ErrStyle("url", "https://github.com/kubernetes/minikube/issues/new")
}
