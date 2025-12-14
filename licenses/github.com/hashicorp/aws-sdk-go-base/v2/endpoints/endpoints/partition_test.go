// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package endpoints_test

import (
	"testing"

	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
)

func TestDefaultPartitions(t *testing.T) {
	t.Parallel()

	got := endpoints.DefaultPartitions()
	if len(got) == 0 {
		t.Fatalf("expected partitions, got none")
	}
}

func TestPartitionForRegion(t *testing.T) {
	t.Parallel()

	testcases := map[string]struct {
		expectedFound bool
		expectedID    string
	}{
		"us-east-1": {
			expectedFound: true,
			expectedID:    "aws",
		},
		"us-gov-west-1": {
			expectedFound: true,
			expectedID:    "aws-us-gov",
		},
		"not-found": {
			expectedFound: false,
		},
		"us-east-17": {
			expectedFound: true,
			expectedID:    "aws",
		},
	}

	ps := endpoints.DefaultPartitions()
	for region, testcase := range testcases {
		gotID, gotFound := endpoints.PartitionForRegion(ps, region)

		if gotFound != testcase.expectedFound {
			t.Errorf("expected PartitionFound %t for Region %q, got %t", testcase.expectedFound, region, gotFound)
		}
		if gotID.ID() != testcase.expectedID {
			t.Errorf("expected PartitionID %q for Region %q, got %q", testcase.expectedID, region, gotID.ID())
		}
	}
}

func TestPartitionRegions(t *testing.T) {
	t.Parallel()

	testcases := map[string]struct {
		expectedRegions bool
	}{
		"us-east-1": {
			expectedRegions: true,
		},
		"us-gov-west-1": {
			expectedRegions: true,
		},
		"not-found": {
			expectedRegions: false,
		},
	}

	ps := endpoints.DefaultPartitions()
	for region, testcase := range testcases {
		gotID, _ := endpoints.PartitionForRegion(ps, region)

		if got, want := len(gotID.Regions()) > 0, testcase.expectedRegions; got != want {
			t.Errorf("expected Regions %t for Region %q, got %t", want, region, got)
		}
	}
}

func TestPartitionServices(t *testing.T) {
	t.Parallel()

	testcases := map[string]struct {
		expectedServices bool
	}{
		"us-east-1": {
			expectedServices: true,
		},
		"us-gov-west-1": {
			expectedServices: true,
		},
		"not-found": {
			expectedServices: false,
		},
	}

	ps := endpoints.DefaultPartitions()
	for region, testcase := range testcases {
		gotID, _ := endpoints.PartitionForRegion(ps, region)

		if got, want := len(gotID.Services()) > 0, testcase.expectedServices; got != want {
			t.Errorf("expected services %t for Region %q, got %t", want, region, got)
		}
	}
}
