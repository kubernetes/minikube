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

package download

import (
	"fmt"
	"io/fs"
	"os"
	"sync"
	"testing"
	"time"

	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/localpath"
)

// Force download tests to run in serial.
func TestDownload(t *testing.T) {
	t.Run("BinaryDownloadPreventsMultipleDownload", testBinaryDownloadPreventsMultipleDownload)
	t.Run("PreloadDownloadPreventsMultipleDownload", testPreloadDownloadPreventsMultipleDownload)
	t.Run("ImageToCache", testImageToCache)
	t.Run("PreloadNotExists", testPreloadNotExists)
	t.Run("PreloadExistsCaching", testPreloadExistsCaching)
	t.Run("PreloadWithCachedSizeZero", testPreloadWithCachedSizeZero)
}

// Returns a mock function that sleeps before incrementing `downloadsCounter` and creates the requested file.
func mockSleepDownload(downloadsCounter *int) func(src, dst string) error {
	return func(src, dst string) error {
		// Sleep for 200ms to assure locking must have occurred.
		time.Sleep(time.Millisecond * 200)
		*downloadsCounter++
		return CreateDstDownloadMock(src, dst)
	}
}

//	point each subtest at an isolated MINIKUBE_HOME, pre-create the preload cache directory,
//
// and automatically restore the global download/preload mocks after each run.
// Applied the helper across all download-related tests
func setupTestMiniHome(t *testing.T) {
	t.Helper()
	tmpHome := t.TempDir()
	t.Setenv(localpath.MinikubeHome, tmpHome)
	if err := os.MkdirAll(targetDir(), 0o755); err != nil {
		t.Fatalf("failed to create preload cache dir: %v", err)
	}
	origDownloadMock := DownloadMock
	origCheckCache := checkCache
	origCheckPreloadExists := checkPreloadExists
	origGetChecksumGCS := getChecksumGCS
	t.Cleanup(func() {
		DownloadMock = origDownloadMock
		checkCache = origCheckCache
		checkPreloadExists = origCheckPreloadExists
		getChecksumGCS = origGetChecksumGCS
	})
}

func testBinaryDownloadPreventsMultipleDownload(t *testing.T) {
	setupTestMiniHome(t)
	downloadNum := 0
	DownloadMock = mockSleepDownload(&downloadNum)

	checkCache = func(_ string) (fs.FileInfo, error) {
		if downloadNum == 0 {
			return nil, fmt.Errorf("some error")
		}
		return nil, nil
	}

	var group sync.WaitGroup
	group.Add(2)
	dlCall := func() {
		if _, err := Binary("kubectl", "v1.20.2", "linux", "amd64", ""); err != nil {
			t.Errorf("Failed to download binary: %+v", err)
		}
		group.Done()
	}

	go dlCall()
	go dlCall()

	group.Wait()

	if downloadNum != 1 {
		t.Errorf("Expected only 1 download attempt but got %v!", downloadNum)
	}
}

func testPreloadDownloadPreventsMultipleDownload(t *testing.T) {
	setupTestMiniHome(t)

	downloadNum := 0
	DownloadMock = mockSleepDownload(&downloadNum)
	f, err := os.CreateTemp("", "preload")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(f.Name())
	if _, err := f.Write([]byte("data")); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}

	checkCache = func(_ string) (fs.FileInfo, error) {
		if downloadNum == 0 {
			return nil, fmt.Errorf("some error")
		}
		return os.Stat(f.Name())
	}
	checkPreloadExists = func(_, _, _ string, _ ...bool) bool { return true }
	getChecksumGCS = func(_, _ string) ([]byte, error) { return []byte("check"), nil }

	var group sync.WaitGroup
	group.Add(2)
	dlCall := func() {
		if err := Preload(constants.DefaultKubernetesVersion, constants.Docker, "docker"); err != nil {
			t.Logf("Failed to download preload: %+v (may be ok)", err)
		}
		group.Done()
	}

	go dlCall()
	go dlCall()

	group.Wait()

	if downloadNum != 1 {
		t.Errorf("Expected only 1 download attempt but got %v!", downloadNum)
	}
}

func testPreloadNotExists(t *testing.T) {
	setupTestMiniHome(t)
	downloadNum := 0
	DownloadMock = mockSleepDownload(&downloadNum)

	checkCache = func(_ string) (fs.FileInfo, error) { return nil, fmt.Errorf("cache not found") }
	checkPreloadExists = func(_, _, _ string, _ ...bool) bool { return false }
	getChecksumGCS = func(_, _ string) ([]byte, error) { return []byte("check"), nil }

	err := Preload(constants.DefaultKubernetesVersion, constants.Docker, "docker")
	if err != nil {
		t.Errorf("Expected no error when preload exists")
	}

	if downloadNum != 0 {
		t.Errorf("Expected no download attempt but got %v!", downloadNum)
	}
}

func testImageToCache(t *testing.T) {
	setupTestMiniHome(t)
	downloadNum := 0
	DownloadMock = mockSleepDownload(&downloadNum)

	checkImageExistsInCache = func(_ string) bool { return downloadNum > 0 }

	var group sync.WaitGroup
	group.Add(2)
	dlCall := func() {
		if err := ImageToCache("testimg"); err != nil {
			t.Errorf("Failed to download preload: %+v", err)
		}
		group.Done()
	}

	go dlCall()
	go dlCall()

	group.Wait()

	if downloadNum != 1 {
		t.Errorf("Expected only 1 download attempt but got %v!", downloadNum)
	}
}

// Validates that preload existence checks correctly caches values retrieved by remote checks.
// testPreloadExistsCaching verifies the caching semantics of PreloadExists when
// the local cache is absent and remote existence checks are required.
// In summary, this test enforces that:
// - PreloadExists performs remote checks only on cache misses.
// - Negative and positive results are cached per (k8sVersion, containerVersion, runtime) key.
// - GitHub is only consulted when GCS reports the preload as not existing.
// - Global state is correctly restored after the test.
func testPreloadExistsCaching(t *testing.T) {
	setupTestMiniHome(t)
	checkCache = func(_ string) (fs.FileInfo, error) {
		return nil, fmt.Errorf("cache not found")
	}
	doesPreloadExist := false
	gcsCheckCalls := 0
	ghCheckCalls := 0
	savedGCSCheck := checkRemotePreloadExistsGCS
	savedGHCheck := checkRemotePreloadExistsGitHub
	preloadStates = make(map[string]map[string]preloadState)
	checkRemotePreloadExistsGCS = func(_, _ string) bool {
		gcsCheckCalls++
		return doesPreloadExist
	}
	checkRemotePreloadExistsGitHub = func(_, _ string) bool {
		ghCheckCalls++
		return false
	}
	t.Cleanup(func() {
		checkRemotePreloadExistsGCS = savedGCSCheck
		checkRemotePreloadExistsGitHub = savedGHCheck
		preloadStates = make(map[string]map[string]preloadState)
	})

	existence := PreloadExists("v1", "c1", "docker", true)
	if existence || gcsCheckCalls != 1 || ghCheckCalls != 1 {
		t.Errorf("Expected preload not to exist and checks to be performed. Existence: %v, GCS Calls: %d, GH Calls: %d", existence, gcsCheckCalls, ghCheckCalls)
	}
	gcsCheckCalls = 0
	ghCheckCalls = 0
	existence = PreloadExists("v1", "c1", "docker", true)
	if existence || gcsCheckCalls != 0 || ghCheckCalls != 0 {
		t.Errorf("Expected preload not to exist and no checks to be performed. Existence: %v, GCS Calls: %d, GH Calls: %d", existence, gcsCheckCalls, ghCheckCalls)
	}
	doesPreloadExist = true
	gcsCheckCalls = 0
	ghCheckCalls = 0
	existence = PreloadExists("v2", "c1", "docker", true)
	if !existence || gcsCheckCalls != 1 || ghCheckCalls != 0 {
		t.Errorf("Expected preload to exist via GCS. Existence: %v, GCS Calls: %d, GH Calls: %d", existence, gcsCheckCalls, ghCheckCalls)
	}
	gcsCheckCalls = 0
	ghCheckCalls = 0
	existence = PreloadExists("v2", "c2", "docker", true)
	if !existence || gcsCheckCalls != 1 || ghCheckCalls != 0 {
		t.Errorf("Expected preload to exist via GCS for new runtime. Existence: %v, GCS Calls: %d, GH Calls: %d", existence, gcsCheckCalls, ghCheckCalls)
	}
	gcsCheckCalls = 0
	ghCheckCalls = 0
	existence = PreloadExists("v2", "c2", "docker", true)
	if !existence || gcsCheckCalls != 0 || ghCheckCalls != 0 {
		t.Errorf("Expected preload to exist and no checks to be performed. Existence: %v, GCS Calls: %d, GH Calls: %d", existence, gcsCheckCalls, ghCheckCalls)
	}
}

func testPreloadWithCachedSizeZero(t *testing.T) {
	setupTestMiniHome(t)

	downloadNum := 0
	DownloadMock = mockSleepDownload(&downloadNum)
	f, err := os.CreateTemp("", "preload")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	checkCache = func(_ string) (fs.FileInfo, error) { return os.Stat(f.Name()) }
	checkPreloadExists = func(_, _, _ string, _ ...bool) bool { return true }
	getChecksumGCS = func(_, _ string) ([]byte, error) { return []byte("check"), nil }

	if err := Preload(constants.DefaultKubernetesVersion, constants.Docker, "docker"); err != nil {
		t.Errorf("Expected no error with cached preload of size zero")
	}

	if downloadNum != 1 {
		t.Errorf("Expected only 1 download attempt but got %v!", downloadNum)
	}
}
