// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package diag

var _ Diagnostic = WarningDiagnostic{}

// WarningDiagnostic is a generic diagnostic with warning severity.
type WarningDiagnostic struct {
	detail  string
	summary string
}

// NewWarningDiagnostic returns a new warning severity diagnostic with the given summary and detail.
func NewWarningDiagnostic(summary string, detail string) WarningDiagnostic {
	return WarningDiagnostic{
		detail:  detail,
		summary: summary,
	}
}

// Severity returns the diagnostic severity.
func (d WarningDiagnostic) Severity() Severity {
	return SeverityWarning
}

// Summary returns the diagnostic summary.
func (d WarningDiagnostic) Summary() string {
	return d.summary
}

// Detail returns the diagnostic detail.
func (d WarningDiagnostic) Detail() string {
	return d.detail
}

// Equal returns true if the other diagnostic is wholly equivalent.
func (d WarningDiagnostic) Equal(other Diagnostic) bool {
	wd, ok := other.(WarningDiagnostic)

	if !ok {
		return false
	}

	return wd.Summary() == d.Summary() && wd.Detail() == d.Detail()
}
