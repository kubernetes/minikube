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

package drvs

import (
	// Register all of the drvs we know of
	_ "k8s.io/minikube/pkg/minikube/registry/drvs/docker"
	_ "k8s.io/minikube/pkg/minikube/registry/drvs/hyperkit"
	_ "k8s.io/minikube/pkg/minikube/registry/drvs/hyperv"
	_ "k8s.io/minikube/pkg/minikube/registry/drvs/kvm2"
	_ "k8s.io/minikube/pkg/minikube/registry/drvs/none"
	_ "k8s.io/minikube/pkg/minikube/registry/drvs/parallels"
	_ "k8s.io/minikube/pkg/minikube/registry/drvs/podman"
	_ "k8s.io/minikube/pkg/minikube/registry/drvs/qemu2"
	_ "k8s.io/minikube/pkg/minikube/registry/drvs/ssh"
	_ "k8s.io/minikube/pkg/minikube/registry/drvs/vfkit"
	_ "k8s.io/minikube/pkg/minikube/registry/drvs/virtualbox"
	_ "k8s.io/minikube/pkg/minikube/registry/drvs/vmware"
)
