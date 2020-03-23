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

package cmd

import (
	"fmt"

	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
)

// Return a minikube command containing the current profile name
func minikubeCmd() string {
	cname := ClusterFlagValue()
	if cname != constants.DefaultClusterName {
		return fmt.Sprintf("minikube -p %s", cname)
	}
	return "minikube"
}

// ClusterFlagValue returns the current cluster name based on flags
func ClusterFlagValue() string {
	return viper.GetString(config.ProfileName)
}
