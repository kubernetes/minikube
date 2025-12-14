// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package diag

import "fmt"

var _ DiagnosticWithErr = NativeErrorDiagnostic{}

// NativeErrorDiagnostic is a diagnostic with error severity which wraps a Go error.
type NativeErrorDiagnostic struct {
	// detail  string
	// summary string
	err error
}

// NewNativeErrorDiagnostic returns a new error severity diagnostic with the given error.
func NewNativeErrorDiagnostic(err error) NativeErrorDiagnostic {
	return NativeErrorDiagnostic{
		err: err,
	}
}

// Severity returns the diagnostic severity.
func (d NativeErrorDiagnostic) Severity() Severity {
	return SeverityError
}

// Summary returns the diagnostic summary.
func (d NativeErrorDiagnostic) Summary() string {
	return d.err.Error()
}

// Detail returns the diagnostic detail.
func (d NativeErrorDiagnostic) Detail() string {
	return ""
}

func (d NativeErrorDiagnostic) Err() error {
	return d.err
}

// Equal returns true if the other diagnostic is wholly equivalent.
func (d NativeErrorDiagnostic) Equal(other Diagnostic) bool {
	ed, ok := other.(NativeErrorDiagnostic)

	if !ok {
		return false
	}

	return ed.Summary() == d.Summary() && ed.Detail() == d.Detail()
}

func (d NativeErrorDiagnostic) GoString() string {
	return fmt.Sprintf("NativeErrorDiagnostic: err: %s", d.err.Error())
}
