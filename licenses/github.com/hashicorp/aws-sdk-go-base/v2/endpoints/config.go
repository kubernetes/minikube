// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package awsbase

import (
	"github.com/hashicorp/aws-sdk-go-base/v2/internal/config"
)

// Config, APNInfo, APNProduct, and AssumeRole are aliased to an internal package to break a dependency cycle
// in internal/httpclient.

type Config = config.Config

type APNInfo = config.APNInfo

type AssumeRole = config.AssumeRole

type AssumeRoleWithWebIdentity = config.AssumeRoleWithWebIdentity

type UserAgentProducts = config.UserAgentProducts

type UserAgentProduct = config.UserAgentProduct

const (
	EC2MetadataEndpointModeIPv4 = "IPv4"
	EC2MetadataEndpointModeIPv6 = "IPv6"
)

func EC2MetadataEndpointMode_Values() []string {
	return []string{
		EC2MetadataEndpointModeIPv4,
		EC2MetadataEndpointModeIPv6,
	}
}

const (
	HTTPProxyModeLegacy   = config.HTTPProxyModeLegacy
	HTTPProxyModeSeparate = config.HTTPProxyModeSeparate
)
