// Package problem helps deliver actionable feedback to a user based on an error message.
package problem

import (
	"regexp"

	"k8s.io/minikube/pkg/minikube/console"
)

const issueBase = "https://github.com/kubernetes/minikube/issue"

// Problem represents a known problem in minikube.
type Problem struct {
	ID       string
	Err      error
	Solution string
	URL      string
	Issues   []int
}

// match maps a regular expression to problem metadata.
type match struct {
	Regexp   *regexp.Regexp
	Solution string
	URL      string
	Issues   []int
}

// Display problem metadata to the console
func (p *Problem) Display() {
	console.ErrStyle("solution", "Error:         [%s] %v", p.ID, p.Err)
	console.ErrStyle("solution", "Solution:      %s", p.Solution)
	console.ErrStyle("solution", "Documentation: %s", p.URL)
	if len(p.Issues) == 0 {
		return
	}
	issues := p.Issues
	if len(issues) > 3 {
		issues = issues[0:3]
	}
	console.ErrStyle("solution", "Related issues:")
	for _, i := range issues {
		console.ErrStyle("related-issue", "%s/%d", issueBase, i)
	}
}

// FromError returns a known problem from an error.
func FromError(err error) *Problem {
	maps := []map[string]match{
		vmProblems,
		netProblems,
		deployProblems,
	}
	for _, m := range maps {
		for k, v := range m {
			if v.Regexp.MatchString(err.Error()) {
				return &Problem{
					Err:      err,
					Solution: v.Solution,
					URL:      v.URL,
					ID:       k,
					Issues:   v.Issues,
				}
			}
		}
	}
	return nil
}
