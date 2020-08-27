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

package exit

import (
	"fmt"
	"regexp"
	"strings"

	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/out/register"
	"k8s.io/minikube/pkg/minikube/translate"
)

const issueBase = "https://github.com/kubernetes/minikube/issues"

// KnownIssue represents a known issue in minikube.
type KnownIssue struct {
	// ID is an arbitrary unique and stable string describing this issue
	ID string
	// Regexp is which regular expression this issue matches
	Regexp *regexp.Regexp
	// Operating systems this error is specific to
	GOOS []string

	// Advice is actionable text that the user should follow
	Advice string
	// URL is a reference URL for more information
	URL string
	// Issues are a list of related issues to this KnownIssue
	Issues []int
	// Show the new issue link
	ShowNewIssueLink bool
	// ExitCode to be used (defaults to 1)
	ExitCode int
}

func (ki *KnownIssue) issueURLs() []string {
	is := []string{}
	for _, i := range ki.Issues {
		is = append(is, fmt.Sprintf("%s/%d", issueBase, i))
	}
	return is
}

// Display a known issue
func (ki *KnownIssue) Display() {
	out.ErrT(out.Tip, "Suggestion: {{.advice}}", out.V{"advice": translate.T(ki.Advice)})
	if ki.URL != "" {
		out.ErrT(out.Documentation, "Documentation: {{.url}}", out.V{"url": ki.URL})
	}

	issueURLs := ki.issueURLs()
	if len(issueURLs) > 0 {
		out.ErrT(out.Issues, "Related issues:")
		for _, i := range issueURLs {
			out.ErrT(out.Issue, "{{.url}}", out.V{"url": i})
		}
	}

	if ki.ShowNewIssueLink {
		out.ErrT(out.Empty, "")
		out.ErrT(out.Sad, "If the above advice does not help, please let us know: ")
		out.ErrT(out.URL, "https://github.com/kubernetes/minikube/issues/new/choose")
	}
}

func (ki *KnownIssue) DisplayJSON(format string, a ...out.V) {
	msg := out.ApplyTemplateFormatting(out.FailureType, false, format, a...)
	register.PrintErrorExitCode(strings.TrimSpace(msg), ki.ExitCode, map[string]string{
		"name":   ki.ID,
		"advice": ki.Advice,
		"url":    ki.URL,
		"issues": strings.Join(ki.issueURLs(), ","),
	})
}

func knownIssues() []KnownIssue {
	kis := []KnownIssue{}
	// This is intentionally in dependency order
	kis = append(kis, ProgramIssues...)
	kis = append(kis, ResourceIssues...)
	kis = append(kis, HostIssues...)
	kis = append(kis, ProviderIssues...)
	kis = append(kis, DriverIssues...)
	kis = append(kis, LocalNetworkIssues...)
	kis = append(kis, InternetIssues...)
	kis = append(kis, GuestIssues...)
	kis = append(kis, RuntimeIssues...)
	kis = append(kis, ControlPlaneIssues...)
	kis = append(kis, ServiceIssues...)
	return kis
}

// KnownIssueFromError returns a known issue from an error on an OS
func KnownIssueFromError(err error, goos string) *KnownIssue {
	var genericMatch *KnownIssue

	for _, ki := range knownIssues() {
		ki := ki
		if !ki.Regexp.MatchString(err.Error()) {
			continue
		}

		// Does this match require an OS matchup?
		if len(ki.GOOS) > 0 {
			for _, o := range ki.GOOS {
				if o == goos {
					return &ki
				}
			}
		}
		if genericMatch == nil {
			genericMatch = &ki
		}
	}

	return genericMatch
}
