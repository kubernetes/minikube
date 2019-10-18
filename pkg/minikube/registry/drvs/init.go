package drvs

import (
	// Register all of the drvs we know of
	_ "k8s.io/minikube/pkg/minikube/registry/drvs/hyperkit"
	_ "k8s.io/minikube/pkg/minikube/registry/drvs/hyperv"
	_ "k8s.io/minikube/pkg/minikube/registry/drvs/kvm2"
	_ "k8s.io/minikube/pkg/minikube/registry/drvs/none"
	_ "k8s.io/minikube/pkg/minikube/registry/drvs/parallels"
	_ "k8s.io/minikube/pkg/minikube/registry/drvs/virtualbox"
	_ "k8s.io/minikube/pkg/minikube/registry/drvs/vmware"
	_ "k8s.io/minikube/pkg/minikube/registry/drvs/vmwarefusion"
)
