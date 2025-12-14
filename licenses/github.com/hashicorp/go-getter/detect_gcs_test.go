// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package getter

import (
	"testing"
)

func TestGCSDetector(t *testing.T) {
	cases := []struct {
		Input  string
		Output string
	}{
		{
			"www.googleapis.com/storage/v1/bucket/foo",
			"gcs::https://www.googleapis.com/storage/v1/bucket/foo",
		},
		{
			"www.googleapis.com/storage/v1/bucket/foo/bar",
			"gcs::https://www.googleapis.com/storage/v1/bucket/foo/bar",
		},
		{
			"www.googleapis.com/storage/v1/foo/bar.baz",
			"gcs::https://www.googleapis.com/storage/v1/foo/bar.baz",
		},
		{
			"www.googleapis.com/storage/v2/foo/bar/toor.baz",
			"gcs::https://www.googleapis.com/storage/v2/foo/bar/toor.baz",
		},
	}

	pwd := "/pwd"
	f := new(GCSDetector)
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

func TestGCSDetector_MalformedDetectHTTP(t *testing.T) {
	cases := []struct {
		Name     string
		Input    string
		Expected string
		Output   string
	}{
		{
			"valid url",
			"www.googleapis.com/storage/v1/my-bucket/foo/bar",
			"",
			"gcs::https://www.googleapis.com/storage/v1/my-bucket/foo/bar",
		},
		{
			"empty url",
			"",
			"",
			"",
		},
		{
			"not valid url",
			"storage/v1/my-bucket/foo/bar",
			"error parsing GCS URL",
			"",
		},
		{
			"not valid url domain",
			"www.googleapis.com.invalid/storage/v1/",
			"URL is not a valid GCS URL",
			"",
		},
		{
			"not valid url length",
			"http://www.googleapis.com/storage",
			"URL is not a valid GCS URL",
			"",
		},
	}

	pwd := "/pwd"
	f := new(GCSDetector)
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
