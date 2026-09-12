/*
Copyright 2024 The Kubernetes Authors All rights reserved.

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

package addons

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// minimalAuthorizedUserJSON is a valid authorized_user JSON accepted by google.CredentialsFromJSON.
// authorized_user credentials do not require a cryptographic key, making them suitable for unit tests.
const minimalAuthorizedUserJSON = `{
  "type": "authorized_user",
  "client_id": "test-client-id.apps.googleusercontent.com",
  "client_secret": "test-client-secret",
  "refresh_token": "test-refresh-token"
}`

func TestCredentialsFromCloudShellADC_EnvUnset(t *testing.T) {
	t.Setenv("CLOUDSDK_CONFIG", "")
	_, err := credentialsFromCloudShellADC(context.Background())
	if err == nil {
		t.Fatal("expected error when CLOUDSDK_CONFIG is unset, got nil")
	}
}

func TestCredentialsFromCloudShellADC_FileMissing(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLOUDSDK_CONFIG", dir)
	// No application_default_credentials.json written
	_, err := credentialsFromCloudShellADC(context.Background())
	if err == nil {
		t.Fatal("expected error when credentials file is missing, got nil")
	}
}

func TestCredentialsFromCloudShellADC_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLOUDSDK_CONFIG", dir)
	if err := os.WriteFile(filepath.Join(dir, "application_default_credentials.json"), []byte("not valid json"), 0600); err != nil {
		t.Fatalf("writing test file: %v", err)
	}
	_, err := credentialsFromCloudShellADC(context.Background())
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestCredentialsFromCloudShellADC_ValidCredentials(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CLOUDSDK_CONFIG", dir)
	data := []byte(minimalAuthorizedUserJSON)
	if err := os.WriteFile(filepath.Join(dir, "application_default_credentials.json"), data, 0600); err != nil {
		t.Fatalf("writing test file: %v", err)
	}
	creds, err := credentialsFromCloudShellADC(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if creds == nil {
		t.Fatal("expected non-nil credentials, got nil")
	}
	if string(creds.JSON) != string(data) {
		t.Errorf("creds.JSON does not match file contents\ngot:  %s\nwant: %s", creds.JSON, data)
	}
}
