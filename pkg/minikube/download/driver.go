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
	"github.com/golang/glog"
	"github.com/hashicorp/go-getter"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/out"
)

func driverWithChecksumURL(name string, v semver.Version) string {
	base := fmt.Sprintf("https://github.com/kubernetes/minikube/releases/download/v%s/%s", v, name)
	return fmt.Sprintf("%s?checksum=file:%s.sha256", base, base)
}

// Driver downloads an arbitrary driver
func Driver(name string, destination string, v semver.Version) error {
	out.T(out.FileDownload, "Downloading driver {{.driver}}:", out.V{"driver": name})

	tmpDst := destination + ".download"

	url := driverWithChecksumURL(name, v)
	client := &getter.Client{
		Src:     url,
		Dst:     tmpDst,
		Mode:    getter.ClientModeFile,
		Options: []getter.ClientOption{getter.WithProgress(DefaultProgressBar)},
	}

	glog.Infof("Downloading: %+v", client)
	if err := client.Get(); err != nil {
		return errors.Wrapf(err, "download failed: %s", url)
	}
	// Give downloaded drivers a baseline decent file permission
	err := os.Chmod(tmpDst, 0755)
	if err != nil {
		return err
	}
	return os.Rename(tmpDst, destination)
}
