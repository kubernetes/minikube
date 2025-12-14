// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package awsbase

import (
	"testing"

	"github.com/hashicorp/aws-sdk-go-base/v2/diag"
)

func TestIsCannotAssumeRoleError(t *testing.T) {
	testCases := []struct {
		Name     string
		Diag     diag.Diagnostic
		Expected bool
	}{
		{
			Name: "nil error",
		},
		{
			Name: "Top-level NoValidCredentialSourcesError",
			Diag: NoValidCredentialSourcesError{},
		},
		{
			Name:     "Top-level CannotAssumeRoleError",
			Diag:     cannotAssumeRoleError{},
			Expected: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			got := IsCannotAssumeRoleError(testCase.Diag)

			if got != testCase.Expected {
				t.Errorf("got %t, expected %t", got, testCase.Expected)
			}
		})
	}
}

func TestIsNoValidCredentialSourcesError(t *testing.T) {
	testCases := []struct {
		Name     string
		Diag     diag.Diagnostic
		Expected bool
	}{
		{
			Name: "nil error",
		},
		{
			Name: "Top-level CannotAssumeRoleError",
			Diag: cannotAssumeRoleError{},
		},
		{
			Name:     "Top-level NoValidCredentialSourcesError",
			Diag:     NoValidCredentialSourcesError{},
			Expected: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			got := IsNoValidCredentialSourcesError(testCase.Diag)

			if got != testCase.Expected {
				t.Errorf("got %t, expected %t", got, testCase.Expected)
			}
		})
	}
}
