/*
Copyright 2020 The Kubernetes Authors All rights reserved.

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

/*
Copyright 2019 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY knd, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package out provides a mechanism for sending localized, stylized output to the console.
package out

import (
	"strings"

	"k8s.io/minikube/pkg/minikube/out/register"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/minikube/style"
)

// Error shows an an error reason
func Error(k reason.Kind, format string, a ...V) {
	if JSON {
		msg := Fmt(format, a...)
		register.PrintErrorExitCode(strings.TrimSpace(msg), k.ExitCode, map[string]string{
			"name":   k.ID,
			"advice": k.Advice,
			"url":    k.URL,
			"issues": strings.Join(k.IssueURLs(), ","),
		})
	} else {
		displayText(k, format, a...)
	}
}

// WarnReason shows a warning reason
func WarnReason(k reason.Kind, format string, a ...V) {
	if JSON {
		msg := Fmt(format, a...)
		register.PrintWarning(msg)
	} else {
		displayText(k, format, a...)
	}
}

// indentMultiLine indents a message if it contains multiple lines
func indentMultiLine(s string) string {
	if !strings.Contains(s, "\n") {
		return s
	}

	cleaned := []string{"\n"}
	for _, sn := range strings.Split(s, "\n") {
		cleaned = append(cleaned, style.Indented+strings.TrimSpace(sn))
	}
	return strings.Join(cleaned, "\n")
}

func displayText(k reason.Kind, format string, a ...V) {
	Ln("")
	st := k.Style

	if st == style.None {
		st = style.KnownIssue
	}

	ErrT(st, format, a...)

	if k.Advice != "" {

		advice := indentMultiLine(Fmt(k.Advice, a...))
		ErrT(style.Tip, Fmt("Suggestion: {{.advice}}", V{"advice": advice}))
	}

	if k.URL != "" {
		ErrT(style.Documentation, "Documentation: {{.url}}", V{"url": k.URL})
	}

	issueURLs := k.IssueURLs()
	if len(issueURLs) > 0 {
		ErrT(style.Issues, "Related issues:")
		for _, i := range issueURLs {
			ErrT(style.Issue, "{{.url}}", V{"url": i})
		}
	}

	if k.NewIssueLink {
		ErrT(style.Empty, "")
		ErrT(style.Sad, "If the above advice does not help, please let us know: ")
		ErrT(style.URL, "https://github.com/kubernetes/minikube/issues/new/choose")
	}
	Ln("")
}
