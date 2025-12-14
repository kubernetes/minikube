// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package test

import (
	"slices"
	"testing"

	"github.com/hashicorp/aws-sdk-go-base/v2/diag"
)

type DiagsValidator func(*testing.T, diag.Diagnostics)

type ErrValidator func(error) bool

type DiagValidator func(diag.Diagnostic) bool

func ExpectNoDiags(t *testing.T, diags diag.Diagnostics) {
	expectDiagsCount(t, diags, 0)
}

func ExpectErrDiagValidator(msg string, ev ErrValidator) DiagsValidator {
	return func(t *testing.T, diags diag.Diagnostics) {
		// Check for the correct type of error before checking for single diagnostic
		if !expectDiagsContainsErr(diags, ev) {
			t.Fatalf("expected %s, got %#v", msg, diags)
		}

		expectDiagsCount(t, diags, 1)
	}
}

func ExpectDiagValidator(msg string, dv DiagValidator) DiagsValidator {
	return func(t *testing.T, diags diag.Diagnostics) {
		// Check for the correct type of error before checking for single diagnostic
		if !expectDiagsContainsDiagFunc(diags, dv) {
			t.Fatalf("expected %s, got %#v", msg, diags)
		}

		expectDiagsCount(t, diags, 1)
	}
}

func expectDiagsCount(t *testing.T, diags diag.Diagnostics, c int) {
	if l := diags.Count(); l != c {
		t.Fatalf("Diagnostics: expected %d element, got %d\n%#v", c, l, diags)
	}
}

func expectDiagsContainsErr(diags diag.Diagnostics, ev ErrValidator) bool {
	for _, d := range diags.Errors() {
		if e, ok := d.(diag.DiagnosticWithErr); ok {
			if ev(e.Err()) {
				return true
			}
		}
	}
	return false
}

func expectDiagsContainsDiagFunc(diags diag.Diagnostics, dv DiagValidator) bool {
	return slices.ContainsFunc(diags, dv)
}
