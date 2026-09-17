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
	"archive/tar"
	"bytes"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReleaseArchURL(t *testing.T) {
	tests := []struct {
		goarch string
		want   string
	}{
		{"amd64", "https://storage.googleapis.com/gvisor/releases/release/latest/x86_64/"},
		{"arm64", "https://storage.googleapis.com/gvisor/releases/release/latest/aarch64/"},
	}
	for _, tc := range tests {
		if got := releaseArchURL(tc.goarch); got != tc.want {
			t.Errorf("releaseArchURL(%q) = %q, want %q", tc.goarch, got, tc.want)
		}
	}
	if !strings.HasSuffix(gvisorTarballURL(), "gvisor.tar.bz2") {
		t.Errorf("gvisorTarballURL() = %q, want suffix gvisor.tar.bz2 (individual binaries no longer published, see #23709)", gvisorTarballURL())
	}
	if strings.HasSuffix(gvisorTarballURL(), "containerd-shim-runsc-v1") || strings.HasSuffix(gvisorTarballURL(), "/runsc") {
		t.Errorf("gvisorTarballURL() = %q, must not point at removed individual binaries", gvisorTarballURL())
	}
	for _, tc := range []struct {
		goarch string
		arch   string
	}{
		{"amd64", "x86_64"},
		{"arm64", "aarch64"},
	} {
		url := releaseArchURL(tc.goarch) + "gvisor.tar.bz2"
		if !strings.Contains(url, "/"+tc.arch+"/") {
			t.Errorf("tarball url for %q = %q, want arch path %q", tc.goarch, url, tc.arch)
		}
	}
}

func buildTestTar(t *testing.T, files map[string]string) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	for name, body := range files {
		if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0o755, Size: int64(len(body)), Typeflag: tar.TypeReg}); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	return &buf
}

func TestExtractTarInstallsBinaries(t *testing.T) {
	dir := t.TempDir()
	var clean bytes.Buffer
	tw := tar.NewWriter(&clean)
	for _, e := range []struct{ name, body string }{
		{"runsc", "runsc-binary"},
		{"containerd-shim-runsc-v1", "shim-binary"},
		{"gvisor-bin/gvisor_sentry", "sentry"},
		{"unrelated.txt", "ignored"},
		{"gvisor-bin/nested/evil", "ignored-nested"},
	} {
		if err := tw.WriteHeader(&tar.Header{Name: e.name, Mode: 0o755, Size: int64(len(e.body)), Typeflag: tar.TypeReg}); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(e.body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := extractTar(&clean, dir); err != nil {
		t.Fatalf("extractTar() = %v", err)
	}
	for _, name := range []string{"runsc", "containerd-shim-runsc-v1", "gvisor-bin/gvisor_sentry"} {
		p := filepath.Join(dir, name)
		data, err := os.ReadFile(p)
		if err != nil {
			t.Errorf("expected %s installed: %v", name, err)
			continue
		}
		if len(data) == 0 {
			t.Errorf("%s installed but empty", name)
		}
		fi, err := os.Stat(p)
		if err != nil {
			t.Fatal(err)
		}
		if fi.Mode().Perm()&0o111 == 0 {
			t.Errorf("%s not executable: %v", name, fi.Mode())
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "unrelated.txt")); !os.IsNotExist(err) {
		t.Errorf("unrelated.txt should be skipped")
	}
	if _, err := os.Stat(filepath.Join(dir, "gvisor-bin/nested/evil")); !os.IsNotExist(err) {
		t.Errorf("nested gvisor-bin entries should be skipped")
	}
}

func TestExtractTarMissingBinaryFailsClosed(t *testing.T) {
	dir := t.TempDir()
	buf := buildTestTar(t, map[string]string{"runsc": "only-runsc"})
	if err := extractTar(buf, dir); err == nil {
		t.Fatal("extractTar() with missing shim = nil, want error")
	}
}

func TestExtractTarRejectsTraversal(t *testing.T) {
	dir := t.TempDir()
	buf := buildTestTar(t, map[string]string{
		"runsc":                    "runsc-binary",
		"containerd-shim-runsc-v1": "shim-binary",
		"../evil":                  "evil",
	})
	if err := extractTar(buf, dir); err == nil {
		t.Fatal("extractTar() with traversal entry = nil, want error")
	}
}

func TestDownloadToTempFailsClosedOn404XML(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `<?xml version='1.0' encoding='UTF-8'?><Error><Code>NoSuchKey</Code><Message>The specified key does not exist.</Message></Error>`)
	}))
	defer srv.Close()
	if _, err := downloadToTemp(srv.URL + "/containerd-shim-runsc-v1"); err == nil {
		t.Fatal("downloadToTemp() on 404 XML = nil, want error (regression for #23709 exec format error)")
	}
}

func TestDownloadToTempSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "binary-bytes")
	}))
	defer srv.Close()
	p, err := downloadToTemp(srv.URL + "/gvisor.tar.bz2")
	if err != nil {
		t.Fatalf("downloadToTemp() = %v", err)
	}
	defer os.Remove(p)
	data, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "binary-bytes" {
		t.Errorf("downloaded body = %q", data)
	}
}

func TestExtractTarRejectsSymlink(t *testing.T) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	for _, e := range []struct {
		name     string
		body     string
		flag     byte
		linkname string
	}{
		{"runsc", "runsc-binary", tar.TypeReg, ""},
		{"containerd-shim-runsc-v1", "/etc/passwd", tar.TypeSymlink, "/etc/passwd"},
	} {
		if err := tw.WriteHeader(&tar.Header{Name: e.name, Mode: 0o755, Size: int64(len(e.body)), Typeflag: e.flag, Linkname: e.linkname}); err != nil {
			t.Fatal(err)
		}
		if e.flag == tar.TypeReg {
			if _, err := tw.Write([]byte(e.body)); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := extractTar(&buf, t.TempDir()); err == nil {
		t.Fatal("extractTar() with symlink shim = nil, want error")
	}
}

func TestDotDotComponentCheck(t *testing.T) {
	if !hasDotDotComponent("../evil") {
		t.Errorf("hasDotDotComponent(../evil) = false, want true")
	}
	if !hasDotDotComponent("a/../../evil") {
		t.Errorf("hasDotDotComponent(a/../../evil) = false, want true")
	}
	if hasDotDotComponent("foo..bar") {
		t.Errorf("hasDotDotComponent(foo..bar) = true, want false (not a path traversal)")
	}
	if hasDotDotComponent("gvisor-bin/gvisor_sentry") {
		t.Errorf("hasDotDotComponent(gvisor-bin/gvisor_sentry) = true, want false")
	}
}

// gvisorTestTarBz2Base64 is the former testdata/gvisor-test.tar.bz2 fixture
// (runsc, containerd-shim-runsc-v1, gvisor-bin/gvisor_sentry), inlined as
// base64 so the change stays text-reviewable instead of a binary blob.
const gvisorTestTarBz2Base64 = "QlpoOTFBWSZTWW2U0u8AAdj/t/64Q4RQA//jO2b/eP/v/+AAQY4QCAAAAYAgCAhQA14AAAAcZME0MhkZGTQ0AaDIwgGg0aZDENADjJgmhkMjIyaGgDQZGEA0GjTIYhoAcZME0MhkZGTQ0AaDIwgGg0aZDENADjJgmhkMjIyaGgDQZGEA0GjTIYhoAFSSASEzQaI0mTJ6EA00aANDI0aaYmIaehPjNR+hPAT0nR/qqVf3q6VvaV5zrdSS85SVPYcZ2Vx5PIcBxHpFgcpRJzFHCchiXIHGVIlipIPCcpY9woJv/Bw7Xvvgswq7E2BN8liG2KQm4UfEUTdGhCxkWHqKnYKvP9HSPIfJv8vWyJOyUTxjWJsJROG1Sqit4nfMZN0m1t8lXOS1VZhgdc3T8TAwHodJMfN76p/MtdO6ev2fF7cyaDsE8R3jty+VXJJ0e1wBOQnVOiXAmwn0koTw/e6H5PzeP9P38WhN3bEsT82+jskpfgSob9TmolHNYnqmEbML11bKmZLrfDqsdw3DPTTHkyiafQWdDFnmTAomFHgMdD1mejRuLrqxzyheS++xUNplf95u54kqYYVkSxNZfnhL9Gha5flV24S+W0uhcS7NnzmGrPAtoWmVy/LGX31btUTUuLjK81M6L5mZWMNDrk33AfseE8Q7Z4D0nlMzxj9TtHXOYYcptFix5nbq7+TslHlMS+8vLZF5oXReV7x8ksTynGfZP+J3CdgmZPrwJ3SYP8JsHnO2NU/so28iec2xyenn0cE9W0az0Eo1lHuHf4Ce6ayfUTY4TE4Z3zznKOLi4vr96199Zk6ZNqN43D1DZN6GsOcnIc5PmJY5jIlkxJSxKlE3pLmMUdGDch0yaiZE6hNh85dE0TeGotKmROHi3OPqzQmkpXQJxyZTFia4ZEzJjGRKdQ0FjAbs1yVzk1TeJ44azEn/E2GhxDpmCdg2xzj2Cx6yie0e0WesUH/i7kinChINsppd4A=="

func gvisorTestTarBz2(t *testing.T) []byte {
	t.Helper()
	blob, err := base64.StdEncoding.DecodeString(gvisorTestTarBz2Base64)
	if err != nil {
		t.Fatal(err)
	}
	return blob
}

func TestExtractGvisorTarballFromBz2Fixture(t *testing.T) {
	dir := t.TempDir()
	tarball := filepath.Join(t.TempDir(), "gvisor.tar.bz2")
	if err := os.WriteFile(tarball, gvisorTestTarBz2(t), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := extractGvisorTarball(tarball, dir); err != nil {
		t.Fatalf("extractGvisorTarball(fixture) = %v", err)
	}
	for name, want := range map[string]string{
		"runsc":                    "runsc-binary-content-0123456789",
		"containerd-shim-runsc-v1": "shim-binary-content-0123456789",
		"gvisor-bin/gvisor_sentry": "sentry-sidecar",
	} {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Errorf("expected %s installed: %v", name, err)
			continue
		}
		if string(data) != want {
			t.Errorf("%s = %q, want %q", name, data, want)
		}
	}
}

func TestDownloadBinariesFromEndToEnd(t *testing.T) {
	blob := gvisorTestTarBz2(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(blob)
	}))
	defer srv.Close()
	dir := t.TempDir()
	if err := downloadBinariesFrom(srv.URL+"/gvisor.tar.bz2", dir); err != nil {
		t.Fatalf("downloadBinariesFrom() = %v", err)
	}
	for _, name := range []string{"runsc", "containerd-shim-runsc-v1", "gvisor-bin/gvisor_sentry"} {
		fi, err := os.Stat(filepath.Join(dir, name))
		if err != nil {
			t.Errorf("expected %s installed: %v", name, err)
			continue
		}
		if fi.Mode().Perm()&0o111 == 0 {
			t.Errorf("%s not executable: %v", name, fi.Mode())
		}
	}
}
