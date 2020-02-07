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

package podman

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/docker/machine/libmachine/drivers"
	"k8s.io/minikube/pkg/drivers/kic"
	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/registry"
)

func init() {
	if err := registry.Register(registry.DriverDef{
		Name:     driver.Podman,
		Config:   configure,
		Init:     func() drivers.Driver { return kic.NewDriver(kic.Config{OCIBinary: oci.Podman}) },
		Status:   status,
		Priority: registry.Experimental,
	}); err != nil {
		panic(fmt.Sprintf("register failed: %v", err))
	}
}

func configure(mc config.MachineConfig) interface{} {
	return kic.NewDriver(kic.Config{
		MachineName:   mc.Name,
		StorePath:     localpath.MiniPath(),
		ImageDigest:   strings.Split(kic.BaseImage, "@")[0], // for podman does not support docker images references with both a tag and digest.
		CPU:           mc.CPUs,
		Memory:        mc.Memory,
		OCIBinary:     oci.Podman,
		APIServerPort: mc.Nodes[0].Port,
	})

}

func status() registry.State {
	_, err := exec.LookPath(oci.Podman)
	if err != nil {
		return registry.State{Error: err, Installed: false, Healthy: false, Fix: "Podman is required.", Doc: "https://minikube.sigs.k8s.io/docs/reference/drivers/podman/"}
	}

	_, err := exec.Cmd()
	// if err != nil {
	// 	return registry.State{Error: err, Installed: false, Healthy: false, Fix: "Podman is required.", Doc: "https://minikube.sigs.k8s.io/docs/reference/drivers/podman/"}
	// }

	// Allow no more than 2 seconds for querying state
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err = exec.CommandContext(ctx, oci.Podman, "info").Run()
	if err != nil {
		return registry.State{Error: err, Installed: true, Healthy: false, Fix: "Podman is not running or taking too long to respond. Try: restarting podman."}
	}

	return registry.State{Installed: true, Healthy: true}
}
