// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package diag_test

import (
	"testing"

	"github.com/hashicorp/aws-sdk-go-base/v2/diag"
)

func TestErrorDiagnosticEqual(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		diag     diag.ErrorDiagnostic
		other    diag.Diagnostic
		expected bool
	}{
		"matching": {
			diag:     diag.NewErrorDiagnostic("test summary", "test detail"),
			other:    diag.NewErrorDiagnostic("test summary", "test detail"),
			expected: true,
		},
		"nil": {
			diag:     diag.NewErrorDiagnostic("test summary", "test detail"),
			other:    nil,
			expected: false,
		},
		"different-detail": {
			diag:     diag.NewErrorDiagnostic("test summary", "test detail"),
			other:    diag.NewErrorDiagnostic("test summary", "different detail"),
			expected: false,
		},
		"different-summary": {
			diag:     diag.NewErrorDiagnostic("test summary", "test detail"),
			other:    diag.NewErrorDiagnostic("different summary", "test detail"),
			expected: false,
		},
		"different-type": {
			diag:     diag.NewErrorDiagnostic("test summary", "test detail"),
			other:    diag.NewWarningDiagnostic("test summary", "test detail"),
			expected: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := tc.diag.Equal(tc.other)

			if got != tc.expected {
				t.Errorf("Unexpected response: got: %t, wanted: %t", got, tc.expected)
			}
		})
	}
}
