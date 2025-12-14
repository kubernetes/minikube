// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package diag

// Severity represents the level of feedback for a diagnostic.
//
// Each severity implies behavior changes for the feedback and potentially the
// further execution of logic.
type Severity int

const (
	// SeverityInvalid represents an undefined severity.
	//
	// It should not be used directly in implementations.
	SeverityInvalid Severity = 0

	// SeverityError represents a terminating condition.
	//
	// This can cause a failing status code for command line programs.
	//
	// Most implementations should return early when encountering an error.
	SeverityError Severity = 1

	// SeverityWarning represents a condition with explicit feedback.
	//
	// Most implementations should continue when encountering a warning.
	SeverityWarning Severity = 2
)

// String returns a textual representation of the severity.
func (s Severity) String() string {
	switch s {
	case SeverityError:
		return "Error"
	case SeverityWarning:
		return "Warning"
	default:
		return "Invalid"
	}
}
