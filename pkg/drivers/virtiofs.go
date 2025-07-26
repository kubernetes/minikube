/*
Copyright 2025 The Kubernetes Authors All rights reserved.

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

package drivers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// VirtiofsShare is a directory on the host shared with the guest using
// virtiofs.
type VirtiofsShare struct {
	SharedDir string
	MountTag  string
}

// ParseVirtiofsShare parses --virtiofs-share argument and returns a validated VirtiofsShare.
// - SharedDir is an absolute path to existing directory to share with the guest.
// - MountTag is a non-empty string without leading or trailing whitespace.
func ParseVirtiofsShare(s string) (*VirtiofsShare, error) {
	pair := strings.SplitN(s, ":", 2)
	if len(pair) != 2 {
		return nil, fmt.Errorf("invalid virtiofsShare string %q: (expected '/shared-dir:mount-tag')", s)
	}

	sharedDir := pair[0]
	mountTag := pair[1]

	if !filepath.IsAbs(sharedDir) {
		return nil, fmt.Errorf("shared directory %q is not absolute path", sharedDir)
	}

	if fs, err := os.Stat(sharedDir); err != nil {
		return nil, fmt.Errorf("failed to validate shared directory %q: %w", sharedDir, err)
	} else if !fs.IsDir() {
		return nil, fmt.Errorf("shared directory %q is not a directory", sharedDir)
	}

	if mountTag == "" || mountTag != strings.TrimSpace(mountTag) {
		return nil, fmt.Errorf("invalid mount tag: %q", mountTag)
	}

	return &VirtiofsShare{SharedDir: sharedDir, MountTag: mountTag}, nil
}
