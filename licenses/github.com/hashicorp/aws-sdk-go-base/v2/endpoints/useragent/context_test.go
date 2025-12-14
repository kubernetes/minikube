// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package useragent

import (
	"context"
	"testing"

	"github.com/hashicorp/aws-sdk-go-base/v2/internal/config"
)

func TestFromContext(t *testing.T) {
	testcases := map[string]struct {
		setup    func() context.Context
		expected string
	}{
		"empty": {
			setup: func() context.Context {
				return context.Background()
			},
			expected: "",
		},
		"UserAgentProducts": {
			setup: func() context.Context {
				return Context(context.Background(), config.UserAgentProducts{
					{
						Name:    "first",
						Version: "1.2.3",
					},
					{
						Name:    "second",
						Version: "1.0.2",
						Comment: "a comment",
					},
				})
			},
			expected: "first/1.2.3 second/1.0.2 (a comment)",
		},
		"[]UserAgentProduct": {
			setup: func() context.Context {
				return Context(context.Background(), []config.UserAgentProduct{
					{
						Name:    "first",
						Version: "1.2.3",
					},
					{
						Name:    "second",
						Version: "1.0.2",
						Comment: "a comment",
					},
				})
			},
			expected: "first/1.2.3 second/1.0.2 (a comment)",
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			ctx := testcase.setup()

			v := BuildFromContext(ctx)

			if v != testcase.expected {
				t.Errorf("expected %q, got %q", testcase.expected, v)
			}
		})
	}
}
