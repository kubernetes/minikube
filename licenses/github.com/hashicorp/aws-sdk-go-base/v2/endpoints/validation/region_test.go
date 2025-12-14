// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validation

import (
	"testing"
)

func TestSupportedRegion(t *testing.T) {
	var testCases = []struct {
		Region      string
		ExpectError bool
	}{
		{
			Region:      "us-east-1",
			ExpectError: false,
		},
		{
			Region:      "ap-northeast-3",
			ExpectError: false,
		},
		{
			Region:      "us-gov-west-1",
			ExpectError: false,
		},
		{
			Region:      "cn-northwest-1",
			ExpectError: false,
		},
		{
			Region:      "invalid",
			ExpectError: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Region, func(t *testing.T) {
			err := SupportedRegion(testCase.Region)
			if err != nil && !testCase.ExpectError {
				t.Fatalf("Expected no error, received error: %s", err)
			}
			if err == nil && testCase.ExpectError {
				t.Fatal("Expected error, received none")
			}
		})
	}
}
