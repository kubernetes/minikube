// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package expand_test

import (
	"os"
	"testing"

	"github.com/hashicorp/aws-sdk-go-base/v2/internal/expand"
	"github.com/hashicorp/aws-sdk-go-base/v2/servicemocks"
)

func TestExpandFilePath(t *testing.T) {
	testcases := map[string]struct {
		path     string
		expected string
		envvars  map[string]string
	}{
		"filename": {
			path:     "file",
			expected: "file",
		},
		"file in current dir": {
			path:     "./file",
			expected: "./file",
		},
		"file with tilde": {
			path:     "~/file",
			expected: "/my/home/dir/file",
			envvars: map[string]string{
				"HOME": "/my/home/dir",
			},
		},
		"file with envvar": {
			path:     "$HOME/file",
			expected: "/home/dir/file",
			envvars: map[string]string{
				"HOME": "/home/dir",
			},
		},
		"full file in envvar": {
			path:     "$CONF_FILE",
			expected: "/path/to/conf/file",
			envvars: map[string]string{
				"CONF_FILE": "/path/to/conf/file",
			},
		},
	}

	for name, testcase := range testcases {
		t.Run(name, func(t *testing.T) {
			servicemocks.StashEnv(t)

			for k, v := range testcase.envvars {
				os.Setenv(k, v)
			}

			a, err := expand.FilePath(testcase.path)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			if a != testcase.expected {
				t.Errorf("expected expansion to %q, got %q", testcase.expected, a)
			}
		})
	}
}
