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

package common

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/diskfs/go-diskfs"
	"github.com/diskfs/go-diskfs/filesystem"
)

// ExtractFile extracts a file from an ISO.
// It first attempts an exact path match. If that fails, it performs a case-insensitive search
// to locate the file, handling inconsistencies in ISO naming conventions (e.g. uppercase vs lowercase).
func ExtractFile(isoPath, srcPath, destPath string) error {
	disk, err := diskfs.Open(isoPath, diskfs.WithOpenMode(diskfs.ReadOnly))
	if err != nil {
		return err
	}
	defer disk.Close()

	fs, err := disk.GetFilesystem(0)
	if err != nil {
		return err
	}

	f, err := fs.OpenFile(srcPath, os.O_RDONLY)
	if err != nil {
		// Fallback: case-insensitive search
		actualPath, err2 := findCaseInsensitivePath(fs, srcPath)
		if err2 != nil {
			return err
		}
		f, err = fs.OpenFile(actualPath, os.O_RDONLY)
		if err != nil {
			return err
		}
	}
	defer f.Close()

	dst, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, f)
	return err
}

// findCaseInsensitivePath searches for a file in the ISO filesystem in a case-insensitive manner.
// This is necessary because ISO 9660 file names are often stored in uppercase (or with other normalizations),
// while the requested path might be in lowercase.
func findCaseInsensitivePath(fs filesystem.FileSystem, targetPath string) (string, error) {
	parts := strings.Split(targetPath, "/")
	currentPath := "/"

	for _, part := range parts {
		if part == "" || part == "." {
			continue
		}

		infos, err := fs.ReadDir(currentPath)
		if err != nil {
			return "", err
		}

		found := false
		for _, info := range infos {
			name := info.Name()
			if strings.EqualFold(name, part) || strings.EqualFold(strings.TrimSuffix(name, "."), part) {
				currentPath = path.Join(currentPath, name)
				found = true
				break
			}
		}
		if !found {
			return "", fmt.Errorf("file not found: %s in %s", part, currentPath)
		}
	}
	return currentPath, nil
}
