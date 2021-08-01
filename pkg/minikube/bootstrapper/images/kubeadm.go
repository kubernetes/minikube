/*
Copyright 2019 The Kubernetes Authors All rights reserved.

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

package images

import (
	"fmt"
	"strings"

	"github.com/blang/semver/v4"
	"github.com/pkg/errors"
)

// Kubeadm returns a list of images necessary to bootstrap kubeadm
func Kubeadm(mirror string, version string) ([]string, error) {
	v, err := semver.Make(strings.TrimPrefix(version, "v"))
	if err != nil {
		return nil, errors.Wrap(err, "semver")
	}
	if v.Major > 1 {
		return nil, fmt.Errorf("version too new: %v", v)
	}
	if semver.MustParseRange("<1.12.0-alpha.0")(v) {
		return nil, fmt.Errorf("version too old: %v", v)
	}
	imgs := essentials(mirror, v)
	imgs = append(imgs, auxiliary(mirror)...)
	return imgs, nil
}
