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

package kic

import (
	"fmt"

	"k8s.io/minikube/pkg/drivers/kic/oci"
)

const (
	// DefaultNetwork is the Docker default bridge network named "bridge"
	// (https://docs.docker.com/network/bridge/#use-the-default-bridge-network)
	DefaultNetwork = "bridge"
	// DefaultPodCIDR is The CIDR to be used for pods inside the node.
	DefaultPodCIDR = "10.244.0.0/16"

	// Version is the current version of kic
	Version = "v0.0.7"
	// SHA of the kic base image
	baseImageSHA = "a6f288de0e5863cdeab711fa6bafa38ee7d8d285ca14216ecf84fcfb07c7d176"

	// OverlayImage is the cni plugin used for overlay image, created by kind.
	// CNI plugin image used for kic drivers created by kind.
	OverlayImage = "kindest/kindnetd:0.5.3"
)

var (
	// BaseImage is the base image is used to spin up kic containers. it uses same base-image as kind.
	BaseImage = fmt.Sprintf("gcr.io/k8s-minikube/kicbase:%s@sha256:%s", Version, baseImageSHA)
)

// Config is configuration for the kic driver used by registry
type Config struct {
	MachineName   string            // maps to the container name being created
	CPU           int               // Number of CPU cores assigned to the container
	Memory        int               // max memory in MB
	StorePath     string            // libmachine store path
	OCIBinary     string            // oci tool to use (docker, podman,...)
	ImageDigest   string            // image name with sha to use for the node
	Mounts        []oci.Mount       // mounts
	APIServerPort int               // kubernetes api server port inside the container
	PortMappings  []oci.PortMapping // container port mappings
	Envs          map[string]string // key,value of environment variables passed to the node
}
