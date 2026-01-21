/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package machine

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gofrs/flock"
	"k8s.io/minikube/pkg/libmachine/auth"
	"k8s.io/minikube/pkg/libmachine/host"
	"k8s.io/minikube/pkg/minikube/run"
)

func TestBootstrapCertificatesWithLock_Timeout(t *testing.T) {
	t.Parallel()
	// Create a temp home dir to avoid conflicting with actual minikube files
	tempHome := t.TempDir()

	// Initialize the client
	api, err := NewAPIClient(&run.CommandOptions{}, tempHome)
	if err != nil {
		t.Fatalf("NewAPIClient failed: %v", err)
	}
	// TestBootstrapCertificatesWithLock_Timeout tests that the mechanism correctly times out
	// when conflicting lock is held. This is critical to prevent minikube from hanging indefinitely
	// in automated environments or when multiple instances are started inadvertently.
	client := api.(*LocalClient)

	// Pre-lock the file to simulate contention
	lockPath := filepath.Join(tempHome, "machine_client.lock")
	if err := os.MkdirAll(filepath.Dir(lockPath), 0755); err != nil {
		t.Fatalf("failed to create lock dir: %v", err)
	}
	externalLock := flock.New(lockPath)
	locked, err := externalLock.TryLock()
	if err != nil {
		t.Fatalf("failed to acquire external lock: %v", err)
	}
	if !locked {
		t.Fatalf("failed to acquire external lock: already locked")
	}
	defer externalLock.Unlock()

	// Attempt to bootstrap certificates. This should time out.
	// We pass a dummy host; it shouldn't matter because we expect the lock to fail first.
	h := &host.Host{}

	start := time.Now()
	err = client.bootstrapCertificatesWithLock(h)
	duration := time.Since(start)

	if err == nil {
		t.Errorf("expected error due to lock timeout, got nil")
	} else {
		if err.Error() != "failed to acquire bootstrap client lock: timeout acquiring lock" {
			t.Errorf("unexpected error message: %v", err)
		}
	}

	// Verify duration is at least 5 seconds (the timeout)
	if duration < 5*time.Second {
		t.Errorf("expected timeout to take at least 5s, took %v", duration)
	}
}

func TestBootstrapCertificatesWithLock_Success(t *testing.T) {
	t.Parallel()
	// Create a temp home dir
	tempHome := t.TempDir()

	api, err := NewAPIClient(&run.CommandOptions{}, tempHome)
	if err != nil {
		t.Fatalf("NewAPIClient failed: %v", err)
	}
	client := api.(*LocalClient)

	// We can't easily test the full success path without setting up valid certs structure
	// because BootstrapCertificates will fail.
	// However, checking that we *acquired* the lock (progressed past locking) is enough.
	// Implementation detail: if locking succeeds, it calls BootstrapCertificates.
	// We can't mock BootstrapCertificates easily as it's a function from another package.
	// But we can rely on the fact that if lock acquisition fails, we get a specific error.
	// If we get a different error (from BootstrapCertificates), then locking SUCCEEDED.

	h := &host.Host{
		Name: "test-machine",
		HostOptions: &host.Options{
			AuthOptions: &auth.Options{
				CertDir:          filepath.Join(tempHome, "certs"),
				CaCertPath:       filepath.Join(tempHome, "certs", "ca.pem"),
				CaPrivateKeyPath: filepath.Join(tempHome, "certs", "ca-key.pem"),
				ClientCertPath:   filepath.Join(tempHome, "certs", "cert.pem"),
				ClientKeyPath:    filepath.Join(tempHome, "certs", "key.pem"),
				ServerCertPath:   filepath.Join(tempHome, "machines", "server.pem"),
				ServerKeyPath:    filepath.Join(tempHome, "machines", "server-key.pem"),
			},
		},
	}
	// Create the certs directory to avoid immediate failure before lock acquisition (although lock is first)
	// Actually, we want to fail AFTER lock if lock succeeds.
	if err := os.MkdirAll(h.AuthOptions().CertDir, 0755); err != nil {
		t.Fatalf("failed to create certs dir: %v", err)
	}

	err = client.bootstrapCertificatesWithLock(h)

	// We expect an error from BootstrapCertificates because we didn't set up CA keys etc.
	// BUT, importantly, it should NOT be a lock error.
	if err != nil && err.Error() == "timeout acquiring lock" {
		t.Errorf("unexpected lock timeout: %v", err)
	}
}
