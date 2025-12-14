// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package awsbase

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/sso"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/hashicorp/aws-sdk-go-base/v2/logging"
)

// This endpoint resolver is needed when authenticating because the AWS SDK makes internal
// calls to STS. The resolver should not be attached to the aws.Config returned to the
// client, since it should configure its own overrides
func credentialsEndpointResolver(ctx context.Context, c *Config) aws.EndpointResolverWithOptions {
	logger := logging.RetrieveLogger(ctx)

	resolver := func(service, region string, options ...any) (aws.Endpoint, error) {
		switch service {
		case iam.ServiceID:
			if endpoint := c.IamEndpoint; endpoint != "" {
				logger.Info(ctx, "Credentials resolution: setting custom IAM endpoint", map[string]any{
					"tf_aws.iam_client.endpoint": endpoint,
				})
				return aws.Endpoint{
					URL:           endpoint,
					Source:        aws.EndpointSourceCustom,
					SigningRegion: region,
				}, nil
			}
		case sso.ServiceID:
			if endpoint := c.SsoEndpoint; endpoint != "" {
				logger.Info(ctx, "Credentials resolution: setting custom SSO endpoint", map[string]any{
					"tf_aws.sso_client.endpoint": endpoint,
				})
				return aws.Endpoint{
					URL:           endpoint,
					Source:        aws.EndpointSourceCustom,
					SigningRegion: region,
				}, nil
			}
		case sts.ServiceID:
			if endpoint := c.StsEndpoint; endpoint != "" {
				fields := map[string]any{
					"tf_aws.sts_client.endpoint": endpoint,
				}
				if c.StsRegion != "" {
					fields["tf_aws.sts_client.signing_region"] = c.StsRegion
					region = c.StsRegion
				}
				logger.Info(ctx, "Credentials resolution: setting custom STS endpoint", fields)
				return aws.Endpoint{
					URL:           endpoint,
					Source:        aws.EndpointSourceCustom,
					SigningRegion: region,
				}, nil
			}
		}

		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	}

	return aws.EndpointResolverWithOptionsFunc(resolver)
}
