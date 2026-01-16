/*
Copyright 2025 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package gh

import (
	"strings"
	"testing"

	"github.com/google/go-github/v81/github"
)

func TestAssetSHA256(t *testing.T) {
	t.Run("found_with_sha256_prefix", func(t *testing.T) {
		assets := []*github.ReleaseAsset{
			{
				Name:               github.Ptr("minikube-linux-amd64"),
				Digest:             github.Ptr("sha256:abcdef123456"),
				ID:                 github.Ptr(int64(101)),
				BrowserDownloadURL: github.Ptr("http://example/minikube-linux-amd64"),
			},
		}
		got, err := AssetSHA256("minikube-linux-amd64", assets)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if string(got) != "abcdef123456" {
			t.Fatalf("expected digest %q, got %q", "abcdef123456", string(got))
		}
	})

	t.Run("found_without_prefix", func(t *testing.T) {
		assets := []*github.ReleaseAsset{
			{
				Name:               github.Ptr("minikube-darwin-arm64"),
				Digest:             github.Ptr("1234abcd"),
				ID:                 github.Ptr(int64(102)),
				BrowserDownloadURL: github.Ptr("http://example/minikube-darwin-arm64"),
			},
		}
		got, err := AssetSHA256("minikube-darwin-arm64", assets)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if string(got) != "1234abcd" {
			t.Fatalf("expected digest %q, got %q", "1234abcd", string(got))
		}
	})

	t.Run("asset_missing_digest", func(t *testing.T) {
		assets := []*github.ReleaseAsset{
			{
				Name:               github.Ptr("minikube-windows-amd64.exe"),
				Digest:             github.Ptr(""),
				ID:                 github.Ptr(int64(103)),
				BrowserDownloadURL: github.Ptr("http://example/minikube-windows-amd64.exe"),
			},
		}
		_, err := AssetSHA256("minikube-windows-amd64.exe", assets)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "has no digest") {
			t.Fatalf("unexpected error message: %v", err)
		}
	})

	t.Run("asset_not_found", func(t *testing.T) {
		assets := []*github.ReleaseAsset{
			{
				Name:               github.Ptr("unrelated"),
				Digest:             github.Ptr("sha256:deadbeef"),
				ID:                 github.Ptr(int64(104)),
				BrowserDownloadURL: github.Ptr("http://example/unrelated"),
			},
		}
		_, err := AssetSHA256("missing", assets)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if !strings.Contains(err.Error(), `asset "missing" not found`) {
			t.Fatalf("unexpected error message: %v", err)
		}
	})
}
