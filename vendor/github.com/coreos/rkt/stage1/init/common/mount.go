// Copyright 2015 The rkt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package common

import (
	"github.com/appc/spec/schema"
	"github.com/appc/spec/schema/types"
)

func isMPReadOnly(mountPoints []types.MountPoint, name types.ACName) bool {
	for _, mp := range mountPoints {
		if mp.Name == name {
			return mp.ReadOnly
		}
	}

	return false
}

// IsMountReadOnly returns if a mount should be readOnly.
// If the readOnly flag in the pod manifest is not nil, it overrides the
// readOnly flag in the image manifest.
func IsMountReadOnly(vol types.Volume, mountPoints []types.MountPoint) bool {
	if vol.ReadOnly != nil {
		return *vol.ReadOnly
	}

	return isMPReadOnly(mountPoints, vol.Name)
}

// GenerateMounts maps MountPoint paths to volumes, returning a list of Mounts.
func GenerateMounts(ra *schema.RuntimeApp, volumes map[types.ACName]types.Volume) []schema.Mount {
	app := ra.App

	mnts := make(map[string]schema.Mount)
	for _, m := range ra.Mounts {
		mnts[m.Path] = m
	}

	for _, mp := range app.MountPoints {
		// there's already an injected mount for this target path, skip
		if _, ok := mnts[mp.Path]; ok {
			continue
		}
		vol, ok := volumes[mp.Name]
		// there is no volume for this mount point, creating an "empty" volume
		// implicitly
		if !ok {
			defaultMode := "0755"
			defaultUID := 0
			defaultGID := 0
			emptyVol := types.Volume{
				Name: mp.Name,
				Kind: "empty",
				Mode: &defaultMode,
				UID:  &defaultUID,
				GID:  &defaultGID,
			}

			log.Printf("warning: no volume specified for mount point %q, implicitly creating an \"empty\" volume. This volume will be removed when the pod is garbage-collected.", mp.Name)

			volumes[mp.Name] = emptyVol
			ra.Mounts = append(ra.Mounts, schema.Mount{Volume: mp.Name, Path: mp.Path})
		} else {
			ra.Mounts = append(ra.Mounts, schema.Mount{Volume: vol.Name, Path: mp.Path})
		}
	}

	return ra.Mounts
}
