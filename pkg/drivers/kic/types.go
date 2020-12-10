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
	// Version is the current version of kic
	Version = "v0.0.15-snapshot4"
	// SHA of the kic base image
	baseImageSHA = "ef1f485b5a1cfa4c989bc05e153f0a8525968ec999e242efff871cbb31649c16"
)

var (
	// BaseImage is the base image is used to spin up kic containers. it uses same base-image as kind.
	BaseImage = fmt.Sprintf("gcr.io/k8s-minikube/kicbase:%s@sha256:%s", Version, baseImageSHA)

	// FallbackImages are backup base images in case gcr isn't available
	FallbackImages = []string{
		// the fallback of BaseImage in case gcr.io is not available. stored in docker hub
		// same image is push to https://github.com/kicbase/stable
		fmt.Sprintf("kicbase/stable:%s@sha256:%s", Version, baseImageSHA),

		// the fallback of BaseImage in case gcr.io is not available. stored in github packages https://github.com/kubernetes/minikube/packages/206071
		// github packages docker does _NOT_ support pulling by sha as mentioned in the docs:
		// https://help.github.com/en/packages/using-github-packages-with-your-projects-ecosystem/configuring-docker-for-use-with-github-packages
		fmt.Sprintf("docker.pkg.github.com/kubernetes/minikube/kicbase:%s", Version),
	}
)

// Config is configuration for the kic driver used by registry
type Config struct {
	ClusterName       string            // The cluster the container belongs to
	MachineName       string            // maps to the container name being created
	CPU               int               // Number of CPU cores assigned to the container
	Memory            int               // max memory in MB
	StorePath         string            // libmachine store path
	OCIBinary         string            // oci tool to use (docker, podman,...)
	ImageDigest       string            // image name with sha to use for the node
	Mounts            []oci.Mount       // mounts
	APIServerPort     int               // Kubernetes api server port inside the container
	PortMappings      []oci.PortMapping // container port mappings
	Envs              map[string]string // key,value of environment variables passed to the node
	KubernetesVersion string            // Kubernetes version to install
	ContainerRuntime  string            // container runtime kic is running
	ExtraArgs         []string          // a list of any extra option to pass to oci binary during creation time, for example --expose 8080...
}
