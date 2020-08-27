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

	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/out/register"
	"k8s.io/minikube/pkg/minikube/translate"
)

const issueBase = "https://github.com/kubernetes/minikube/issues"

// KnownIssue represents a known issue in minikube.
type KnownIssue struct {
	// ID is an arbitrary unique and stable string describing this issue
	ID string
	// Err is the original error
	Err error
	// Advice is actionable text that the user should follow
	Advice string
	// URL is a reference URL for more information
	URL string
	// Issues are a list of related issues to this KnownIssue
	Issues []int
	// Hide the new issue link: it isn't our KnownIssue, and we won't be able to suggest additional assistance.
	ShowIssueLink bool
	// ExitCode to be used (defaults to 1)
	ExitCode int
}

// match maps a regular expression to KnownIssue metadata.
type match struct {
	Regexp *regexp.Regexp
	Advice string
	URL    string
	Issues []int
	// GOOS is what platforms this KnownIssue may be specific to, when disambiguation is necessary.
	GOOS []string
	// Hide the new issue link: it isn't our KnownIssue, and we won't be able to suggest additional assistance.
	ShowIssueLink bool
	// ExitCode to be used (defaults to 1)
	ExitCode int
}

// Display KnownIssue metadata to the console
func (ki *KnownIssue) Display() {
	out.ErrT(out.Tip, "Suggestion: {{.advice}}", out.V{"advice": translate.T(ki.Advice)})
	if ki.URL != "" {
		out.ErrT(out.Documentation, "Documentation: {{.url}}", out.V{"url": ki.URL})
	}
	if len(ki.Issues) == 0 {
		return
	}

	if len(ki.Issues) == 1 {
		out.ErrT(out.Issues, "Related issue: {{.url}}", out.V{"url": fmt.Sprintf("%s/%d", issueBase, ki.Issues[0])})
		return
	}

	out.ErrT(out.Issues, "Related issues:")
	issues := ki.Issues
	if len(issues) > 3 {
		issues = issues[0:3]
	}
	for _, i := range issues {
		out.ErrT(out.Issue, "{{.url}}", out.V{"url": fmt.Sprintf("%s/%d", issueBase, i)})
	}
}

// DisplayJSON displays KnownIssue metadata in JSON format
func (ki *KnownIssue) DisplayJSON(exitcode int) {
	var issues string
	for _, i := range ki.Issues {
		issues += fmt.Sprintf("https://github.com/kubernetes/minikube/issues/%v,", i)
	}
	extraArgs := map[string]string{
		"name":   ki.ID,
		"advice": ki.Advice,
		"url":    ki.URL,
		"issues": issues,
	}
	register.PrintErrorExitCode(ki.Err.Error(), exitcode, extraArgs)
}


func KnownIssueMap() map[string]match{} {
	maps := []map[string]match){
		ProgramErrors,
		ResourceErrors,
		HostErrors,
	}
}


// KnownIssueFromError returns a known issue from an error on an OS
func KnownIssueFromError(err error, goos string) *KnownIssue {
	var osMatch *KnownIssue
	var genericMatch *KnownIssue

	for _, m := range maps {
		for id, match := range m {
			if !match.Regexp.MatchString(err.Error()) {
				continue
			}

			// Does this match require an OS matchup?
			if len(match.GOOS) > 0 {
				foundOS := false
				for _, o := range match.GOOS {
					if o == goos {
						foundOS = true
					}
				}
				if !foundOS {
					continue
				}
			}

			p := &KnownIssue{
				Err:           err,
				Advice:        match.Advice,
				URL:           match.URL,
				ID:            id,
				Issues:        match.Issues,
				ShowIssueLink: match.ShowIssueLink,
			}

			if len(match.GOOS) > 0 {
				osMatch = p
			} else {
				genericMatch = p
			}
		}
	}

	// Prioritize operating-system specific matches over general ones
	if osMatch != nil {
		return osMatch
	}
	return genericMatch
}
