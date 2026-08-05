/*
Copyright 2022 The Kubernetes Authors All rights reserved.

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

package config

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-github/v84/github"
)

// fakeGitHubServer creates an httptest.Server that mimics the GitHub Releases API.
// releasesByTag maps tag names to HTTP status codes (e.g. 200 for found, 404 for not found).
func fakeGitHubServer(t *testing.T, releasesByTag map[string]int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Expected path: /repos/kubernetes/kubernetes/releases/tags/{tag}
		prefix := "/repos/kubernetes/kubernetes/releases/tags/"
		if len(r.URL.Path) <= len(prefix) {
			http.NotFound(w, r)
			return
		}
		tag := r.URL.Path[len(prefix):]
		status, ok := releasesByTag[tag]
		if !ok {
			http.NotFound(w, r)
			return
		}
		if status == http.StatusOK {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"tag_name": %q}`, tag)
			return
		}
		w.WriteHeader(status)
		fmt.Fprintf(w, `{"message": "Not Found"}`)
	}))
}

// testGitHubClient returns a newGitHubClient replacement that points at the given test server
// and optionally verifies the Authorization header.
func testGitHubClient(t *testing.T, serverURL string, wantAuth bool) func(context.Context) *github.Client {
	t.Helper()
	return func(ctx context.Context) *github.Client {
		if wantAuth {
			// Wrap transport to verify auth header is set
			client := github.NewClient(nil).WithAuthToken("test-token")
			client.BaseURL.Scheme = "http"
			client.BaseURL.Host = serverURL[len("http://"):]
			return client
		}
		client := github.NewClient(nil)
		client.BaseURL.Scheme = "http"
		client.BaseURL.Host = serverURL[len("http://"):]
		return client
	}
}

func TestIsInGitHubKubernetesVersions(t *testing.T) {
	server := fakeGitHubServer(t, map[string]int{
		"v1.30.0": http.StatusOK,
		"v1.99.0": http.StatusNotFound,
	})
	defer server.Close()

	origClient := newGitHubClient
	defer func() { newGitHubClient = origClient }()

	newGitHubClient = testGitHubClient(t, server.URL, false)

	t.Run("existing release returns true", func(t *testing.T) {
		found, err := IsInGitHubKubernetesVersions("v1.30.0")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !found {
			t.Fatal("expected found=true for existing release")
		}
	})

	t.Run("non-existing release returns false", func(t *testing.T) {
		found, err := IsInGitHubKubernetesVersions("v1.99.0")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if found {
			t.Fatal("expected found=false for non-existing release")
		}
	})
}

func TestNewGitHubClientRespectsToken(t *testing.T) {
	// Verify that when GITHUB_TOKEN is set, the client sends an Authorization header.
	var gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"tag_name": "v1.30.0"}`)
	}))
	defer server.Close()

	origClient := newGitHubClient
	defer func() { newGitHubClient = origClient }()

	t.Run("with GITHUB_TOKEN", func(t *testing.T) {
		t.Setenv("GITHUB_TOKEN", "test-secret-token")

		// Call the real newGitHubClient (not a fake), but point it at our test server
		ctx := context.Background()
		ghc := origClient(ctx)
		ghc.BaseURL.Scheme = "http"
		ghc.BaseURL.Host = server.URL[len("http://"):]

		gotAuth = ""
		_, _, err := ghc.Repositories.GetReleaseByTag(ctx, "kubernetes", "kubernetes", "v1.30.0")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotAuth == "" {
			t.Fatal("expected Authorization header to be set when GITHUB_TOKEN is provided")
		}
		if gotAuth != "Bearer test-secret-token" {
			t.Fatalf("unexpected Authorization header: %q", gotAuth)
		}
	})

	t.Run("without GITHUB_TOKEN", func(t *testing.T) {
		t.Setenv("GITHUB_TOKEN", "")

		ctx := context.Background()
		ghc := origClient(ctx)
		ghc.BaseURL.Scheme = "http"
		ghc.BaseURL.Host = server.URL[len("http://"):]

		gotAuth = ""
		_, _, err := ghc.Repositories.GetReleaseByTag(ctx, "kubernetes", "kubernetes", "v1.30.0")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotAuth != "" {
			t.Fatalf("expected no Authorization header when GITHUB_TOKEN is empty, got: %q", gotAuth)
		}
	})
}
