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

import "k8s.io/minikube/pkg/drivers/kic/oci"

const (
	// Docker default bridge network is named "bridge" (https://docs.docker.com/network/bridge/#use-the-default-bridge-network)
	DefaultNetwork = "bridge"
	// DefaultPodCIDR is The CIDR to be used for pods inside the node.
	DefaultPodCIDR = "10.244.0.0/16"
	// DefaultBindIPV4 is The default IP the container will bind to.
	DefaultBindIPV4 = "127.0.0.1"
	// BaseImage is the base image is used to spin up kic containers created by kind.
	BaseImage = "gcr.io/k8s-minikube/kicbase:v0.0.3@sha256:34db5e30f8830c0d5e49b62f3ea6b2844f805980592fe0084cbea799bfb12664" // OverlayImage is the cni plugin used for overlay image, created by kind.
	// CNI plugin image used for kic drivers created by kind.
	OverlayImage = "kindest/kindnetd:0.5.3"
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

type createParams struct {
	Name          string            // used for container name and hostname
	Image         string            // container image to use to create the node.
	ClusterLabel  string            // label the containers we create using minikube so we can clean up
	Role          string            // currently only role supported is control-plane
	Mounts        []oci.Mount       // volume mounts
	APIServerPort int               // kubernetes api server port
	PortMappings  []oci.PortMapping // ports to map to container from host
	CPUs          string            // number of cpu cores assign to container
	Memory        string            // memory (mbs) to assign to the container
	Envs          map[string]string // environment variables to pass to the container
	ExtraArgs     []string          // a list of any extra option to pass to oci binary during creation time, for example --expose 8080...
	OCIBinary     string            // docker or podman
}
