// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package config

import (
	"fmt"

	"github.com/hashicorp/aws-sdk-go-base/v2/diag"
)

// CannotAssumeRoleWithWebIdentityError occurs when AssumeRoleWithWebIdentity cannot complete.
type CannotAssumeRoleWithWebIdentityError struct {
	Config *Config
	err    error
}

func (e CannotAssumeRoleWithWebIdentityError) Severity() diag.Severity {
	return diag.SeverityError
}

func (e CannotAssumeRoleWithWebIdentityError) Summary() string {
	return "Cannot assume IAM Role with web identity"
}

func (e CannotAssumeRoleWithWebIdentityError) Detail() string {
	if e.Config == nil || e.Config.AssumeRoleWithWebIdentity == nil {
		return fmt.Sprintf("cannot assume role with web identity: %s", e.err)
	}

	return fmt.Sprintf(`IAM Role (%s) cannot be assumed with web identity token.

There are a number of possible causes of this - the most common are:
  * The web identity token used in order to assume the role is invalid
  * The web identity token does not have appropriate permission to assume the role
  * The role ARN is not valid

Error: %s
`, e.Config.AssumeRoleWithWebIdentity.RoleARN, e.err)
}

func (e CannotAssumeRoleWithWebIdentityError) Equal(other diag.Diagnostic) bool {
	ed, ok := other.(CannotAssumeRoleWithWebIdentityError)
	if !ok {
		return false
	}

	return ed.Summary() == e.Summary() && ed.Detail() == e.Detail()
}

func (e CannotAssumeRoleWithWebIdentityError) Err() error {
	return e.err
}

func (c *Config) NewCannotAssumeRoleWithWebIdentityError(err error) CannotAssumeRoleWithWebIdentityError {
	return CannotAssumeRoleWithWebIdentityError{
		Config: c,
		err:    err,
	}
}

var _ diag.DiagnosticWithErr = CannotAssumeRoleWithWebIdentityError{}

// NoValidCredentialSourcesError occurs when all credential lookup methods have been exhausted without results.
type NoValidCredentialSourcesError struct {
	Config *Config
	err    error
}

func (e NoValidCredentialSourcesError) Severity() diag.Severity {
	return diag.SeverityError
}

func (e NoValidCredentialSourcesError) Summary() string {
	return "No valid credential sources found"
}

func (e NoValidCredentialSourcesError) Detail() string {
	if e.Config == nil {
		return e.err.Error()
	}

	return fmt.Sprintf(`Please see %[1]s
for more information about providing credentials.

Error: %[2]s
`, e.Config.CallerDocumentationURL, e.err)
}

func (e NoValidCredentialSourcesError) Equal(other diag.Diagnostic) bool {
	ed, ok := other.(NoValidCredentialSourcesError)
	if !ok {
		return false
	}

	return ed.Summary() == e.Summary() && ed.Detail() == e.Detail()
}

func (e NoValidCredentialSourcesError) Err() error {
	return e.err
}

func (c *Config) NewNoValidCredentialSourcesError(err error) NoValidCredentialSourcesError {
	return NoValidCredentialSourcesError{
		Config: c,
		err:    err,
	}
}

var _ diag.DiagnosticWithErr = NoValidCredentialSourcesError{}
