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
)

// UsageT outputs a templated usage error and exits with error code 64
func UsageT(format string, a ...out.V) {
	exitcode := ProgramUsage
	out.ErrWithExitCode(out.Usage, format, exitcode, a...)
	os.Exit(exitcode)
}

// WithCodeT outputs a templated fatal error message and exits with the supplied error code.
func WithCodeT(code int, format string, a ...out.V) {
	out.ErrWithExitCode(out.FatalType, format, code, a...)
	os.Exit(code)
}

// WithError outputs an error and exits.
func WithError(code int, msg string, err error) {
	glog.Infof("WithError(%s, %v) called from:\n%s", msg, err, debug.Stack())
	ki := KnownIssueFromError(err, runtime.GOOS)
	if ki != nil {
		WithKnownIssue(*ki, "Error: {{.err}}", out.V{"err": err})
	} else {
		out.DisplayError(msg, err)
		os.Exit(code)
	}
}

// WithKnownIssue outputs an error and exits.
func WithKnownIssue(ki KnownIssue, format string, a ...out.V) {
	glog.Infof("WithKnownIssue(%+v, %s, %s) called from:\n%s", ki, format, a, debug.Stack())

	if ki.ExitCode == 0 {
		ki.ExitCode = ProgramError
	}

	if out.JSON {
		ki.DisplayJSON(format, a...)
	} else {
		out.ErrT(out.FailureType, format, a...)
		ki.Display()
	}

	os.Exit(ki.ExitCode)
}
