// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package getter

import (
	"context"
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
)

// Note for external contributors: In order to run the S3 test suite, you will only be able to be run
// in GitHub Actions when you open a PR.

func TestS3Getter_impl(t *testing.T) {
	var _ Getter = new(S3Getter)
}

func TestS3Getter(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test that requires AWS credentials in short mode")
	}

	g := new(S3Getter)
	dst := t.TempDir()

	// With a dir that doesn't exist
	err := g.Get(
		dst, testURL("https://s3.amazonaws.com/hc-go-getter-test/go-getter/folder"))
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the main file exists
	mainPath := filepath.Join(dst, "main.tf")
	if _, err := os.Stat(mainPath); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestS3Getter_subdir(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test that requires AWS credentials in short mode")
	}

	g := new(S3Getter)
	dst := t.TempDir()

	// With a dir that doesn't exist
	err := g.Get(
		dst, testURL("https://s3.amazonaws.com/hc-go-getter-test/go-getter/folder/subfolder"))
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the main file exists
	subPath := filepath.Join(dst, "sub.tf")
	if _, err := os.Stat(subPath); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestS3Getter_GetFile(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test that requires AWS credentials in short mode")
	}

	g := new(S3Getter)
	dst := filepath.Join(t.TempDir(), "test-file")

	// Download
	err := g.GetFile(
		dst, testURL("https://s3.amazonaws.com/hc-go-getter-test/go-getter/folder/main.tf"))
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Verify the main file exists
	if _, err := os.Stat(dst); err != nil {
		t.Fatalf("err: %s", err)
	}
	assertContents(t, dst, "# Main\n")
}

func TestS3Getter_GetFile_badParams(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test that requires AWS credentials in short mode")
	}

	g := new(S3Getter)
	dst := filepath.Join(t.TempDir(), "test-file")

	// Download
	err := g.GetFile(
		dst,
		testURL("https://s3.amazonaws.com/hc-go-getter-test/go-getter/folder/main.tf?aws_access_key_id=foo&aws_access_key_secret=bar&aws_access_token=baz"))
	if err == nil {
		t.Fatalf("expected error, got none")
	}

	var respErr *awshttp.ResponseError
	if errors.As(err, &respErr) && respErr.HTTPStatusCode() != 403 {
		t.Fatalf("expected InvalidAccessKeyId error")
	}
}

func TestS3Getter_GetFile_notfound(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test that requires AWS credentials in short mode")
	}

	g := new(S3Getter)
	dst := filepath.Join(t.TempDir(), "test-file")

	// Download
	err := g.GetFile(
		dst, testURL("https://s3.amazonaws.com/hc-go-getter-test/go-getter/folder/404.tf"))
	if err == nil {
		t.Fatalf("expected error, got none")
	}
}

func TestS3Getter_ClientMode_dir(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test that requires AWS credentials in short mode")
	}

	g := new(S3Getter)

	// Check client mode on a key prefix with only a single key.
	mode, err := g.ClientMode(
		testURL("https://s3.amazonaws.com/hc-go-getter-test/go-getter/folder"))
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if mode != ClientModeDir {
		t.Fatal("expect ClientModeDir")
	}
}

func TestS3Getter_ClientMode_file(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test that requires AWS credentials in short mode")
	}

	g := new(S3Getter)

	// Check client mode on a key prefix which contains sub-keys.
	mode, err := g.ClientMode(
		testURL("https://s3.amazonaws.com/hc-go-getter-test/go-getter/folder/main.tf"))
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if mode != ClientModeFile {
		t.Fatal("expect ClientModeFile")
	}
}

func TestS3Getter_ClientMode_notfound(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test that requires AWS credentials in short mode")
	}

	g := new(S3Getter)

	// Check the client mode when a non-existent key is looked up. This does not
	// return an error, but rather should just return the file mode so that S3
	// can return an appropriate error later on. This also checks that the
	// prefix is handled properly (e.g., "/fold" and "/folder" don't put the
	// client mode into "dir".
	mode, err := g.ClientMode(
		testURL("https://s3.amazonaws.com/hc-go-getter-test/go-getter/fold"))
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if mode != ClientModeFile {
		t.Fatal("expect ClientModeFile")
	}
}

func TestS3Getter_ClientMode_collision(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test that requires AWS credentials in short mode")
	}

	g := new(S3Getter)

	// Check that the client mode is "file" if there is both an object and a
	// folder with a common prefix (i.e., a "collision" in the namespace).
	mode, err := g.ClientMode(
		testURL("https://s3.amazonaws.com/hc-go-getter-test/go-getter/collision/foo"))
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if mode != ClientModeFile {
		t.Fatal("expect ClientModeFile")
	}
}

func TestS3Getter_Url(t *testing.T) {
	var s3tests = []struct {
		name        string
		url         string
		region      string
		bucket      string
		path        string
		version     string
		accessKey   string
		secretKey   string
		expectedErr string
	}{
		{
			name:    "AWSv1234",
			url:     "s3::https://s3-eu-west-1.amazonaws.com/bucket/foo/bar.baz?version=1234",
			region:  "eu-west-1",
			bucket:  "bucket",
			path:    "foo/bar.baz",
			version: "1234",
		},
		{
			name:    "AWSVhostDot",
			url:     "s3::https://bucket.s3.eu-west-1.amazonaws.com/foo/bar.baz?version=1234",
			region:  "eu-west-1",
			bucket:  "bucket",
			path:    "foo/bar.baz",
			version: "1234",
		},
		{
			name:    "AWSVhostDash",
			url:     "s3::https://bucket.s3-eu-west-1.amazonaws.com/foo/bar.baz?version=1234",
			region:  "eu-west-1",
			bucket:  "bucket",
			path:    "foo/bar.baz",
			version: "1234",
		},
		{
			name:      "localhost-1",
			url:       "s3::http://127.0.0.1:9000/test-bucket/hello.txt?aws_access_key_id=TESTID&aws_access_key_secret=TestSecret&region=us-east-2&version=1",
			region:    "us-east-2",
			bucket:    "test-bucket",
			path:      "hello.txt",
			version:   "1",
			accessKey: "TESTID",
			secretKey: "TestSecret",
		},
		{
			name:      "localhost-2",
			url:       "s3::http://127.0.0.1:9000/test-bucket/hello.txt?aws_access_key_id=TESTID&aws_access_key_secret=TestSecret&version=1",
			region:    "us-east-1",
			bucket:    "test-bucket",
			path:      "hello.txt",
			version:   "1",
			accessKey: "TESTID",
			secretKey: "TestSecret",
		},
		{
			name:      "localhost-3",
			url:       "s3::http://127.0.0.1:9000/test-bucket/hello.txt?aws_access_key_id=TESTID&aws_access_key_secret=TestSecret",
			region:    "us-east-1",
			bucket:    "test-bucket",
			path:      "hello.txt",
			version:   "",
			accessKey: "TESTID",
			secretKey: "TestSecret",
		},
		{
			name:        "malformed s3 url",
			url:         "s3::https://s3.amazonaws.com/bucket",
			expectedErr: "URL is not a valid S3 URL",
		},
	}

	for i, pt := range s3tests {
		t.Run(pt.name, func(t *testing.T) {
			g := new(S3Getter)
			forced, src := getForcedGetter(pt.url)
			u, err := url.Parse(src)

			if err != nil {
				t.Errorf("test %d: unexpected error: %s", i, err)
			}
			if forced != "s3" {
				t.Fatalf("expected forced protocol to be s3")
			}

			region, bucket, path, version, creds, err := g.parseUrl(u)

			if err != nil {
				if pt.expectedErr == "" {
					t.Fatalf("err: %s", err)
				}
				if err.Error() != pt.expectedErr {
					t.Fatalf("expected %s, got %s", pt.expectedErr, err.Error())
				}
				return
			} else if pt.expectedErr != "" {
				t.Fatalf("expected error, got none")
			}
			if region != pt.region {
				t.Fatalf("expected %s, got %s", pt.region, region)
			}
			if bucket != pt.bucket {
				t.Fatalf("expected %s, got %s", pt.bucket, bucket)
			}
			if path != pt.path {
				t.Fatalf("expected %s, got %s", pt.path, path)
			}
			if version != pt.version {
				t.Fatalf("expected %s, got %s", pt.version, version)
			}

			if creds == nil {
				if pt.accessKey != "" || pt.secretKey != "" {
					t.Fatal("expected credentials, got nil")
				}
				return
			}

			credV, err := creds.Retrieve(context.Background())
			if err != nil {
				t.Fatalf("failed to get credentials: %s", err)
			}

			if credV.AccessKeyID != pt.accessKey {
				t.Fatalf("expected %s, got %s", pt.accessKey, credV.AccessKeyID)
			}

			if credV.SecretAccessKey != pt.secretKey {
				t.Fatalf("expected %s, got %s", pt.secretKey, credV.SecretAccessKey)
			}
		})
	}
}

func Test_S3Getter_ParseUrl_Malformed(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "path style",
			input:    "https://s3.amazonaws.com/bucket",
			expected: "URL is not a valid S3 URL",
		},
		{
			name:     "vhost-style, dash region indication",
			input:    "https://bucket.s3-us-east-1.amazonaws.com",
			expected: "URL is not a valid S3 URL",
		},
		{
			name:     "vhost-style, dot region indication",
			input:    "https://bucket.s3.us-east-1.amazonaws.com",
			expected: "URL is not a valid S3 URL",
		},
		{
			name:     "invalid host parts",
			input:    "https://invalid.host.parts.lenght.s3.us-east-1.amazonaws.com",
			expected: "URL is not a valid S3 URL",
		},
		{
			name:     "invalid host suffix",
			input:    "https://bucket.s3.amazonaws.com.invalid",
			expected: "URL is not a valid S3 compliant URL",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := new(S3Getter)
			u, err := url.Parse(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			_, _, _, _, _, err = g.parseUrl(u)
			if err == nil {
				t.Fatalf("expected error, got none")
			}
			if err.Error() != tt.expected {
				t.Fatalf("expected error '%s', got %s for %s", tt.expected, err.Error(), tt.name)
			}
		})
	}

}
