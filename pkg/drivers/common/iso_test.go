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

package common

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/diskfs/go-diskfs"
	"github.com/diskfs/go-diskfs/disk"
	"github.com/diskfs/go-diskfs/filesystem"
	"github.com/diskfs/go-diskfs/filesystem/iso9660"
)

// TestExtractFile verifies that we can extract files from an ISO 9660 image.
// It dynamically creates an ISO with uppercase filenames to explicitly test the case-insensitive lookup logic.
// This is critical because some ISOs (like boot2docker) store files in uppercase (e.g., BZIMAGE),
// but the driver code requests them in lowercase (e.g., bzimage).
func TestExtractFile(t *testing.T) {
	// Create a temporary ISO file
	tmpIso, err := os.CreateTemp("", "test*.iso")
	if err != nil {
		t.Fatalf("failed to create temp iso: %v", err)
	}
	isoPath := tmpIso.Name()
	tmpIso.Close()
	os.Remove(isoPath) // diskfs.Create requires the file to not exist or we handle it differently?
	// actually let's just use the name and remove it.
	defer os.Remove(isoPath)

	// Create 10MB disk image
	size := int64(10 * 1024 * 1024)
	d, err := diskfs.Create(isoPath, size, 2048)
	if err != nil {
		t.Fatalf("failed to create disk: %v", err)
	}
	defer d.Close()

	// Create ISO filesystem
	fs, err := d.CreateFilesystem(disk.FilesystemSpec{
		Partition: 0,
		FSType:    filesystem.TypeISO9660,
	})
	if err != nil {
		t.Fatalf("failed to create filesystem: %v", err)
	}

	// Add files to ISO
	// 1. Root file with uppercase name (to test case insensitivity)
	rw, err := fs.OpenFile("/TEST1.TXT", os.O_CREATE|os.O_RDWR)
	if err != nil {
		t.Fatalf("failed to create TEST1.TXT: %v", err)
	}

	if _, err := rw.Write([]byte("content1")); err != nil {
		t.Fatalf("failed to write to TEST1.TXT: %v", err)
	}
	if err := rw.Close(); err != nil {
		t.Fatalf("failed to close TEST1.TXT: %v", err)
	}

	// 2. Nested directory and file
	if err := fs.Mkdir("/TEST2"); err != nil {
		t.Fatalf("failed to mkdir TEST2: %v", err)
	}
	rw2, err := fs.OpenFile("/TEST2/TEST2.TXT", os.O_CREATE|os.O_RDWR)
	if err != nil {
		t.Fatalf("failed to create TEST2/TEST2.TXT: %v", err)
	}

	if _, err := rw2.Write([]byte("content2")); err != nil {
		t.Fatalf("failed to write to TEST2.TXT: %v", err)
	}
	if err := rw2.Close(); err != nil {
		t.Fatalf("failed to close TEST2/TEST2.TXT: %v", err)
	}

	// Finalize ISO
	isoFs, ok := fs.(*iso9660.FileSystem)
	if !ok {
		t.Fatalf("fs is not iso9660")
	}
	if err := isoFs.Finalize(iso9660.FinalizeOptions{}); err != nil {
		t.Fatalf("failed to finalize iso: %v", err)
	}

	testDir := t.TempDir()

	tests := []struct {
		name          string
		isoPath       string
		srcPath       string
		destPath      string
		expectedError bool
		checkContent  string // content to expect in dest file if success
	}{
		{
			name:          "case insensitive root file (test1.txt -> TEST1.TXT)",
			isoPath:       isoPath,
			srcPath:       "/test1.txt",
			destPath:      filepath.Join(testDir, "extracted1.txt"),
			expectedError: false,
			checkContent:  "content1",
		},
		{
			name:          "case insensitive nested file (test2/test2.txt -> TEST2/TEST2.TXT)",
			isoPath:       isoPath,
			srcPath:       "/test2/test2.txt",
			destPath:      filepath.Join(testDir, "extracted2.txt"),
			expectedError: false,
			checkContent:  "content2",
		},
		{
			name:          "exact match (TEST1.TXT)",
			isoPath:       isoPath,
			srcPath:       "/TEST1.TXT",
			destPath:      filepath.Join(testDir, "extracted3.txt"),
			expectedError: false,
			checkContent:  "content1",
		},
		{
			name:          "missing file",
			isoPath:       isoPath,
			srcPath:       "/missing.txt",
			destPath:      filepath.Join(testDir, "missing.txt"),
			expectedError: true,
		},
		{
			name:          "iso path error",
			isoPath:       "/nonexistent.iso",
			srcPath:       "/test1.txt",
			destPath:      filepath.Join(testDir, "fail.txt"),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ExtractFile(tt.isoPath, tt.srcPath, tt.destPath)
			if (err != nil) != tt.expectedError {
				t.Errorf("expectedError = %v, get = %v", tt.expectedError, err)
				return
			}
			if !tt.expectedError && tt.checkContent != "" {
				content, err := os.ReadFile(tt.destPath)
				if err != nil {
					t.Errorf("failed to read dest file: %v", err)
				}
				if string(content) != tt.checkContent {
					t.Errorf("content mismatch: expected %q, got %q", tt.checkContent, string(content))
				}
			}
		})
	}
}
