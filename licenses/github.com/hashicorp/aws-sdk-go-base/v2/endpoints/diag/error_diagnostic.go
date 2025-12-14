// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package diag

var _ Diagnostic = ErrorDiagnostic{}

// ErrorDiagnostic is a generic diagnostic with error severity.
type ErrorDiagnostic struct {
	detail  string
	summary string
}

// NewErrorDiagnostic returns a new error severity diagnostic with the given summary and detail.
func NewErrorDiagnostic(summary string, detail string) ErrorDiagnostic {
	return ErrorDiagnostic{
		detail:  detail,
		summary: summary,
	}
}

// Severity returns the diagnostic severity.
func (d ErrorDiagnostic) Severity() Severity {
	return SeverityError
}

// Summary returns the diagnostic summary.
func (d ErrorDiagnostic) Summary() string {
	return d.summary
}

// Detail returns the diagnostic detail.
func (d ErrorDiagnostic) Detail() string {
	return d.detail
}

// Equal returns true if the other diagnostic is wholly equivalent.
func (d ErrorDiagnostic) Equal(other Diagnostic) bool {
	ed, ok := other.(ErrorDiagnostic)

	if !ok {
		return false
	}

	return ed.Summary() == d.Summary() && ed.Detail() == d.Detail()
}
