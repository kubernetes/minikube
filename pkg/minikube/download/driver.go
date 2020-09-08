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
	"os"

	"github.com/blang/semver"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/style"
)

func driverWithChecksumURL(name string, v semver.Version) string {
	base := fmt.Sprintf("https://github.com/kubernetes/minikube/releases/download/v%s/%s", v, name)
	return fmt.Sprintf("%s?checksum=file:%s.sha256", base, base)
}

// Driver downloads an arbitrary driver
func Driver(name string, destination string, v semver.Version) error {
	out.T(style.FileDownload, "Downloading driver {{.driver}}:", out.V{"driver": name})
	if err := download(driverWithChecksumURL(name, v), destination); err != nil {
		return errors.Wrap(err, "download")
	}

	// Give downloaded drivers a baseline decent file permission
	return os.Chmod(destination, 0o755)
}
