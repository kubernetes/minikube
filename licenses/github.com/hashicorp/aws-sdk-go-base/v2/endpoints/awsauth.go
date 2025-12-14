// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package awsbase

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/smithy-go"
	"github.com/hashicorp/aws-sdk-go-base/v2/logging"
	multierror "github.com/hashicorp/go-multierror"
)

// getAccountIDAndPartition gets the account ID and associated partition.
func getAccountIDAndPartition(ctx context.Context, iamClient *iam.Client, stsClient *sts.Client, authProviderName string) (string, string, error) {
	var accountID, partition string
	var err, errors error

	if authProviderName == ec2rolecreds.ProviderName {
		accountID, partition, err = getAccountIDAndPartitionFromEC2Metadata(ctx)
	} else {
		accountID, partition, err = getAccountIDAndPartitionFromIAMGetUser(ctx, iamClient)
	}
	if accountID != "" {
		return accountID, partition, nil
	}
	errors = multierror.Append(errors, err)

	accountID, partition, err = getAccountIDAndPartitionFromSTSGetCallerIdentity(ctx, stsClient)
	if accountID != "" {
		return accountID, partition, nil
	}
	errors = multierror.Append(errors, err)

	accountID, partition, err = getAccountIDAndPartitionFromIAMListRoles(ctx, iamClient)
	if accountID != "" {
		return accountID, partition, nil
	}
	errors = multierror.Append(errors, err)

	return accountID, partition, errors
}

// getAccountIDAndPartitionFromEC2Metadata gets the account ID and associated
// partition from EC2 metadata.
func getAccountIDAndPartitionFromEC2Metadata(ctx context.Context) (accountID string, partition string, err error) {
	logger := logging.RetrieveLogger(ctx)

	logger.Debug(ctx, "Retrieving account information from EC2 Metadata")

	cfg := aws.Config{}

	metadataClient := imds.NewFromConfig(cfg)
	info, err := metadataClient.GetIAMInfo(ctx, &imds.GetIAMInfoInput{})
	if err != nil {
		// We can end up here if there's an issue with the instance metadata service
		// or if we're getting credentials from AdRoll's Hologram (in which case IAMInfo will
		// error out).
		logger.Debug(ctx, "Unable to retrieve account information from EC2 Metadata", map[string]any{
			"error": err,
		})
		return "", "", fmt.Errorf("retrieving account information via EC2 Metadata IAM information: %w", err)
	}

	accountID, partition, err = parseAccountIDAndPartitionFromARN(info.InstanceProfileArn)
	if err != nil {
		logger.Debug(ctx, "Unable to retrieve account information from EC2 Metadata", map[string]any{
			"error": err,
		})
		return "", "", fmt.Errorf("retrieving account information from EC2 Metadata: %w", err)
	} else {
		logger.Info(ctx, "Retrieved account information from EC2 Metadata")
	}
	return
}

// getAccountIDAndPartitionFromIAMGetUser gets the account ID and associated
// partition from IAM.
func getAccountIDAndPartitionFromIAMGetUser(ctx context.Context, iamClient iam.GetUserAPIClient) (accountID string, partition string, err error) {
	logger := logging.RetrieveLogger(ctx)

	logger.Debug(ctx, "Retrieving account information via iam:GetUser")

	output, err := iamClient.GetUser(ctx, &iam.GetUserInput{})
	if err != nil {
		// AccessDenied and ValidationError can be raised
		// if credentials belong to federated profile, so we ignore these
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) {
			switch apiErr.ErrorCode() {
			case "AccessDenied", "InvalidClientTokenId", "ValidationError":
				logger.Debug(ctx, "Retrieving account information via iam:GetUser: ignoring error", map[string]any{
					"error": err,
				})
				return "", "", nil
			}
		}
		logger.Debug(ctx, "Unable to retrieve account information via iam:GetUser", map[string]any{
			"error": err,
		})
		return "", "", fmt.Errorf("retrieving account information via iam:GetUser: %w", err)
	}

	if output == nil || output.User == nil {
		logger.Debug(ctx, "Unable to retrieve account information via iam:GetUser", map[string]any{
			"error": "empty response",
		})
		return "", "", errors.New("retrieving account information via iam:GetUser: empty response")
	}

	accountID, partition, err = parseAccountIDAndPartitionFromARN(aws.ToString(output.User.Arn))
	if err != nil {
		logger.Debug(ctx, "Unable to retrieve account information via iam:GetUser", map[string]any{
			"error": err,
		})
		return "", "", fmt.Errorf("retrieving account information via iam:GetUser: %w", err)
	} else {
		logger.Info(ctx, "Retrieved account information via iam:GetUser")
	}
	return
}

// getAccountIDAndPartitionFromIAMListRoles gets the account ID and associated
// partition from listing IAM roles.
func getAccountIDAndPartitionFromIAMListRoles(ctx context.Context, iamClient iam.ListRolesAPIClient) (accountID string, partition string, err error) {
	logger := logging.RetrieveLogger(ctx)

	logger.Debug(ctx, "Retrieving account information via iam:ListRoles")

	output, err := iamClient.ListRoles(ctx, &iam.ListRolesInput{
		MaxItems: aws.Int32(1),
	})
	if err != nil {
		logger.Debug(ctx, "Unable to retrieve account information via iam:ListRoles", map[string]any{
			"error": err,
		})
		return "", "", fmt.Errorf("retrieving account information via iam:ListRoles: %w", err)
	}

	if output == nil || len(output.Roles) < 1 {
		logger.Debug(ctx, "Unable to retrieve account information via iam:ListRoles", map[string]any{
			"error": "empty response",
		})
		return "", "", errors.New("retrieving account information via iam:ListRoles: empty response")
	}

	accountID, partition, err = parseAccountIDAndPartitionFromARN(aws.ToString(output.Roles[0].Arn))
	if err != nil {
		logger.Debug(ctx, "Unable to retrieve account information via iam:ListRoles", map[string]any{
			"error": err,
		})
		return "", "", fmt.Errorf("retrieving account information via iam:ListRoles: %w", err)
	} else {
		logger.Info(ctx, "Retrieved account information via iam:ListRoles")
	}
	return
}

// getAccountIDAndPartitionFromSTSGetCallerIdentity gets the account ID and associated
// partition from STS caller identity.
func getAccountIDAndPartitionFromSTSGetCallerIdentity(ctx context.Context, stsClient *sts.Client) (accountID string, partition string, err error) {
	logger := logging.RetrieveLogger(ctx)

	logger.Debug(ctx, "Retrieving caller identity from STS")

	output, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		logger.Debug(ctx, "Unable to retrieve caller identity from STS", map[string]any{
			"error": err,
		})
		return "", "", fmt.Errorf("retrieving caller identity from STS: %w", err)
	}

	if output == nil || output.Arn == nil {
		logger.Debug(ctx, "Unable to retrieve caller identity from STS", map[string]any{
			"error": "empty response",
		})
		return "", "", errors.New("retrieving caller identity from STS: empty response")
	}

	accountID, partition, err = parseAccountIDAndPartitionFromARN(aws.ToString(output.Arn))
	if err != nil {
		logger.Debug(ctx, "Unable to retrieve caller identity from STS", map[string]any{
			"error": err,
		})
		return "", "", fmt.Errorf("retrieving caller identity from STS: %w", err)
	} else {
		logger.Info(ctx, "Retrieved caller identity from STS")
	}
	return
}

func parseAccountIDAndPartitionFromARN(inputARN string) (string, string, error) {
	arn, err := arn.Parse(inputARN)
	if err != nil {
		return "", "", fmt.Errorf("parsing ARN (%s): %s", inputARN, err)
	}
	return arn.AccountID, arn.Partition, nil
}
