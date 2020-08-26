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
	"k8s.io/minikube/pkg/minikube/exitcode"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/problem"
)

// UsageT outputs a templated usage error and exits with error code 64
func UsageT(format string, a ...out.V) {
	out.ErrWithExitCode(out.Usage, format, exitcode.ProgramUsage, a...)
	os.Exit(exitcode.ProgramUsage)
}

// WithCodeT outputs a templated fatal error message and exits with the supplied error code.
func WithCodeT(code int, format string, a ...out.V) {
	out.ErrWithExitCode(out.FatalType, format, code, a...)
	os.Exit(code)
}

// WithError outputs an error and exits.
func WithError(msg string, err error) {
	glog.Infof("WithError(%s, %v) called from:\n%s", msg, err, debug.Stack())
	p := problem.FromError(err, runtime.GOOS)
	exitcode := exitcode.ProgramError
	if p == nil {
		out.DisplayError(msg, err)
		os.Exit(exitcode)
	}

	if p.exitcode > 0 {
		exitcode = p.ExitCode
	}

	if out.JSON {
		p.DisplayJSON()
	} else {
		showProblem(msg, err, p)
	}
	os.Exit(exitcode)
}

// WithProblem outputs an error specifically associated to a problem
func WithProblem(id string, msg string) {
	glog.Infof("WithProblem(%s, %v) called from:\n%s", id, msg, debug.Stack())
}

// showProblem outputs info related to a known problem
func showProblem(msg string, err error, p *problem.Problem) {
	out.ErrT(out.Empty, "")
	out.FailureT("[{{.id}}] {{.msg}} {{.error}}", out.V{"msg": msg, "id": p.ID, "error": p.Err})
	p.Display()
	if p.ShowIssueLink {
		out.ErrT(out.Empty, "")
		out.ErrT(out.Sad, "If the above advice does not help, please let us know: ")
		out.ErrT(out.URL, "https://github.com/kubernetes/minikube/issues/new/choose")
	}
}
