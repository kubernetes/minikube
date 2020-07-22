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
	"os"
	"runtime"
	"runtime/debug"

	"github.com/golang/glog"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/problem"
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
)

// UsageT outputs a templated usage error and exits with error code 64
func UsageT(format string, a ...out.V) {
	out.ErrWithExitCode(out.Usage, format, BadUsage, a...)
	os.Exit(BadUsage)
}

// WithCodeT outputs a templated fatal error message and exits with the supplied error code.
func WithCodeT(code int, format string, a ...out.V) {
	out.ErrWithExitCode(out.FatalType, format, code, a...)
	os.Exit(code)
}

// WithError outputs an error and exits.
func WithError(msg string, err error) {
	glog.Infof("WithError(%s)=%v called from:\n%s", msg, err, debug.Stack())
	p := problem.FromError(err, runtime.GOOS)
	if p != nil && out.JSON {
		p.DisplayJSON(Config)
		os.Exit(Config)
	} else {
		WithProblem(msg, err, p)
		os.Exit(Config)
	}
	out.DisplayError(msg, err)
	os.Exit(Software)
}

// WithProblem outputs info related to a known problem and exits.
func WithProblem(msg string, err error, p *problem.Problem) {
	out.ErrT(out.Empty, "")
	out.FailureT("[{{.id}}] {{.msg}} {{.error}}", out.V{"msg": msg, "id": p.ID, "error": p.Err})
	p.Display()
	if p.ShowIssueLink {
		out.ErrT(out.Empty, "")
		out.ErrT(out.Sad, "If the above advice does not help, please let us know: ")
		out.ErrT(out.URL, "https://github.com/kubernetes/minikube/issues/new/choose")
	}
}
