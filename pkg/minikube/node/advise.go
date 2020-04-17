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

package node

import (
	"runtime"

	"github.com/golang/glog"
	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/minikube/bootstrapper/bsutil/kverify"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/out"
)

// adviseNodePressure will advise the user what to do with difference pressure errors based on their environment
func adviseNodePressure(err error, name string, drv string) {
	if diskErr, ok := err.(*kverify.ErrDiskPressure); ok {
		out.ErrLn("")
		glog.Warning(diskErr)
		out.WarningT("The node {{.name}} has ran out of disk space.", out.V{"name": name})
		// generic advice for all drivers
		out.T(out.Tip, "Please free up disk or prune images.")
		if driver.IsVM(drv) {
			out.T(out.Stopped, "Consider creating a cluster with bigger disk size: `minikube start --disk SIZE_MB` ")
		} else if drv == oci.Docker && runtime.GOOS != "linux" {
			out.T(out.Stopped, "Consider increasing Docker Desktop's disk size.")
			if runtime.GOOS == "darwin" {
				out.T(out.Documentation, "Documentation: {{.url}}", out.V{"url": "https://docs.docker.com/docker-for-mac/space/"})
			}
			if runtime.GOOS == "windows" {
				out.T(out.Documentation, "Documentation: {{.url}}", out.V{"url": "https://docs.docker.com/docker-for-windows/"})
			}
		}
		out.ErrLn("")
	}

	if memErr, ok := err.(*kverify.ErrMemoryPressure); ok {
		out.ErrLn("")
		glog.Warning(memErr)
		out.WarningT("The node {{.name}} has ran out of memory.", out.V{"name": name})
		out.T(out.Tip, "Please free up memory on the cluster.")
		if driver.IsVM(drv) {
			out.T(out.Stopped, "Consider creating a cluster with larger memory size using `minikube start --memory SIZE_MB` ")
		} else if drv == oci.Docker && runtime.GOOS != "linux" {
			out.T(out.Stopped, "Consider increasing Docker Desktop's memory size.")
			if runtime.GOOS == "darwin" {
				out.T(out.Documentation, "Documentation: {{.url}}", out.V{"url": "https://docs.docker.com/docker-for-mac/space/"})
			}
			if runtime.GOOS == "windows" {
				out.T(out.Documentation, "Documentation: {{.url}}", out.V{"url": "https://docs.docker.com/docker-for-windows/"})
			}
		}
		out.ErrLn("")
	}

	if pidErr, ok := err.(*kverify.ErrPIDPressure); ok {
		glog.Warning(pidErr)
		out.ErrLn("")
		out.WarningT("The node {{.name}} has ran out of available PIDs.", out.V{"name": name})
		out.ErrLn("")
	}

	if netErr, ok := err.(*kverify.ErrNetworkNotReady); ok {
		glog.Warning(netErr)
		out.ErrLn("")
		out.WarningT("The node {{.name}} network is not available. Please verify network settings.", out.V{"name": name})
		out.ErrLn("")
	}
}