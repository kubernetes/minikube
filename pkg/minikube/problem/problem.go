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

// Package problem helps deliver actionable feedback to a user based on an error message.
package problem

import (
	"fmt"
	"regexp"

	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/out/register"
	"k8s.io/minikube/pkg/minikube/translate"
)

const issueBase = "https://github.com/kubernetes/minikube/issues"

// Problem represents a known problem in minikube.
type Problem struct {
	// ID is an arbitrary unique and stable string describing this issue
	ID string
	// Err is the original error
	Err error
	// Advice is actionable text that the user should follow
	Advice string
	// URL is a reference URL for more information
	URL string
	// Issues are a list of related issues to this problem
	Issues []int
	// Hide the new issue link: it isn't our problem, and we won't be able to suggest additional assistance.
	ShowIssueLink bool
}

// match maps a regular expression to problem metadata.
type match struct {
	Regexp *regexp.Regexp
	Advice string
	URL    string
	Issues []int
	// GOOS is what platforms this problem may be specific to, when disambiguation is necessary.
	GOOS []string
	// Hide the new issue link: it isn't our problem, and we won't be able to suggest additional assistance.
	ShowIssueLink bool
}

// Display problem metadata to the console
func (p *Problem) Display() {
	out.ErrT(out.Tip, "Suggestion: {{.advice}}", out.V{"advice": translate.T(p.Advice)})
	if p.URL != "" {
		out.ErrT(out.Documentation, "Documentation: {{.url}}", out.V{"url": p.URL})
	}
	if len(p.Issues) == 0 {
		return
	}

	if len(p.Issues) == 1 {
		out.ErrT(out.Issues, "Related issue: {{.url}}", out.V{"url": fmt.Sprintf("%s/%d", issueBase, p.Issues[0])})
		return
	}

	out.ErrT(out.Issues, "Related issues:")
	issues := p.Issues
	if len(issues) > 3 {
		issues = issues[0:3]
	}
	for _, i := range issues {
		out.ErrT(out.Issue, "{{.url}}", out.V{"url": fmt.Sprintf("%s/%d", issueBase, i)})
	}
}

// DisplayJSON displays problem metadata in JSON format
func (p *Problem) DisplayJSON(exitcode int) {
	var issues string
	for _, i := range p.Issues {
		issues += fmt.Sprintf("https://github.com/kubernetes/minikube/issues/%v,", i)
	}
	extraArgs := map[string]string{
		"name":   p.ID,
		"advice": p.Advice,
		"url":    p.URL,
		"issues": issues,
	}
	register.PrintErrorExitCode(p.Err.Error(), exitcode, extraArgs)
}

// FromError returns a known problem from an error on an OS
func FromError(err error, goos string) *Problem {
	maps := []map[string]match{
		osProblems,
		vmProblems,
		netProblems,
		deployProblems,
		stateProblems,
		dockerProblems,
	}

	var osMatch *Problem
	var genericMatch *Problem

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

			p := &Problem{
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
