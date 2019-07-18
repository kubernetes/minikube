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
	"runtime"

	"github.com/golang/glog"
	"k8s.io/minikube/pkg/minikube/console"
	"k8s.io/minikube/pkg/minikube/problem"
	"k8s.io/minikube/pkg/minikube/translate"
)

// Exit codes based on sysexits(3)
const (
	Failure     = 1  // Failure represents a general failure code
	Interrupted = 2  // Ctrl-C (SIGINT)
	BadUsage    = 64 // Usage represents an incorrect command line
	Data        = 65 // Data represents incorrect data supplied by the user
	NoInput     = 66 // NoInput represents that the input file did not exist or was not readable
	Unavailable = 69 // Unavailable represents when a service was unavailable
	Software    = 70 // Software represents an internal software error.
	IO          = 74 // IO represents an I/O error
	Config      = 78 // Config represents an unconfigured or misconfigured state
	Permissions = 77 // Permissions represents a permissions error

	// MaxLogEntries controls the number of log entries to show for each source
	MaxLogEntries = 3
)

// Usage outputs a usage error and exits with error code 64
func Usage(format string, a ...interface{}) {
	console.ErrStyle(console.Usage, format, a...)
	os.Exit(BadUsage)
}

// UsageT outputs a templated usage error and exits with error code 64
func UsageT(format string, a ...console.Arg) {
	console.ErrT(console.Usage, format, a...)
	os.Exit(BadUsage)
}

// WithCode outputs a fatal error message and exits with a supplied error code.
func WithCode(code int, format string, a ...interface{}) {
	// use Warning because Error will display a duplicate message to stderr
	glog.Warningf(format, a...)
	console.Fatal(format, a...)
	os.Exit(code)
}

// WithCodeT outputs a templated fatal error message and exits with the supplied error code.
func WithCodeT(code int, format string, a ...console.Arg) {
	console.FatalT(format, a...)
	os.Exit(code)
}

// WithError outputs an error and exits.
func WithError(msg string, err error) {
	p := problem.FromError(err, runtime.GOOS)
	if p != nil {
		WithProblem(msg, p)
	}
	displayError(msg, err)
	os.Exit(Software)
}

// WithProblem outputs info related to a known problem and exits.
func WithProblem(msg string, p *problem.Problem) {
	console.ErrT(console.Empty, "")
	console.FatalT(msg)
	p.Display()
	console.ErrT(console.Empty, "")
	console.ErrT(console.Sad, "If the above advice does not help, please let us know: ")
	console.ErrT(console.URL, "https://github.com/kubernetes/minikube/issues/new/choose")
	os.Exit(Config)
}

// WithLogEntries outputs an error along with any important log entries, and exits.
func WithLogEntries(msg string, err error, entries map[string][]string) {
	displayError(msg, err)

	for name, lines := range entries {
		console.OutT(console.FailureType, "Problems detected in {{.entry}}:", console.Arg{"entry": name})
		if len(lines) > MaxLogEntries {
			lines = lines[:MaxLogEntries]
		}
		for _, l := range lines {
			console.OutT(console.LogEntry, l)
		}
	}
	os.Exit(Software)
}

func displayError(msg string, err error) {
	// use Warning because Error will display a duplicate message to stderr
	glog.Warningf(fmt.Sprintf("%s: %v", msg, err))
	console.ErrT(console.Empty, "")
	console.FatalT("{{.msg}}: {{.err}}", console.Arg{"msg": translate.T(msg), "err": err})
	console.ErrT(console.Empty, "")
	console.ErrT(console.Sad, "Sorry that minikube crashed. If this was unexpected, we would love to hear from you:")
	console.ErrT(console.URL, "https://github.com/kubernetes/minikube/issues/new/choose")
}
