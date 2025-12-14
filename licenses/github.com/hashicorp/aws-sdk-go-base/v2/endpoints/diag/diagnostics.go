// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package diag

// Diagnostics represents a collection of diagnostics.
//
// While this collection is ordered, the order is not guaranteed as reliable
// or consistent.
type Diagnostics []Diagnostic

// AddError adds a generic error diagnostic to the collection.
func (diags Diagnostics) AddError(summary string, detail string) Diagnostics {
	return diags.Append(NewErrorDiagnostic(summary, detail))
}

func (diags Diagnostics) AddSimpleError(err error) Diagnostics {
	return diags.Append(NewNativeErrorDiagnostic(err))
}

// AddWarning adds a generic warning diagnostic to the collection.
func (diags Diagnostics) AddWarning(summary string, detail string) Diagnostics {
	return diags.Append(NewWarningDiagnostic(summary, detail))
}

// Append adds non-empty and non-duplicate diagnostics to the collection.
func (diags Diagnostics) Append(in ...Diagnostic) Diagnostics {
	for _, diag := range in {
		if diag == nil {
			continue
		}

		if diags.Contains(diag) {
			continue
		}

		if diags == nil {
			diags = Diagnostics{diag}
		} else {
			diags = append(diags, diag)
		}
	}

	return diags
}

// Contains returns true if the collection contains an equal Diagnostic.
func (diags Diagnostics) Contains(in Diagnostic) bool {
	for _, diag := range diags {
		if diag.Equal(in) {
			return true
		}
	}

	return false
}

// Equal returns true if all given diagnostics are equivalent in order and
// content, based on the underlying (Diagnostic).Equal() method of each.
func (diags Diagnostics) Equal(other Diagnostics) bool {
	if len(diags) != len(other) {
		return false
	}

	for diagIndex, diag := range diags {
		if !diag.Equal(other[diagIndex]) {
			return false
		}
	}

	return true
}

// HasError returns true if the collection has an error severity Diagnostic.
func (diags Diagnostics) HasError() bool {
	for _, diag := range diags {
		if diag.Severity() == SeverityError {
			return true
		}
	}

	return false
}

func (diags Diagnostics) Count() int {
	return len(diags)
}

// ErrorsCount returns the number of Diagnostic in Diagnostics that are SeverityError.
func (diags Diagnostics) ErrorsCount() int {
	return len(diags.Errors())
}

// WarningsCount returns the number of Diagnostic in Diagnostics that are SeverityWarning.
func (diags Diagnostics) WarningsCount() int {
	return len(diags.Warnings())
}

// Errors returns all the Diagnostic in Diagnostics that are SeverityError.
func (diags Diagnostics) Errors() Diagnostics {
	dd := Diagnostics{}

	for _, d := range diags {
		if SeverityError == d.Severity() {
			dd = append(dd, d)
		}
	}

	return dd
}

// Warnings returns all the Diagnostic in Diagnostics that are SeverityWarning.
func (diags Diagnostics) Warnings() Diagnostics {
	dd := Diagnostics{}

	for _, d := range diags {
		if SeverityWarning == d.Severity() {
			dd = append(dd, d)
		}
	}

	return dd
}
