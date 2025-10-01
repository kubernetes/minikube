/*
Copyright 2022 The Kubernetes Authors All rights reserved.

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

package mcndockerclient

import (
	"fmt"

	"k8s.io/minikube/pkg/libmachine/auth"
)

type URLer interface {
	// URL returns the Docker host URL
	URL() (string, error)
}

type AuthOptionser interface {
	// AuthOptions returns the authOptions
	AuthOptions() *auth.Options
}

type DockerHost interface {
	URLer
	AuthOptionser
}

type RemoteDocker struct {
	HostURL    string
	AuthOption *auth.Options
}

// URL returns the Docker host URL
func (rd *RemoteDocker) URL() (string, error) {
	if rd.HostURL == "" {
		return "", fmt.Errorf("Docker Host URL not set")
	}

	return rd.HostURL, nil
}

// AuthOptions returns the authOptions
func (rd *RemoteDocker) AuthOptions() *auth.Options {
	return rd.AuthOption
}
