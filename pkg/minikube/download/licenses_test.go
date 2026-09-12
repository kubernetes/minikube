package download

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func tarGzWithLicense() []byte {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	content := []byte("Apache License 2.0")
	hdr := &tar.Header{
		Name: "licenses/cloud.google.com/go/compute/metadata/LICENSE",
		Mode: 0600,
		Size: int64(len(content)),
	}
	_ = tw.WriteHeader(hdr)
	_, _ = tw.Write(content)
	_ = tw.Close()
	_ = gz.Close()
	return buf.Bytes()
}

func TestLicensesTarballURLForVersion_Fallback(t *testing.T) {
	orig := httpHead
	defer func() { httpHead = orig }()
	httpHead = func(_ string) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusNotFound}, nil
	}
	url := licensesTarballURLForVersion("v1.39.0")
	if !strings.Contains(url, "storage.googleapis.com") {
		t.Fatalf("expected GCS fallback, got %q", url)
	}
	if !strings.Contains(url, "v1.39.0") {
		t.Fatalf("expected version in URL, got %q", url)
	}
}

func TestLicensesTarballURLForVersion_GitHubHit(t *testing.T) {
	orig := httpHead
	defer func() { httpHead = orig }()
	httpHead = func(url string) (*http.Response, error) {
		if strings.Contains(url, "github.com") {
			return &http.Response{StatusCode: http.StatusOK}, nil
		}
		return &http.Response{StatusCode: http.StatusNotFound}, nil
	}
	url := licensesTarballURLForVersion("v1.39.0")
	if !strings.Contains(url, "github.com") {
		t.Fatalf("expected github URL, got %q", url)
	}
}

func TestDownloadAndExtractLicenses_Success(t *testing.T) {
	data := tarGzWithLicense()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	}))
	defer srv.Close()

	orig := httpGet
	httpGet = srv.Client().Get
	defer func() { httpGet = orig }()

	dir := t.TempDir()
	if err := downloadAndExtractLicenses(srv.URL, dir); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	b, err := os.ReadFile(dir + "/licenses/cloud.google.com/go/compute/metadata/LICENSE")
	if err != nil {
		t.Fatalf("failed to read extracted license: %v", err)
	}
	if !strings.Contains(string(b), "Apache License") {
		t.Fatalf("expected Apache License, got %q", string(b))
	}
}

func TestDownloadAndExtractLicenses_404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	orig := httpGet
	httpGet = srv.Client().Get
	defer func() { httpGet = orig }()

	if err := downloadAndExtractLicenses(srv.URL, t.TempDir()); err == nil {
		t.Fatal("expected error for 404, got nil")
	} else if !strings.Contains(err.Error(), "404") {
		t.Fatalf("expected 404 in error, got %v", err)
	}
}

func TestLicenses_EndToEnd_404Repro(t *testing.T) {
	origHead, origGet := httpHead, httpGet
	defer func() { httpHead = origHead; httpGet = origGet }()

	httpHead = func(_ string) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusNotFound}, nil
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer srv.Close()
	httpGet = func(url string) (*http.Response, error) {
		return srv.Client().Get(srv.URL)
	}

	err := downloadAndExtractLicenses(licensesTarballURLForVersion("v9.99.99-unreleased"), t.TempDir())
	if err == nil || !strings.Contains(err.Error(), "404") {
		t.Fatalf("should fail with 404 for unreleased version, got %v", err)
	}
}
