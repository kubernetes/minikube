// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package awsbase

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/hashicorp/aws-sdk-go-base/v2/internal/test"
	"github.com/hashicorp/aws-sdk-go-base/v2/servicemocks"
)

// TestAWSGetCredentials_ec2Imds is designed to test the scenario of running Terraform
// from an EC2 instance, without environment variables or manually supplied
// credentials.
func TestAWSGetCredentials_ec2Imds(t *testing.T) {
	// clear AWS_* environment variables
	resetEnv := servicemocks.UnsetEnv(t)
	defer resetEnv()

	ctx := test.Context(t)

	// capture the test server's close method, to call after the test returns
	ts := servicemocks.AwsMetadataApiMock(append(
		servicemocks.Ec2metadata_securityCredentialsEndpoints,
		servicemocks.Ec2metadata_instanceIdEndpoint,
		servicemocks.Ec2metadata_iamInfoEndpoint,
	))
	defer ts()

	// An empty config, no key supplied
	cfg := Config{}

	creds, source, err := getCredentialsProvider(ctx, &cfg)
	if err != nil {
		t.Fatalf("unexpected '%[1]T' error getting credentials provider: %[1]s", err)
	}

	if a, e := source, ec2rolecreds.ProviderName; a != e {
		t.Errorf("Expected initial source to be %q, %q given", e, a)
	}

	validateCredentialsProvider(ctx, creds, "Ec2MetadataAccessKey", "Ec2MetadataSecretKey", "Ec2MetadataSessionToken", ec2rolecreds.ProviderName, t)
	testCredentialsProviderWrappedWithCache(creds, t)
}

func TestAWSGetCredentials_shouldErrorWithInvalidEc2ImdsEndpoint(t *testing.T) {
	ctx := test.Context(t)

	resetEnv := servicemocks.UnsetEnv(t)
	defer resetEnv()
	// capture the test server's close method, to call after the test returns
	ts := servicemocks.InvalidEC2MetadataEndpoint(t)
	defer ts()

	// An empty config, no key supplied
	cfg := Config{}

	_, _, diags := getCredentialsProvider(ctx, &cfg)
	if diags == nil {
		t.Fatal("expected error returned when getting creds w/ invalid EC2 IMDS endpoint")
	}
	if !ContainsNoValidCredentialSourcesError(diags) {
		t.Fatalf("expected NoValidCredentialSourcesError, got '%[1]T': %[1]s", diags)
	}
}

func TestAWSGetCredentials_sharedCredentialsFile(t *testing.T) {
	ctx := test.Context(t)

	resetEnv := servicemocks.UnsetEnv(t)
	defer resetEnv()

	t.Setenv("AWS_PROFILE", "myprofile")

	fileEnvName := writeCredentialsFile(credentialsFileContentsEnv, t)
	defer os.Remove(fileEnvName)

	fileParamName := writeCredentialsFile(credentialsFileContentsParam, t)
	defer os.Remove(fileParamName)

	t.Setenv("AWS_SHARED_CREDENTIALS_FILE", fileEnvName)

	// Confirm AWS_SHARED_CREDENTIALS_FILE is working
	credsEnv, source, err := getCredentialsProvider(ctx, &Config{
		Profile: "myprofile",
	})
	if err != nil {
		t.Fatalf("unexpected '%[1]T' error getting credentials provider from environment: %[1]s", err)
	}
	if a, e := source, sharedConfigCredentialsSource(fileEnvName); a != e {
		t.Errorf("Expected initial source to be %q, %q given", e, a)
	}
	validateCredentialsProvider(ctx, credsEnv, "accesskey1", "secretkey1", "", sharedConfigCredentialsSource(fileEnvName), t)

	// Confirm CredsFilename overrides AWS_SHARED_CREDENTIALS_FILE
	credsParam, source, err := getCredentialsProvider(ctx, &Config{
		Profile:                "myprofile",
		SharedCredentialsFiles: []string{fileParamName},
	})
	if err != nil {
		t.Fatalf("unexpected '%[1]T' error getting credentials provider from configuration: %[1]s", err)
	}
	if a, e := source, sharedConfigCredentialsSource(fileParamName); a != e {
		t.Errorf("Expected initial source to be %q, %q given", e, a)
	}
	validateCredentialsProvider(ctx, credsParam, "accesskey2", "secretkey2", "", sharedConfigCredentialsSource(fileParamName), t)
}

func TestAWSGetCredentials_webIdentityToken(t *testing.T) {
	ctx := test.Context(t)

	cfg := Config{
		AssumeRoleWithWebIdentity: &AssumeRoleWithWebIdentity{
			RoleARN:          servicemocks.MockStsAssumeRoleWithWebIdentityArn,
			SessionName:      servicemocks.MockStsAssumeRoleWithWebIdentitySessionName,
			WebIdentityToken: servicemocks.MockWebIdentityToken,
		},
	}

	ts := servicemocks.MockAwsApiServer("STS", []*servicemocks.MockEndpoint{
		servicemocks.MockStsAssumeRoleWithWebIdentityValidEndpoint,
		servicemocks.MockStsGetCallerIdentityValidAssumedRoleEndpoint,
	})
	defer ts.Close()
	cfg.StsEndpoint = ts.URL

	creds, source, err := getCredentialsProvider(ctx, &cfg)
	if err != nil {
		t.Fatalf("unexpected '%[1]T' error getting credentials provider: %[1]s", err)
	}

	if a, e := source, stscreds.WebIdentityProviderName; a != e {
		t.Errorf("Expected initial source to be %q, %q given", e, a)
	}

	validateCredentialsProvider(ctx, creds,
		servicemocks.MockStsAssumeRoleWithWebIdentityAccessKey,
		servicemocks.MockStsAssumeRoleWithWebIdentitySecretKey,
		servicemocks.MockStsAssumeRoleWithWebIdentitySessionToken,
		stscreds.WebIdentityProviderName, t)
	testCredentialsProviderWrappedWithCache(creds, t)
}

var credentialsFileContentsEnv = `[myprofile]
aws_access_key_id = accesskey1
aws_secret_access_key = secretkey1
`

var credentialsFileContentsParam = `[myprofile]
aws_access_key_id = accesskey2
aws_secret_access_key = secretkey2
`

func writeCredentialsFile(credentialsFileContents string, t *testing.T) string {
	file, err := os.CreateTemp(os.TempDir(), "terraform_aws_cred")
	if err != nil {
		t.Fatalf("Error writing temporary credentials file: %s", err)
	}
	_, err = file.WriteString(credentialsFileContents)
	if err != nil {
		t.Fatalf("Error writing temporary credentials to file: %s", err)
	}
	err = file.Close()
	if err != nil {
		t.Fatalf("Error closing temporary credentials file: %s", err)
	}
	return file.Name()
}

func validateCredentialsProvider(ctx context.Context, creds aws.CredentialsProvider, accesskey, secretkey, token, source string, t *testing.T) {
	v, err := creds.Retrieve(ctx)
	if err != nil {
		t.Fatalf("Error retrieving credentials: %s", err)
	}

	if v.AccessKeyID != accesskey {
		t.Errorf("AccessKeyID mismatch, expected: %q, got %q", accesskey, v.AccessKeyID)
	}
	if v.SecretAccessKey != secretkey {
		t.Errorf("SecretAccessKey mismatch, expected: %q, got %q", secretkey, v.SecretAccessKey)
	}
	if v.SessionToken != token {
		t.Errorf("SessionToken mismatch, expected: %q, got %q", token, v.SessionToken)
	}
	if v.Source != source {
		t.Errorf("Expected provider name to be %q, %q given", source, v.Source)
	}
}

func testCredentialsProviderWrappedWithCache(creds aws.CredentialsProvider, t *testing.T) {
	switch creds.(type) {
	case *aws.CredentialsCache:
		break
	default:
		t.Error("expected credentials provider to be wrapped with aws.CredentialsCache")
	}
}

func sharedConfigCredentialsSource(filename string) string {
	return fmt.Sprintf(sharedConfigCredentialsProvider+": %s", filename)
}
