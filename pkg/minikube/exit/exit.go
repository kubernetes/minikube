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

	"github.com/golang/glog"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/minikube/style"
)

// Message outputs a templated fatal error message and exits with the supplied error code.
func Message(r reason.Kind, format string, args ...out.V) {
	if r.ID == "" {
		glog.Errorf("supplied reason has no ID: %+v", r)
	}

	if r.Style == style.None {
		r.Style = style.Failure
	}

	if r.ExitCode == 0 {
		r.ExitCode = reason.ExProgramError
	}

	if len(args) == 0 {
		args = append(args, out.V{})
	}
	args[0]["fatal_msg"] = out.Fmt(format, args...)
	args[0]["fatal_code"] = r.ID
	out.Error(r, "Exiting due to {{.fatal_code}}: {{.fatal_msg}}", args...)
	os.Exit(r.ExitCode)
}

// Advice is syntactic sugar to output a message with dynamically generated advice
func Advice(r reason.Kind, msg string, advice string, a ...out.V) {
	r.Advice = out.Fmt(advice, a...)
	Message(r, msg, a...)
}

// Error outputs an error and exits.
func Error(r reason.Kind, msg string, err error) {
	ki := reason.MatchKnownIssue(r, err, runtime.GOOS)
	if ki != nil {
		Message(*ki, err.Error())
	}
	// By default, unmatched errors should show a link
	r.NewIssueLink = true
	Message(r, err.Error())
}
