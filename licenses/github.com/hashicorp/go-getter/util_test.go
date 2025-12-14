// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package getter

import (
	"io"
	"os"
	"strings"
	"testing"
)

// tempEnv sets the env var temporarily and returns a function that should
// be deferred to clean it up.
func tempEnv(t *testing.T, k, v string) func() {
	old := os.Getenv(k)

	// Set env
	if err := os.Setenv(k, v); err != nil {
		t.Fatalf("err: %s", err)
	}

	// Easy cleanup
	return func() {
		if err := os.Setenv(k, old); err != nil {
			t.Fatalf("err: %s", err)
		}
	}
}

// tempFileContents writes a temporary file and returns the path and a function
// to clean it up.
func tempFileContents(t *testing.T, contents string) (string, func()) {
	tf, err := os.CreateTemp("", "getter")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if _, err := io.Copy(tf, strings.NewReader(contents)); err != nil {
		t.Fatalf("err: %s", err)
	}

	_ = tf.Close()

	path := tf.Name()
	return path, func() {
		if err := os.Remove(path); err != nil {
			t.Fatalf("err: %s", err)
		}
	}
}
