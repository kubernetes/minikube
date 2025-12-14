// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package configtesting

import (
	"context"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/hashicorp/aws-sdk-go-base/v2/servicemocks"
)

// haredConfigFileParsing prevents regression in shared config file parsing
// * https://github.com/aws/aws-sdk-go-v2/issues/2349: indented keys
func SharedConfigFileParsing(t *testing.T, driver TestDriver) {
	t.Helper()

	driver.Init(TestModeLocal)

	testcases := map[string]struct {
		Configuration           []ConfigFunc
		SharedConfigurationFile string
		Check                   func(t *testing.T, thing Thing)
	}{
		"leading newline": {
			SharedConfigurationFile: `
[default]
region = us-west-2
`,
			Check: func(t *testing.T, thing Thing) {
				region := thing.GetRegion()
				if a, e := region, "us-west-2"; a != e {
					t.Errorf("expected region %q, got %q", e, a)
				}
			},
		},

		"leading whitespace": {
			// Do not "fix" indentation!
			SharedConfigurationFile: `	[default]
	region = us-west-2
	`,
			Check: func(t *testing.T, thing Thing) {
				region := thing.GetRegion()
				if a, e := region, "us-west-2"; a != e {
					t.Errorf("expected region %q, got %q", e, a)
				}
			},
		},

		"leading newline and whitespace": {
			// Do not "fix" indentation!
			SharedConfigurationFile: `
	[default]
	region = us-west-2
		`,
			Check: func(t *testing.T, thing Thing) {
				region := thing.GetRegion()
				if a, e := region, "us-west-2"; a != e {
					t.Errorf("expected region %q, got %q", e, a)
				}
			},
		},

		"named profile after leading newline and whitespace": {
			Configuration: []ConfigFunc{
				WithProfile("test"),
			},
			// Do not "fix" indentation!
			SharedConfigurationFile: `
[default]
region = us-west-2

	[profile test]
	region = us-east-1
			`,
			Check: func(t *testing.T, thing Thing) {
				region := thing.GetRegion()
				if a, e := region, "us-east-1"; a != e {
					t.Errorf("expected region %q, got %q", e, a)
				}
			},
		},

		"named profile": {
			Configuration: []ConfigFunc{
				WithProfile("test"),
			},
			SharedConfigurationFile: `
[default]
region = us-west-2

[profile test]
region = us-east-1
`,
			Check: func(t *testing.T, thing Thing) {
				region := thing.GetRegion()
				if a, e := region, "us-east-1"; a != e {
					t.Errorf("expected region %q, got %q", e, a)
				}
			},
		},

		"trailing hash": {
			SharedConfigurationFile: `
[default]
sso_start_url = https://d-123456789a.awsapps.com/start#
`,
			Check: func(t *testing.T, thing Thing) {
				ct, ok := thing.(AwsConfigThing)
				if !ok {
					t.Skipf("Not an AwsConfigThing")
				}

				awsConfig := ct.GetAwsConfig()
				var ssoStartUrl string
				for _, source := range awsConfig.ConfigSources {
					if shared, ok := source.(config.SharedConfig); ok {
						ssoStartUrl = shared.SSOStartURL
					}
				}
				if a, e := ssoStartUrl, "https://d-123456789a.awsapps.com/start#"; a != e {
					t.Errorf("expected sso_start_url %q, got %q", e, a)
				}
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			ctx := context.TODO()

			caseDriver := driver.TestCase()

			servicemocks.InitSessionTestEnv(t)

			config := caseDriver.Configuration(tc.Configuration)

			config.SetAccessKey(servicemocks.MockStaticAccessKey)
			config.SetSecretKey(servicemocks.MockStaticSecretKey)

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

			_, thing := caseDriver.Apply(ctx, t)

			tc.Check(t, thing)
		})
	}
}
