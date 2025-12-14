// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package awsbase

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"maps"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/smithy-go"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/aws-sdk-go-base/v2/configtesting"
	"github.com/hashicorp/aws-sdk-go-base/v2/diag"
	"github.com/hashicorp/aws-sdk-go-base/v2/internal/awsconfig"
	"github.com/hashicorp/aws-sdk-go-base/v2/internal/constants"
	"github.com/hashicorp/aws-sdk-go-base/v2/internal/test"
	"github.com/hashicorp/aws-sdk-go-base/v2/logging"
	"github.com/hashicorp/aws-sdk-go-base/v2/mockdata"
	"github.com/hashicorp/aws-sdk-go-base/v2/servicemocks"
	"github.com/hashicorp/aws-sdk-go-base/v2/useragent"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-log/tflogtest"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

const (
	// Shockingly, this is not defined in the SDK
	sharedConfigCredentialsProvider = "SharedConfigCredentials"
)

func TestGetAwsConfig(t *testing.T) {
	testCases := map[string]struct {
		Config                     *Config
		EnableEc2MetadataServer    bool
		EnableEcsCredentialsServer bool
		EnableWebIdentityEnvVars   bool
		EnableWebIdentityConfig    bool
		EnvironmentVariables       map[string]string
		ExpectedCredentialsValue   aws.Credentials
		ExpectedRegion             string
		MockStsEndpoints           []*servicemocks.MockEndpoint
		SharedConfigurationFile    string
		SharedCredentialsFile      string
		ValidateDiags              test.DiagsValidator
	}{
		"no configuration or credentials": {
			Config:        &Config{},
			ValidateDiags: test.ExpectDiagValidator("NoValidCredentialSourcesError", IsNoValidCredentialSourcesError),
		},
		"config AccessKey": {
			Config: &Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				Region:    "us-east-1",
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			ExpectedCredentialsValue: mockdata.MockStaticCredentials,
			ExpectedRegion:           "us-east-1",
			MockStsEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsGetCallerIdentityValidEndpoint,
			},
		},
		"config AccessKey config AssumeRoleARN access key": {
			Config: &Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				AssumeRole: []AssumeRole{{
					RoleARN:     servicemocks.MockStsAssumeRoleArn,
					SessionName: servicemocks.MockStsAssumeRoleSessionName,
				}},
				Region:    "us-east-1",
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			ExpectedCredentialsValue: mockdata.MockStsAssumeRoleCredentials,
			ExpectedRegion:           "us-east-1",
			MockStsEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsAssumeRoleValidEndpoint,
				servicemocks.MockStsGetCallerIdentityValidEndpoint,
			},
		},
		"ExpiredToken": {
			Config: &Config{
				AccessKey:  servicemocks.MockStaticAccessKey,
				Region:     "us-east-1",
				SecretKey:  servicemocks.MockStaticSecretKey,
				MaxRetries: 100,
			},
			ExpectedRegion: "us-east-1",
			ValidateDiags: test.ExpectErrDiagValidator("ExpiredToken", func(err error) bool {
				return strings.Contains(err.Error(), "ExpiredToken")
			}),
			MockStsEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsGetCallerIdentityInvalidBodyExpiredToken,
			},
		},
		"ExpiredTokenException": {
			Config: &Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				Region:    "us-east-1",
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			ExpectedRegion: "us-east-1",
			ValidateDiags: test.ExpectErrDiagValidator("ExpiredTokenException", func(err error) bool {
				return strings.Contains(err.Error(), "ExpiredTokenException")
			}),
			MockStsEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsGetCallerIdentityInvalidBodyExpiredTokenException,
			},
		},
		"RequestExpired": {
			Config: &Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				Region:    "us-east-1",
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			ExpectedRegion: "us-east-1",
			ValidateDiags: test.ExpectErrDiagValidator("RequestExpired", func(err error) bool {
				return strings.Contains(err.Error(), "RequestExpired")
			}),
			MockStsEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsGetCallerIdentityInvalidBodyRequestExpired,
			},
		},
		"config AssumeRoleDuration": {
			Config: &Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				AssumeRole: []AssumeRole{{
					RoleARN:     servicemocks.MockStsAssumeRoleArn,
					Duration:    1 * time.Hour,
					SessionName: servicemocks.MockStsAssumeRoleSessionName,
				}},
				Region:    "us-east-1",
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			ExpectedCredentialsValue: mockdata.MockStsAssumeRoleCredentials,
			ExpectedRegion:           "us-east-1",
			MockStsEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsAssumeRoleValidEndpointWithOptions(map[string]string{"DurationSeconds": "3600"}),
				servicemocks.MockStsGetCallerIdentityValidEndpoint,
			},
		},
		"config AssumeRoleExternalID": {
			Config: &Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				AssumeRole: []AssumeRole{{
					RoleARN:     servicemocks.MockStsAssumeRoleArn,
					ExternalID:  servicemocks.MockStsAssumeRoleExternalId,
					SessionName: servicemocks.MockStsAssumeRoleSessionName,
				}},
				Region:    "us-east-1",
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			ExpectedCredentialsValue: mockdata.MockStsAssumeRoleCredentials,
			ExpectedRegion:           "us-east-1",
			MockStsEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsAssumeRoleValidEndpointWithOptions(map[string]string{"ExternalId": servicemocks.MockStsAssumeRoleExternalId}),
				servicemocks.MockStsGetCallerIdentityValidEndpoint,
			},
		},
		"config AssumeRolePolicy": {
			Config: &Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				AssumeRole: []AssumeRole{{
					RoleARN:     servicemocks.MockStsAssumeRoleArn,
					Policy:      servicemocks.MockStsAssumeRolePolicy,
					SessionName: servicemocks.MockStsAssumeRoleSessionName,
				}},
				Region:    "us-east-1",
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			ExpectedCredentialsValue: mockdata.MockStsAssumeRoleCredentials,
			ExpectedRegion:           "us-east-1",
			MockStsEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsAssumeRoleValidEndpointWithOptions(map[string]string{"Policy": servicemocks.MockStsAssumeRolePolicy}),
				servicemocks.MockStsGetCallerIdentityValidEndpoint,
			},
		},
		"config AssumeRolePolicyARNs": {
			Config: &Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				AssumeRole: []AssumeRole{{
					RoleARN:     servicemocks.MockStsAssumeRoleArn,
					PolicyARNs:  []string{servicemocks.MockStsAssumeRolePolicyArn},
					SessionName: servicemocks.MockStsAssumeRoleSessionName,
				}},
				Region:    "us-east-1",
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			ExpectedCredentialsValue: mockdata.MockStsAssumeRoleCredentials,
			ExpectedRegion:           "us-east-1",
			MockStsEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsAssumeRoleValidEndpointWithOptions(map[string]string{"PolicyArns.member.1.arn": servicemocks.MockStsAssumeRolePolicyArn}),
				servicemocks.MockStsGetCallerIdentityValidEndpoint,
			},
		},
		"config AssumeRoleTags": {
			Config: &Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				AssumeRole: []AssumeRole{{
					RoleARN:     servicemocks.MockStsAssumeRoleArn,
					SessionName: servicemocks.MockStsAssumeRoleSessionName,
					Tags: map[string]string{
						servicemocks.MockStsAssumeRoleTagKey: servicemocks.MockStsAssumeRoleTagValue,
					},
				}},
				Region:    "us-east-1",
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			ExpectedCredentialsValue: mockdata.MockStsAssumeRoleCredentials,
			ExpectedRegion:           "us-east-1",
			MockStsEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsAssumeRoleValidEndpointWithOptions(map[string]string{"Tags.member.1.Key": servicemocks.MockStsAssumeRoleTagKey, "Tags.member.1.Value": servicemocks.MockStsAssumeRoleTagValue}),
				servicemocks.MockStsGetCallerIdentityValidEndpoint,
			},
		},
		"config AssumeRoleTransitiveTagKeys": {
			Config: &Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				AssumeRole: []AssumeRole{{
					RoleARN:     servicemocks.MockStsAssumeRoleArn,
					SessionName: servicemocks.MockStsAssumeRoleSessionName,
					Tags: map[string]string{
						servicemocks.MockStsAssumeRoleTagKey: servicemocks.MockStsAssumeRoleTagValue,
					},
					TransitiveTagKeys: []string{servicemocks.MockStsAssumeRoleTagKey},
				}},
				Region:    "us-east-1",
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			ExpectedCredentialsValue: mockdata.MockStsAssumeRoleCredentials,
			ExpectedRegion:           "us-east-1",
			MockStsEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsAssumeRoleValidEndpointWithOptions(map[string]string{"Tags.member.1.Key": servicemocks.MockStsAssumeRoleTagKey, "Tags.member.1.Value": servicemocks.MockStsAssumeRoleTagValue, "TransitiveTagKeys.member.1": servicemocks.MockStsAssumeRoleTagKey}),
				servicemocks.MockStsGetCallerIdentityValidEndpoint,
			},
		},
		"config AssumeRoleSourceIdentity": {
			Config: &Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				AssumeRole: []AssumeRole{{
					RoleARN:        servicemocks.MockStsAssumeRoleArn,
					SessionName:    servicemocks.MockStsAssumeRoleSessionName,
					SourceIdentity: servicemocks.MockStsAssumeRoleSourceIdentity,
				}},
				Region:    "us-east-1",
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			ExpectedCredentialsValue: mockdata.MockStsAssumeRoleCredentials,
			ExpectedRegion:           "us-east-1",
			MockStsEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsAssumeRoleValidEndpointWithOptions(map[string]string{"SourceIdentity": servicemocks.MockStsAssumeRoleSourceIdentity}),
				servicemocks.MockStsGetCallerIdentityValidEndpoint,
			},
		},
		"config Profile shared credentials profile aws_access_key_id": {
			Config: &Config{
				Profile: "SharedCredentialsProfile",
				Region:  "us-east-1",
			},
			ExpectedCredentialsValue: aws.Credentials{
				AccessKeyID:     "ProfileSharedCredentialsAccessKey",
				SecretAccessKey: "ProfileSharedCredentialsSecretKey",
				Source:          sharedConfigCredentialsProvider,
			},
			ExpectedRegion: "us-east-1",
			MockStsEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsGetCallerIdentityValidEndpoint,
			},
			SharedCredentialsFile: `
[default]
aws_access_key_id = DefaultSharedCredentialsAccessKey
aws_secret_access_key = DefaultSharedCredentialsSecretKey

[SharedCredentialsProfile]
aws_access_key_id = ProfileSharedCredentialsAccessKey
aws_secret_access_key = ProfileSharedCredentialsSecretKey
`,
		},
		"config Profile shared configuration credential_source Ec2InstanceMetadata": {
			Config: &Config{
				Profile: "SharedConfigurationProfile",
				Region:  "us-east-1",
			},
			EnableEc2MetadataServer:  true,
			ExpectedCredentialsValue: mockdata.MockStsAssumeRoleCredentials,
			ExpectedRegion:           "us-east-1",
			MockStsEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsAssumeRoleValidEndpoint,
				servicemocks.MockStsGetCallerIdentityValidEndpoint,
			},
			SharedConfigurationFile: fmt.Sprintf(`
[profile SharedConfigurationProfile]
credential_source = Ec2InstanceMetadata
role_arn = %[1]s
role_session_name = %[2]s
`, servicemocks.MockStsAssumeRoleArn, servicemocks.MockStsAssumeRoleSessionName),
		},
		// 		"config Profile shared configuration credential_source EcsContainer": {
		// 			Config: &Config{
		// 				Profile: "SharedConfigurationProfile",
		// 				Region:  "us-east-1",
		// 			},
		// 			EnvironmentVariables: map[string]string{
		// 				"AWS_CONTAINER_CREDENTIALS_RELATIVE_URI": "/creds",
		// 			},
		// 			EnableEc2MetadataServer:    true,
		// 			EnableEcsCredentialsServer: true,
		// 			ExpectedCredentialsValue:   mockdata.MockStsAssumeRoleCredentialsV2,
		// 			ExpectedRegion:             "us-east-1",
		// 			MockStsEndpoints: []*servicemocks.MockEndpoint{
		// 				servicemocks.MockStsAssumeRoleValidEndpoint,
		// 				servicemocks.MockStsGetCallerIdentityValidEndpoint,
		// 			},
		// 			SharedConfigurationFile: fmt.Sprintf(`
		// [profile SharedConfigurationProfile]
		// credential_source = EcsContainer
		// role_arn = %[1]s
		// role_session_name = %[2]s
		// `, servicemocks.MockStsAssumeRoleArn, servicemocks.MockStsAssumeRoleSessionName),
		// 		},
		"config Profile shared configuration source_profile": {
			Config: &Config{
				Profile: "SharedConfigurationProfile",
				Region:  "us-east-1",
			},
			ExpectedCredentialsValue: mockdata.MockStsAssumeRoleCredentials,
			ExpectedRegion:           "us-east-1",
			MockStsEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsAssumeRoleValidEndpoint,
				servicemocks.MockStsGetCallerIdentityValidEndpoint,
			},
			SharedConfigurationFile: fmt.Sprintf(`
[profile SharedConfigurationProfile]
role_arn = %[1]s
role_session_name = %[2]s
source_profile = SharedConfigurationSourceProfile

[profile SharedConfigurationSourceProfile]
aws_access_key_id = SharedConfigurationSourceAccessKey
aws_secret_access_key = SharedConfigurationSourceSecretKey
`, servicemocks.MockStsAssumeRoleArn, servicemocks.MockStsAssumeRoleSessionName),
		},
		"environment AWS_ACCESS_KEY_ID": {
			Config: &Config{
				Region: "us-east-1",
			},
			EnvironmentVariables: map[string]string{
				"AWS_ACCESS_KEY_ID":     servicemocks.MockEnvAccessKey,
				"AWS_SECRET_ACCESS_KEY": servicemocks.MockEnvSecretKey,
			},
			ExpectedCredentialsValue: mockdata.MockEnvCredentials,
			ExpectedRegion:           "us-east-1",
			MockStsEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsGetCallerIdentityValidEndpoint,
			},
		},
		"environment AWS_ACCESS_KEY_ID config AssumeRoleARN access key": {
			Config: &Config{
				AssumeRole: []AssumeRole{{
					RoleARN:     servicemocks.MockStsAssumeRoleArn,
					SessionName: servicemocks.MockStsAssumeRoleSessionName,
				}},
				Region: "us-east-1",
			},
			EnvironmentVariables: map[string]string{
				"AWS_ACCESS_KEY_ID":     servicemocks.MockEnvAccessKey,
				"AWS_SECRET_ACCESS_KEY": servicemocks.MockEnvSecretKey,
			},
			ExpectedCredentialsValue: mockdata.MockStsAssumeRoleCredentials,
			ExpectedRegion:           "us-east-1",
			MockStsEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsAssumeRoleValidEndpoint,
				servicemocks.MockStsGetCallerIdentityValidEndpoint,
			},
		},
		"environment AWS_PROFILE shared credentials profile aws_access_key_id": {
			Config: &Config{
				Region: "us-east-1",
			},
			EnvironmentVariables: map[string]string{
				"AWS_PROFILE": "SharedCredentialsProfile",
			},
			ExpectedCredentialsValue: aws.Credentials{
				AccessKeyID:     "ProfileSharedCredentialsAccessKey",
				SecretAccessKey: "ProfileSharedCredentialsSecretKey",
				Source:          sharedConfigCredentialsProvider,
			},
			ExpectedRegion: "us-east-1",
			MockStsEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsGetCallerIdentityValidEndpoint,
			},
			SharedCredentialsFile: `
[default]
aws_access_key_id = DefaultSharedCredentialsAccessKey
aws_secret_access_key = DefaultSharedCredentialsSecretKey

[SharedCredentialsProfile]
aws_access_key_id = ProfileSharedCredentialsAccessKey
aws_secret_access_key = ProfileSharedCredentialsSecretKey
`,
		},
		"environment AWS_PROFILE shared configuration credential_source Ec2InstanceMetadata": {
			Config: &Config{
				Region: "us-east-1",
			},
			EnableEc2MetadataServer: true,
			EnvironmentVariables: map[string]string{
				"AWS_PROFILE": "SharedConfigurationProfile",
			},
			ExpectedCredentialsValue: mockdata.MockStsAssumeRoleCredentials,
			ExpectedRegion:           "us-east-1",
			MockStsEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsAssumeRoleValidEndpoint,
				servicemocks.MockStsGetCallerIdentityValidEndpoint,
			},
			SharedConfigurationFile: fmt.Sprintf(`
[profile SharedConfigurationProfile]
credential_source = Ec2InstanceMetadata
role_arn = %[1]s
role_session_name = %[2]s
`, servicemocks.MockStsAssumeRoleArn, servicemocks.MockStsAssumeRoleSessionName),
		},
		// 		"environment AWS_PROFILE shared configuration credential_source EcsContainer": {
		// 			Config: &Config{
		// 				Region: "us-east-1",
		// 			},
		// 			EnableEc2MetadataServer:    true,
		// 			EnableEcsCredentialsServer: true,
		// 			EnvironmentVariables: map[string]string{
		// 				"AWS_CONTAINER_CREDENTIALS_RELATIVE_URI": "/creds",
		// 				"AWS_PROFILE":                            "SharedConfigurationProfile",
		// 			},
		// 			ExpectedCredentialsValue: mockdata.MockStsAssumeRoleCredentialsV2,
		// 			ExpectedRegion:           "us-east-1",
		// 			MockStsEndpoints: []*servicemocks.MockEndpoint{
		// 				servicemocks.MockStsAssumeRoleValidEndpoint,
		// 				servicemocks.MockStsGetCallerIdentityValidEndpoint,
		// 			},
		// 			SharedConfigurationFile: fmt.Sprintf(`
		// [profile SharedConfigurationProfile]
		// credential_source = EcsContainer
		// role_arn = %[1]s
		// role_session_name = %[2]s
		// `, servicemocks.MockStsAssumeRoleArn, servicemocks.MockStsAssumeRoleSessionName),
		// 		},
		"environment AWS_PROFILE shared configuration source_profile": {
			Config: &Config{
				Region: "us-east-1",
			},
			EnvironmentVariables: map[string]string{
				"AWS_PROFILE": "SharedConfigurationProfile",
			},
			ExpectedCredentialsValue: mockdata.MockStsAssumeRoleCredentials,
			ExpectedRegion:           "us-east-1",
			MockStsEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsAssumeRoleValidEndpoint,
				servicemocks.MockStsGetCallerIdentityValidEndpoint,
			},
			SharedConfigurationFile: fmt.Sprintf(`
[profile SharedConfigurationProfile]
role_arn = %[1]s
role_session_name = %[2]s
source_profile = SharedConfigurationSourceProfile

[profile SharedConfigurationSourceProfile]
aws_access_key_id = SharedConfigurationSourceAccessKey
aws_secret_access_key = SharedConfigurationSourceSecretKey
`, servicemocks.MockStsAssumeRoleArn, servicemocks.MockStsAssumeRoleSessionName),
		},
		"environment AWS_SESSION_TOKEN": {
			Config: &Config{
				Region: "us-east-1",
			},
			EnvironmentVariables: map[string]string{
				"AWS_ACCESS_KEY_ID":     servicemocks.MockEnvAccessKey,
				"AWS_SECRET_ACCESS_KEY": servicemocks.MockEnvSecretKey,
				"AWS_SESSION_TOKEN":     servicemocks.MockEnvSessionToken,
			},
			ExpectedCredentialsValue: mockdata.MockEnvCredentialsWithSessionToken,
			ExpectedRegion:           "us-east-1",
			MockStsEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsGetCallerIdentityValidEndpoint,
			},
		},
		"shared credentials default aws_access_key_id": {
			Config: &Config{
				Region: "us-east-1",
			},
			ExpectedCredentialsValue: aws.Credentials{
				AccessKeyID:     "DefaultSharedCredentialsAccessKey",
				SecretAccessKey: "DefaultSharedCredentialsSecretKey",
				Source:          sharedConfigCredentialsProvider,
			},
			ExpectedRegion: "us-east-1",
			MockStsEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsGetCallerIdentityValidEndpoint,
			},
			SharedCredentialsFile: `
[default]
aws_access_key_id = DefaultSharedCredentialsAccessKey
aws_secret_access_key = DefaultSharedCredentialsSecretKey
`,
		},
		"shared credentials default aws_access_key_id config AssumeRoleARN access key": {
			Config: &Config{
				AssumeRole: []AssumeRole{{
					RoleARN:     servicemocks.MockStsAssumeRoleArn,
					SessionName: servicemocks.MockStsAssumeRoleSessionName,
				}},
				Region: "us-east-1",
			},
			ExpectedCredentialsValue: mockdata.MockStsAssumeRoleCredentials,
			ExpectedRegion:           "us-east-1",
			MockStsEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsAssumeRoleValidEndpoint,
				servicemocks.MockStsGetCallerIdentityValidEndpoint,
			},
			SharedCredentialsFile: `
[default]
aws_access_key_id = DefaultSharedCredentialsAccessKey
aws_secret_access_key = DefaultSharedCredentialsSecretKey
`,
		},
		"web identity token access key": {
			Config: &Config{
				Region: "us-east-1",
			},
			EnableWebIdentityEnvVars: true,
			ExpectedCredentialsValue: mockdata.MockStsAssumeRoleWithWebIdentityCredentials,
			ExpectedRegion:           "us-east-1",
			MockStsEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsAssumeRoleWithWebIdentityValidEndpoint,
				servicemocks.MockStsGetCallerIdentityValidEndpoint,
			},
		},
		"EC2 metadata access key": {
			Config: &Config{
				Region: "us-east-1",
			},
			EnableEc2MetadataServer:  true,
			ExpectedCredentialsValue: mockdata.MockEc2MetadataCredentials,
			ExpectedRegion:           "us-east-1",
			MockStsEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsGetCallerIdentityValidEndpoint,
			},
		},
		"EC2 metadata access key config AssumeRoleARN access key": {
			Config: &Config{
				AssumeRole: []AssumeRole{{
					RoleARN:     servicemocks.MockStsAssumeRoleArn,
					SessionName: servicemocks.MockStsAssumeRoleSessionName,
				}},
				Region: "us-east-1",
			},
			EnableEc2MetadataServer:  true,
			ExpectedCredentialsValue: mockdata.MockStsAssumeRoleCredentials,
			ExpectedRegion:           "us-east-1",
			MockStsEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsAssumeRoleValidEndpoint,
				servicemocks.MockStsGetCallerIdentityValidEndpoint,
			},
		},
		"ECS credentials access key": {
			Config: &Config{
				Region: "us-east-1",
			},
			EnableEcsCredentialsServer: true,
			ExpectedCredentialsValue:   mockdata.MockEcsCredentialsCredentials,
			ExpectedRegion:             "us-east-1",
			MockStsEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsGetCallerIdentityValidEndpoint,
			},
		},
		"ECS credentials access key config AssumeRoleARN access key": {
			Config: &Config{
				AssumeRole: []AssumeRole{{
					RoleARN:     servicemocks.MockStsAssumeRoleArn,
					SessionName: servicemocks.MockStsAssumeRoleSessionName,
				}},
				Region: "us-east-1",
			},
			EnableEcsCredentialsServer: true,
			ExpectedCredentialsValue:   mockdata.MockStsAssumeRoleCredentials,
			ExpectedRegion:             "us-east-1",
			MockStsEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsAssumeRoleValidEndpoint,
				servicemocks.MockStsGetCallerIdentityValidEndpoint,
			},
		},
		"AssumeWebIdentity envvar AssumeRoleARN access key": {
			Config: &Config{
				AssumeRole: []AssumeRole{{
					RoleARN:     servicemocks.MockStsAssumeRoleArn,
					SessionName: servicemocks.MockStsAssumeRoleSessionName,
				}},
				Region: "us-east-1",
			},
			EnableWebIdentityEnvVars: true,
			ExpectedCredentialsValue: mockdata.MockStsAssumeRoleCredentials,
			ExpectedRegion:           "us-east-1",
			MockStsEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsAssumeRoleWithWebIdentityValidEndpoint,
				servicemocks.MockStsAssumeRoleValidEndpoint,
				servicemocks.MockStsGetCallerIdentityValidEndpoint,
			},
		},
		"AssumeWebIdentity config AssumeRoleARN access key": {
			Config: &Config{
				AssumeRole: []AssumeRole{{
					RoleARN:     servicemocks.MockStsAssumeRoleArn,
					SessionName: servicemocks.MockStsAssumeRoleSessionName,
				}},
				Region: "us-east-1",
			},
			EnableWebIdentityConfig:  true,
			ExpectedCredentialsValue: mockdata.MockStsAssumeRoleCredentials,
			ExpectedRegion:           "us-east-1",
			MockStsEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsAssumeRoleWithWebIdentityValidEndpoint,
				servicemocks.MockStsAssumeRoleValidEndpoint,
				servicemocks.MockStsGetCallerIdentityValidEndpoint,
			},
		},
		"config AccessKey over environment AWS_ACCESS_KEY_ID": {
			Config: &Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				Region:    "us-east-1",
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			EnvironmentVariables: map[string]string{
				"AWS_ACCESS_KEY_ID":     servicemocks.MockEnvAccessKey,
				"AWS_SECRET_ACCESS_KEY": servicemocks.MockEnvSecretKey,
			},
			ExpectedCredentialsValue: mockdata.MockStaticCredentials,
			ExpectedRegion:           "us-east-1",
			MockStsEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsGetCallerIdentityValidEndpoint,
			},
		},
		"config AccessKey over shared credentials default aws_access_key_id": {
			Config: &Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				Region:    "us-east-1",
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			ExpectedCredentialsValue: mockdata.MockStaticCredentials,
			ExpectedRegion:           "us-east-1",
			MockStsEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsGetCallerIdentityValidEndpoint,
			},
			SharedCredentialsFile: `
[default]
aws_access_key_id = DefaultSharedCredentialsAccessKey
aws_secret_access_key = DefaultSharedCredentialsSecretKey
`,
		},
		"config AccessKey over EC2 metadata access key": {
			Config: &Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				Region:    "us-east-1",
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			ExpectedCredentialsValue: mockdata.MockStaticCredentials,
			ExpectedRegion:           "us-east-1",
			MockStsEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsGetCallerIdentityValidEndpoint,
			},
		},
		"config AccessKey over ECS credentials access key": {
			Config: &Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				Region:    "us-east-1",
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			EnableEcsCredentialsServer: true,
			ExpectedCredentialsValue:   mockdata.MockStaticCredentials,
			ExpectedRegion:             "us-east-1",
			MockStsEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsGetCallerIdentityValidEndpoint,
			},
		},
		"environment AWS_ACCESS_KEY_ID over shared credentials default aws_access_key_id": {
			Config: &Config{
				Region: "us-east-1",
			},
			EnvironmentVariables: map[string]string{
				"AWS_ACCESS_KEY_ID":     servicemocks.MockEnvAccessKey,
				"AWS_SECRET_ACCESS_KEY": servicemocks.MockEnvSecretKey,
			},
			ExpectedCredentialsValue: mockdata.MockEnvCredentials,
			ExpectedRegion:           "us-east-1",
			MockStsEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsGetCallerIdentityValidEndpoint,
			},
			SharedCredentialsFile: `
[default]
aws_access_key_id = DefaultSharedCredentialsAccessKey
aws_secret_access_key = DefaultSharedCredentialsSecretKey
`,
		},
		"environment AWS_ACCESS_KEY_ID over EC2 metadata access key": {
			Config: &Config{
				Region: "us-east-1",
			},
			EnvironmentVariables: map[string]string{
				"AWS_ACCESS_KEY_ID":     servicemocks.MockEnvAccessKey,
				"AWS_SECRET_ACCESS_KEY": servicemocks.MockEnvSecretKey,
			},
			ExpectedCredentialsValue: mockdata.MockEnvCredentials,
			ExpectedRegion:           "us-east-1",
			MockStsEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsGetCallerIdentityValidEndpoint,
			},
		},
		"environment AWS_ACCESS_KEY_ID over ECS credentials access key": {
			Config: &Config{
				Region: "us-east-1",
			},
			EnvironmentVariables: map[string]string{
				"AWS_ACCESS_KEY_ID":     servicemocks.MockEnvAccessKey,
				"AWS_SECRET_ACCESS_KEY": servicemocks.MockEnvSecretKey,
			},
			EnableEcsCredentialsServer: true,
			ExpectedCredentialsValue:   mockdata.MockEnvCredentials,
			ExpectedRegion:             "us-east-1",
			MockStsEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsGetCallerIdentityValidEndpoint,
			},
		},
		"shared credentials default aws_access_key_id over EC2 metadata access key": {
			Config: &Config{
				Region: "us-east-1",
			},
			ExpectedCredentialsValue: aws.Credentials{
				AccessKeyID:     "DefaultSharedCredentialsAccessKey",
				SecretAccessKey: "DefaultSharedCredentialsSecretKey",
				Source:          sharedConfigCredentialsProvider,
			},
			ExpectedRegion: "us-east-1",
			MockStsEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsGetCallerIdentityValidEndpoint,
			},
			SharedCredentialsFile: `
[default]
aws_access_key_id = DefaultSharedCredentialsAccessKey
aws_secret_access_key = DefaultSharedCredentialsSecretKey
`,
		},
		"shared credentials default aws_access_key_id over ECS credentials access key": {
			Config: &Config{
				Region: "us-east-1",
			},
			EnableEcsCredentialsServer: true,
			ExpectedCredentialsValue: aws.Credentials{
				AccessKeyID:     "DefaultSharedCredentialsAccessKey",
				SecretAccessKey: "DefaultSharedCredentialsSecretKey",
				Source:          sharedConfigCredentialsProvider,
			},
			ExpectedRegion: "us-east-1",
			MockStsEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsGetCallerIdentityValidEndpoint,
			},
			SharedCredentialsFile: `
[default]
aws_access_key_id = DefaultSharedCredentialsAccessKey
aws_secret_access_key = DefaultSharedCredentialsSecretKey
`,
		},
		"ECS credentials access key over EC2 metadata access key": {
			Config: &Config{
				Region: "us-east-1",
			},
			EnableEcsCredentialsServer: true,
			ExpectedCredentialsValue:   mockdata.MockEcsCredentialsCredentials,
			ExpectedRegion:             "us-east-1",
			MockStsEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsGetCallerIdentityValidEndpoint,
			},
		},
		"retrieve region from shared configuration file": {
			Config: &Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			ExpectedCredentialsValue: mockdata.MockStaticCredentials,
			ExpectedRegion:           "us-east-1",
			MockStsEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsGetCallerIdentityValidEndpoint,
			},
			SharedConfigurationFile: `
[default]
region = us-east-1
`,
		},
		"assume role error": {
			Config: &Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				AssumeRole: []AssumeRole{{
					RoleARN:     servicemocks.MockStsAssumeRoleArn,
					SessionName: servicemocks.MockStsAssumeRoleSessionName,
				}},
				Region:    "us-east-1",
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			ValidateDiags:  test.ExpectDiagValidator("CannotAssumeRoleError", IsCannotAssumeRoleError),
			ExpectedRegion: "us-east-1",
			MockStsEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsAssumeRoleInvalidEndpointInvalidClientTokenId,
				servicemocks.MockStsGetCallerIdentityValidEndpoint,
			},
		},
		// "credential validation error": {
		// 	Config: &Config{
		// 		AccessKey: servicemocks.MockStaticAccessKey,
		// 		Region:    "us-east-1",
		// 		SecretKey: servicemocks.MockStaticSecretKey,
		// 	},
		// 	ExpectedError: func(err error) bool {
		// 		return tfawserr.ErrCodeEquals(err, "AccessDenied")
		// 	},
		// 	MockStsEndpoints: []*servicemocks.MockEndpoint{
		// 		servicemocks.MockStsGetCallerIdentityInvalidEndpointAccessDenied,
		// 	},
		// },
		"session creation error": {
			Config: &Config{
				Profile: "SharedConfigurationProfile",
				Region:  "us-east-1",
			},
			ValidateDiags: test.ExpectErrDiagValidator("CredentialRequiresARNError", func(err error) bool {
				var e config.CredentialRequiresARNError
				return errors.As(err, &e)
			}),
			SharedConfigurationFile: `
[profile SharedConfigurationProfile]
source_profile = SourceSharedCredentials
`,
		},
		"skip credentials validation": {
			Config: &Config{
				AccessKey:           servicemocks.MockStaticAccessKey,
				Region:              "us-east-1",
				SecretKey:           servicemocks.MockStaticSecretKey,
				SkipCredsValidation: true,
			},
			ExpectedCredentialsValue: mockdata.MockStaticCredentials,
			ExpectedRegion:           "us-east-1",
		},
		"skip EC2 Metadata API check": {
			Config: &Config{
				Region:                        "us-east-1",
				EC2MetadataServiceEnableState: imds.ClientDisabled,
			},
			ValidateDiags:  test.ExpectDiagValidator("NoValidCredentialSourcesError", IsNoValidCredentialSourcesError),
			ExpectedRegion: "us-east-1",
			// The IMDS server must be enabled so that auth will succeed if the IMDS is called
			EnableEc2MetadataServer: true,
			MockStsEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsGetCallerIdentityValidEndpoint,
			},
		},
		"invalid profile name from envvar": {
			Config: &Config{
				Region: "us-east-1",
			},
			EnvironmentVariables: map[string]string{
				"AWS_PROFILE": "no-such-profile",
			},
			ValidateDiags: test.ExpectErrDiagValidator("SharedConfigProfileNotExistError", func(err error) bool {
				var e config.SharedConfigProfileNotExistError
				return errors.As(err, &e)
			}),
			SharedCredentialsFile: `
[some-profile]
aws_access_key_id = DefaultSharedCredentialsAccessKey
aws_secret_access_key = DefaultSharedCredentialsSecretKey
`,
		},
		"invalid profile name from config": {
			Config: &Config{
				Profile: "no-such-profile",
				Region:  "us-east-1",
			},
			ValidateDiags: test.ExpectErrDiagValidator("SharedConfigProfileNotExistError", func(err error) bool {
				var e config.SharedConfigProfileNotExistError
				return errors.As(err, &e)
			}),
			SharedCredentialsFile: `
[some-profile]
aws_access_key_id = DefaultSharedCredentialsAccessKey
aws_secret_access_key = DefaultSharedCredentialsSecretKey
`,
		},
		"AWS_ACCESS_KEY_ID overrides AWS_PROFILE": {
			Config: &Config{},
			EnvironmentVariables: map[string]string{
				"AWS_ACCESS_KEY_ID":     servicemocks.MockEnvAccessKey,
				"AWS_SECRET_ACCESS_KEY": servicemocks.MockEnvSecretKey,
				"AWS_PROFILE":           "SharedCredentialsProfile",
			},
			SharedCredentialsFile: `
[default]
aws_access_key_id = DefaultSharedCredentialsAccessKey
aws_secret_access_key = DefaultSharedCredentialsSecretKey

[SharedCredentialsProfile]
aws_access_key_id = ProfileSharedCredentialsAccessKey
aws_secret_access_key = ProfileSharedCredentialsSecretKey
`,
			MockStsEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsGetCallerIdentityValidEndpoint,
			},
			ExpectedCredentialsValue: mockdata.MockEnvCredentials,
		},
		"AWS_ACCESS_KEY_ID does not override invalid profile name from envvar": {
			Config: &Config{
				Region: "us-east-1",
			},
			EnvironmentVariables: map[string]string{
				"AWS_ACCESS_KEY_ID":     servicemocks.MockEnvAccessKey,
				"AWS_SECRET_ACCESS_KEY": servicemocks.MockEnvSecretKey,
				"AWS_PROFILE":           "no-such-profile",
			},
			ValidateDiags: test.ExpectErrDiagValidator("SharedConfigProfileNotExistError", func(err error) bool {
				var e config.SharedConfigProfileNotExistError
				return errors.As(err, &e)
			}),
			SharedCredentialsFile: `
[some-profile]
aws_access_key_id = DefaultSharedCredentialsAccessKey
aws_secret_access_key = DefaultSharedCredentialsSecretKey
`,
		},
		"environment AWS_ACCESS_KEY_ID does not override config Profile": {
			Config: &Config{
				Profile: "SharedCredentialsProfile",
				Region:  "us-east-1",
			},
			EnvironmentVariables: map[string]string{
				"AWS_ACCESS_KEY_ID":     servicemocks.MockEnvAccessKey,
				"AWS_SECRET_ACCESS_KEY": servicemocks.MockEnvSecretKey,
			},
			ExpectedCredentialsValue: aws.Credentials{
				AccessKeyID:     "ProfileSharedCredentialsAccessKey",
				SecretAccessKey: "ProfileSharedCredentialsSecretKey",
				Source:          sharedConfigCredentialsProvider,
			},
			ExpectedRegion: "us-east-1",
			MockStsEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsGetCallerIdentityValidEndpoint,
			},
			SharedCredentialsFile: `
[default]
aws_access_key_id = DefaultSharedCredentialsAccessKey
aws_secret_access_key = DefaultSharedCredentialsSecretKey

[SharedCredentialsProfile]
aws_access_key_id = ProfileSharedCredentialsAccessKey
aws_secret_access_key = ProfileSharedCredentialsSecretKey
`,
		},
	}

	for testName, testCase := range testCases {
		if testCase.ValidateDiags == nil {
			testCase.ValidateDiags = test.ExpectNoDiags
		}

		t.Run(testName, func(t *testing.T) {
			servicemocks.InitSessionTestEnv(t)

			if testCase.EnableEc2MetadataServer {
				closeEc2Metadata := servicemocks.AwsMetadataApiMock(append(
					servicemocks.Ec2metadata_securityCredentialsEndpoints,
					servicemocks.Ec2metadata_instanceIdEndpoint,
					servicemocks.Ec2metadata_iamInfoEndpoint,
				))
				defer closeEc2Metadata()
			}

			if testCase.EnableEcsCredentialsServer {
				closeEcsCredentials := servicemocks.EcsCredentialsApiMock()
				defer closeEcsCredentials()
			}

			if testCase.EnableWebIdentityEnvVars || testCase.EnableWebIdentityConfig {
				file, err := os.CreateTemp("", "aws-sdk-go-base-web-identity-token-file")
				if err != nil {
					t.Fatalf("unexpected error creating temporary web identity token file: %s", err)
				}

				defer os.Remove(file.Name())

				err = os.WriteFile(file.Name(), []byte(servicemocks.MockWebIdentityToken), 0600)

				if err != nil {
					t.Fatalf("unexpected error writing web identity token file: %s", err)
				}

				if testCase.EnableWebIdentityEnvVars {
					t.Setenv("AWS_ROLE_ARN", servicemocks.MockStsAssumeRoleWithWebIdentityArn)
					t.Setenv("AWS_ROLE_SESSION_NAME", servicemocks.MockStsAssumeRoleWithWebIdentitySessionName)
					t.Setenv("AWS_WEB_IDENTITY_TOKEN_FILE", file.Name())
				} else if testCase.EnableWebIdentityConfig {
					testCase.Config.AssumeRoleWithWebIdentity = &AssumeRoleWithWebIdentity{
						RoleARN:              servicemocks.MockStsAssumeRoleWithWebIdentityArn,
						SessionName:          servicemocks.MockStsAssumeRoleWithWebIdentitySessionName,
						WebIdentityTokenFile: file.Name(),
					}
				}
			}

			closeSts, _, stsEndpoint := mockdata.GetMockedAwsApiSession("STS", testCase.MockStsEndpoints)
			defer closeSts()

			testCase.Config.StsEndpoint = stsEndpoint

			if testCase.SharedConfigurationFile != "" {
				file, err := os.CreateTemp("", "aws-sdk-go-base-shared-configuration-file")

				if err != nil {
					t.Fatalf("unexpected error creating temporary shared configuration file: %s", err)
				}

				defer os.Remove(file.Name())

				err = os.WriteFile(file.Name(), []byte(testCase.SharedConfigurationFile), 0600)

				if err != nil {
					t.Fatalf("unexpected error writing shared configuration file: %s", err)
				}

				testCase.Config.SharedConfigFiles = []string{file.Name()}
			}

			if testCase.SharedCredentialsFile != "" {
				file, err := os.CreateTemp("", "aws-sdk-go-base-shared-credentials-file")

				if err != nil {
					t.Fatalf("unexpected error creating temporary shared credentials file: %s", err)
				}

				defer os.Remove(file.Name())

				err = os.WriteFile(file.Name(), []byte(testCase.SharedCredentialsFile), 0600)

				if err != nil {
					t.Fatalf("unexpected error writing shared credentials file: %s", err)
				}

				testCase.Config.SharedCredentialsFiles = []string{file.Name()}
				if testCase.ExpectedCredentialsValue.Source == sharedConfigCredentialsProvider {
					testCase.ExpectedCredentialsValue.Source = sharedConfigCredentialsSource(file.Name())
				}
			}

			for k, v := range testCase.EnvironmentVariables {
				t.Setenv(k, v)
			}

			ctx, awsConfig, diags := GetAwsConfig(context.Background(), testCase.Config)

			testCase.ValidateDiags(t, diags)
			if diags.HasError() {
				return
			}

			credentialsValue, err := awsConfig.Credentials.Retrieve(ctx)

			if err != nil {
				t.Fatalf("unexpected credentials Retrieve() error: %s", err)
			}

			if diff := cmp.Diff(credentialsValue, testCase.ExpectedCredentialsValue, cmpopts.IgnoreFields(aws.Credentials{}, "Expires")); diff != "" {
				t.Fatalf("unexpected credentials: (- got, + expected)\n%s", diff)
			}
			// TODO: test credentials.Expires

			if expected, actual := testCase.ExpectedRegion, awsConfig.Region; expected != actual {
				t.Fatalf("expected region (%s), got: %s", expected, actual)
			}
		})
	}
}

func TestUserAgentProducts(t *testing.T) {
	test.TestUserAgentProducts(t, awsSdkGoUserAgent, testUserAgentProducts)
}

func testUserAgentProducts(t *testing.T, testCase test.UserAgentTestCase) {
	var (
		httpUserAgent string
		httpSdkAgent  string
	)

	readUserAgent := cancelRequestMiddleware(t, "ReadUserAgent", func(_ *testing.T, request *smithyhttp.Request) {
		httpUserAgent = request.UserAgent()
		httpSdkAgent = request.Header.Get("X-Amz-User-Agent")
	})

	ctx, awsConfig, diags := GetAwsConfig(context.Background(), testCase.Config)
	if diags.HasError() {
		t.Fatalf("error in GetAwsConfig(): %v", diags)
	}

	client := stsClient(ctx, awsConfig, testCase.Config)

	if testCase.Context != nil {
		ctx = useragent.Context(ctx, testCase.Context)
	}

	_, err := client.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{},
		func(opts *sts.Options) {
			opts.APIOptions = append(opts.APIOptions, func(stack *middleware.Stack) error {
				return stack.Finalize.Add(readUserAgent, middleware.Before)
			})
		},
	)
	if err == nil {
		t.Fatal("Expected an error, got none")
	} else if !errors.Is(err, errCancelOperation) {
		t.Fatalf("Unexpected error: %s", err)
	}

	cleanedUserAgent := cleanUserAgent(httpUserAgent)

	if testCase.ExpectedUserAgent != cleanedUserAgent {
		t.Errorf("expected User-Agent %q, got %q", testCase.ExpectedUserAgent, cleanedUserAgent)
	}

	// The header X-Amz-User-Agent was disabled but not removed in v1.3.0 (2021-03-18)
	if httpSdkAgent != "" {
		t.Errorf("expected header X-Amz-User-Agent to not be set, got %q", httpSdkAgent)
	}
}

var errCancelOperation = errors.New("Test: Cancelling request")

// cancelRequestMiddleware creates a Smithy middleware that intercepts the request before sending and cancels it
func cancelRequestMiddleware(t *testing.T, id string, f func(t *testing.T, request *smithyhttp.Request)) middleware.FinalizeMiddleware {
	return middleware.FinalizeMiddlewareFunc(
		fmt.Sprintf("Test: Cancel Request: %s", id),
		func(_ context.Context, in middleware.FinalizeInput, next middleware.FinalizeHandler) (middleware.FinalizeOutput, middleware.Metadata, error) {
			request, ok := in.Request.(*smithyhttp.Request)
			if !ok {
				t.Fatalf("Expected *github.com/aws/smithy-go/transport/http.Request, got %s", fullTypeName(in.Request))
			}

			f(t, request)

			return middleware.FinalizeOutput{}, middleware.Metadata{}, errCancelOperation
		})
}

func fullTypeName(i any) string {
	return fullValueTypeName(reflect.ValueOf(i))
}

func fullValueTypeName(v reflect.Value) string {
	if v.Kind() == reflect.Ptr {
		return "*" + fullValueTypeName(reflect.Indirect(v))
	}

	requestType := v.Type()
	return fmt.Sprintf("%s.%s", requestType.PkgPath(), requestType.Name())
}

func TestRegion(t *testing.T) {
	testCases := map[string]struct {
		Config                  *Config
		EnvironmentVariables    map[string]string
		IMDSRegion              string
		SharedConfigurationFile string
		ExpectedRegion          string
	}{
		"no configuration": {
			Config: &Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			ExpectedRegion: "",
		},

		"config": {
			Config: &Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				Region:    "us-east-1",
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			ExpectedRegion: "us-east-1",
		},

		"AWS_REGION": {
			Config: &Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			EnvironmentVariables: map[string]string{
				"AWS_REGION": "us-east-1",
			},
			ExpectedRegion: "us-east-1",
		},
		"AWS_DEFAULT_REGION": {
			Config: &Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			EnvironmentVariables: map[string]string{
				"AWS_DEFAULT_REGION": "us-east-1",
			},
			ExpectedRegion: "us-east-1",
		},
		"AWS_REGION overrides AWS_DEFAULT_REGION": {
			Config: &Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			EnvironmentVariables: map[string]string{
				"AWS_REGION":         "us-east-1",
				"AWS_DEFAULT_REGION": "us-west-2",
			},
			ExpectedRegion: "us-east-1",
		},

		"shared configuration file": {
			Config: &Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			SharedConfigurationFile: `
[default]
region = us-east-1
`,
			ExpectedRegion: "us-east-1",
		},

		"IMDS": {
			Config:         &Config{},
			IMDSRegion:     "us-east-1",
			ExpectedRegion: "us-east-1",
		},

		"config overrides AWS_REGION": {
			Config: &Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				SecretKey: servicemocks.MockStaticSecretKey,
				Region:    "us-east-1",
			},
			EnvironmentVariables: map[string]string{
				"AWS_REGION": "us-west-2",
			},
			ExpectedRegion: "us-east-1",
		},
		"config overrides AWS_DEFAULT_REGION": {
			Config: &Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				SecretKey: servicemocks.MockStaticSecretKey,
				Region:    "us-east-1",
			},
			EnvironmentVariables: map[string]string{
				"AWS_DEFAULT_REGION": "us-west-2",
			},
			ExpectedRegion: "us-east-1",
		},

		"config overrides IMDS": {
			Config: &Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				SecretKey: servicemocks.MockStaticSecretKey,
				Region:    "us-west-2",
			},
			IMDSRegion:     "us-east-1",
			ExpectedRegion: "us-west-2",
		},

		"AWS_REGION overrides shared configuration": {
			Config: &Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			EnvironmentVariables: map[string]string{
				"AWS_REGION": "us-east-1",
			},
			SharedConfigurationFile: `
[default]
region = us-west-2
`,
			ExpectedRegion: "us-east-1",
		},
		"AWS_DEFAULT_REGION overrides shared configuration": {
			Config: &Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			EnvironmentVariables: map[string]string{
				"AWS_DEFAULT_REGION": "us-east-1",
			},
			SharedConfigurationFile: `
[default]
region = us-west-2
`,
			ExpectedRegion: "us-east-1",
		},

		"AWS_REGION overrides IMDS": {
			Config: &Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			EnvironmentVariables: map[string]string{
				"AWS_REGION": "us-east-1",
			},
			IMDSRegion:     "us-west-2",
			ExpectedRegion: "us-east-1",
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			servicemocks.InitSessionTestEnv(t)

			for k, v := range testCase.EnvironmentVariables {
				t.Setenv(k, v)
			}

			if testCase.IMDSRegion != "" {
				closeEc2Metadata := servicemocks.AwsMetadataApiMock(append(
					servicemocks.Ec2metadata_securityCredentialsEndpoints,
					servicemocks.Ec2metadata_instanceIdEndpoint,
					servicemocks.Ec2metadata_iamInfoEndpoint,
					servicemocks.Ec2metadata_instanceIdentityEndpoint(testCase.IMDSRegion),
				))
				defer closeEc2Metadata()
			}

			if testCase.SharedConfigurationFile != "" {
				file, err := os.CreateTemp("", "aws-sdk-go-base-shared-configuration-file")

				if err != nil {
					t.Fatalf("unexpected error creating temporary shared configuration file: %s", err)
				}

				defer os.Remove(file.Name())

				err = os.WriteFile(file.Name(), []byte(testCase.SharedConfigurationFile), 0600)

				if err != nil {
					t.Fatalf("unexpected error writing shared configuration file: %s", err)
				}

				testCase.Config.SharedConfigFiles = []string{file.Name()}
			}

			testCase.Config.SkipCredsValidation = true

			_, awsConfig, diags := GetAwsConfig(context.Background(), testCase.Config)
			if diags.HasError() {
				t.Fatalf("error in GetAwsConfig(): %v", diags)
			}

			if a, e := awsConfig.Region, testCase.ExpectedRegion; a != e {
				t.Errorf("expected Region %q, got: %q", e, a)
			}
		})
	}
}

func TestMaxAttempts(t *testing.T) {
	testCases := map[string]struct {
		Config                  *Config
		EnvironmentVariables    map[string]string
		SharedConfigurationFile string
		ExpectedMaxAttempts     int
	}{
		"no configuration": {
			Config: &Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			ExpectedMaxAttempts: retry.DefaultMaxAttempts,
		},

		"config": {
			Config: &Config{
				AccessKey:  servicemocks.MockStaticAccessKey,
				SecretKey:  servicemocks.MockStaticSecretKey,
				MaxRetries: 5,
			},
			ExpectedMaxAttempts: 5,
		},

		"AWS_MAX_ATTEMPTS": {
			Config: &Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			EnvironmentVariables: map[string]string{
				"AWS_MAX_ATTEMPTS": "5",
			},
			ExpectedMaxAttempts: 5,
		},

		"shared configuration file": {
			Config: &Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			SharedConfigurationFile: `
[default]
max_attempts = 5
`,
			ExpectedMaxAttempts: 5,
		},

		"config overrides AWS_MAX_ATTEMPTS": {
			Config: &Config{
				AccessKey:  servicemocks.MockStaticAccessKey,
				SecretKey:  servicemocks.MockStaticSecretKey,
				MaxRetries: 10,
			},
			EnvironmentVariables: map[string]string{
				"AWS_MAX_ATTEMPTS": "5",
			},
			ExpectedMaxAttempts: 10,
		},

		"AWS_MAX_ATTEMPTS overrides shared configuration": {
			Config: &Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			EnvironmentVariables: map[string]string{
				"AWS_MAX_ATTEMPTS": "5",
			},
			SharedConfigurationFile: `
[default]
max_attempts = 10
`,
			ExpectedMaxAttempts: 5,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			servicemocks.InitSessionTestEnv(t)

			for k, v := range testCase.EnvironmentVariables {
				t.Setenv(k, v)
			}

			if testCase.SharedConfigurationFile != "" {
				file, err := os.CreateTemp("", "aws-sdk-go-base-shared-configuration-file")

				if err != nil {
					t.Fatalf("unexpected error creating temporary shared configuration file: %s", err)
				}

				defer os.Remove(file.Name())

				err = os.WriteFile(file.Name(), []byte(testCase.SharedConfigurationFile), 0600)

				if err != nil {
					t.Fatalf("unexpected error writing shared configuration file: %s", err)
				}

				testCase.Config.SharedConfigFiles = []string{file.Name()}
			}

			testCase.Config.SkipCredsValidation = true

			_, awsConfig, diags := GetAwsConfig(context.Background(), testCase.Config)
			if diags.HasError() {
				t.Fatalf("error in GetAwsConfig(): %v", diags)
			}

			retryer := awsConfig.Retryer()
			if retryer == nil {
				t.Fatal("no retryer set")
			}
			if a, e := retryer.MaxAttempts(), testCase.ExpectedMaxAttempts; a != e {
				t.Errorf(`expected MaxAttempts "%d", got: "%d"`, e, a)
			}
		})
	}
}

func TestRetryMode(t *testing.T) {
	var (
		standardRetryer = reflect.TypeOf((*retry.Standard)(nil))
		adaptiveRetryer = reflect.TypeOf((*retry.AdaptiveMode)(nil))
	)

	testCases := map[string]struct {
		Config                  *Config
		EnvironmentVariables    map[string]string
		SharedConfigurationFile string
		ExpectedRetryMode       aws.RetryMode
		RetyerType              reflect.Type
	}{
		"no configuration": {
			Config: &Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			ExpectedRetryMode: "",
			RetyerType:        standardRetryer,
		},

		"config": {
			Config: &Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				SecretKey: servicemocks.MockStaticSecretKey,
				RetryMode: aws.RetryModeAdaptive,
			},
			ExpectedRetryMode: aws.RetryModeAdaptive,
			RetyerType:        adaptiveRetryer,
		},

		"AWS_RETRY_MODE": {
			Config: &Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			EnvironmentVariables: map[string]string{
				"AWS_RETRY_MODE": "adaptive",
			},
			ExpectedRetryMode: aws.RetryModeAdaptive,
			RetyerType:        adaptiveRetryer,
		},

		"shared configuration file": {
			Config: &Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			SharedConfigurationFile: `
[default]
retry_mode = adaptive
`,
			ExpectedRetryMode: aws.RetryModeAdaptive,
			RetyerType:        adaptiveRetryer,
		},

		"config overrides AWS_RETRY_MODE": {
			Config: &Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				SecretKey: servicemocks.MockStaticSecretKey,
				RetryMode: aws.RetryModeStandard,
			},
			EnvironmentVariables: map[string]string{
				"AWS_RETRY_MODE": "adaptive",
			},
			ExpectedRetryMode: aws.RetryModeStandard,
			RetyerType:        standardRetryer,
		},

		"AWS_RETRY_MODE overrides shared configuration": {
			Config: &Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			EnvironmentVariables: map[string]string{
				"AWS_RETRY_MODE": "standard",
			},
			SharedConfigurationFile: `
[default]
retry_mode = adaptive
`,
			ExpectedRetryMode: aws.RetryModeStandard,
			RetyerType:        standardRetryer,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			servicemocks.InitSessionTestEnv(t)

			for k, v := range testCase.EnvironmentVariables {
				t.Setenv(k, v)
			}

			if testCase.SharedConfigurationFile != "" {
				file, err := os.CreateTemp("", "aws-sdk-go-base-shared-configuration-file")

				if err != nil {
					t.Fatalf("unexpected error creating temporary shared configuration file: %s", err)
				}

				defer os.Remove(file.Name())

				err = os.WriteFile(file.Name(), []byte(testCase.SharedConfigurationFile), 0600)

				if err != nil {
					t.Fatalf("unexpected error writing shared configuration file: %s", err)
				}

				testCase.Config.SharedConfigFiles = []string{file.Name()}
			}

			testCase.Config.SkipCredsValidation = true

			_, awsConfig, err := GetAwsConfig(context.Background(), testCase.Config)
			if err != nil {
				t.Fatalf("error in GetAwsConfig() '%[1]T': %[1]s", err)
			}

			retryMode := awsConfig.RetryMode
			if a, e := retryMode, testCase.ExpectedRetryMode; a != e {
				t.Errorf(`expected RetryMode "%s", got: "%s"`, e.String(), a.String())
			}

			retryer := awsConfig.Retryer()
			if retryer == nil {
				t.Fatal("no retryer set")
			}

			nes, ok := retryer.(*networkErrorShortcutter)
			if !ok {
				t.Fatalf(`expected type "*networkErrorShortcutter", got "%T"`, retryer)
			}

			retryer = nes.RetryerV2
			if a, e := reflect.TypeOf(retryer), testCase.RetyerType; a != e {
				t.Errorf(`expected type "%s", got: "%s"`, e, a)
			}
		})
	}
}

func TestServiceEndpointTypes(t *testing.T) {
	testCases := map[string]struct {
		Config                            *Config
		EnvironmentVariables              map[string]string
		SharedConfigurationFile           string
		ExpectedUseFIPSEndpointState      aws.FIPSEndpointState
		ExpectedUseDualStackEndpointState aws.DualStackEndpointState
	}{
		"normal endpoint": {
			Config: &Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			ExpectedUseFIPSEndpointState:      aws.FIPSEndpointStateUnset,
			ExpectedUseDualStackEndpointState: aws.DualStackEndpointStateUnset,
		},

		// FIPS Endpoint
		"FIPS endpoint config": {
			Config: &Config{
				AccessKey:       servicemocks.MockStaticAccessKey,
				SecretKey:       servicemocks.MockStaticSecretKey,
				UseFIPSEndpoint: true,
			},
			ExpectedUseFIPSEndpointState: aws.FIPSEndpointStateEnabled,
		},
		"FIPS endpoint envvar": {
			Config: &Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			EnvironmentVariables: map[string]string{
				"AWS_USE_FIPS_ENDPOINT": "true",
			},
			ExpectedUseFIPSEndpointState: aws.FIPSEndpointStateEnabled,
		},
		"FIPS endpoint shared configuration file": {
			Config: &Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			SharedConfigurationFile: `
[default]
use_fips_endpoint = true
`,
			ExpectedUseFIPSEndpointState: aws.FIPSEndpointStateEnabled,
		},
		"FIPS endpoint config overrides env var": {
			Config: &Config{
				AccessKey:       servicemocks.MockStaticAccessKey,
				SecretKey:       servicemocks.MockStaticSecretKey,
				UseFIPSEndpoint: true,
			},
			EnvironmentVariables: map[string]string{
				"AWS_USE_FIPS_ENDPOINT": "true",
			},
			ExpectedUseFIPSEndpointState: aws.FIPSEndpointStateEnabled,
		},
		"FIPS endpoint env var overrides shared configuration file": {
			Config: &Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			EnvironmentVariables: map[string]string{
				"AWS_USE_FIPS_ENDPOINT": "true",
			},
			SharedConfigurationFile: `
[default]
use_fips_endpoint = false
`,
			ExpectedUseFIPSEndpointState: aws.FIPSEndpointStateEnabled,
		},

		// DualStack Endpoint
		"DualStack endpoint config": {
			Config: &Config{
				AccessKey:            servicemocks.MockStaticAccessKey,
				SecretKey:            servicemocks.MockStaticSecretKey,
				UseDualStackEndpoint: true,
			},
			ExpectedUseDualStackEndpointState: aws.DualStackEndpointStateEnabled,
		},
		"DualStack endpoint envvar": {
			Config: &Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			EnvironmentVariables: map[string]string{
				"AWS_USE_DUALSTACK_ENDPOINT": "true",
			},
			ExpectedUseDualStackEndpointState: aws.DualStackEndpointStateEnabled,
		},
		"DualStack endpoint shared configuration file": {
			Config: &Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			SharedConfigurationFile: `
[default]
use_dualstack_endpoint = true
`,
			ExpectedUseDualStackEndpointState: aws.DualStackEndpointStateEnabled,
		},
		"DualStack endpoint config overrides env var": {
			Config: &Config{
				AccessKey:            servicemocks.MockStaticAccessKey,
				SecretKey:            servicemocks.MockStaticSecretKey,
				UseDualStackEndpoint: true,
			},
			EnvironmentVariables: map[string]string{
				"AWS_USE_DUALSTACK_ENDPOINT": "true",
			},
			ExpectedUseDualStackEndpointState: aws.DualStackEndpointStateEnabled,
		},
		"DualStack endpoint env var overrides shared configuration file": {
			Config: &Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			EnvironmentVariables: map[string]string{
				"AWS_USE_DUALSTACK_ENDPOINT": "true",
			},
			SharedConfigurationFile: `
[default]
use_dualstack_endpoint = false
`,
			ExpectedUseDualStackEndpointState: aws.DualStackEndpointStateEnabled,
		},

		// FIPS and DualStack Endpoint
		"Both endpoints config": {
			Config: &Config{
				AccessKey:            servicemocks.MockStaticAccessKey,
				SecretKey:            servicemocks.MockStaticSecretKey,
				UseDualStackEndpoint: true,
				UseFIPSEndpoint:      true,
			},
			ExpectedUseFIPSEndpointState:      aws.FIPSEndpointStateEnabled,
			ExpectedUseDualStackEndpointState: aws.DualStackEndpointStateEnabled,
		},
		"Both endpoints FIPS config DualStack envvar": {
			Config: &Config{
				AccessKey:       servicemocks.MockStaticAccessKey,
				SecretKey:       servicemocks.MockStaticSecretKey,
				UseFIPSEndpoint: true,
			},
			EnvironmentVariables: map[string]string{
				"AWS_USE_DUALSTACK_ENDPOINT": "true",
			},
			ExpectedUseFIPSEndpointState:      aws.FIPSEndpointStateEnabled,
			ExpectedUseDualStackEndpointState: aws.DualStackEndpointStateEnabled,
		},
		"Both endpoints FIPS shared configuration file DualStack config": {
			Config: &Config{
				AccessKey:            servicemocks.MockStaticAccessKey,
				SecretKey:            servicemocks.MockStaticSecretKey,
				UseDualStackEndpoint: true,
			},
			SharedConfigurationFile: `
[default]
use_fips_endpoint = true
`,
			ExpectedUseFIPSEndpointState:      aws.FIPSEndpointStateEnabled,
			ExpectedUseDualStackEndpointState: aws.DualStackEndpointStateEnabled,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			servicemocks.InitSessionTestEnv(t)

			for k, v := range testCase.EnvironmentVariables {
				t.Setenv(k, v)
			}

			if testCase.SharedConfigurationFile != "" {
				file, err := os.CreateTemp("", "aws-sdk-go-base-shared-configuration-file")

				if err != nil {
					t.Fatalf("unexpected error creating temporary shared configuration file: %s", err)
				}

				defer os.Remove(file.Name())

				err = os.WriteFile(file.Name(), []byte(testCase.SharedConfigurationFile), 0600)

				if err != nil {
					t.Fatalf("unexpected error writing shared configuration file: %s", err)
				}

				testCase.Config.SharedConfigFiles = []string{file.Name()}
			}

			testCase.Config.SkipCredsValidation = true

			ctx, awsConfig, diags := GetAwsConfig(context.Background(), testCase.Config)
			if diags.HasError() {
				t.Fatalf("error in GetAwsConfig(): %v", diags)
			}

			useFIPSState, _, err := awsconfig.ResolveUseFIPSEndpoint(ctx, awsConfig.ConfigSources)
			if err != nil {
				t.Fatalf("error in ResolveUseFIPSEndpoint: %s", err)
			}
			if a, e := useFIPSState, testCase.ExpectedUseFIPSEndpointState; a != e {
				t.Errorf("expected UseFIPSEndpoint %q, got: %q", awsconfig.FIPSEndpointStateString(e), awsconfig.FIPSEndpointStateString(a))
			}

			useDualStackState, _, err := awsconfig.ResolveUseDualStackEndpoint(ctx, awsConfig.ConfigSources)
			if err != nil {
				t.Fatalf("error in ResolveUseDualStackEndpoint: %s", err)
			}
			if a, e := useDualStackState, testCase.ExpectedUseDualStackEndpointState; a != e {
				t.Errorf("expected UseDualStackEndpoint %q, got: %q", awsconfig.DualStackEndpointStateString(e), awsconfig.DualStackEndpointStateString(a))
			}
		})
	}
}

func TestEC2MetadataServiceClientEnableState(t *testing.T) {
	testCases := map[string]struct {
		Config                                      *Config
		EnvironmentVariables                        map[string]string
		SharedConfigurationFile                     string
		ExpectedEC2MetadataServiceClientEnableState imds.ClientEnableState
	}{
		"no configuration": {
			Config: &Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			ExpectedEC2MetadataServiceClientEnableState: imds.ClientDefaultEnableState,
		},

		"config enabled": {
			Config: &Config{
				AccessKey:                     servicemocks.MockStaticAccessKey,
				SecretKey:                     servicemocks.MockStaticSecretKey,
				EC2MetadataServiceEnableState: imds.ClientEnabled,
			},
			ExpectedEC2MetadataServiceClientEnableState: imds.ClientEnabled,
		},
		"config disabled": {
			Config: &Config{
				AccessKey:                     servicemocks.MockStaticAccessKey,
				SecretKey:                     servicemocks.MockStaticSecretKey,
				EC2MetadataServiceEnableState: imds.ClientDisabled,
			},
			ExpectedEC2MetadataServiceClientEnableState: imds.ClientDisabled,
		},

		"envvar true": {
			Config: &Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			EnvironmentVariables: map[string]string{
				"AWS_EC2_METADATA_DISABLED": "true",
			},
			ExpectedEC2MetadataServiceClientEnableState: imds.ClientDisabled,
		},
		"envvar false": {
			Config: &Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			EnvironmentVariables: map[string]string{
				"AWS_EC2_METADATA_DISABLED": "false",
			},
			ExpectedEC2MetadataServiceClientEnableState: imds.ClientEnabled,
		},

		"config enabled envvar true": {
			Config: &Config{
				AccessKey:                     servicemocks.MockStaticAccessKey,
				SecretKey:                     servicemocks.MockStaticSecretKey,
				EC2MetadataServiceEnableState: imds.ClientEnabled,
			},
			EnvironmentVariables: map[string]string{
				"AWS_EC2_METADATA_DISABLED": "true",
			},
			ExpectedEC2MetadataServiceClientEnableState: imds.ClientEnabled,
		},
		"config disabled envvar false": {
			Config: &Config{
				AccessKey:                     servicemocks.MockStaticAccessKey,
				SecretKey:                     servicemocks.MockStaticSecretKey,
				EC2MetadataServiceEnableState: imds.ClientDisabled,
			},
			EnvironmentVariables: map[string]string{
				"AWS_EC2_METADATA_DISABLED": "false",
			},
			ExpectedEC2MetadataServiceClientEnableState: imds.ClientDisabled,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			servicemocks.InitSessionTestEnv(t)

			for k, v := range testCase.EnvironmentVariables {
				t.Setenv(k, v)
			}

			if testCase.SharedConfigurationFile != "" {
				file, err := os.CreateTemp("", "aws-sdk-go-base-shared-configuration-file")

				if err != nil {
					t.Fatalf("unexpected error creating temporary shared configuration file: %s", err)
				}

				defer os.Remove(file.Name())

				err = os.WriteFile(file.Name(), []byte(testCase.SharedConfigurationFile), 0600)

				if err != nil {
					t.Fatalf("unexpected error writing shared configuration file: %s", err)
				}

				testCase.Config.SharedConfigFiles = []string{file.Name()}
			}

			testCase.Config.SkipCredsValidation = true

			_, awsConfig, diags := GetAwsConfig(context.Background(), testCase.Config)
			if diags.HasError() {
				t.Fatalf("error in GetAwsConfig(): %v", diags)
			}

			ec2MetadataServiceClientEnableState, _, err := awsconfig.ResolveEC2IMDSClientEnableState(awsConfig.ConfigSources)
			if err != nil {
				t.Fatalf("error in ResolveEC2IMDSClientEnableState: %s", err)
			}
			if a, e := ec2MetadataServiceClientEnableState, testCase.ExpectedEC2MetadataServiceClientEnableState; a != e {
				t.Errorf("expected EC2MetadataServiceClientEnableState %q, got: %q", awsconfig.EC2IMDSClientEnableStateString(e), awsconfig.EC2IMDSClientEnableStateString(a))
			}
		})
	}
}

func TestEC2MetadataServiceEndpoint(t *testing.T) {
	testCases := map[string]struct {
		Config                             *Config
		EnvironmentVariables               map[string]string
		SharedConfigurationFile            string
		ExpectedEC2MetadataServiceEndpoint string
		ExpectedDiags                      diag.Diagnostics
	}{
		"no configuration": {
			Config: &Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			ExpectedEC2MetadataServiceEndpoint: "",
		},

		"config": {
			Config: &Config{
				AccessKey:                  servicemocks.MockStaticAccessKey,
				SecretKey:                  servicemocks.MockStaticSecretKey,
				EC2MetadataServiceEndpoint: "https://127.0.0.1:1234",
			},
			ExpectedEC2MetadataServiceEndpoint: "https://127.0.0.1:1234",
		},

		"envvar": {
			Config: &Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			EnvironmentVariables: map[string]string{
				"AWS_EC2_METADATA_SERVICE_ENDPOINT": "https://127.0.0.1:1234",
			},
			ExpectedEC2MetadataServiceEndpoint: "https://127.0.0.1:1234",
		},
		"deprecated envvar": {
			Config: &Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			EnvironmentVariables: map[string]string{
				"AWS_METADATA_URL": "https://127.0.0.1:1234",
			},
			ExpectedEC2MetadataServiceEndpoint: "https://127.0.0.1:1234",
			ExpectedDiags: diag.Diagnostics{
				diag.NewWarningDiagnostic(
					"Deprecated Environment Variable",
					`The environment variable "AWS_METADATA_URL" is deprecated. Use "AWS_EC2_METADATA_SERVICE_ENDPOINT" instead.`,
				),
			},
		},
		"envvar overrides deprecated envvar": {
			Config: &Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			EnvironmentVariables: map[string]string{
				"AWS_METADATA_URL":                  "https://127.1.1.1:1111",
				"AWS_EC2_METADATA_SERVICE_ENDPOINT": "https://127.0.0.1:1234",
			},
			ExpectedEC2MetadataServiceEndpoint: "https://127.0.0.1:1234",
			ExpectedDiags: diag.Diagnostics{
				diag.NewWarningDiagnostic(
					"Deprecated Environment Variable",
					`The environment variable "AWS_METADATA_URL" is deprecated. Use "AWS_EC2_METADATA_SERVICE_ENDPOINT" instead.`+"\n"+
						`"AWS_EC2_METADATA_SERVICE_ENDPOINT" is set to "https://127.0.0.1:1234" and "AWS_METADATA_URL" is set to "https://127.1.1.1:1111". Ignoring "AWS_METADATA_URL".`,
				),
			},
		},

		"shared configuration file": {
			Config: &Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			SharedConfigurationFile: `
[default]
ec2_metadata_service_endpoint = https://127.0.0.1:1234
`,
			ExpectedEC2MetadataServiceEndpoint: "https://127.0.0.1:1234",
		},

		"config overrides envvar": {
			Config: &Config{
				AccessKey:                  servicemocks.MockStaticAccessKey,
				SecretKey:                  servicemocks.MockStaticSecretKey,
				EC2MetadataServiceEndpoint: "https://127.0.0.1:1234",
			},
			EnvironmentVariables: map[string]string{
				"AWS_EC2_METADATA_SERVICE_ENDPOINT": "https://127.1.1.1:1111",
			},
			ExpectedEC2MetadataServiceEndpoint: "https://127.0.0.1:1234",
		},
		"config overrides deprecated envvar": {
			Config: &Config{
				AccessKey:                  servicemocks.MockStaticAccessKey,
				SecretKey:                  servicemocks.MockStaticSecretKey,
				EC2MetadataServiceEndpoint: "https://127.0.0.1:1234",
			},
			EnvironmentVariables: map[string]string{
				"AWS_METADATA_URL": "https://127.1.1.1:1111",
			},
			ExpectedEC2MetadataServiceEndpoint: "https://127.0.0.1:1234",
		},

		"envvar overrides shared configuration": {
			Config: &Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			EnvironmentVariables: map[string]string{
				"AWS_EC2_METADATA_SERVICE_ENDPOINT": "https://127.0.0.1:1234",
			},
			SharedConfigurationFile: `
[default]
ec2_metadata_service_endpoint = https://127.1.1.1:1111
`,
			ExpectedEC2MetadataServiceEndpoint: "https://127.0.0.1:1234",
		},
		"deprecated envvar overrides shared configuration": {
			Config: &Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			EnvironmentVariables: map[string]string{
				"AWS_METADATA_URL": "https://127.0.0.1:1234",
			},
			SharedConfigurationFile: `
[default]
ec2_metadata_service_endpoint = https://127.1.1.1:1111
`,
			ExpectedEC2MetadataServiceEndpoint: "https://127.0.0.1:1234",
			ExpectedDiags: diag.Diagnostics{
				diag.NewWarningDiagnostic(
					"Deprecated Environment Variable",
					`The environment variable "AWS_METADATA_URL" is deprecated. Use "AWS_EC2_METADATA_SERVICE_ENDPOINT" instead.`,
				),
			},
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			servicemocks.InitSessionTestEnv(t)

			for k, v := range testCase.EnvironmentVariables {
				t.Setenv(k, v)
			}

			if testCase.SharedConfigurationFile != "" {
				file, err := os.CreateTemp("", "aws-sdk-go-base-shared-configuration-file")

				if err != nil {
					t.Fatalf("unexpected error creating temporary shared configuration file: %s", err)
				}

				defer os.Remove(file.Name())

				err = os.WriteFile(file.Name(), []byte(testCase.SharedConfigurationFile), 0600)

				if err != nil {
					t.Fatalf("unexpected error writing shared configuration file: %s", err)
				}

				testCase.Config.SharedConfigFiles = []string{file.Name()}
			}

			testCase.Config.SkipCredsValidation = true

			_, awsConfig, diags := GetAwsConfig(context.Background(), testCase.Config)

			if diff := cmp.Diff(diags, testCase.ExpectedDiags); diff != "" {
				t.Errorf("Unexpected response (+wanted, -got): %s", diff)
			}
			if diags.HasError() {
				return
			}

			ec2MetadataServiceEndpoint, _, err := awsconfig.ResolveEC2IMDSEndpointConfig(awsConfig.ConfigSources)
			if err != nil {
				t.Fatalf("error in ResolveEC2IMDSEndpointConfig: %s", err)
			}
			if a, e := ec2MetadataServiceEndpoint, testCase.ExpectedEC2MetadataServiceEndpoint; a != e {
				t.Errorf("expected EC2MetadataServiceEndpoint %q, got: %q", e, a)
			}
		})
	}
}

func TestEC2MetadataServiceEndpointMode(t *testing.T) {
	testCases := map[string]struct {
		Config                                 *Config
		EnvironmentVariables                   map[string]string
		SharedConfigurationFile                string
		ExpectedEC2MetadataServiceEndpointMode imds.EndpointModeState
	}{
		"no configuration": {
			Config: &Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			ExpectedEC2MetadataServiceEndpointMode: imds.EndpointModeStateUnset,
		},

		"config": {
			Config: &Config{
				AccessKey:                      servicemocks.MockStaticAccessKey,
				SecretKey:                      servicemocks.MockStaticSecretKey,
				EC2MetadataServiceEndpointMode: EC2MetadataEndpointModeIPv4,
			},
			ExpectedEC2MetadataServiceEndpointMode: imds.EndpointModeStateIPv4,
		},

		"envvar": {
			Config: &Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			EnvironmentVariables: map[string]string{
				"AWS_EC2_METADATA_SERVICE_ENDPOINT_MODE": EC2MetadataEndpointModeIPv6,
			},
			ExpectedEC2MetadataServiceEndpointMode: imds.EndpointModeStateIPv6,
		},

		"shared configuration file": {
			Config: &Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			SharedConfigurationFile: `
[default]
ec2_metadata_service_endpoint_mode = IPv6
`,
			ExpectedEC2MetadataServiceEndpointMode: imds.EndpointModeStateIPv6,
		},

		"config overrides envvar": {
			Config: &Config{
				AccessKey:                      servicemocks.MockStaticAccessKey,
				SecretKey:                      servicemocks.MockStaticSecretKey,
				EC2MetadataServiceEndpointMode: EC2MetadataEndpointModeIPv4,
			},
			EnvironmentVariables: map[string]string{
				"AWS_EC2_METADATA_SERVICE_ENDPOINT_MODE": EC2MetadataEndpointModeIPv6,
			},
			ExpectedEC2MetadataServiceEndpointMode: imds.EndpointModeStateIPv4,
		},

		"envvar overrides shared configuration": {
			Config: &Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			EnvironmentVariables: map[string]string{
				"AWS_EC2_METADATA_SERVICE_ENDPOINT_MODE": EC2MetadataEndpointModeIPv6,
			},
			SharedConfigurationFile: `
[default]
ec2_metadata_service_endpoint_mode = IPv4
`,
			ExpectedEC2MetadataServiceEndpointMode: imds.EndpointModeStateIPv6,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			servicemocks.InitSessionTestEnv(t)

			for k, v := range testCase.EnvironmentVariables {
				t.Setenv(k, v)
			}

			if testCase.SharedConfigurationFile != "" {
				file, err := os.CreateTemp("", "aws-sdk-go-base-shared-configuration-file")

				if err != nil {
					t.Fatalf("unexpected error creating temporary shared configuration file: %s", err)
				}

				defer os.Remove(file.Name())

				err = os.WriteFile(file.Name(), []byte(testCase.SharedConfigurationFile), 0600)

				if err != nil {
					t.Fatalf("unexpected error writing shared configuration file: %s", err)
				}

				testCase.Config.SharedConfigFiles = []string{file.Name()}
			}

			testCase.Config.SkipCredsValidation = true

			_, awsConfig, diags := GetAwsConfig(context.Background(), testCase.Config)
			if diags.HasError() {
				t.Fatalf("error in GetAwsConfig(): %v", diags)
			}

			ec2MetadataServiceEndpointMode, _, err := awsconfig.ResolveEC2IMDSEndpointModeConfig(awsConfig.ConfigSources)
			if err != nil {
				t.Fatalf("error in ResolveEC2IMDSEndpointConfig: %s", err)
			}
			if a, e := ec2MetadataServiceEndpointMode, testCase.ExpectedEC2MetadataServiceEndpointMode; a != e {
				t.Errorf("expected EC2MetadataServiceEndpointMode %q, got: %q", awsconfig.EC2IMDSEndpointModeString(e), awsconfig.EC2IMDSEndpointModeString(a))
			}
		})
	}
}

func TestCustomCABundle(t *testing.T) {
	testCases := map[string]struct {
		Config                              *Config
		SetConfig                           bool
		SetEnvironmentVariable              bool
		SetSharedConfigurationFile          bool
		SetSharedConfigurationFileToInvalid bool
		ExpandEnvVars                       bool
		EnvironmentVariables                map[string]string
		ExpectTLSClientConfigRootCAsSet     bool
	}{
		"no configuration": {
			Config: &Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			ExpectTLSClientConfigRootCAsSet: false,
		},

		"config": {
			Config: &Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			SetConfig:                       true,
			ExpectTLSClientConfigRootCAsSet: true,
		},

		"expanded config": {
			Config: &Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			SetConfig:                       true,
			ExpandEnvVars:                   true,
			ExpectTLSClientConfigRootCAsSet: true,
		},

		"envvar": {
			Config: &Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			SetEnvironmentVariable:          true,
			ExpectTLSClientConfigRootCAsSet: true,
		},

		"shared configuration file": {
			Config: &Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			SetSharedConfigurationFile:      true,
			ExpectTLSClientConfigRootCAsSet: true,
		},

		"config overrides envvar": {
			Config: &Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			SetConfig: true,
			EnvironmentVariables: map[string]string{
				"AWS_CA_BUNDLE": "no-such-file",
			},
			ExpectTLSClientConfigRootCAsSet: true,
		},

		"envvar overrides shared configuration": {
			Config: &Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			SetEnvironmentVariable:              true,
			SetSharedConfigurationFileToInvalid: true,
			ExpectTLSClientConfigRootCAsSet:     true,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			servicemocks.InitSessionTestEnv(t)

			for k, v := range testCase.EnvironmentVariables {
				t.Setenv(k, v)
			}

			tempdir, err := os.MkdirTemp("", "temp")
			if err != nil {
				t.Fatalf("error creating temp dir: %s", err)
			}
			defer os.Remove(tempdir)
			t.Setenv("TMPDIR", tempdir)

			pemFile, err := servicemocks.TempPEMFile()
			defer os.Remove(pemFile)
			if err != nil {
				t.Fatalf("error creating PEM file: %s", err)
			}

			if testCase.ExpandEnvVars {
				tmpdir := os.Getenv("TMPDIR")
				rel, err := filepath.Rel(tmpdir, pemFile)
				if err != nil {
					t.Fatalf("error making path relative: %s", err)
				}
				t.Logf("relative: %s", rel)
				pemFile = filepath.Join("$TMPDIR", rel)
				t.Logf("env tempfile: %s", pemFile)
			}

			if testCase.SetConfig {
				testCase.Config.CustomCABundle = pemFile
			}

			if testCase.SetEnvironmentVariable {
				t.Setenv("AWS_CA_BUNDLE", pemFile)
			}

			if testCase.SetSharedConfigurationFile {
				file, err := os.CreateTemp("", "aws-sdk-go-base-shared-configuration-file")

				if err != nil {
					t.Fatalf("unexpected error creating temporary shared configuration file: %s", err)
				}

				defer os.Remove(file.Name())

				err = os.WriteFile(
					file.Name(),
					fmt.Appendf(nil, `
[default]
ca_bundle = %s
`, pemFile),
					0600)

				if err != nil {
					t.Fatalf("unexpected error writing shared configuration file: %s", err)
				}

				testCase.Config.SharedConfigFiles = []string{file.Name()}
			}

			if testCase.SetSharedConfigurationFileToInvalid {
				file, err := os.CreateTemp("", "aws-sdk-go-base-shared-configuration-file")

				if err != nil {
					t.Fatalf("unexpected error creating temporary shared configuration file: %s", err)
				}

				defer os.Remove(file.Name())

				err = os.WriteFile(
					file.Name(),
					[]byte(`
[default]
ca_bundle = no-such-file
`),
					0600)

				if err != nil {
					t.Fatalf("unexpected error writing shared configuration file: %s", err)
				}

				testCase.Config.SharedConfigFiles = []string{file.Name()}
			}

			testCase.Config.SkipCredsValidation = true

			_, awsConfig, diags := GetAwsConfig(context.Background(), testCase.Config)
			if diags.HasError() {
				t.Fatalf("error in GetAwsConfig(): %v", diags)
			}

			type transportGetter interface {
				GetTransport() *http.Transport
			}

			trGetter := awsConfig.HTTPClient.(transportGetter)
			tr := trGetter.GetTransport()

			if a, e := tr.TLSClientConfig.RootCAs != nil, testCase.ExpectTLSClientConfigRootCAsSet; a != e {
				t.Errorf("expected(%t) CA Bundle, got: %t", e, a)
			}
		})
	}
}

func TestAssumeRole(t *testing.T) {
	testCases := map[string]struct {
		Config                   *Config
		SharedConfigurationFile  string
		ExpectedCredentialsValue aws.Credentials
		ExpectedDiags            diag.Diagnostics
		MockStsEndpoints         []*servicemocks.MockEndpoint
	}{
		"config single": {
			Config: &Config{
				AssumeRole: []AssumeRole{{
					RoleARN:     servicemocks.MockStsAssumeRoleArn,
					SessionName: servicemocks.MockStsAssumeRoleSessionName,
				}},
				AccessKey: servicemocks.MockStaticAccessKey,
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			ExpectedCredentialsValue: mockdata.MockStsAssumeRoleCredentials,
			MockStsEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsAssumeRoleValidEndpoint,
			},
		},

		"config multiple": {
			Config: &Config{
				AssumeRole: []AssumeRole{
					{
						RoleARN:     servicemocks.MockStsAssumeRoleArn,
						SessionName: servicemocks.MockStsAssumeRoleSessionName,
					},
					{
						RoleARN:     servicemocks.MockStsAssumeRoleArn2,
						SessionName: servicemocks.MockStsAssumeRoleSessionName2,
					},
				},
				AccessKey: servicemocks.MockStaticAccessKey,
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			ExpectedCredentialsValue: mockdata.MockStsAssumeRoleCredentials,
			MockStsEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsAssumeRoleValidEndpoint,
				servicemocks.MockStsAssumeRoleValidEndpointWithOptions(map[string]string{
					"RoleArn":         servicemocks.MockStsAssumeRoleArn2,
					"RoleSessionName": servicemocks.MockStsAssumeRoleSessionName2,
				}),
			},
		},

		"shared configuration file": {
			Config: &Config{},
			SharedConfigurationFile: fmt.Sprintf(`
[default]
role_arn = %[1]s
role_session_name = %[2]s
source_profile = SharedConfigurationSourceProfile

[profile SharedConfigurationSourceProfile]
aws_access_key_id = SharedConfigurationSourceAccessKey
aws_secret_access_key = SharedConfigurationSourceSecretKey
`, servicemocks.MockStsAssumeRoleArn, servicemocks.MockStsAssumeRoleSessionName),
			ExpectedCredentialsValue: mockdata.MockStsAssumeRoleCredentials,
			MockStsEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsAssumeRoleValidEndpoint,
			},
		},

		"config overrides shared configuration": {
			Config: &Config{
				AssumeRole: []AssumeRole{{
					RoleARN:     servicemocks.MockStsAssumeRoleArn,
					SessionName: servicemocks.MockStsAssumeRoleSessionName,
				}},
				AccessKey: servicemocks.MockStaticAccessKey,
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			SharedConfigurationFile: fmt.Sprintf(`
[default]
role_arn = %[1]s
role_session_name = %[2]s
source_profile = SharedConfigurationSourceProfile

[profile SharedConfigurationSourceProfile]
aws_access_key_id = SharedConfigurationSourceAccessKey
aws_secret_access_key = SharedConfigurationSourceSecretKey
`, servicemocks.MockStsAssumeRoleArn, servicemocks.MockStsAssumeRoleSessionName),
			ExpectedCredentialsValue: mockdata.MockStsAssumeRoleCredentials,
			MockStsEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsAssumeRoleValidEndpoint,
			},
		},

		"with duration": {
			Config: &Config{
				AssumeRole: []AssumeRole{{
					RoleARN:     servicemocks.MockStsAssumeRoleArn,
					SessionName: servicemocks.MockStsAssumeRoleSessionName,
					Duration:    1 * time.Hour,
				}},
				AccessKey: servicemocks.MockStaticAccessKey,
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			ExpectedCredentialsValue: mockdata.MockStsAssumeRoleCredentials,
			MockStsEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsAssumeRoleValidEndpointWithOptions(map[string]string{"DurationSeconds": "3600"}),
			},
		},

		"with policy": {
			Config: &Config{
				AssumeRole: []AssumeRole{{
					RoleARN:     servicemocks.MockStsAssumeRoleArn,
					SessionName: servicemocks.MockStsAssumeRoleSessionName,
					Policy:      "{}",
				}},
				AccessKey: servicemocks.MockStaticAccessKey,
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			ExpectedCredentialsValue: mockdata.MockStsAssumeRoleCredentials,
			MockStsEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsAssumeRoleValidEndpointWithOptions(map[string]string{"Policy": "{}"}),
			},
		},

		"invalid empty single config": {
			Config: &Config{
				AssumeRole: []AssumeRole{
					{},
				},
				AccessKey: servicemocks.MockStaticAccessKey,
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			ExpectedCredentialsValue: mockdata.MockStsAssumeRoleCredentials,
			ExpectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Cannot assume IAM Role",
					"IAM Role ARN not set in assume role 1 of 1",
				),
			},
		},

		"invalid empty last config": {
			Config: &Config{
				AssumeRole: []AssumeRole{
					{
						RoleARN:     servicemocks.MockStsAssumeRoleArn,
						SessionName: servicemocks.MockStsAssumeRoleSessionName,
					},
					{},
				},
				AccessKey: servicemocks.MockStaticAccessKey,
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			ExpectedCredentialsValue: mockdata.MockStsAssumeRoleCredentials,
			ExpectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Cannot assume IAM Role",
					"IAM Role ARN not set in assume role 2 of 2",
				),
			},
			MockStsEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsAssumeRoleValidEndpoint,
			},
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			servicemocks.InitSessionTestEnv(t)

			closeSts, _, stsEndpoint := mockdata.GetMockedAwsApiSession("STS", testCase.MockStsEndpoints)
			defer closeSts()

			testCase.Config.StsEndpoint = stsEndpoint

			tempdir, err := os.MkdirTemp("", "temp")
			if err != nil {
				t.Fatalf("error creating temp dir: %s", err)
			}
			defer os.Remove(tempdir)
			t.Setenv("TMPDIR", tempdir)

			if testCase.SharedConfigurationFile != "" {
				file, err := os.CreateTemp("", "aws-sdk-go-base-shared-configuration-file")

				if err != nil {
					t.Fatalf("unexpected error creating temporary shared configuration file: %s", err)
				}

				defer os.Remove(file.Name())

				err = os.WriteFile(file.Name(), []byte(testCase.SharedConfigurationFile), 0600)

				if err != nil {
					t.Fatalf("unexpected error writing shared configuration file: %s", err)
				}

				testCase.Config.SharedConfigFiles = []string{file.Name()}
			}

			testCase.Config.SkipCredsValidation = true

			ctx, awsConfig, diags := GetAwsConfig(context.Background(), testCase.Config)

			if diff := cmp.Diff(diags, testCase.ExpectedDiags); diff != "" {
				t.Errorf("Unexpected response (+wanted, -got): %s", diff)
			}
			if diags.HasError() {
				return
			}

			credentialsValue, err := awsConfig.Credentials.Retrieve(ctx)

			if err != nil {
				t.Fatalf("unexpected credentials Retrieve() error: %s", err)
			}

			if diff := cmp.Diff(credentialsValue, testCase.ExpectedCredentialsValue, cmpopts.IgnoreFields(aws.Credentials{}, "Expires")); diff != "" {
				t.Fatalf("unexpected credentials: (- got, + expected)\n%s", diff)
			}
		})
	}
}

func TestAssumeRoleWithWebIdentity(t *testing.T) {
	testCases := map[string]struct {
		Config                          *Config
		SetConfig                       bool
		ExpandEnvVars                   bool
		EnvironmentVariables            map[string]string
		SetTokenFileEnvironmentVariable bool
		SharedConfigurationFile         string
		SetSharedConfigurationFile      bool
		ExpectedCredentialsValue        aws.Credentials
		ExpectedDiags                   diag.Diagnostics
		MockStsEndpoints                []*servicemocks.MockEndpoint
	}{
		"config with inline token": {
			Config: &Config{
				AssumeRoleWithWebIdentity: &AssumeRoleWithWebIdentity{
					RoleARN:          servicemocks.MockStsAssumeRoleWithWebIdentityArn,
					SessionName:      servicemocks.MockStsAssumeRoleWithWebIdentitySessionName,
					WebIdentityToken: servicemocks.MockWebIdentityToken,
				},
			},
			ExpectedCredentialsValue: mockdata.MockStsAssumeRoleWithWebIdentityCredentials,
			MockStsEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsAssumeRoleWithWebIdentityValidEndpoint,
			},
		},

		"config with token file": {
			Config: &Config{
				AssumeRoleWithWebIdentity: &AssumeRoleWithWebIdentity{
					RoleARN:     servicemocks.MockStsAssumeRoleWithWebIdentityArn,
					SessionName: servicemocks.MockStsAssumeRoleWithWebIdentitySessionName,
				},
			},
			SetConfig:                true,
			ExpectedCredentialsValue: mockdata.MockStsAssumeRoleWithWebIdentityCredentials,
			MockStsEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsAssumeRoleWithWebIdentityValidEndpoint,
			},
		},

		"config with expanded path": {
			Config: &Config{
				AssumeRoleWithWebIdentity: &AssumeRoleWithWebIdentity{
					RoleARN:     servicemocks.MockStsAssumeRoleWithWebIdentityArn,
					SessionName: servicemocks.MockStsAssumeRoleWithWebIdentitySessionName,
				},
			},
			SetConfig:                true,
			ExpandEnvVars:            true,
			ExpectedCredentialsValue: mockdata.MockStsAssumeRoleWithWebIdentityCredentials,
			MockStsEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsAssumeRoleWithWebIdentityValidEndpoint,
			},
		},

		"envvar": {
			Config: &Config{},
			EnvironmentVariables: map[string]string{
				"AWS_ROLE_ARN":          servicemocks.MockStsAssumeRoleWithWebIdentityArn,
				"AWS_ROLE_SESSION_NAME": servicemocks.MockStsAssumeRoleWithWebIdentitySessionName,
			},
			SetTokenFileEnvironmentVariable: true,
			ExpectedCredentialsValue:        mockdata.MockStsAssumeRoleWithWebIdentityCredentials,
			MockStsEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsAssumeRoleWithWebIdentityValidEndpoint,
			},
		},

		"shared configuration file": {
			Config: &Config{},
			SharedConfigurationFile: fmt.Sprintf(`
[default]
role_arn = %[1]s
role_session_name = %[2]s
`, servicemocks.MockStsAssumeRoleWithWebIdentityArn, servicemocks.MockStsAssumeRoleWithWebIdentitySessionName),
			SetSharedConfigurationFile: true,
			ExpectedCredentialsValue:   mockdata.MockStsAssumeRoleWithWebIdentityCredentials,
			MockStsEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsAssumeRoleWithWebIdentityValidEndpoint,
			},
		},

		"config overrides envvar": {
			Config: &Config{
				AssumeRoleWithWebIdentity: &AssumeRoleWithWebIdentity{
					RoleARN:          servicemocks.MockStsAssumeRoleWithWebIdentityArn,
					SessionName:      servicemocks.MockStsAssumeRoleWithWebIdentitySessionName,
					WebIdentityToken: servicemocks.MockWebIdentityToken,
				},
			},
			EnvironmentVariables: map[string]string{
				"AWS_ROLE_ARN":                servicemocks.MockStsAssumeRoleWithWebIdentityAlternateArn,
				"AWS_ROLE_SESSION_NAME":       servicemocks.MockStsAssumeRoleWithWebIdentityAlternateSessionName,
				"AWS_WEB_IDENTITY_TOKEN_FILE": "no-such-file",
			},
			ExpectedCredentialsValue: mockdata.MockStsAssumeRoleWithWebIdentityCredentials,
			MockStsEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsAssumeRoleWithWebIdentityValidEndpoint,
			},
		},

		// "config with file envvar": {
		// 	Config: &Config{
		// 		AssumeRoleWithWebIdentity: &AssumeRoleWithWebIdentity{
		// 			RoleARN:     servicemocks.MockStsAssumeRoleWithWebIdentityArn,
		// 			SessionName: servicemocks.MockStsAssumeRoleWithWebIdentitySessionName,
		// 		},
		// 	},
		// 	SetTokenFileEnvironmentVariable: true,
		// 	ExpectedCredentialsValue:        mockdata.MockStsAssumeRoleWithWebIdentityCredentials,
		// 	MockStsEndpoints: []*servicemocks.MockEndpoint{
		// 		servicemocks.MockStsAssumeRoleWithWebIdentityValidEndpoint,
		// 	},
		// },

		"envvar overrides shared configuration": {
			Config: &Config{},
			EnvironmentVariables: map[string]string{
				"AWS_ROLE_ARN":          servicemocks.MockStsAssumeRoleWithWebIdentityArn,
				"AWS_ROLE_SESSION_NAME": servicemocks.MockStsAssumeRoleWithWebIdentitySessionName,
			},
			SetTokenFileEnvironmentVariable: true,
			SharedConfigurationFile: fmt.Sprintf(`
[default]
role_arn = %[1]s
role_session_name = %[2]s
web_identity_token_file = no-such-file
`, servicemocks.MockStsAssumeRoleWithWebIdentityAlternateArn, servicemocks.MockStsAssumeRoleWithWebIdentityAlternateSessionName),
			ExpectedCredentialsValue: mockdata.MockStsAssumeRoleWithWebIdentityCredentials,
			MockStsEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsAssumeRoleWithWebIdentityValidEndpoint,
			},
		},

		"config overrides shared configuration": {
			Config: &Config{
				AssumeRoleWithWebIdentity: &AssumeRoleWithWebIdentity{
					RoleARN:          servicemocks.MockStsAssumeRoleWithWebIdentityArn,
					SessionName:      servicemocks.MockStsAssumeRoleWithWebIdentitySessionName,
					WebIdentityToken: servicemocks.MockWebIdentityToken,
				},
			},
			SharedConfigurationFile: fmt.Sprintf(`
[default]
role_arn = %[1]s
role_session_name = %[2]s
web_identity_token_file = no-such-file
`, servicemocks.MockStsAssumeRoleWithWebIdentityAlternateArn, servicemocks.MockStsAssumeRoleWithWebIdentityAlternateSessionName),
			ExpectedCredentialsValue: mockdata.MockStsAssumeRoleWithWebIdentityCredentials,
			MockStsEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsAssumeRoleWithWebIdentityValidEndpoint,
			},
		},

		"with duration": {
			Config: &Config{
				AssumeRoleWithWebIdentity: &AssumeRoleWithWebIdentity{
					RoleARN:          servicemocks.MockStsAssumeRoleWithWebIdentityArn,
					SessionName:      servicemocks.MockStsAssumeRoleWithWebIdentitySessionName,
					WebIdentityToken: servicemocks.MockWebIdentityToken,
					Duration:         1 * time.Hour,
				},
			},
			ExpectedCredentialsValue: mockdata.MockStsAssumeRoleWithWebIdentityCredentials,
			MockStsEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsAssumeRoleWithWebIdentityValidWithOptions(map[string]string{"DurationSeconds": "3600"}),
			},
		},

		"with policy": {
			Config: &Config{
				AssumeRoleWithWebIdentity: &AssumeRoleWithWebIdentity{
					RoleARN:          servicemocks.MockStsAssumeRoleWithWebIdentityArn,
					SessionName:      servicemocks.MockStsAssumeRoleWithWebIdentitySessionName,
					WebIdentityToken: servicemocks.MockWebIdentityToken,
					Policy:           "{}",
				},
			},
			ExpectedCredentialsValue: mockdata.MockStsAssumeRoleWithWebIdentityCredentials,
			MockStsEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsAssumeRoleWithWebIdentityValidWithOptions(map[string]string{"Policy": "{}"}),
			},
		},

		"invalid empty config": {
			Config: &Config{
				AssumeRoleWithWebIdentity: &AssumeRoleWithWebIdentity{},
			},
			ExpectedCredentialsValue: mockdata.MockStsAssumeRoleWithWebIdentityCredentials,
			ExpectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Assume Role With Web Identity",
					"Role ARN was not set",
				),
			},
		},

		"invalid no token": {
			Config: &Config{
				AssumeRoleWithWebIdentity: &AssumeRoleWithWebIdentity{
					RoleARN: servicemocks.MockStsAssumeRoleWithWebIdentityArn,
				},
			},
			ExpectedCredentialsValue: mockdata.MockStsAssumeRoleWithWebIdentityCredentials,
			ExpectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Assume Role With Web Identity",
					"One of WebIdentityToken, WebIdentityTokenFile must be set",
				),
			},
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			ctx := context.Background()

			servicemocks.InitSessionTestEnv(t)

			for k, v := range testCase.EnvironmentVariables {
				t.Setenv(k, v)
			}

			closeSts, _, stsEndpoint := mockdata.GetMockedAwsApiSession("STS", testCase.MockStsEndpoints)
			defer closeSts()

			testCase.Config.StsEndpoint = stsEndpoint

			tempdir, err := os.MkdirTemp("", "temp")
			if err != nil {
				t.Fatalf("error creating temp dir: %s", err)
			}
			defer os.Remove(tempdir)
			t.Setenv("TMPDIR", tempdir)

			tokenFile, err := os.CreateTemp("", "aws-sdk-go-base-web-identity-token-file")
			if err != nil {
				t.Fatalf("unexpected error creating temporary web identity token file: %s", err)
			}
			tokenFileName := tokenFile.Name()

			defer os.Remove(tokenFileName)

			err = os.WriteFile(tokenFileName, []byte(servicemocks.MockWebIdentityToken), 0600)

			if err != nil {
				t.Fatalf("unexpected error writing web identity token file: %s", err)
			}

			if testCase.ExpandEnvVars {
				tmpdir := os.Getenv("TMPDIR")
				rel, err := filepath.Rel(tmpdir, tokenFileName)
				if err != nil {
					t.Fatalf("error making path relative: %s", err)
				}
				t.Logf("relative: %s", rel)
				tokenFileName = filepath.Join("$TMPDIR", rel)
				t.Logf("env tempfile: %s", tokenFileName)
			}

			if testCase.SetConfig {
				testCase.Config.AssumeRoleWithWebIdentity.WebIdentityTokenFile = tokenFileName
			}

			if testCase.SetTokenFileEnvironmentVariable {
				t.Setenv("AWS_WEB_IDENTITY_TOKEN_FILE", tokenFileName)
			}

			if testCase.SharedConfigurationFile != "" {
				file, err := os.CreateTemp("", "aws-sdk-go-base-shared-configuration-file")

				if err != nil {
					t.Fatalf("unexpected error creating temporary shared configuration file: %s", err)
				}

				defer os.Remove(file.Name())

				if testCase.SetSharedConfigurationFile {
					testCase.SharedConfigurationFile += fmt.Sprintf("web_identity_token_file = %s\n", tokenFileName)
				}

				err = os.WriteFile(file.Name(), []byte(testCase.SharedConfigurationFile), 0600)

				if err != nil {
					t.Fatalf("unexpected error writing shared configuration file: %s", err)
				}

				testCase.Config.SharedConfigFiles = []string{file.Name()}
			}

			testCase.Config.SkipCredsValidation = true

			ctx, awsConfig, diags := GetAwsConfig(ctx, testCase.Config)

			if diff := cmp.Diff(diags, testCase.ExpectedDiags); diff != "" {
				t.Errorf("Unexpected response (+wanted, -got): %s", diff)
			}
			if diags.HasError() {
				return
			}

			credentialsValue, err := awsConfig.Credentials.Retrieve(ctx)

			if err != nil {
				t.Fatalf("unexpected credentials Retrieve() error: %s", err)
			}

			if diff := cmp.Diff(credentialsValue, testCase.ExpectedCredentialsValue, cmpopts.IgnoreFields(aws.Credentials{}, "Expires")); diff != "" {
				t.Fatalf("unexpected credentials: (- got, + expected)\n%s", diff)
			}
		})
	}
}

func TestStsEndpoint(t *testing.T) {
	type settype int
	const (
		setNone settype = iota
		setValid
		setInvalid
	)
	testcases := map[string]struct {
		Config             Config
		SetServiceEndpoint settype
		SetEnv             string
		SetInvalidEnv      string
		// Use string at index 1 for valid endpoint url and index 2 for invalid endpoint url
		ConfigFile          string
		ExpectedCredentials aws.Credentials
	}{
		"service config": {
			Config: Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				Region:    "us-east-1",
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			SetServiceEndpoint:  setValid,
			ExpectedCredentials: mockdata.MockStaticCredentials,
		},

		"service config overrides service envvar": {
			Config: Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				Region:    "us-east-1",
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			SetServiceEndpoint:  setValid,
			SetInvalidEnv:       "AWS_ENDPOINT_URL_STS",
			ExpectedCredentials: mockdata.MockStaticCredentials,
		},

		"service config overrides base envvar": {
			Config: Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				Region:    "us-east-1",
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			SetServiceEndpoint:  setValid,
			SetInvalidEnv:       "AWS_ENDPOINT_URL",
			ExpectedCredentials: mockdata.MockStaticCredentials,
		},

		"service config overrides service config_file": {
			Config: Config{
				Profile: "default",
			},
			ConfigFile: `
[default]
aws_access_key_id = DefaultSharedCredentialsAccessKey
aws_secret_access_key = DefaultSharedCredentialsSecretKey
services = sts-test

[services sts-test]
sts =
	endpoint_url = %[2]s
`,
			SetServiceEndpoint: setValid,
			ExpectedCredentials: aws.Credentials{
				AccessKeyID:     "DefaultSharedCredentialsAccessKey",
				SecretAccessKey: "DefaultSharedCredentialsSecretKey",
				Source:          sharedConfigCredentialsProvider,
			},
		},

		"service config overrides base config_file": {
			Config: Config{
				Profile: "default",
			},
			ConfigFile: `
[default]
aws_access_key_id = DefaultSharedCredentialsAccessKey
aws_secret_access_key = DefaultSharedCredentialsSecretKey
endpoint_url = %[2]s
`,
			SetServiceEndpoint: setValid,
			ExpectedCredentials: aws.Credentials{
				AccessKeyID:     "DefaultSharedCredentialsAccessKey",
				SecretAccessKey: "DefaultSharedCredentialsSecretKey",
				Source:          sharedConfigCredentialsProvider,
			},
		},

		"service envvar": {
			Config: Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				Region:    "us-east-1",
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			SetEnv:              "AWS_ENDPOINT_URL_STS",
			ExpectedCredentials: mockdata.MockStaticCredentials,
		},

		"base envvar": {
			Config: Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				Region:    "us-east-1",
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			SetEnv:              "AWS_ENDPOINT_URL",
			ExpectedCredentials: mockdata.MockStaticCredentials,
		},

		"service envvar overrides base envvar": {
			Config: Config{
				AccessKey: servicemocks.MockStaticAccessKey,
				Region:    "us-east-1",
				SecretKey: servicemocks.MockStaticSecretKey,
			},
			SetEnv:              "AWS_ENDPOINT_URL_STS",
			SetInvalidEnv:       "AWS_ENDPOINT_URL",
			ExpectedCredentials: mockdata.MockStaticCredentials,
		},

		"service config_file": {
			Config: Config{
				Profile: "default",
			},
			ConfigFile: `
[default]
aws_access_key_id = DefaultSharedCredentialsAccessKey
aws_secret_access_key = DefaultSharedCredentialsSecretKey
services = sts-test

[services sts-test]
sts =
	endpoint_url = %[1]s
`,
			ExpectedCredentials: aws.Credentials{
				AccessKeyID:     "DefaultSharedCredentialsAccessKey",
				SecretAccessKey: "DefaultSharedCredentialsSecretKey",
				Source:          sharedConfigCredentialsProvider,
			},
		},

		"service config_file overrides base config_file": {
			Config: Config{
				Profile: "default",
			},
			ConfigFile: `
[default]
aws_access_key_id = DefaultSharedCredentialsAccessKey
aws_secret_access_key = DefaultSharedCredentialsSecretKey
services = sts-test
endpoint_url = %[2]s

[services sts-test]
sts =
	endpoint_url = %[1]s
`,
			ExpectedCredentials: aws.Credentials{
				AccessKeyID:     "DefaultSharedCredentialsAccessKey",
				SecretAccessKey: "DefaultSharedCredentialsSecretKey",
				Source:          sharedConfigCredentialsProvider,
			},
		},

		"service envvar overrides service config_file": {
			Config: Config{
				Profile: "default",
			},
			SetEnv: "AWS_ENDPOINT_URL_STS",
			ConfigFile: `
[default]
aws_access_key_id = DefaultSharedCredentialsAccessKey
aws_secret_access_key = DefaultSharedCredentialsSecretKey
services = sts-test

[services sts-test]
sts =
	endpoint_url = %[2]s
`,
			ExpectedCredentials: aws.Credentials{
				AccessKeyID:     "DefaultSharedCredentialsAccessKey",
				SecretAccessKey: "DefaultSharedCredentialsSecretKey",
				Source:          sharedConfigCredentialsProvider,
			},
		},

		"base envvar overrides service config_file": {
			Config: Config{
				Profile: "default",
			},
			SetEnv: "AWS_ENDPOINT_URL",
			ConfigFile: `
[default]
aws_access_key_id = DefaultSharedCredentialsAccessKey
aws_secret_access_key = DefaultSharedCredentialsSecretKey
services = sts-test

[services sts-test]
sts =
	endpoint_url = %[2]s
`,
			ExpectedCredentials: aws.Credentials{
				AccessKeyID:     "DefaultSharedCredentialsAccessKey",
				SecretAccessKey: "DefaultSharedCredentialsSecretKey",
				Source:          sharedConfigCredentialsProvider,
			},
		},

		"base config_file": {
			Config: Config{
				Profile: "default",
			},
			ConfigFile: `
[default]
aws_access_key_id = DefaultSharedCredentialsAccessKey
aws_secret_access_key = DefaultSharedCredentialsSecretKey
endpoint_url = %[1]s
`,
			ExpectedCredentials: aws.Credentials{
				AccessKeyID:     "DefaultSharedCredentialsAccessKey",
				SecretAccessKey: "DefaultSharedCredentialsSecretKey",
				Source:          sharedConfigCredentialsProvider,
			},
		},

		"base envvar overrides base config_file": {
			Config: Config{
				Profile: "default",
			},
			SetEnv: "AWS_ENDPOINT_URL",
			ConfigFile: `
[default]
aws_access_key_id = DefaultSharedCredentialsAccessKey
aws_secret_access_key = DefaultSharedCredentialsSecretKey
endpoint_url = %[2]s
`,
			ExpectedCredentials: aws.Credentials{
				AccessKeyID:     "DefaultSharedCredentialsAccessKey",
				SecretAccessKey: "DefaultSharedCredentialsSecretKey",
				Source:          sharedConfigCredentialsProvider,
			},
		},

		"service envvar overrides base config_file": {
			Config: Config{
				Profile: "default",
			},
			SetEnv: "AWS_ENDPOINT_URL_STS",
			ConfigFile: `
[default]
aws_access_key_id = DefaultSharedCredentialsAccessKey
aws_secret_access_key = DefaultSharedCredentialsSecretKey
endpoint_url = %[2]s
`,
			ExpectedCredentials: aws.Credentials{
				AccessKeyID:     "DefaultSharedCredentialsAccessKey",
				SecretAccessKey: "DefaultSharedCredentialsSecretKey",
				Source:          sharedConfigCredentialsProvider,
			},
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			servicemocks.InitSessionTestEnv(t)

			ctx := context.Background()

			ts := servicemocks.MockAwsApiServer("STS", []*servicemocks.MockEndpoint{
				servicemocks.MockStsGetCallerIdentityValidEndpoint,
			})
			defer ts.Close()
			stsEndpoint := ts.URL

			invalidTS := servicemocks.MockAwsApiServer("STS", []*servicemocks.MockEndpoint{
				servicemocks.MockStsGetCallerIdentityInvalidEndpointAccessDenied,
			})
			defer invalidTS.Close()
			stsInvalidEndpoint := invalidTS.URL

			if testcase.SetServiceEndpoint == setValid {
				testcase.Config.StsEndpoint = stsEndpoint
			}
			if testcase.SetEnv != "" {
				t.Setenv(testcase.SetEnv, stsEndpoint)
			}
			if testcase.SetInvalidEnv != "" {
				t.Setenv(testcase.SetInvalidEnv, stsInvalidEndpoint)
			}
			if testcase.ConfigFile != "" {
				tempDir := t.TempDir()
				filename := writeSharedConfigFile(t, &testcase.Config, tempDir, fmt.Sprintf(testcase.ConfigFile, stsEndpoint, stsInvalidEndpoint))
				testcase.ExpectedCredentials.Source = sharedConfigCredentialsSource(filename)
			}

			ctx, awsConfig, diags := GetAwsConfig(ctx, &testcase.Config)

			if diff := cmp.Diff(diags, diag.Diagnostics{}); diff != "" {
				t.Errorf("Unexpected response (+wanted, -got): %s", diff)
			}
			if diags.HasError() {
				return
			}

			credentialsValue, err := awsConfig.Credentials.Retrieve(ctx)
			if err != nil {
				t.Fatalf("unexpected credentials Retrieve() error: %s", err)
			}

			if diff := cmp.Diff(credentialsValue, testcase.ExpectedCredentials, cmpopts.IgnoreFields(aws.Credentials{}, "Expires")); diff != "" {
				t.Fatalf("unexpected credentials: (- got, + expected)\n%s", diff)
			}
		})
	}
}

var _ configtesting.TestDriver = &testDriver{}

type testDriver struct {
	mode configtesting.TestMode
}

func (t *testDriver) Init(mode configtesting.TestMode) {
	t.mode = mode
}

func (t testDriver) TestCase() configtesting.TestCaseDriver {
	if t.mode == configtesting.TestModeInvalid {
		panic("TestDriver not initialized")
	}
	return &testCaseDriver{
		mode: t.mode,
	}
}

var _ configtesting.TestCaseDriver = &testCaseDriver{}

type testCaseDriver struct {
	mode   configtesting.TestMode
	config configurer
}

func (d *testCaseDriver) Configuration(fs []configtesting.ConfigFunc) configtesting.Configurer {
	for _, f := range fs {
		f(&d.config)
	}
	return &d.config
}

func (d testCaseDriver) Setup(_ *testing.T) {
	// Noop
}

// TODO: Make work with expected diffs
func (d testCaseDriver) Apply(ctx context.Context, t *testing.T) (context.Context, configtesting.Thing) {
	t.Helper()

	if d.mode == configtesting.TestModeLocal {
		d.config.setSkipCredsValidation(true)
	}

	config := Config(d.config)
	ctx, awsConfig, diags := GetAwsConfig(ctx, &config)

	if diff := cmp.Diff(diags, diag.Diagnostics{}); diff != "" {
		t.Errorf("unexpected diagnostics difference: %s", diff)
	}

	if diags.HasError() {
		t.FailNow()
	}

	return ctx, thing(awsConfig)
}

var _ configtesting.Configurer = &configurer{}

type configurer Config

func (c *configurer) SetAccessKey(s string) {
	c.AccessKey = s
}

func (c *configurer) SetSecretKey(s string) {
	c.SecretKey = s
}

func (c *configurer) SetProfile(s string) {
	c.Profile = s
}

func (c *configurer) SetUseFIPSEndpoint(b bool) {
	c.UseFIPSEndpoint = b
}

func (c *configurer) AddEndpoint(k, v string) {
	switch k {
	case "sso":
		c.SsoEndpoint = v
	default:
		panic(fmt.Sprintf(`invalid endpoint "%s"`, k))
	}
}

func (c *configurer) AddSharedConfigFile(f string) {
	c.SharedConfigFiles = append(c.SharedConfigFiles, f)
}

func (c *configurer) setSkipCredsValidation(b bool) {
	c.SkipCredsValidation = b
}

var _ configtesting.Thing = thing{}

type thing aws.Config

func (t thing) GetCredentials() aws.CredentialsProvider {
	return t.Credentials
}

func (t thing) GetRegion() string {
	return t.Region
}

func (t thing) GetAwsConfig() aws.Config {
	return aws.Config(t)
}

func TestSSO(t *testing.T) {
	configtesting.SSO(t, &testDriver{})
}

func TestLegacySSO(t *testing.T) {
	configtesting.LegacySSO(t, &testDriver{})
}

func TestGetAwsConfigWithAccountIDAndPartition(t *testing.T) {
	servicemocks.InitSessionTestEnv(t)

	testCases := []struct {
		desc              string
		config            *Config
		expectedAcctID    string
		expectedPartition string
		expectError       bool
		mockStsEndpoints  []*servicemocks.MockEndpoint
		ExpectedDiags     diag.Diagnostics
	}{
		{
			desc: "StandardProvider_Config",
			config: &Config{
				AccessKey: "MockAccessKey",
				SecretKey: "MockSecretKey",
				Region:    "us-west-2"},
			expectedAcctID: "222222222222", expectedPartition: "aws",
			mockStsEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsGetCallerIdentityValidEndpoint,
			},
		},
		{
			desc: "SkipCredsValidation_Config",
			config: &Config{
				AccessKey:           "MockAccessKey",
				SecretKey:           "MockSecretKey",
				Region:              "us-west-2",
				SkipCredsValidation: true},
			expectedAcctID: "222222222222", expectedPartition: "aws",
			mockStsEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsGetCallerIdentityValidEndpoint,
			},
		},
		{
			desc: "SkipRequestingAccountId_Config",
			config: &Config{
				AccessKey:               "MockAccessKey",
				SecretKey:               "MockSecretKey",
				Region:                  "us-west-2",
				SkipCredsValidation:     true,
				SkipRequestingAccountId: true},
			expectedAcctID: "", expectedPartition: "aws",
			mockStsEndpoints: []*servicemocks.MockEndpoint{},
		},
		{
			desc: "WithAssumeRole",
			config: &Config{
				AccessKey: "MockAccessKey",
				SecretKey: "MockSecretKey",
				Region:    "us-west-2",
				AssumeRole: []AssumeRole{{
					RoleARN:     servicemocks.MockStsAssumeRoleArn,
					SessionName: servicemocks.MockStsAssumeRoleSessionName,
				}},
			},
			expectedAcctID: "555555555555", expectedPartition: "aws",
			mockStsEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsAssumeRoleValidEndpoint,
				servicemocks.MockStsGetCallerIdentityValidAssumedRoleEndpoint,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.desc, func(t *testing.T) {
			ts := servicemocks.MockAwsApiServer("STS", testCase.mockStsEndpoints)
			defer ts.Close()
			testCase.config.StsEndpoint = ts.URL

			ctx, awsConfig, diags := GetAwsConfig(context.Background(), testCase.config)
			if diags.HasError() {
				t.Fatalf("error in GetAwsConfig(): %v", diags)
			}

			acctID, part, diags := GetAwsAccountIDAndPartition(ctx, awsConfig, testCase.config)

			if diff := cmp.Diff(diags, testCase.ExpectedDiags); diff != "" {
				t.Errorf("Unexpected response (+wanted, -got): %s", diff)
			}
			if diags.HasError() {
				return
			}

			if acctID != testCase.expectedAcctID {
				t.Errorf("expected account ID (%s), got: %s", testCase.expectedAcctID, acctID)
			}

			if part != testCase.expectedPartition {
				t.Errorf("expected partition (%s), got: %s", testCase.expectedPartition, part)
			}
		})
	}
}

type mockRetryableError struct{ b bool }

func (m mockRetryableError) RetryableError() bool { return m.b }
func (m mockRetryableError) Error() string {
	return fmt.Sprintf("mock retryable %t", m.b)
}

func TestRetryHandlers(t *testing.T) {
	const maxRetries = 10

	testcases := map[string]struct {
		NextHandler   func() middleware.FinalizeHandler
		ExpectResults retry.AttemptResults
		Err           error
	}{
		"stops at maxRetries for retryable errors": {
			NextHandler: func() middleware.FinalizeHandler {
				num := 0
				reqsErrs := make([]error, maxRetries)
				for i := range maxRetries {
					reqsErrs[i] = mockRetryableError{b: true}
				}
				return middleware.FinalizeHandlerFunc(func(ctx context.Context, in middleware.FinalizeInput) (out middleware.FinalizeOutput, metadata middleware.Metadata, err error) {
					if num >= len(reqsErrs) {
						err = fmt.Errorf("more requests than expected")
					} else {
						err = reqsErrs[num]
						num++
					}
					return out, metadata, err
				})
			},
			Err: fmt.Errorf("exceeded maximum number of attempts"),
			ExpectResults: func() retry.AttemptResults {
				results := retry.AttemptResults{
					Results: make([]retry.AttemptResult, maxRetries),
				}
				for i := range maxRetries - 1 {
					results.Results[i] = retry.AttemptResult{
						Err:       mockRetryableError{b: true},
						Retryable: true,
						Retried:   true,
					}
				}
				results.Results[maxRetries-1] = retry.AttemptResult{
					Err:       &retry.MaxAttemptsError{Attempt: maxRetries, Err: mockRetryableError{b: true}},
					Retryable: true,
				}
				return results
			}(),
		},
		"stops at MaxNetworkRetryCount for 'no such host' errors": {
			NextHandler: func() middleware.FinalizeHandler {
				num := 0
				reqsErrs := make([]error, constants.MaxNetworkRetryCount)
				for i := range constants.MaxNetworkRetryCount {
					reqsErrs[i] = &net.OpError{Op: "dial", Err: errors.New("no such host")}
				}
				return middleware.FinalizeHandlerFunc(func(ctx context.Context, in middleware.FinalizeInput) (out middleware.FinalizeOutput, metadata middleware.Metadata, err error) {
					if num >= len(reqsErrs) {
						err = fmt.Errorf("more requests than expected")
					} else {
						err = reqsErrs[num]
						num++
					}
					return out, metadata, err
				})
			},
			Err: fmt.Errorf("exceeded maximum number of attempts"),
			ExpectResults: func() retry.AttemptResults {
				results := retry.AttemptResults{
					Results: make([]retry.AttemptResult, constants.MaxNetworkRetryCount),
				}
				for i := range constants.MaxNetworkRetryCount - 1 {
					results.Results[i] = retry.AttemptResult{
						Err:       &net.OpError{Op: "dial", Err: errors.New("no such host")},
						Retryable: true,
						Retried:   true,
					}
				}
				results.Results[constants.MaxNetworkRetryCount-1] = retry.AttemptResult{
					Err:       &retry.MaxAttemptsError{Attempt: constants.MaxNetworkRetryCount, Err: &net.OpError{Op: "dial", Err: errors.New("no such host")}},
					Retryable: true,
				}
				return results
			}(),
		},
		"stops at MaxNetworkRetryCount for 'connection refused' errors": {
			NextHandler: func() middleware.FinalizeHandler {
				num := 0
				reqsErrs := make([]error, constants.MaxNetworkRetryCount)
				for i := range constants.MaxNetworkRetryCount {
					reqsErrs[i] = &net.OpError{Op: "dial", Err: errors.New("connection refused")}
				}
				return middleware.FinalizeHandlerFunc(func(ctx context.Context, in middleware.FinalizeInput) (out middleware.FinalizeOutput, metadata middleware.Metadata, err error) {
					if num >= len(reqsErrs) {
						err = fmt.Errorf("more requests than expected")
					} else {
						err = reqsErrs[num]
						num++
					}
					return out, metadata, err
				})
			},
			Err: fmt.Errorf("exceeded maximum number of attempts"),
			ExpectResults: func() retry.AttemptResults {
				results := retry.AttemptResults{
					Results: make([]retry.AttemptResult, constants.MaxNetworkRetryCount),
				}
				for i := range constants.MaxNetworkRetryCount - 1 {
					results.Results[i] = retry.AttemptResult{
						Err:       &net.OpError{Op: "dial", Err: errors.New("connection refused")},
						Retryable: true,
						Retried:   true,
					}
				}
				results.Results[constants.MaxNetworkRetryCount-1] = retry.AttemptResult{
					Err:       &retry.MaxAttemptsError{Attempt: constants.MaxNetworkRetryCount, Err: &net.OpError{Op: "dial", Err: errors.New("connection refused")}},
					Retryable: true,
				}
				return results
			}(),
		},
		"no retries for ExpiredToken": {
			NextHandler: func() middleware.FinalizeHandler {
				num := 0
				reqsErrs := make([]error, 2)
				for i := range 2 {
					reqsErrs[i] = &smithy.OperationError{
						ServiceID:     "STS",
						OperationName: "GetCallerIdentity",
						Err: &smithyhttp.ResponseError{
							Response: &smithyhttp.Response{
								Response: &http.Response{
									StatusCode: 403,
								},
							},
							Err: &smithy.GenericAPIError{
								Code:    "ExpiredToken",
								Message: "The security token included in the request is expired",
							},
						},
					}
				}
				return middleware.FinalizeHandlerFunc(func(ctx context.Context, in middleware.FinalizeInput) (out middleware.FinalizeOutput, metadata middleware.Metadata, err error) {
					if num >= len(reqsErrs) {
						err = fmt.Errorf("more requests than expected")
					} else {
						err = reqsErrs[num]
						num++
					}
					return out, metadata, err
				})
			},
			Err: &smithy.OperationError{
				ServiceID:     "STS",
				OperationName: "GetCallerIdentity",
				Err: &smithyhttp.ResponseError{
					Response: &smithyhttp.Response{
						Response: &http.Response{
							StatusCode: 403,
						},
					},
					Err: &smithy.GenericAPIError{
						Code:    "ExpiredToken",
						Message: "The security token included in the request is expired",
					},
				},
			},
			ExpectResults: func() retry.AttemptResults {
				results := retry.AttemptResults{
					Results: make([]retry.AttemptResult, 1),
				}
				results.Results[0] = retry.AttemptResult{
					Err: &smithy.OperationError{
						ServiceID:     "STS",
						OperationName: "GetCallerIdentity",
						Err: &smithyhttp.ResponseError{
							Response: &smithyhttp.Response{
								Response: &http.Response{
									StatusCode: 403,
								},
							},
							Err: &smithy.GenericAPIError{
								Code:    "ExpiredToken",
								Message: "The security token included in the request is expired",
							},
						},
					},
				}
				return results
			}(),
		},
		"stops at maxRetries for other network errors": {
			NextHandler: func() middleware.FinalizeHandler {
				num := 0
				reqsErrs := make([]error, maxRetries)
				for i := range maxRetries {
					reqsErrs[i] = &net.OpError{Op: "dial", Err: errors.New("other error")}
				}
				return middleware.FinalizeHandlerFunc(func(ctx context.Context, in middleware.FinalizeInput) (out middleware.FinalizeOutput, metadata middleware.Metadata, err error) {
					if num >= len(reqsErrs) {
						err = fmt.Errorf("more requests than expected")
					} else {
						err = reqsErrs[num]
						num++
					}
					return out, metadata, err
				})
			},
			Err: fmt.Errorf("exceeded maximum number of attempts"),
			ExpectResults: func() retry.AttemptResults {
				results := retry.AttemptResults{
					Results: make([]retry.AttemptResult, maxRetries),
				}
				for i := range maxRetries - 1 {
					results.Results[i] = retry.AttemptResult{
						Err:       &net.OpError{Op: "dial", Err: errors.New("other error")},
						Retryable: true,
						Retried:   true,
					}
				}
				results.Results[maxRetries-1] = retry.AttemptResult{
					Err:       &retry.MaxAttemptsError{Attempt: maxRetries, Err: &net.OpError{Op: "dial", Err: errors.New("other error")}},
					Retryable: true,
				}
				return results
			}(),
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			servicemocks.InitSessionTestEnv(t)

			config := &Config{
				AccessKey:           servicemocks.MockStaticAccessKey,
				Region:              "us-east-1",
				MaxRetries:          maxRetries,
				SecretKey:           servicemocks.MockStaticSecretKey,
				SkipCredsValidation: true,
			}
			ctx, awsConfig, diags := GetAwsConfig(context.Background(), config)
			if diags.HasError() {
				t.Fatalf("error in GetAwsConfig(): %v", diags)
			}

			if awsConfig.Retryer == nil {
				t.Fatal("No Retryer configured on awsConfig")
			}

			am := retry.NewAttemptMiddleware(&withNoDelay{
				Retryer: awsConfig.Retryer(),
			}, func(i any) any {
				return i
			})
			_, metadata, err := am.HandleFinalize(ctx, middleware.FinalizeInput{Request: nil}, testcase.NextHandler())
			if err != nil && testcase.Err == nil {
				t.Errorf("expect no error, got %v", err)
			} else if err == nil && testcase.Err != nil {
				t.Errorf("expect error, got none")
			} else if err != nil && testcase.Err != nil {
				if !strings.Contains(err.Error(), testcase.Err.Error()) {
					t.Errorf("expect %v, got %v", testcase.Err, err)
				}
			}

			attemptResults, ok := retry.GetAttemptResults(metadata)
			if !ok {
				t.Fatalf("expected metadata to contain attempt results, got none")
			}
			if e, a := testcase.ExpectResults, attemptResults; !reflect.DeepEqual(e, a) {
				t.Fatalf("expected %v, got %v", e, a)
			}

			for i, attempt := range attemptResults.Results {
				_, ok := retry.GetAttemptResults(attempt.ResponseMetadata)
				if ok {
					t.Errorf("expect no attempt to include AttemptResults metadata, %v does, %#v", i, attempt)
				}
			}
		})
	}
}

// TestSharedConfigFileParsing prevents regression in shared config file parsing
// * https://github.com/aws/aws-sdk-go-v2/issues/2349: indented keys
// * https://github.com/aws/aws-sdk-go-v2/issues/2363: leading whitespace
// * https://github.com/aws/aws-sdk-go-v2/issues/2369: trailing `#` in, e.g. SSO Start URLs
func TestSharedConfigFileParsing(t *testing.T) {
	configtesting.SharedConfigFileParsing(t, &testDriver{})
}

type withNoDelay struct {
	aws.Retryer
}

func (r *withNoDelay) RetryDelay(attempt int, err error) (time.Duration, error) {
	delay, delayErr := r.Retryer.RetryDelay(attempt, err)
	if delayErr != nil {
		return delay, delayErr
	}

	return 0 * time.Second, nil
}

func TestLogger_TfLog(t *testing.T) {
	ctx := context.Background()
	var buf bytes.Buffer
	ctx = tflogtest.RootLogger(ctx, &buf)

	servicemocks.InitSessionTestEnv(t)

	ctx, logger := logging.NewTfLogger(ctx)

	config := &Config{
		AccessKey: servicemocks.MockStaticAccessKey,
		Logger:    logger,
		Region:    "us-east-1",
		SecretKey: servicemocks.MockStaticSecretKey,
	}

	ts := servicemocks.MockAwsApiServer("STS", []*servicemocks.MockEndpoint{
		servicemocks.MockStsGetCallerIdentityValidEndpoint,
	})
	defer ts.Close()
	config.StsEndpoint = ts.URL

	expectedName := fmt.Sprintf("provider.%s", loggerName)

	ctx, awsConfig, diags := GetAwsConfig(ctx, config)
	if diags.HasError() {
		t.Fatalf("error in GetAwsConfig(): %v", diags)
	}

	lines, err := tflogtest.MultilineJSONDecode(&buf)
	if err != nil {
		t.Fatalf("GetAwsConfig: decoding log lines: %s", err)
	}

	if len(lines) == 0 {
		t.Fatalf("expected log entries, had none")
	}
	for i, line := range lines {
		if a, e := line["@module"], expectedName; a != e {
			t.Errorf("GetAwsConfig: line %d: expected module %q, got %q", i+1, e, a)
		}
	}
	var requestLines []map[string]any
	var responseLines []map[string]any
	for _, line := range lines {
		if line["@message"] == "HTTP Request Sent" {
			requestLines = append(requestLines, line)
		} else if line["@message"] == "HTTP Response Received" {
			responseLines = append(responseLines, line)
		}
	}
	if len(requestLines) != 1 {
		t.Fatalf("expected 1 request line, got %d", len(requestLines))
	}
	requestLine := requestLines[0]
	maps.DeleteFunc(requestLine, func(k string, _ any) bool {
		return strings.HasPrefix(k, "@")
	})

	for _, k := range []string{
		string(semconv.HTTPUserAgentKey),
		string(logging.RequestHeaderAttributeKey("Amz-Sdk-Invocation-Id")),
		string(logging.RequestHeaderAttributeKey("Authorization")),
		string(logging.RequestHeaderAttributeKey("X-Amz-Date")),
	} {
		_, ok := requestLine[k]
		if !ok {
			t.Errorf("expected a value for request attribute %q", k)
		}
		delete(requestLine, k)
	}

	tsUrl, _ := url.Parse(ts.URL)
	if tsUrl.Path == "" {
		tsUrl.Path = "/"
	}

	port, err := strconv.ParseFloat(tsUrl.Port(), 32)
	if err != nil {
		t.Errorf("error parsing URL port %q: %s", tsUrl.Port(), err)
	}

	requestBody := "Action=GetCallerIdentity&Version=2011-06-15"

	expectedRequest := map[string]any{
		// AWS attributes
		string(semconv.RPCSystemKey):  otelaws.AWSSystemVal,
		string(semconv.RPCServiceKey): sts.ServiceID,
		string(otelaws.RegionKey):     "us-east-1",
		string(semconv.RPCMethodKey):  "GetCallerIdentity",
		// Custom attributes
		string(logging.AwsSdkKey):         awsSdkGoV2Val,
		string(logging.CustomEndpointKey): true,
		"http.request.body":               requestBody + "\n",
		// HTTP attributes
		string(semconv.HTTPMethodKey):                                "POST",
		string(logging.RequestHeaderAttributeKey("Amz-Sdk-Request")): "attempt=1; max=3",
		string(logging.RequestHeaderAttributeKey("Content-Type")):    "application/x-www-form-urlencoded",
		string(semconv.HTTPRequestContentLengthKey):                  float64(len(requestBody)),
		string(semconv.HTTPURLKey):                                   tsUrl.String(),
		// Net attributes
		string(semconv.NetPeerNameKey): tsUrl.Hostname(),
		string(semconv.NetPeerPortKey): port,
	}

	if diff := cmp.Diff(requestLine, expectedRequest); diff != "" {
		t.Fatalf("unexpected request attributes: (- got, + expected)\n%s", diff)
	}

	if len(responseLines) != 1 {
		t.Fatalf("expected 1 response line, got %d", len(responseLines))
	}
	responseLine := responseLines[0]
	maps.DeleteFunc(responseLine, func(k string, _ any) bool {
		return strings.HasPrefix(k, "@")
	})

	for _, k := range []string{
		string("http.duration"),
		string("http.response.body"),
		string(logging.ResponseHeaderAttributeKey("Date")),
		string(semconv.HTTPResponseContentLengthKey),
		string(logging.ResponseHeaderAttributeKey("X-Amzn-Requestid")),
	} {
		_, ok := responseLine[k]
		if !ok {
			t.Errorf("expected a value for response attribute %q", k)
		}
		delete(responseLine, k)
	}

	expectedResponse := map[string]any{
		// AWS attributes
		string(semconv.RPCSystemKey):  otelaws.AWSSystemVal,
		string(semconv.RPCServiceKey): sts.ServiceID,
		string(otelaws.RegionKey):     "us-east-1",
		string(semconv.RPCMethodKey):  "GetCallerIdentity",
		// Custom attributes
		string(logging.AwsSdkKey):         awsSdkGoV2Val,
		string(logging.CustomEndpointKey): true,
		// HTTP attributes
		string(semconv.HTTPStatusCodeKey):                          float64(http.StatusOK),
		string(logging.ResponseHeaderAttributeKey("Content-Type")): "text/xml",
	}

	if diff := cmp.Diff(responseLine, expectedResponse); diff != "" {
		t.Fatalf("unexpected response attributes: (- got, + expected)\n%s", diff)
	}

	_, _, diags = GetAwsAccountIDAndPartition(ctx, awsConfig, config)
	if diags.HasError() {
		t.Fatalf("GetAwsAccountIDAndPartition: unexpected '%[1]T': %[1]s", err)
	}

	lines, err = tflogtest.MultilineJSONDecode(&buf)
	if err != nil {
		t.Fatalf("GetAwsAccountIDAndPartition: decoding log lines: %s", err)
	}

	if len(lines) == 0 {
		t.Fatalf("expected log entries, had none")
	}
	for i, line := range lines {
		if a, e := line["@module"], expectedName; a != e {
			t.Errorf("GetAwsAccountIDAndPartition: line %d: expected module %q, got %q", i+1, e, a)
		}
	}
}

func TestLoggerDefaultMasking_TfLog(t *testing.T) {
	ctx := context.Background()
	var buf bytes.Buffer
	ctx = tflogtest.RootLogger(ctx, &buf)

	servicemocks.InitSessionTestEnv(t)

	config := &Config{
		AccessKey: servicemocks.MockStaticAccessKey,
		Region:    "us-east-1",
		SecretKey: servicemocks.MockStaticSecretKey,
	}

	ts := servicemocks.MockAwsApiServer("STS", []*servicemocks.MockEndpoint{
		servicemocks.MockStsGetCallerIdentityValidEndpoint,
	})
	defer ts.Close()
	config.StsEndpoint = ts.URL

	ctx, _, diags := GetAwsConfig(ctx, config)
	if diags.HasError() {
		t.Fatalf("error in GetAwsConfig(): %v", diags)
	}

	buf.Reset()

	tflog.Info(ctx, "message", map[string]any{
		"id": "AKIAI44QH8DHBEXAMPLE",
	})

	lines, err := tflogtest.MultilineJSONDecode(&buf)
	if err != nil {
		t.Fatalf("decoding log lines: %s", err)
	}

	if l := len(lines); l != 1 {
		t.Fatalf("expected 1 log entry, got %d", l)
	}

	line := lines[0]
	if a, e := line["id"], "***"; a != e {
		t.Errorf("expected %q, got %q", e, a)
	}
}

func TestLogger_HcLog(t *testing.T) {
	ctx := context.Background()

	rootName := "hc-log-test"
	expectedName := rootName + "." + loggerName

	var buf bytes.Buffer
	hclogger := configureHcLogger(rootName, &buf)

	servicemocks.InitSessionTestEnv(t)

	ctx, logger := logging.NewHcLogger(ctx, hclogger)

	config := &Config{
		AccessKey: servicemocks.MockStaticAccessKey,
		Logger:    logger,
		Region:    "us-east-1",
		SecretKey: servicemocks.MockStaticSecretKey,
	}

	ts := servicemocks.MockAwsApiServer("STS", []*servicemocks.MockEndpoint{
		servicemocks.MockStsGetCallerIdentityValidEndpoint,
	})
	defer ts.Close()
	config.StsEndpoint = ts.URL

	ctx, awsConfig, diags := GetAwsConfig(ctx, config)
	if diags.HasError() {
		t.Fatalf("error in GetAwsConfig(): %v", diags)
	}

	lines, err := tflogtest.MultilineJSONDecode(&buf)
	if err != nil {
		t.Fatalf("GetAwsConfig: decoding log lines: %s", err)
	}

	if len(lines) == 0 {
		t.Fatalf("expected log entries, had none")
	}
	for i, line := range lines {
		if a, e := line["@module"], expectedName; a != e {
			t.Errorf("GetAwsConfig: line %d: expected module %q, got %q", i+1, e, a)
		}
	}
	var requestLines []map[string]any
	var responseLines []map[string]any
	for _, line := range lines {
		if line["@message"] == "HTTP Request Sent" {
			requestLines = append(requestLines, line)
		} else if line["@message"] == "HTTP Response Received" {
			responseLines = append(responseLines, line)
		}
	}
	if len(requestLines) != 1 {
		t.Fatalf("expected 1 request line, got %d", len(requestLines))
	}
	requestLine := requestLines[0]
	maps.DeleteFunc(requestLine, func(k string, _ any) bool {
		return strings.HasPrefix(k, "@")
	})

	for _, k := range []string{
		string(semconv.HTTPUserAgentKey),
		string(logging.RequestHeaderAttributeKey("Amz-Sdk-Invocation-Id")),
		string(logging.RequestHeaderAttributeKey("Authorization")),
		string(logging.RequestHeaderAttributeKey("X-Amz-Date")),
	} {
		_, ok := requestLine[k]
		if !ok {
			t.Errorf("expected a value for request attribute %q", k)
		}
		delete(requestLine, k)
	}

	tsUrl, _ := url.Parse(ts.URL)
	if tsUrl.Path == "" {
		tsUrl.Path = "/"
	}

	port, err := strconv.ParseFloat(tsUrl.Port(), 32)
	if err != nil {
		t.Errorf("error parsing URL port %q: %s", tsUrl.Port(), err)
	}

	requestBody := "Action=GetCallerIdentity&Version=2011-06-15"

	expectedRequest := map[string]any{
		// AWS attributes
		string(semconv.RPCSystemKey):  otelaws.AWSSystemVal,
		string(semconv.RPCServiceKey): sts.ServiceID,
		string(otelaws.RegionKey):     "us-east-1",
		string(semconv.RPCMethodKey):  "GetCallerIdentity",
		// Custom attributes
		string(logging.AwsSdkKey):         awsSdkGoV2Val,
		string(logging.CustomEndpointKey): true,
		"http.request.body":               requestBody + "\n",
		// HTTP attributes
		string(semconv.HTTPMethodKey):                                "POST",
		string(logging.RequestHeaderAttributeKey("Amz-Sdk-Request")): "attempt=1; max=3",
		string(logging.RequestHeaderAttributeKey("Content-Type")):    "application/x-www-form-urlencoded",
		string(semconv.HTTPRequestContentLengthKey):                  float64(len(requestBody)),
		string(semconv.HTTPURLKey):                                   tsUrl.String(),
		// Net attributes
		string(semconv.NetPeerNameKey): tsUrl.Hostname(),
		string(semconv.NetPeerPortKey): port,
	}

	if diff := cmp.Diff(requestLine, expectedRequest); diff != "" {
		t.Fatalf("unexpected request attributes: (- got, + expected)\n%s", diff)
	}

	if len(responseLines) != 1 {
		t.Fatalf("expected 1 response line, got %d", len(responseLines))
	}
	responseLine := responseLines[0]
	maps.DeleteFunc(responseLine, func(k string, _ any) bool {
		return strings.HasPrefix(k, "@")
	})

	for _, k := range []string{
		string("http.duration"),
		string("http.response.body"),
		string(logging.ResponseHeaderAttributeKey("Date")),
		string(semconv.HTTPResponseContentLengthKey),
		string(logging.ResponseHeaderAttributeKey("X-Amzn-Requestid")),
	} {
		_, ok := responseLine[k]
		if !ok {
			t.Errorf("expected a value for response attribute %q", k)
		}
		delete(responseLine, k)
	}

	expectedResponse := map[string]any{
		// AWS attributes
		string(semconv.RPCSystemKey):  otelaws.AWSSystemVal,
		string(semconv.RPCServiceKey): sts.ServiceID,
		string(otelaws.RegionKey):     "us-east-1",
		string(semconv.RPCMethodKey):  "GetCallerIdentity",
		// Custom attributes
		string(logging.AwsSdkKey):         awsSdkGoV2Val,
		string(logging.CustomEndpointKey): true,
		// HTTP attributes
		string(semconv.HTTPStatusCodeKey):                          float64(http.StatusOK),
		string(logging.ResponseHeaderAttributeKey("Content-Type")): "text/xml",
	}

	if diff := cmp.Diff(responseLine, expectedResponse); diff != "" {
		t.Fatalf("unexpected response attributes: (- got, + expected)\n%s", diff)
	}

	_, _, diags = GetAwsAccountIDAndPartition(ctx, awsConfig, config)
	if diags.HasError() {
		t.Fatalf("GetAwsAccountIDAndPartition: unexpected '%[1]T': %[1]s", err)
	}

	lines, err = tflogtest.MultilineJSONDecode(&buf)
	if err != nil {
		t.Fatalf("GetAwsAccountIDAndPartition: decoding log lines: %s", err)
	}

	if len(lines) == 0 {
		t.Fatalf("expected log entries, had none")
	}
	for i, line := range lines {
		if a, e := line["@module"], expectedName; a != e {
			t.Errorf("GetAwsAccountIDAndPartition: line %d: expected module %q, got %q", i+1, e, a)
		}
	}
}

// configureHcLogger configures the default logger with settings suitable for testing:
//
//   - Log level set to TRACE
//   - Written to the io.Writer passed in, such as a bytes.Buffer
//   - Log entries are in JSON format, and can be decoded using multilineJSONDecode
//   - Caller information is not included
//   - Timestamp is not included
func configureHcLogger(name string, output io.Writer) hclog.Logger {
	logger := hclog.NewInterceptLogger(&hclog.LoggerOptions{
		Name:              name,
		Level:             hclog.Trace,
		Output:            output,
		IndependentLevels: true,
		JSONFormat:        true,
		IncludeLocation:   false,
		DisableTime:       true,
	})

	return logger
}

func writeSharedConfigFile(t *testing.T, config *Config, tempDir, content string) string {
	t.Helper()

	file, err := os.Create(filepath.Join(tempDir, "aws-sdk-go-base-shared-configuration-file"))
	if err != nil {
		t.Fatalf("creating shared configuration file: %s", err)
	}

	_, err = file.WriteString(content)
	if err != nil {
		t.Fatalf(" writing shared configuration file: %s", err)
	}

	config.SharedConfigFiles = append(config.SharedConfigFiles, file.Name())

	return file.Name()
}
