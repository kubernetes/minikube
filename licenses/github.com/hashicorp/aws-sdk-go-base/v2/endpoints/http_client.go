// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package awsbase

import (
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/hashicorp/aws-sdk-go-base/v2/internal/config"
)

func defaultHttpClient(c *config.Config) (*awshttp.BuildableClient, error) {
	opts, err := c.HTTPTransportOptions()
	if err != nil {
		return nil, err
	}

	httpClient := awshttp.NewBuildableClient().WithTransportOptions(opts)

	return httpClient, err
}
