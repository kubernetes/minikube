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

package image

import (
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/google/go-containerregistry/pkg/v1/tarball"
)

// TestSaveToTarFileMemory measures peak heap usage when saving a daemon image to cache.
// Before the fix, this allocates roughly the full image size in RAM (buffered daemon read).
// After the fix, allocations should be near zero (streaming directly to disk).
//
// Requires Docker daemon with redis:7 pulled locally:
//
//	docker pull redis:7
func TestSaveToTarFileMemory(t *testing.T) {
	if _, err := os.Stat("/var/run/docker.sock"); err != nil {
		if _, err2 := os.Stat(os.ExpandEnv("$HOME/.docker/run/docker.sock")); err2 != nil {
			t.Skip("Docker daemon not available")
		}
	}

	UseDaemon(true)
	UseRemote(false)
	t.Cleanup(func() {
		UseDaemon(true)
		UseRemote(true)
	})

	tmpDir := t.TempDir()
	dst := filepath.Join(tmpDir, "redis_7")

	runtime.GC()
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)

	if err := saveToTarFile("redis:7", dst, true); err != nil {
		t.Fatalf("saveToTarFile: %v", err)
	}

	runtime.GC()
	runtime.ReadMemStats(&after)

	fi, err := os.Stat(dst)
	if err != nil {
		t.Fatalf("output file missing: %v", err)
	}

	heapAllocMB := float64(after.TotalAlloc-before.TotalAlloc) / 1024 / 1024
	fileSizeMB := float64(fi.Size()) / 1024 / 1024

	t.Logf("Image size on disk:  %.1f MB", fileSizeMB)
	t.Logf("Heap allocated:      %.1f MB", heapAllocMB)
	t.Logf("Ratio (heap/disk):   %.2fx", heapAllocMB/fileSizeMB)

	// After the fix the heap allocation should be a small fraction of the image size.
	// Before the fix it will be >= image size (the entire image is buffered in []byte).
	if heapAllocMB >= fileSizeMB*0.5 {
		t.Errorf("EXCESSIVE MEMORY: allocated %.1f MB for a %.1f MB image — full image is being buffered in RAM (bug #17945)", heapAllocMB, fileSizeMB)
	}
}

// TestStreamedTarIsValidDockerFormat verifies that the output of streamImageFromDaemon
// (raw docker save output) is a valid Docker image tar that minikube's downstream
// consumers — tarball.LoadManifest and transferAndLoadImage — can read correctly.
// This matters because the temp-dir path skips go-containerregistry's tarball.Write
// and writes docker save output directly.
//
// Requires Docker daemon with redis:7 pulled locally.
func TestStreamedTarIsValidDockerFormat(t *testing.T) {
	if _, err := os.Stat("/var/run/docker.sock"); err != nil {
		if _, err2 := os.Stat(os.ExpandEnv("$HOME/.docker/run/docker.sock")); err2 != nil {
			t.Skip("Docker daemon not available")
		}
	}

	UseDaemon(true)
	UseRemote(false)
	t.Cleanup(func() {
		UseDaemon(true)
		UseRemote(true)
	})

	tmpDir := t.TempDir()
	dst := filepath.Join(tmpDir, "redis_7")

	if err := saveToTarFile("redis:7", dst, true); err != nil {
		t.Fatalf("saveToTarFile: %v", err)
	}

	manifest, err := tarball.LoadManifest(func() (io.ReadCloser, error) {
		return os.Open(dst)
	})
	if err != nil {
		t.Fatalf("output is not a valid Docker tar: %v", err)
	}
	if len(manifest) == 0 {
		t.Fatal("manifest is empty")
	}
	if len(manifest[0].RepoTags) == 0 {
		t.Fatal("manifest has no repo tags")
	}
	t.Logf("manifest repo tags: %v", manifest[0].RepoTags)
}

func BenchmarkSaveToTarFile_Daemon(b *testing.B) {
	if _, err := os.Stat("/var/run/docker.sock"); err != nil {
		if _, err2 := os.Stat(os.ExpandEnv("$HOME/.docker/run/docker.sock")); err2 != nil {
			b.Skip("Docker daemon not available")
		}
	}

	UseDaemon(true)
	UseRemote(false)
	b.Cleanup(func() {
		UseDaemon(true)
		UseRemote(true)
	})

	for b.Loop() {
		dst := filepath.Join(b.TempDir(), "redis_7")
		if err := saveToTarFile("redis:7", dst, true); err != nil {
			b.Fatalf("saveToTarFile: %v", err)
		}
	}
}
