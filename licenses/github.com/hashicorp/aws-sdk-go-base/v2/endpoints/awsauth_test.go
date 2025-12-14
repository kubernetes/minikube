// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package awsbase

import (
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go-v2/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/hashicorp/aws-sdk-go-base/v2/internal/test"
	"github.com/hashicorp/aws-sdk-go-base/v2/mockdata"
	"github.com/hashicorp/aws-sdk-go-base/v2/servicemocks"
)

func TestGetAccountIDAndPartition(t *testing.T) {
	var testCases = []struct {
		Description          string
		AuthProviderName     string
		EC2MetadataEndpoints []*servicemocks.MetadataResponse
		IAMEndpoints         []*servicemocks.MockEndpoint
		STSEndpoints         []*servicemocks.MockEndpoint
		ErrCount             int
		ExpectedAccountID    string
		ExpectedPartition    string
	}{
		{
			Description:          "EC2 Metadata over iam:GetUser when using EC2 Instance Profile",
			AuthProviderName:     ec2rolecreds.ProviderName,
			EC2MetadataEndpoints: append(servicemocks.Ec2metadata_securityCredentialsEndpoints, servicemocks.Ec2metadata_instanceIdEndpoint, servicemocks.Ec2metadata_iamInfoEndpoint),

			IAMEndpoints: []*servicemocks.MockEndpoint{
				{
					Request:  &servicemocks.MockRequest{Method: "POST", Uri: "/", Body: "Action=GetUser&Version=2010-05-08"},
					Response: &servicemocks.MockResponse{StatusCode: http.StatusOK, Body: servicemocks.IamResponse_GetUser_valid, ContentType: "text/xml"},
				},
			},
			ExpectedAccountID: servicemocks.Ec2metadata_iamInfoEndpoint_expectedAccountID,
			ExpectedPartition: servicemocks.Ec2metadata_iamInfoEndpoint_expectedPartition,
		},
		{
			Description:          "Mimic the metadata service mocked by Hologram (https://github.com/AdRoll/hologram)",
			AuthProviderName:     ec2rolecreds.ProviderName,
			EC2MetadataEndpoints: servicemocks.Ec2metadata_securityCredentialsEndpoints,
			IAMEndpoints: []*servicemocks.MockEndpoint{
				{
					Request:  &servicemocks.MockRequest{Method: "POST", Uri: "/", Body: "Action=GetUser&Version=2010-05-08"},
					Response: &servicemocks.MockResponse{StatusCode: http.StatusForbidden, Body: servicemocks.IamResponse_GetUser_unauthorized, ContentType: "text/xml"},
				},
			},
			STSEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsGetCallerIdentityValidEndpoint,
			},
			ExpectedAccountID: servicemocks.MockStsGetCallerIdentityAccountID,
			ExpectedPartition: servicemocks.MockStsGetCallerIdentityPartition,
		},
		{
			Description: "iam:ListRoles if iam:GetUser AccessDenied and sts:GetCallerIdentity fails",
			IAMEndpoints: []*servicemocks.MockEndpoint{
				{
					Request:  &servicemocks.MockRequest{Method: "POST", Uri: "/", Body: "Action=GetUser&Version=2010-05-08"},
					Response: &servicemocks.MockResponse{StatusCode: http.StatusForbidden, Body: servicemocks.IamResponse_GetUser_unauthorized, ContentType: "text/xml"},
				},
				{
					Request:  &servicemocks.MockRequest{Method: "POST", Uri: "/", Body: "Action=ListRoles&MaxItems=1&Version=2010-05-08"},
					Response: &servicemocks.MockResponse{StatusCode: http.StatusOK, Body: servicemocks.IamResponse_ListRoles_valid, ContentType: "text/xml"},
				},
			},
			STSEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsGetCallerIdentityInvalidEndpointAccessDenied,
			},
			ExpectedAccountID: servicemocks.IamResponse_ListRoles_valid_expectedAccountID,
			ExpectedPartition: servicemocks.IamResponse_ListRoles_valid_expectedPartition,
		},
		{
			Description: "iam:ListRoles if iam:GetUser ValidationError and sts:GetCallerIdentity fails",
			IAMEndpoints: []*servicemocks.MockEndpoint{
				{
					Request:  &servicemocks.MockRequest{Method: "POST", Uri: "/", Body: "Action=GetUser&Version=2010-05-08"},
					Response: &servicemocks.MockResponse{StatusCode: http.StatusBadRequest, Body: servicemocks.IamResponse_GetUser_federatedFailure, ContentType: "text/xml"},
				},
				{
					Request:  &servicemocks.MockRequest{Method: "POST", Uri: "/", Body: "Action=ListRoles&MaxItems=1&Version=2010-05-08"},
					Response: &servicemocks.MockResponse{StatusCode: http.StatusOK, Body: servicemocks.IamResponse_ListRoles_valid, ContentType: "text/xml"},
				},
			},
			STSEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsGetCallerIdentityInvalidEndpointAccessDenied,
			},
			ExpectedAccountID: servicemocks.IamResponse_ListRoles_valid_expectedAccountID,
			ExpectedPartition: servicemocks.IamResponse_ListRoles_valid_expectedPartition,
		},
		{
			Description: "Error when all endpoints fail",
			IAMEndpoints: []*servicemocks.MockEndpoint{
				{
					Request:  &servicemocks.MockRequest{Method: "POST", Uri: "/", Body: "Action=GetUser&Version=2010-05-08"},
					Response: &servicemocks.MockResponse{StatusCode: http.StatusBadRequest, Body: servicemocks.IamResponse_GetUser_federatedFailure, ContentType: "text/xml"},
				},
				{
					Request:  &servicemocks.MockRequest{Method: "POST", Uri: "/", Body: "Action=ListRoles&MaxItems=1&Version=2010-05-08"},
					Response: &servicemocks.MockResponse{StatusCode: http.StatusForbidden, Body: servicemocks.IamResponse_ListRoles_unauthorized, ContentType: "text/xml"},
				},
			},
			STSEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsGetCallerIdentityInvalidEndpointAccessDenied,
			},
			ErrCount: 1,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Description, func(t *testing.T) {
			ctx := test.Context(t)

			resetEnv := servicemocks.UnsetEnv(t)
			defer resetEnv()
			// capture the test server's close method, to call after the test returns
			awsTs := servicemocks.AwsMetadataApiMock(testCase.EC2MetadataEndpoints)
			defer awsTs()

			closeIam, iamConfig, _ := mockdata.GetMockedAwsApiSession("IAM", testCase.IAMEndpoints)
			defer closeIam()

			closeSts, stsConfig, _ := mockdata.GetMockedAwsApiSession("STS", testCase.STSEndpoints)
			defer closeSts()

			iamConn := iam.NewFromConfig(iamConfig)
			stsConn := sts.NewFromConfig(stsConfig)

			accountID, partition, err := getAccountIDAndPartition(ctx, iamConn, stsConn, testCase.AuthProviderName)
			if err != nil && testCase.ErrCount == 0 {
				t.Fatalf("Expected no error, received error: %s", err)
			}
			if err == nil && testCase.ErrCount > 0 {
				t.Fatalf("Expected %d error(s), received none", testCase.ErrCount)
			}
			if accountID != testCase.ExpectedAccountID {
				t.Fatalf("Parsed account ID doesn't match with expected (%q != %q)", accountID, testCase.ExpectedAccountID)
			}
			if partition != testCase.ExpectedPartition {
				t.Fatalf("Parsed partition doesn't match with expected (%q != %q)", partition, testCase.ExpectedPartition)
			}
		})
	}
}

func TestGetAccountIDAndPartitionFromEC2Metadata(t *testing.T) {
	t.Run("EC2 metadata success", func(t *testing.T) {
		ctx := test.Context(t)

		resetEnv := servicemocks.UnsetEnv(t)
		defer resetEnv()

		awsTs := servicemocks.AwsMetadataApiMock(append(
			servicemocks.Ec2metadata_securityCredentialsEndpoints,
			servicemocks.Ec2metadata_instanceIdEndpoint,
			servicemocks.Ec2metadata_iamInfoEndpoint,
		))
		defer awsTs()

		id, partition, err := getAccountIDAndPartitionFromEC2Metadata(ctx)
		if err != nil {
			t.Fatalf("Getting account ID from EC2 metadata API failed: %s", err)
		}

		if id != servicemocks.Ec2metadata_iamInfoEndpoint_expectedAccountID {
			t.Fatalf("Expected account ID: %s, given: %s", servicemocks.Ec2metadata_iamInfoEndpoint_expectedAccountID, id)
		}
		if partition != servicemocks.Ec2metadata_iamInfoEndpoint_expectedPartition {
			t.Fatalf("Expected partition: %s, given: %s", servicemocks.Ec2metadata_iamInfoEndpoint_expectedPartition, partition)
		}
	})
}

func TestGetAccountIDAndPartitionFromIAMGetUser(t *testing.T) {
	var testCases = []struct {
		Description       string
		MockEndpoints     []*servicemocks.MockEndpoint
		ErrCount          int
		ExpectedAccountID string
		ExpectedPartition string
	}{
		{
			Description: "Ignore iam:GetUser failure with federated user",
			MockEndpoints: []*servicemocks.MockEndpoint{
				{
					Request:  &servicemocks.MockRequest{Method: "POST", Uri: "/", Body: "Action=GetUser&Version=2010-05-08"},
					Response: &servicemocks.MockResponse{StatusCode: http.StatusBadRequest, Body: servicemocks.IamResponse_GetUser_federatedFailure, ContentType: "text/xml"},
				},
			},
			ErrCount: 0,
		},
		{
			Description: "Ignore iam:GetUser failure with unauthorized user",
			MockEndpoints: []*servicemocks.MockEndpoint{
				{
					Request:  &servicemocks.MockRequest{Method: "POST", Uri: "/", Body: "Action=GetUser&Version=2010-05-08"},
					Response: &servicemocks.MockResponse{StatusCode: http.StatusForbidden, Body: servicemocks.IamResponse_GetUser_unauthorized, ContentType: "text/xml"},
				},
			},
			ErrCount: 0,
		},
		{
			Description: "iam:GetUser success",
			MockEndpoints: []*servicemocks.MockEndpoint{
				{
					Request:  &servicemocks.MockRequest{Method: "POST", Uri: "/", Body: "Action=GetUser&Version=2010-05-08"},
					Response: &servicemocks.MockResponse{StatusCode: http.StatusOK, Body: servicemocks.IamResponse_GetUser_valid, ContentType: "text/xml"},
				},
			},
			ExpectedAccountID: servicemocks.IamResponse_GetUser_valid_expectedAccountID,
			ExpectedPartition: servicemocks.IamResponse_GetUser_valid_expectedPartition,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Description, func(t *testing.T) {
			ctx := test.Context(t)

			closeIam, config, _ := mockdata.GetMockedAwsApiSession("IAM", testCase.MockEndpoints)
			defer closeIam()

			iamClient := iam.NewFromConfig(config)

			accountID, partition, err := getAccountIDAndPartitionFromIAMGetUser(ctx, iamClient)
			if err != nil && testCase.ErrCount == 0 {
				t.Fatalf("Expected no error, received error: %s", err)
			}
			if err == nil && testCase.ErrCount > 0 {
				t.Fatalf("Expected %d error(s), received none", testCase.ErrCount)
			}
			if accountID != testCase.ExpectedAccountID {
				t.Fatalf("Parsed account ID doesn't match with expected (%q != %q)", accountID, testCase.ExpectedAccountID)
			}
			if partition != testCase.ExpectedPartition {
				t.Fatalf("Parsed partition doesn't match with expected (%q != %q)", partition, testCase.ExpectedPartition)
			}
		})
	}
}

func TestGetAccountIDAndPartitionFromIAMListRoles(t *testing.T) {
	var testCases = []struct {
		Description       string
		MockEndpoints     []*servicemocks.MockEndpoint
		ErrCount          int
		ExpectedAccountID string
		ExpectedPartition string
	}{
		{
			Description: "iam:ListRoles unauthorized",
			MockEndpoints: []*servicemocks.MockEndpoint{
				{
					Request:  &servicemocks.MockRequest{Method: "POST", Uri: "/", Body: "Action=ListRoles&MaxItems=1&Version=2010-05-08"},
					Response: &servicemocks.MockResponse{StatusCode: http.StatusForbidden, Body: servicemocks.IamResponse_ListRoles_unauthorized, ContentType: "text/xml"},
				},
			},
			ErrCount: 1,
		},
		{
			Description: "iam:ListRoles success",
			MockEndpoints: []*servicemocks.MockEndpoint{
				{
					Request:  &servicemocks.MockRequest{Method: "POST", Uri: "/", Body: "Action=ListRoles&MaxItems=1&Version=2010-05-08"},
					Response: &servicemocks.MockResponse{StatusCode: http.StatusOK, Body: servicemocks.IamResponse_ListRoles_valid, ContentType: "text/xml"},
				},
			},
			ExpectedAccountID: servicemocks.IamResponse_ListRoles_valid_expectedAccountID,
			ExpectedPartition: servicemocks.IamResponse_ListRoles_valid_expectedPartition,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Description, func(t *testing.T) {
			ctx := test.Context(t)

			closeIam, config, _ := mockdata.GetMockedAwsApiSession("IAM", testCase.MockEndpoints)
			defer closeIam()

			iamClient := iam.NewFromConfig(config)

			accountID, partition, err := getAccountIDAndPartitionFromIAMListRoles(ctx, iamClient)
			if err != nil && testCase.ErrCount == 0 {
				t.Fatalf("Expected no error, received error: %s", err)
			}
			if err == nil && testCase.ErrCount > 0 {
				t.Fatalf("Expected %d error(s), received none", testCase.ErrCount)
			}
			if accountID != testCase.ExpectedAccountID {
				t.Fatalf("Parsed account ID doesn't match with expected (%q != %q)", accountID, testCase.ExpectedAccountID)
			}
			if partition != testCase.ExpectedPartition {
				t.Fatalf("Parsed partition doesn't match with expected (%q != %q)", partition, testCase.ExpectedPartition)
			}
		})
	}
}

func TestGetAccountIDAndPartitionFromSTSGetCallerIdentity(t *testing.T) {
	var testCases = []struct {
		Description       string
		MockEndpoints     []*servicemocks.MockEndpoint
		ErrCount          int
		ExpectedAccountID string
		ExpectedPartition string
	}{
		{
			Description: "sts:GetCallerIdentity unauthorized",
			MockEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsGetCallerIdentityInvalidEndpointAccessDenied,
			},
			ErrCount: 1,
		},
		{
			Description: "sts:GetCallerIdentity ExpiredToken with invalid JSON response",
			MockEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsGetCallerIdentityInvalidBodyExpiredToken,
			},
			ErrCount: 1,
		},
		{
			Description: "sts:GetCallerIdentity ExpiredToken with valid JSON response",
			MockEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsGetCallerIdentityValidBodyExpiredToken,
			},
			ErrCount: 1,
		},
		{
			Description: "sts:GetCallerIdentity ExpiredTokenException with invalid JSON response",
			MockEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsGetCallerIdentityInvalidBodyExpiredTokenException,
			},
			ErrCount: 1,
		},
		{
			Description: "sts:GetCallerIdentity ExpiredTokenException with valid JSON response",
			MockEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsGetCallerIdentityValidBodyExpiredTokenException,
			},
			ErrCount: 1,
		},
		{
			Description: "sts:GetCallerIdentity RequestExpired with invalid JSON response",
			MockEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsGetCallerIdentityInvalidBodyRequestExpired,
			},
			ErrCount: 1,
		},
		{
			Description: "sts:GetCallerIdentity RequestExpired with valid JSON response",
			MockEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsGetCallerIdentityValidBodyRequestExpired,
			},
			ErrCount: 1,
		},
		{
			Description: "sts:GetCallerIdentity success",
			MockEndpoints: []*servicemocks.MockEndpoint{
				servicemocks.MockStsGetCallerIdentityValidEndpoint,
			},
			ExpectedAccountID: servicemocks.MockStsGetCallerIdentityAccountID,
			ExpectedPartition: servicemocks.MockStsGetCallerIdentityPartition,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Description, func(t *testing.T) {
			ctx := test.Context(t)

			closeSts, config, _ := mockdata.GetMockedAwsApiSession("STS", testCase.MockEndpoints)
			defer closeSts()

			stsClient := sts.NewFromConfig(config)

			accountID, partition, err := getAccountIDAndPartitionFromSTSGetCallerIdentity(ctx, stsClient)
			if err != nil && testCase.ErrCount == 0 {
				t.Fatalf("Expected no error, received error: %s", err)
			}
			if err == nil && testCase.ErrCount > 0 {
				t.Fatalf("Expected %d error(s), received none", testCase.ErrCount)
			}
			if accountID != testCase.ExpectedAccountID {
				t.Fatalf("Parsed account ID doesn't match with expected (%q != %q)", accountID, testCase.ExpectedAccountID)
			}
			if partition != testCase.ExpectedPartition {
				t.Fatalf("Parsed partition doesn't match with expected (%q != %q)", partition, testCase.ExpectedPartition)
			}
		})
	}
}

func TestAWSParseAccountIDAndPartitionFromARN(t *testing.T) {
	var testCases = []struct {
		InputARN          string
		ErrCount          int
		ExpectedAccountID string
		ExpectedPartition string
	}{
		{
			InputARN: "invalid-arn",
			ErrCount: 1,
		},
		{
			InputARN:          "arn:aws:iam::123456789012:instance-profile/name",
			ExpectedAccountID: "123456789012",
			ExpectedPartition: "aws",
		},
		{
			InputARN:          "arn:aws:iam::123456789012:user/name",
			ExpectedAccountID: "123456789012",
			ExpectedPartition: "aws",
		},
		{
			InputARN:          "arn:aws:sts::123456789012:assumed-role/name",
			ExpectedAccountID: "123456789012",
			ExpectedPartition: "aws",
		},
		{
			InputARN:          "arn:aws-us-gov:sts::123456789012:assumed-role/name",
			ExpectedAccountID: "123456789012",
			ExpectedPartition: "aws-us-gov",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.InputARN, func(t *testing.T) {
			accountID, partition, err := parseAccountIDAndPartitionFromARN(testCase.InputARN)
			if err != nil && testCase.ErrCount == 0 {
				t.Fatalf("Expected no error when parsing ARN, received error: %s", err)
			}
			if err == nil && testCase.ErrCount > 0 {
				t.Fatalf("Expected %d error(s) when parsing ARN, received none", testCase.ErrCount)
			}
			if accountID != testCase.ExpectedAccountID {
				t.Fatalf("Parsed account ID doesn't match with expected (%q != %q)", accountID, testCase.ExpectedAccountID)
			}
			if partition != testCase.ExpectedPartition {
				t.Fatalf("Parsed partition doesn't match with expected (%q != %q)", partition, testCase.ExpectedPartition)
			}
		})
	}
}
