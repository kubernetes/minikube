// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package diag

type Diagnostic interface {
	// Severity returns the desired level of feedback for the diagnostic.
	Severity() Severity

	// Summary is a short description for the diagnostic.
	//
	// Typically this is implemented as a title, such as "Invalid Resource Name",
	// or single line sentence.
	Summary() string

	// Detail is a long description for the diagnostic.
	//
	// This should contain all relevant information about why the diagnostic
	// was generated and if applicable, ways to prevent the diagnostic. It
	// should generally be written and formatted for human consumption by
	// practitioners or provider developers.
	Detail() string

	// Equal returns true if the other diagnostic is wholly equivalent.
	Equal(Diagnostic) bool
}

type DiagnosticWithErr interface {
	Diagnostic

	Err() error
}
