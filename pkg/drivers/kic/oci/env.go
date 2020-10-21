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
package oci

import (
	"os"

	"k8s.io/minikube/pkg/minikube/constants"
)

var initialEnvs = make(map[string]string)

func init() {
	for _, env := range constants.DockerDaemonEnvs {
		if v, set := os.LookupEnv(env); set {
			initialEnvs[env] = v
		}
		exEnv := constants.MinikubeExistingPrefix + env
		if v, set := os.LookupEnv(exEnv); set {
			initialEnvs[exEnv] = v
		}
	}
}

// InitialEnv returns the value of the environment variable env before any environment changes made by minikube
func InitialEnv(env string) string {
	return initialEnvs[env]
}

// LookupInitialEnv returns the value of the environment variable env before any environment changes made by minikube
func LookupInitialEnv(env string) (string, bool) {
	v, set := initialEnvs[env]
	return v, set
}
