// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mockdata

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go-v2/credentials/endpointcreds"
	"github.com/aws/aws-sdk-go-v2/credentials/ssocreds"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/hashicorp/aws-sdk-go-base/v2/servicemocks"
)

// GetMockedAwsApiSession establishes an AWS session to a simulated AWS API server for a given service and route endpoints.
func GetMockedAwsApiSession(svcName string, endpoints []*servicemocks.MockEndpoint) (func(), aws.Config, string) {
	ts := servicemocks.MockAwsApiServer(svcName, endpoints)

	sc := credentials.NewStaticCredentialsProvider("accessKey", "secretKey", "")

	awsConfig := aws.Config{
		Credentials: sc,
		Region:      "us-east-1",
		EndpointResolverWithOptions: aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...any) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL:    ts.URL,
				Source: aws.EndpointSourceCustom,
			}, nil
		}),
	}

	return ts.Close, awsConfig, ts.URL
}

var (
	MockEc2MetadataCredentials = aws.Credentials{
		AccessKeyID:     servicemocks.MockEc2MetadataAccessKey,
		Source:          ec2rolecreds.ProviderName,
		SecretAccessKey: servicemocks.MockEc2MetadataSecretKey,
		SessionToken:    servicemocks.MockEc2MetadataSessionToken,
		CanExpire:       true,
	}

	MockEcsCredentialsCredentials = aws.Credentials{
		AccessKeyID:     servicemocks.MockEcsCredentialsAccessKey,
		SecretAccessKey: servicemocks.MockEcsCredentialsSecretKey,
		SessionToken:    servicemocks.MockEcsCredentialsSessionToken,
		CanExpire:       true,
		Source:          endpointcreds.ProviderName,
	}

	MockEnvCredentials = aws.Credentials{
		AccessKeyID:     servicemocks.MockEnvAccessKey,
		SecretAccessKey: servicemocks.MockEnvSecretKey,
		Source:          config.CredentialsSourceName,
	}

	MockEnvCredentialsWithSessionToken = aws.Credentials{
		AccessKeyID:     servicemocks.MockEnvAccessKey,
		SecretAccessKey: servicemocks.MockEnvSecretKey,
		SessionToken:    servicemocks.MockEnvSessionToken,
		Source:          config.CredentialsSourceName,
	}

	MockStaticCredentials = aws.Credentials{
		AccessKeyID:     servicemocks.MockStaticAccessKey,
		SecretAccessKey: servicemocks.MockStaticSecretKey,
		Source:          credentials.StaticCredentialsName,
	}

	MockStsAssumeRoleCredentials = aws.Credentials{
		AccessKeyID:     servicemocks.MockStsAssumeRoleAccessKey,
		AccountID:       "555555555555",
		SecretAccessKey: servicemocks.MockStsAssumeRoleSecretKey,
		SessionToken:    servicemocks.MockStsAssumeRoleSessionToken,
		Source:          stscreds.ProviderName,
		CanExpire:       true,
	}

	MockStsAssumeRoleWithWebIdentityCredentials = aws.Credentials{
		AccessKeyID:     servicemocks.MockStsAssumeRoleWithWebIdentityAccessKey,
		AccountID:       "666666666666",
		SecretAccessKey: servicemocks.MockStsAssumeRoleWithWebIdentitySecretKey,
		SessionToken:    servicemocks.MockStsAssumeRoleWithWebIdentitySessionToken,
		Source:          stscreds.WebIdentityProviderName,
		CanExpire:       true,
	}

	MockSsoCredentials = aws.Credentials{
		AccessKeyID:     servicemocks.MockSsoAccessKeyID,
		AccountID:       "123456789012",
		SecretAccessKey: servicemocks.MockSsoSecretAccessKey,
		SessionToken:    servicemocks.MockSsoSessionToken,
		Source:          ssocreds.ProviderName,
		CanExpire:       true,
	}
)
