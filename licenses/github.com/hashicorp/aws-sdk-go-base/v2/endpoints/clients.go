// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package awsbase

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/hashicorp/aws-sdk-go-base/v2/logging"
)

func iamClient(ctx context.Context, awsConfig aws.Config, c *Config) *iam.Client {
	logger := logging.RetrieveLogger(ctx)

	return iam.NewFromConfig(awsConfig, func(opts *iam.Options) {
		if c.IamEndpoint != "" {
			logger.Info(ctx, "IAM client: setting custom endpoint", map[string]any{
				"tf_aws.iam_client.endpoint": c.IamEndpoint,
			})
			opts.EndpointResolver = iam.EndpointResolverFromURL(c.IamEndpoint) //nolint:staticcheck // The replacement is not documented yet (2023/07/31)
		}
	})
}

func stsClient(ctx context.Context, awsConfig aws.Config, c *Config) *sts.Client {
	logger := logging.RetrieveLogger(ctx)

	return sts.NewFromConfig(awsConfig, func(opts *sts.Options) {
		if c.StsRegion != "" {
			logger.Info(ctx, "STS client: setting region", map[string]any{
				"tf_aws.sts_client.region": c.StsRegion,
			})
			opts.Region = c.StsRegion
		}
		if c.StsEndpoint != "" {
			logger.Info(ctx, "STS client: setting custom endpoint", map[string]any{
				"tf_aws.sts_client.endpoint": c.StsEndpoint,
			})
			opts.EndpointResolver = sts.EndpointResolverFromURL(c.StsEndpoint) //nolint:staticcheck // The replacement is not documented yet (2023/07/31)
		}
	})
}
