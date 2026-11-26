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

	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/minikube/style"
)

var (
	shell bool
)

// SetShell configures if we are doing a shell configuration or not
func SetShell(s bool) {
	shell = s
}

// Message outputs a templated message and exits without interpretation
func Message(r reason.Kind, format string, args ...out.V) {
	if r.ID == "" {
		klog.Errorf("supplied reason has no ID: %+v", r)
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

	// No need to manipulate the message for JSON output
	if out.JSON {
		out.Error(r, format, args...)
	} else {
		args[0]["fatal_msg"] = out.Fmt(format, args...)
		args[0]["fatal_code"] = r.ID
		out.Error(r, "Exiting due to {{.fatal_code}}: {{.fatal_msg}}", args...)
	}

	Code(r.ExitCode)
}

// Code will exit with a code
func Code(code int) {
	if shell {
		out.Outputf(os.Stdout, "false exit code %d\n", code)
	}
	os.Exit(code)
}

// Advice is syntactic sugar to output a message with dynamically generated advice
func Advice(r reason.Kind, msg string, advice string, a ...out.V) {
	r.Advice = out.Fmt(advice, a...)
	Message(r, msg, a...)
}

// Error takes a fatal error and outputs the caller's message for it.
//
// Known-issue matching used to live here, but it silently replaced the
// caller-selected reason, message and exit code whenever the raw error
// text matched a global regex (see #24122): what the developer wrote was
// not necessarily what users saw. Code that needs to change behavior based
// on error text should call reason.MatchKnownIssue explicitly, as
// validateDriver does, so the override is visible at the call site.
func Error(r reason.Kind, msg string, err error) {
	// By default, unmatched errors should show a link
	r.NewIssueLink = true
	Message(r, fmt.Sprintf("%s: %v", msg, err))
}
