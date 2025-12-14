// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logging

import "go.opentelemetry.io/otel/attribute"

const (
	AwsSdkKey         attribute.Key = "tf_aws.sdk"
	SigningRegionKey  attribute.Key = "tf_aws.signing_region"
	CustomEndpointKey attribute.Key = "tf_aws.custom_endpoint"
)

func SigningRegion(region string) attribute.KeyValue {
	return SigningRegionKey.String(region)
}

func CustomEndpoint(custom bool) attribute.KeyValue {
	return CustomEndpointKey.Bool(custom)
}
