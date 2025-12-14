// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validation

import (
	"fmt"
	"slices"

	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
)

type InvalidRegionError struct {
	region string
}

func (e *InvalidRegionError) Error() string {
	return fmt.Sprintf("invalid AWS Region: %s", e.region)
}

// SupportedRegion checks if the given region is a valid AWS region.
func SupportedRegion(region string) error {
	if slices.ContainsFunc(endpoints.DefaultPartitions(), func(p endpoints.Partition) bool {
		_, ok := p.Regions()[region]
		return ok
	}) {
		return nil
	}

	return &InvalidRegionError{
		region: region,
	}
}
