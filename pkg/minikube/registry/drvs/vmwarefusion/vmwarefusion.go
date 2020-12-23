// +build darwin

/*
Copyright 2018 The Kubernetes Authors All rights reserved.

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

// vmwarefusion contains a shell of the deprecated vmware vdriver
package vmwarefusion

import (
	"fmt"

	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/registry"
)

func init() {
	if err := registry.Register(registry.DriverDef{
		Name:     driver.VMwareFusion,
		Status:   status,
		Priority: registry.Obsolete,
	}); err != nil {
		panic(fmt.Sprintf("register: %v", err))
	}
}

func status() registry.State {
	return registry.State{
		Error: fmt.Errorf("The 'vmwarefusion' driver is no longer available"),
		Fix:   "Switch to the newer 'vmware' driver by using '--driver=vmware'. This may require first deleting your existing cluster",
		Doc:   "https://minikube.sigs.k8s.io/docs/drivers/vmware/",
	}
}
