// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package getter

import (
	"testing"
)

func TestS3Detector(t *testing.T) {
	cases := []struct {
		Input  string
		Output string
	}{
		// Virtual hosted style
		{
			"bucket.s3.amazonaws.com/foo",
			"s3::https://s3.amazonaws.com/bucket/foo",
		},
		{
			"bucket.s3.amazonaws.com/foo/bar",
			"s3::https://s3.amazonaws.com/bucket/foo/bar",
		},
		{
			"bucket.s3.amazonaws.com/foo/bar.baz",
			"s3::https://s3.amazonaws.com/bucket/foo/bar.baz",
		},
		{
			"bucket.s3-eu-west-1.amazonaws.com/foo",
			"s3::https://s3-eu-west-1.amazonaws.com/bucket/foo",
		},
		{
			"bucket.s3-eu-west-1.amazonaws.com/foo/bar",
			"s3::https://s3-eu-west-1.amazonaws.com/bucket/foo/bar",
		},
		{
			"bucket.s3-eu-west-1.amazonaws.com/foo/bar.baz",
			"s3::https://s3-eu-west-1.amazonaws.com/bucket/foo/bar.baz",
		},
		// 5 parts Virtual hosted-style
		{
			"bucket.s3.eu-west-1.amazonaws.com/foo/bar.baz",
			"s3::https://s3.eu-west-1.amazonaws.com/bucket/foo/bar.baz",
		},
		// Path style
		{
			"s3.amazonaws.com/bucket/foo",
			"s3::https://s3.amazonaws.com/bucket/foo",
		},
		{
			"s3.amazonaws.com/bucket/foo/bar",
			"s3::https://s3.amazonaws.com/bucket/foo/bar",
		},
		{
			"s3.amazonaws.com/bucket/foo/bar.baz",
			"s3::https://s3.amazonaws.com/bucket/foo/bar.baz",
		},
		{
			"s3-eu-west-1.amazonaws.com/bucket/foo",
			"s3::https://s3-eu-west-1.amazonaws.com/bucket/foo",
		},
		{
			"s3-eu-west-1.amazonaws.com/bucket/foo/bar",
			"s3::https://s3-eu-west-1.amazonaws.com/bucket/foo/bar",
		},
		{
			"s3-eu-west-1.amazonaws.com/bucket/foo/bar.baz",
			"s3::https://s3-eu-west-1.amazonaws.com/bucket/foo/bar.baz",
		},
		// Misc tests
		{
			"s3-eu-west-1.amazonaws.com/bucket/foo/bar.baz?version=1234",
			"s3::https://s3-eu-west-1.amazonaws.com/bucket/foo/bar.baz?version=1234",
		},
	}

	pwd := "/pwd"
	f := new(S3Detector)
	for i, tc := range cases {
		output, ok, err := f.Detect(tc.Input, pwd)
		if err != nil {
			t.Fatalf("err: %s", err)
		}
		if !ok {
			t.Fatal("not ok")
		}

		if output != tc.Output {
			t.Fatalf("%d: bad: %#v", i, output)
		}
	}
}

func TestS3Detector_MalformedDetectHTTP(t *testing.T) {
	cases := []struct {
		Name     string
		Input    string
		Expected string
		Output   string
	}{
		{
			"valid url",
			"s3.amazonaws.com/bucket/foo/bar",
			"",
			"s3::https://s3.amazonaws.com/bucket/foo/bar",
		},
		{
			"empty url",
			"",
			"",
			"",
		},
		{
			"not valid url",
			"bucket/foo/bar",
			"error parsing S3 URL",
			"",
		},
		{
			"not valid url domain",
			"s3.amazonaws.com.invalid/bucket/foo/bar",
			"error parsing S3 URL",
			"",
		},
		{
			"not valid url lenght",
			"http://s3.amazonaws.com",
			"URL is not a valid S3 URL",
			"",
		},
	}

	pwd := "/pwd"
	f := new(S3Detector)
	for _, tc := range cases {
		output, _, err := f.Detect(tc.Input, pwd)
		if err != nil {
			if err.Error() != tc.Expected {
				t.Fatalf("expected error %s, got %s for %s", tc.Expected, err.Error(), tc.Name)
			}
		}

		if output != tc.Output {
			t.Fatalf("expected %s, got %s for %s", tc.Output, output, tc.Name)
		}
	}
}
