// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package awsbase

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/diag"
	"github.com/hashicorp/aws-sdk-go-base/v2/logging"
)

const (
	configSourceProviderConfig      = "provider"
	configSourceEnvironmentVariable = "envvar"
)

func getCredentialsProvider(ctx context.Context, c *Config) (aws.CredentialsProvider, string, diag.Diagnostics) {
	// This function will need to exist for any authentication methods that call STS until the providers use a reasonable default `MaxRetries`.
	// Otherwise, retryable errors will cause the provider to appear to have frozen.
	var diags diag.Diagnostics

	logger := logging.RetrieveLogger(ctx)

	loadOptions, err := commonLoadOptions(ctx, c)
	if err != nil {
		return nil, "", diags.AddSimpleError(err)
	}
	loadOptions = append(
		loadOptions,
		// The endpoint resolver is added here instead of in commonLoadOptions() so that it
		// is not included in the aws.Config returned to the caller
		config.WithEndpointResolverWithOptions(credentialsEndpointResolver(ctx, c)),
	)

	envConfig, err := config.NewEnvConfig()
	if err != nil {
		return nil, "", diags.AddSimpleError(err)
	}

	if c.Profile != "" && os.Getenv("AWS_ACCESS_KEY_ID") != "" && os.Getenv("AWS_SECRET_ACCESS_KEY") != "" {
		diags.AddWarning("Configuration conflict detected",
			`A Profile was specified along with the environment variables "AWS_ACCESS_KEY_ID" and "AWS_SECRET_ACCESS_KEY". `+
				`The Profile is now used instead of the environment variable credentials. This may lead to unexpected behavior.`)
	}

	if profile := c.Profile; profile != "" {
		logger.Debug(ctx, "Setting profile", map[string]any{
			"tf_aws.profile":        profile,
			"tf_aws.profile.source": configSourceProviderConfig,
		})
		loadOptions = append(
			loadOptions,
			config.WithSharedConfigProfile(c.Profile),
		)
	} else if profile := envConfig.SharedConfigProfile; profile != "" {
		logger.Debug(ctx, "Using profile", map[string]any{
			"tf_aws.profile":        profile,
			"tf_aws.profile.source": configSourceEnvironmentVariable,
		})
	}

	logger.Debug(ctx, "Loading configuration")
	cfg, err := config.LoadDefaultConfig(ctx, loadOptions...)
	if err != nil {
		return nil, "", diags.AddSimpleError(err)
	}

	// This can probably be configured directly in commonLoadOptions() once
	// https://github.com/aws/aws-sdk-go-v2/pull/1682 is merged
	if c.AssumeRoleWithWebIdentity != nil {
		if c.AssumeRoleWithWebIdentity.RoleARN == "" {
			return nil, "", diags.AddError("Assume Role With Web Identity", "Role ARN was not set")
		}
		if c.AssumeRoleWithWebIdentity.WebIdentityToken == "" && c.AssumeRoleWithWebIdentity.WebIdentityTokenFile == "" {
			return nil, "", diags.AddError("Assume Role With Web Identity", "One of WebIdentityToken, WebIdentityTokenFile must be set")
		}
		provider, d := webIdentityCredentialsProvider(ctx, cfg, c)
		diags = diags.Append(d...)
		if diags.HasError() {
			return nil, "", diags
		}
		cfg.Credentials = provider
	}

	logger.Debug(ctx, "Retrieving credentials")
	creds, err := cfg.Credentials.Retrieve(ctx)
	if err != nil {
		if c.Profile != "" && os.Getenv("AWS_ACCESS_KEY_ID") != "" && os.Getenv("AWS_SECRET_ACCESS_KEY") != "" {
			err = fmt.Errorf(`A Profile was specified along with the environment variables "AWS_ACCESS_KEY_ID" and "AWS_SECRET_ACCESS_KEY". The Profile is now used instead of the environment variable credentials.

AWS Error: %w`, err)
		}
		return nil, "", diags.Append(c.NewNoValidCredentialSourcesError(err))
	}

	if len(c.AssumeRole) == 0 {
		return cfg.Credentials, creds.Source, diags
	}

	logger.Info(ctx, "Retrieved initial credentials", map[string]any{
		"tf_aws.credentials_source": creds.Source,
	})
	provider, d := assumeRoleCredentialsProvider(ctx, cfg, c)
	diags = diags.Append(d...)
	if diags.HasError() {
		return nil, "", diags
	}

	return provider, creds.Source, diags
}

func webIdentityCredentialsProvider(ctx context.Context, awsConfig aws.Config, c *Config) (aws.CredentialsProvider, diag.Diagnostics) {
	var diags diag.Diagnostics

	logger := logging.RetrieveLogger(ctx)

	ar := c.AssumeRoleWithWebIdentity

	logger.Info(ctx, "Assuming IAM Role With Web Identity", map[string]any{
		"tf_aws.assume_role_with_web_identity.role_arn":     ar.RoleARN,
		"tf_aws.assume_role_with_web_identity.session_name": ar.SessionName,
	})

	// awsConfig now has IMDS creds, remove them before initializing
	// We probably shouldn't be resolving config before setting this up
	awsConfig.Credentials = nil
	client := stsClient(ctx, awsConfig, c)

	appCreds := stscreds.NewWebIdentityRoleProvider(client, ar.RoleARN, ar, func(opts *stscreds.WebIdentityRoleOptions) {
		opts.RoleSessionName = ar.SessionName
		opts.Duration = ar.Duration

		if ar.Policy != "" {
			opts.Policy = aws.String(ar.Policy)
		}

		if len(ar.PolicyARNs) > 0 {
			opts.PolicyARNs = getPolicyDescriptorTypes(ar.PolicyARNs)
		}
	})

	if _, err := appCreds.Retrieve(ctx); err != nil {
		return nil, diags.Append(c.NewCannotAssumeRoleWithWebIdentityError(err))
	}
	return aws.NewCredentialsCache(appCreds), diags
}

func assumeRoleCredentialsProvider(ctx context.Context, awsConfig aws.Config, c *Config) (aws.CredentialsProvider, diag.Diagnostics) {
	var diags diag.Diagnostics

	logger := logging.RetrieveLogger(ctx)

	var creds aws.CredentialsProvider

	total := len(c.AssumeRole)
	for i, ar := range c.AssumeRole {
		if ar.RoleARN == "" {
			return nil, diags.AddError(
				"Cannot assume IAM Role",
				fmt.Sprintf("IAM Role ARN not set in assume role %d of %d", i+1, total),
			)
		}

		logger.Info(ctx, "Assuming IAM Role", map[string]any{
			"tf_aws.assume_role.index":           i,
			"tf_aws.assume_role.role_arn":        ar.RoleARN,
			"tf_aws.assume_role.session_name":    ar.SessionName,
			"tf_aws.assume_role.external_id":     ar.ExternalID,
			"tf_aws.assume_role.source_identity": ar.SourceIdentity,
		})

		// When assuming a role, we need to first authenticate the base credentials above, then assume the desired role
		client := stsClient(ctx, awsConfig, c)

		appCreds := stscreds.NewAssumeRoleProvider(client, ar.RoleARN, func(opts *stscreds.AssumeRoleOptions) {
			opts.RoleSessionName = ar.SessionName
			opts.Duration = ar.Duration

			if ar.ExternalID != "" {
				opts.ExternalID = aws.String(ar.ExternalID)
			}

			if ar.Policy != "" {
				opts.Policy = aws.String(ar.Policy)
			}

			if len(ar.PolicyARNs) > 0 {
				opts.PolicyARNs = getPolicyDescriptorTypes(ar.PolicyARNs)
			}

			if len(ar.Tags) > 0 {
				var tags []types.Tag
				for k, v := range ar.Tags {
					tag := types.Tag{
						Key:   aws.String(k),
						Value: aws.String(v),
					}
					tags = append(tags, tag)
				}

				opts.Tags = tags
			}

			if len(ar.TransitiveTagKeys) > 0 {
				opts.TransitiveTagKeys = ar.TransitiveTagKeys
			}

			if ar.SourceIdentity != "" {
				opts.SourceIdentity = aws.String(ar.SourceIdentity)
			}
		})
		_, err := appCreds.Retrieve(ctx)
		if err != nil {
			return nil, diags.Append(newCannotAssumeRoleError(ar, err))
		}
		creds = aws.NewCredentialsCache(appCreds)
		awsConfig.Credentials = creds
	}
	return creds, nil
}

func getPolicyDescriptorTypes(policyARNs []string) []types.PolicyDescriptorType {
	var policyDescriptorTypes []types.PolicyDescriptorType

	for _, policyARN := range policyARNs {
		policyDescriptorType := types.PolicyDescriptorType{
			Arn: aws.String(policyARN),
		}
		policyDescriptorTypes = append(policyDescriptorTypes, policyDescriptorType)
	}
	return policyDescriptorTypes
}
