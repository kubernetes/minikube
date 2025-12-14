// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package awsbase

import (
	"fmt"
	"slices"

	"github.com/hashicorp/aws-sdk-go-base/v2/diag"
	"github.com/hashicorp/aws-sdk-go-base/v2/internal/config"
)

// cannotAssumeRoleError occurs when AssumeRole cannot complete.
type cannotAssumeRoleError struct {
	ar  config.AssumeRole
	err error
}

func (e cannotAssumeRoleError) Severity() diag.Severity {
	return diag.SeverityError
}

func (e cannotAssumeRoleError) Summary() string {
	return "Cannot assume IAM Role"
}

func (e cannotAssumeRoleError) Detail() string {
	return fmt.Sprintf(`IAM Role (%s) cannot be assumed.

There are a number of possible causes of this - the most common are:
  * The credentials used in order to assume the role are invalid
  * The credentials do not have appropriate permission to assume the role
  * The role ARN is not valid

Error: %s
`, e.ar.RoleARN, e.err)
}

func (e cannotAssumeRoleError) Equal(other diag.Diagnostic) bool {
	ed, ok := other.(cannotAssumeRoleError)
	if !ok {
		return false
	}

	return ed.Summary() == e.Summary() && ed.Detail() == e.Detail()
}

func (e cannotAssumeRoleError) Err() error {
	return e.err
}

func newCannotAssumeRoleError(ar AssumeRole, err error) cannotAssumeRoleError {
	return cannotAssumeRoleError{
		ar:  ar,
		err: err,
	}
}

var _ diag.DiagnosticWithErr = cannotAssumeRoleError{}

// IsCannotAssumeRoleError returns true if the error contains the CannotAssumeRoleError type.
func IsCannotAssumeRoleError(diag diag.Diagnostic) bool {
	_, ok := diag.(cannotAssumeRoleError)
	return ok
}

// NoValidCredentialSourcesError occurs when all credential lookup methods have been exhausted without results.
type NoValidCredentialSourcesError = config.NoValidCredentialSourcesError

// IsNoValidCredentialSourcesError returns true if the diagnostic is a NoValidCredentialSourcesError.
func IsNoValidCredentialSourcesError(diag diag.Diagnostic) bool {
	_, ok := diag.(NoValidCredentialSourcesError)
	return ok
}

// ContainsNoValidCredentialSourcesError returns true if the diagnostics contains a NoValidCredentialSourcesError type.
func ContainsNoValidCredentialSourcesError(diags diag.Diagnostics) bool {
	return slices.ContainsFunc(diags, IsNoValidCredentialSourcesError)
}
