/*
Copyright 2021 The Kubernetes Authors All rights reserved.

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

package delete

import (
	"context"
	"fmt"
	"os/exec"

	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/style"
)

// PossibleLeftOvers deletes KIC & non-KIC drivers left
func PossibleLeftOvers(ctx context.Context, cname string, driverName string) {
	bin := ""
	switch driverName {
	case driver.Docker:
		bin = oci.Docker
	case driver.Podman:
		bin = oci.Podman
	default:
		return
	}

	if _, err := exec.LookPath(bin); err != nil {
		klog.Infof("skipping deletePossibleLeftOvers for %s: %v", bin, err)
		return
	}

	klog.Infof("deleting possible leftovers for %s (driver=%s) ...", cname, driverName)
	delLabel := fmt.Sprintf("%s=%s", oci.ProfileLabelKey, cname)
	cs, err := oci.ListContainersByLabel(ctx, bin, delLabel)
	if err == nil && len(cs) > 0 {
		for _, c := range cs {
			out.Step(style.DeletingHost, `Deleting container "{{.name}}" ...`, out.V{"name": cname})
			err := oci.DeleteContainer(ctx, bin, c)
			if err != nil { // it will error if there is no container to delete
				klog.Errorf("error deleting container %q. You may want to delete it manually :\n%v", cname, err)
			}

		}
	}

	errs := oci.DeleteAllVolumesByLabel(ctx, bin, delLabel)
	if errs != nil { // it will not error if there is nothing to delete
		klog.Warningf("error deleting volumes (might be okay).\nTo see the list of volumes run: 'docker volume ls'\n:%v", errs)
	}

	errs = oci.DeleteKICNetworks(bin)
	if errs != nil {
		klog.Warningf("error deleting leftover networks (might be okay).\nTo see the list of networks: 'docker network ls'\n:%v", errs)
	}

	if bin == oci.Podman {
		// podman prune does not support --filter
		return
	}

	errs = oci.PruneAllVolumesByLabel(ctx, bin, delLabel)
	if len(errs) > 0 { // it will not error if there is nothing to delete
		klog.Warningf("error pruning volume (might be okay):\n%v", errs)
	}
}
