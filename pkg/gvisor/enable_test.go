/*
Copyright 2026 The Kubernetes Authors All rights reserved.

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

package gvisor

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestDownloadFileToDestHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `<?xml version='1.0' encoding='UTF-8'?><Error><Code>NoSuchKey</Code></Error>`)
	}))
	t.Cleanup(srv.Close)

	t.Run("fresh dest", func(t *testing.T) {
		dest := filepath.Join(t.TempDir(), "containerd-shim-runsc-v1")
		if err := downloadFileToDest(srv.URL+"/containerd-shim-runsc-v1", dest); err == nil {
			t.Fatal("downloadFileToDest() on HTTP error = nil, want error (regression for #23709 exec format error)")
		}
		if _, err := os.Stat(dest); err == nil {
			t.Fatal("HTTP error ignored saving payload to destination")
		} else if !os.IsNotExist(err) {
			t.Fatalf("unexpected error: %s", err)
		}
	})

	t.Run("existing dest", func(t *testing.T) {
		dest := filepath.Join(t.TempDir(), "containerd-shim-runsc-v1")
		if err := os.WriteFile(dest, []byte("old-binary"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := downloadFileToDest(srv.URL+"/containerd-shim-runsc-v1", dest); err == nil {
			t.Fatal("downloadFileToDest() on HTTP error = nil, want error (regression for #23709 exec format error)")
		}
		data, err := os.ReadFile(dest)
		if err != nil {
			t.Fatalf("existing dest was deleted by failed download: %s", err)
		}
		if string(data) != "old-binary" {
			t.Fatalf("existing dest was overwritten by failed download: %q", data)
		}
	})
}

func TestDownloadFileToDestSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "binary-bytes")
	}))
	t.Cleanup(srv.Close)
	dest := filepath.Join(t.TempDir(), "runsc")
	if err := downloadFileToDest(srv.URL+"/runsc", dest); err != nil {
		t.Fatalf("downloadFileToDest() = %s", err)
	}
	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("downloaded file cannot be read: %s", err)
	}
	if string(data) != "binary-bytes" {
		t.Fatalf("downloaded body = %q, want %q", data, "binary-bytes")
	}
}
