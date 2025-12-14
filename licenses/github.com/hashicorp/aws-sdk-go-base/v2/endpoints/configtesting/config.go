// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package configtesting

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/aws-sdk-go-base/v2/mockdata"
	"github.com/hashicorp/aws-sdk-go-base/v2/servicemocks"
)

type TestMode int

const (
	TestModeInvalid        TestMode = 0
	TestModeLocal          TestMode = 1
	TestModeAcceptanceTest TestMode = 2
)

type TestDriver interface {
	Init(mode TestMode)
	TestCase() TestCaseDriver
}

type TestCaseDriver interface {
	Configuration(f []ConfigFunc) Configurer
	Setup(t *testing.T)
	Apply(ctx context.Context, t *testing.T) (context.Context, Thing)
}

type Configurer interface {
	SetAccessKey(s string)
	SetSecretKey(s string)
	SetProfile(s string)
	SetUseFIPSEndpoint(b bool)
	AddEndpoint(k, v string)
	AddSharedConfigFile(f string)
}

type Thing interface {
	GetCredentials() aws.CredentialsProvider
	GetRegion() string
}

type AwsConfigThing interface {
	GetAwsConfig() aws.Config
}

type ConfigFunc func(c Configurer)

func WithProfile(s string) ConfigFunc {
	return func(c Configurer) {
		c.SetProfile(s)
	}
}

func WithUseFIPSEndpoint(b bool) ConfigFunc {
	return func(c Configurer) {
		c.SetUseFIPSEndpoint(b)
	}
}

func SSO(t *testing.T, driver TestDriver) {
	t.Helper()

	driver.Init(TestModeLocal)

	const ssoSessionName = "test-sso-session"

	testCases := map[string]struct {
		Configuration            []ConfigFunc
		SharedConfigurationFile  string
		ExpectedCredentialsValue aws.Credentials
	}{
		"shared configuration file": {
			SharedConfigurationFile: fmt.Sprintf(`
[default]
sso_session = %s
sso_account_id = 123456789012
sso_role_name = testRole
region = us-east-1

[sso-session test-sso-session]
sso_region = us-east-1
sso_start_url = https://d-123456789a.awsapps.com/start
sso_registration_scopes = sso:account:access
`, ssoSessionName),
			ExpectedCredentialsValue: mockdata.MockSsoCredentials,
		},

		"use FIPS": {
			Configuration: []ConfigFunc{
				WithUseFIPSEndpoint(true),
			},
			SharedConfigurationFile: fmt.Sprintf(`
[default]
sso_session = %s
sso_account_id = 123456789012
sso_role_name = testRole
region = us-east-1

[sso-session test-sso-session]
sso_region = us-east-1
sso_start_url = https://d-123456789a.awsapps.com/start
sso_registration_scopes = sso:account:access
`, ssoSessionName),
			ExpectedCredentialsValue: mockdata.MockSsoCredentials,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			caseDriver := driver.TestCase()

			servicemocks.InitSessionTestEnv(t)

			ctx := context.TODO()

			err := servicemocks.SsoTestSetup(t, ssoSessionName)
			if err != nil {
				t.Fatalf("setup: %s", err)
			}

			config := caseDriver.Configuration(tc.Configuration)

			closeSso, ssoEndpoint := servicemocks.SsoCredentialsApiMock()
			defer closeSso()
			config.AddEndpoint("sso", ssoEndpoint)

			tempdir, err := os.MkdirTemp("", "temp")
			if err != nil {
				t.Fatalf("error creating temp dir: %s", err)
			}
			defer os.Remove(tempdir)
			t.Setenv("TMPDIR", tempdir)

			if tc.SharedConfigurationFile != "" {
				file, err := os.CreateTemp("", "aws-sdk-go-base-shared-configuration-file")

				if err != nil {
					t.Fatalf("unexpected error creating temporary shared configuration file: %s", err)
				}

				defer os.Remove(file.Name())

				err = os.WriteFile(file.Name(), []byte(tc.SharedConfigurationFile), 0600) //nolint:mnd

				if err != nil {
					t.Fatalf("unexpected error writing shared configuration file: %s", err)
				}

				config.AddSharedConfigFile(file.Name())
			}

			caseDriver.Setup(t)

			ctx, thing := caseDriver.Apply(ctx, t)

			credentials := thing.GetCredentials()
			if credentials == nil {
				t.Fatal("credentials are nil")
			}
			credentialsValue, err := credentials.Retrieve(ctx)

			if err != nil {
				t.Fatalf("retrieving credentials: %s", err)
			}

			if diff := cmp.Diff(credentialsValue, tc.ExpectedCredentialsValue, cmpopts.IgnoreFields(aws.Credentials{}, "Expires")); diff != "" {
				t.Fatalf("unexpected credentials: (- got, + expected)\n%s", diff)
			}
		})
	}
}

func LegacySSO(t *testing.T, driver TestDriver) {
	t.Helper()

	driver.Init(TestModeLocal)

	const ssoStartUrl = "https://d-123456789a.awsapps.com/start"

	testCases := map[string]struct {
		Configuration            []ConfigFunc
		SharedConfigurationFile  string
		ExpectedCredentialsValue aws.Credentials
	}{
		"shared configuration file": {
			SharedConfigurationFile: fmt.Sprintf(`
[default]
sso_start_url = %s
sso_region = us-east-1
sso_account_id = 123456789012
sso_role_name = testRole
region = us-east-1
`, ssoStartUrl),
			ExpectedCredentialsValue: mockdata.MockSsoCredentials,
		},

		"use FIPS": {
			Configuration: []ConfigFunc{
				WithUseFIPSEndpoint(true),
			},
			SharedConfigurationFile: fmt.Sprintf(`
[default]
sso_start_url = %s
sso_region = us-east-1
sso_account_id = 123456789012
sso_role_name = testRole
region = us-east-1
`, ssoStartUrl),
			ExpectedCredentialsValue: mockdata.MockSsoCredentials,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			caseDriver := driver.TestCase()

			servicemocks.InitSessionTestEnv(t)

			ctx := context.TODO()

			err := servicemocks.SsoTestSetup(t, ssoStartUrl)
			if err != nil {
				t.Fatalf("setup: %s", err)
			}

			config := caseDriver.Configuration(tc.Configuration)

			closeSso, ssoEndpoint := servicemocks.SsoCredentialsApiMock()
			defer closeSso()
			config.AddEndpoint("sso", ssoEndpoint)

			tempdir, err := os.MkdirTemp("", "temp")
			if err != nil {
				t.Fatalf("error creating temp dir: %s", err)
			}
			defer os.Remove(tempdir)
			t.Setenv("TMPDIR", tempdir)

			if tc.SharedConfigurationFile != "" {
				file, err := os.CreateTemp("", "aws-sdk-go-base-shared-configuration-file")

				if err != nil {
					t.Fatalf("unexpected error creating temporary shared configuration file: %s", err)
				}

				defer os.Remove(file.Name())

				err = os.WriteFile(file.Name(), []byte(tc.SharedConfigurationFile), 0600) //nolint:mnd

				if err != nil {
					t.Fatalf("unexpected error writing shared configuration file: %s", err)
				}

				config.AddSharedConfigFile(file.Name())
			}

			caseDriver.Setup(t)

			ctx, thing := caseDriver.Apply(ctx, t)

			credentials := thing.GetCredentials()
			if credentials == nil {
				t.Fatal("credentials are nil")
			}
			credentialsValue, err := credentials.Retrieve(ctx)

			if err != nil {
				t.Fatalf("retrieving credentials: %s", err)
			}

			if diff := cmp.Diff(credentialsValue, tc.ExpectedCredentialsValue, cmpopts.IgnoreFields(aws.Credentials{}, "Expires")); diff != "" {
				t.Fatalf("unexpected credentials: (- got, + expected)\n%s", diff)
			}
		})
	}
}
