/*
Copyright 2020 The Kubernetes Authors All rights reserved.

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

package cruntime

import (
	"archive/tar"
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"k8s.io/minikube/pkg/minikube/command"

	"k8s.io/minikube/pkg/version"
)

func TestAddRepoTagToImageName(t *testing.T) {
	var tests = []struct {
		imgName string
		want    string
	}{
		{"gcr.io/k8s-minikube/storage-provisioner:" + version.GetStorageProvisionerVersion(), "gcr.io/k8s-minikube/storage-provisioner:" + version.GetStorageProvisionerVersion()},
	}
	for _, tc := range tests {
		t.Run(tc.imgName, func(t *testing.T) {
			got := addRepoTagToImageName(tc.imgName)
			if got != tc.want {
				t.Errorf("expected image name to be: %q but got %q", tc.want, got)
			}
		})
	}
}

func TestParseContainerdVersion(t *testing.T) {
	var tests = []struct {
		version string
		want    string
	}{
		{"containerd github.com/containerd/containerd v1.2.0 c4446665cb9c30056f4998ed953e6d4ff22c7c39", "1.2.0"},
		{"containerd github.com/containerd/containerd v1.2.1-rc.0 de1f167ab96338a9f5c2b17347abf84bdf1dd411", "1.2.1-rc.0"},
		{"containerd github.com/containerd/containerd 1.4.4-0ubuntu1 ", "1.4.4-0ubuntu1"},
		{"containerd github.com/containerd/containerd 1.5.2-0ubuntu1~21.04.2 ", "1.5.2-0ubuntu1"},
		{"containerd github.com/containerd/containerd 1.5.4~ds1 1.5.4~ds1-1", "1.5.4"},
	}
	for _, tc := range tests {
		t.Run(tc.version, func(t *testing.T) {
			got, err := parseContainerdVersion(tc.version)
			if err != nil {
				t.Fatalf("parse(%s): %v", tc.version, err)
			}
			if got != tc.want {
				t.Errorf("expected version to be: %q but got %q", tc.want, got)
			}
		})
	}
}

// TestDownloadRemote runs the guest download commands locally so that it checks
// HTTP and filesystem behavior without depending on command arguments.
func TestDownloadRemote(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("download commands require a Unix environment")
	}
	for _, tool := range []string{"mktemp", "curl", "tar", "rm"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("%s is required: %v", tool, err)
		}
	}
	// BSD mktemp does not honor TMPDIR without a template. Keep all generated
	// files inside the test directory on both macOS and Linux.
	mktemp, err := exec.LookPath("mktemp")
	if err != nil {
		t.Fatal(err)
	}
	binDir := t.TempDir()
	script := "#!/bin/sh\nexec " + mktemp + " \"$@\" \"$TMPDIR/download.XXXXXXXX\"\n"
	if err := os.WriteFile(filepath.Join(binDir, "mktemp"), []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	var archive bytes.Buffer
	tw := tar.NewWriter(&archive)
	const dockerfile = "FROM scratch\n"
	if err := tw.WriteHeader(&tar.Header{Name: "Dockerfile", Mode: 0600, Size: int64(len(dockerfile))}); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write([]byte(dockerfile)); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name     string
		status   int
		body     []byte
		redirect bool
		wantErr  bool
	}{
		{name: "success", status: http.StatusOK, body: archive.Bytes()},
		{name: "redirect", status: http.StatusOK, body: archive.Bytes(), redirect: true},
		{name: "not found", status: http.StatusNotFound, body: []byte("not found"), wantErr: true},
		// A valid archive in an error response must still be rejected. Otherwise a
		// failure inside tar could hide the missing HTTP status check.
		{name: "server error with archive", status: http.StatusInternalServerError, body: archive.Bytes(), wantErr: true},
		{name: "invalid archive", status: http.StatusOK, body: []byte("not an archive"), wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			t.Setenv("TMPDIR", dir)
			sentinel := filepath.Join(dir, "existing")
			if err := os.WriteFile(sentinel, []byte("keep me"), 0600); err != nil {
				t.Fatal(err)
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tc.redirect && r.URL.Path == "/context" {
					http.Redirect(w, r, "/archive", http.StatusFound)
					return
				}
				w.WriteHeader(tc.status)
				_, _ = w.Write(tc.body)
			}))
			defer server.Close()
			got, err := downloadRemote(command.NewExecRunner(false), server.URL+"/context")
			if (err != nil) != tc.wantErr {
				t.Errorf("downloadRemote() error = %v, want error %v", err, tc.wantErr)
			}
			if tc.wantErr {
				if got != "" {
					t.Errorf("failed download returned context %q", got)
				}
			} else {
				content, err := os.ReadFile(filepath.Join(got, "Dockerfile"))
				if err != nil {
					t.Fatal(err)
				}
				if string(content) != dockerfile {
					t.Errorf("Dockerfile = %q, want %q", content, dockerfile)
				}
				// A successful context belongs to the caller; only the download is temporary.
				if err := os.RemoveAll(got); err != nil {
					t.Fatal(err)
				}
			}
			entries, err := os.ReadDir(dir)
			if err != nil {
				t.Fatal(err)
			}
			if len(entries) != 1 || entries[0].Name() != "existing" {
				t.Errorf("download left temporary files: %v", entries)
			}
			content, err := os.ReadFile(sentinel)
			if err != nil || string(content) != "keep me" {
				t.Errorf("existing file changed: content=%q, error=%v", content, err)
			}
		})
	}
}
