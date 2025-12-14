// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package config

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/aws-sdk-go-base/v2/diag"
	"github.com/hashicorp/aws-sdk-go-base/v2/servicemocks"
)

func TestConfig_VerifyAccountIDAllowed(t *testing.T) {
	tests := []struct {
		name      string
		config    Config
		accountID string
		wantErr   bool
	}{
		{
			"empty",
			Config{},
			"1234",
			false,
		},
		{
			"allowed",
			Config{
				AllowedAccountIds: []string{"1234"},
			},
			"1234",
			false,
		},
		{
			"not allowed",
			Config{
				AllowedAccountIds: []string{"5678"},
			},
			"1234",
			true,
		},
		{
			"forbidden",
			Config{
				ForbiddenAccountIds: []string{"1234"},
			},
			"1234",
			true,
		},
		{
			"not forbidden",
			Config{
				ForbiddenAccountIds: []string{"5678"},
			},
			"1234",
			false,
		},
		{
			// In practice the upstream interfaces (AWS Provider, S3 Backend, etc.) should make
			// these conflict, but documenting the behavior for completeness.
			"allowed and forbidden",
			Config{
				AllowedAccountIds:   []string{"1234"},
				ForbiddenAccountIds: []string{"1234"},
			},
			"1234",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.config.VerifyAccountIDAllowed(tt.accountID); (err != nil) != tt.wantErr {
				t.Errorf("Config.VerifyAccountIDAllowed() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func foo(_ *url.URL, err error) error {
	return err
}

func TestValidateProxyConfig(t *testing.T) {
	testcases := map[string]struct {
		config               Config
		environmentVariables map[string]string
		expectedDiags        diag.Diagnostics
	}{
		"no config": {},

		"invalid HTTP proxy": {
			config: Config{
				HTTPProxy:  aws.String(" http://invalid.test"), // explicit URL parse failure
				HTTPSProxy: aws.String("http://valid.test"),
			},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Invalid HTTP Proxy",
					fmt.Sprintf("Unable to parse URL: %s", foo(url.Parse(" http://invalid.test"))), //nolint:staticcheck
				),
			},
		},

		"invalid HTTPS proxy": {
			config: Config{
				HTTPProxy:  aws.String("http://valid.test"),
				HTTPSProxy: aws.String(" http://invalid.test"), // explicit URL parse failure
			},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Invalid HTTPS Proxy",
					fmt.Sprintf("Unable to parse URL: %s", foo(url.Parse(" http://invalid.test"))), //nolint:staticcheck
				),
			},
		},

		"invalid both proxies": {
			config: Config{
				HTTPProxy:  aws.String(" http://invalid.test"), // explicit URL parse failure
				HTTPSProxy: aws.String(" http://invalid.test"), // explicit URL parse failure
			},
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Invalid HTTP Proxy",
					fmt.Sprintf("Unable to parse URL: %s", foo(url.Parse(" http://invalid.test"))), //nolint:staticcheck
				),
				diag.NewErrorDiagnostic(
					"Invalid HTTPS Proxy",
					fmt.Sprintf("Unable to parse URL: %s", foo(url.Parse(" http://invalid.test"))), //nolint:staticcheck
				),
			},
		},

		"HTTP proxy without HTTPS proxy Legacy": {
			config: Config{
				HTTPProxy:     aws.String("http://valid.test"),
				HTTPProxyMode: HTTPProxyModeLegacy,
			},
			expectedDiags: diag.Diagnostics{
				diag.NewWarningDiagnostic(
					missingHttpsProxyWarningSummary,
					fmt.Sprintf(
						"An HTTP proxy was set but no HTTPS proxy was. Using HTTP proxy %q for HTTPS requests. This behavior may change in future versions.\n\n"+
							"To specify no proxy for HTTPS, set the HTTPS to an empty string.",
						"http://valid.test"),
				),
			},
		},

		"HTTP proxy empty string": {
			config: Config{
				HTTPProxy: aws.String(""),
			},
			expectedDiags: diag.Diagnostics{},
		},

		"HTTP proxy with HTTPS proxy empty string Legacy": {
			config: Config{
				HTTPProxy:     aws.String("http://valid.test"),
				HTTPSProxy:    aws.String(""),
				HTTPProxyMode: HTTPProxyModeLegacy,
			},
			expectedDiags: diag.Diagnostics{},
		},

		"HTTP proxy config with HTTPS_PROXY envvar": {
			config: Config{
				HTTPProxy: aws.String("http://valid.test"),
			},
			environmentVariables: map[string]string{
				"HTTPS_PROXY": "http://envvar-proxy.test:1234",
			},
			expectedDiags: diag.Diagnostics{},
		},

		"HTTP proxy config with https_proxy envvar": {
			config: Config{
				HTTPProxy: aws.String("http://valid.test"),
			},
			environmentVariables: map[string]string{
				"https_proxy": "http://envvar-proxy.test:1234",
			},
			expectedDiags: diag.Diagnostics{},
		},

		"HTTP proxy without HTTPS proxy Separate": {
			config: Config{
				HTTPProxy:     aws.String("http://valid.test"),
				HTTPProxyMode: HTTPProxyModeSeparate,
			},
			expectedDiags: diag.Diagnostics{
				diag.NewWarningDiagnostic(
					missingHttpsProxyWarningSummary,
					"An HTTP proxy was set but no HTTPS proxy was.\n\n"+
						"To specify no proxy for HTTPS, set the HTTPS to an empty string.",
				),
			},
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			servicemocks.InitSessionTestEnv(t)

			for k, v := range testcase.environmentVariables {
				t.Setenv(k, v)
			}

			var diags diag.Diagnostics

			testcase.config.ValidateProxySettings(&diags)

			if diff := cmp.Diff(diags, testcase.expectedDiags); diff != "" {
				t.Errorf("Unexpected response (+wanted, -got): %s", diff)
			}
		})
	}
}
