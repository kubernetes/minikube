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

	"github.com/docker/machine/libmachine/drivers"
	"github.com/google/uuid"
)

// VirtiofsShare is a directory on the host shared with the guest using
// virtiofs.
type VirtiofsShare struct {
	// HostPath is an absolute path to existing directory to share with the guest.
	HostPath string
	// GuestPath is a path in the guest for mounting the shared directory.
	GuestPath string
	// MountTag is a string idenfifying the shared file system in the guest.
	MountTag string
}

// ValidateVirtiofsShares parses and validates --virtiofs-share arguments and
// returns list of validated VirtiofsShare.
func ValidateVirtiofsShares(args []string) ([]*VirtiofsShare, error) {
	var shares []*VirtiofsShare

	seenHost := map[string]*VirtiofsShare{}
	seenGuest := map[string]*VirtiofsShare{}

	for _, s := range args {
		share, err := ParseVirtiofsShare(s)
		if err != nil {
			return nil, err
		}

		if err := share.Validate(); err != nil {
			return nil, err
		}

		if existing, ok := seenHost[share.HostPath]; ok {
			return nil, fmt.Errorf("host path %q is already shared at guest path %q", share.HostPath, existing.GuestPath)
		}
		seenHost[share.HostPath] = share

		if existing, ok := seenGuest[share.GuestPath]; ok {
			return nil, fmt.Errorf("guest path %q is already shared from host path %q", share.GuestPath, existing.HostPath)
		}
		seenGuest[share.GuestPath] = share

		shares = append(shares, share)
	}

	return shares, nil
}

// ParseVirtiofsShare parses a string in the format "/host-path:/guest-path"
// and returns a new VirtiofsShare instance.
func ParseVirtiofsShare(s string) (*VirtiofsShare, error) {
	pair := strings.SplitN(s, ":", 2)
	if len(pair) != 2 {
		return nil, fmt.Errorf("invalid virtiofsShare %q: (expected '/host-path:/guest-path')", s)
	}

	return &VirtiofsShare{
		HostPath:  pair[0],
		GuestPath: pair[1],
		MountTag:  uuid.NewString(),
	}, nil
}

// Validate a virtiofs share.
func (s *VirtiofsShare) Validate() error {
	// "," is a --device configuration separator in vfkit and krunkit.
	if strings.Contains(s.HostPath, ",") {
		return fmt.Errorf("host path %q must not contain ','", s.HostPath)
	}

	if !filepath.IsAbs(s.HostPath) {
		return fmt.Errorf("host path %q is not absolute path", s.HostPath)
	}

	if fs, err := os.Stat(s.HostPath); err != nil {
		return fmt.Errorf("failed to validate host path %q: %w", s.HostPath, err)
	} else if !fs.IsDir() {
		return fmt.Errorf("host path %q is not a directory", s.HostPath)
	}

	if !filepath.IsAbs(s.GuestPath) {
		return fmt.Errorf("guest path %q is not absolute path", s.GuestPath)
	}

	return nil
}

func MountVirtiofsShares(d drivers.Driver, shares []*VirtiofsShare) error {
	var script strings.Builder

	script.WriteString("set -e\n")

	for _, share := range shares {
		script.WriteString(fmt.Sprintf("sudo mkdir -p \"%s\"\n", share.GuestPath))
		script.WriteString(fmt.Sprintf("sudo mount -t virtiofs %s \"%s\"\n", share.MountTag, share.GuestPath))
	}

	if _, err := drivers.RunSSHCommandFromDriver(d, script.String()); err != nil {
		return err
	}

	return nil
}
