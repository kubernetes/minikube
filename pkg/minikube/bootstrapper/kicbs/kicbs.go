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

// bootstrapper for kic
package kicbs

import (
	"net"
	"time"

	"k8s.io/minikube/pkg/minikube/bootstrapper"
	"k8s.io/minikube/pkg/minikube/config"
)

func PullImages(config.KubernetesConfig) error {

	return nil
}
func StartCluster(config.KubernetesConfig) error {

	return nil
}
func UpdateCluster(config.MachineConfig) error {
	return nil
}
func DeleteCluster(config.KubernetesConfig) error {
	return nil
}
func WaitForCluster(config.KubernetesConfig, time.Duration) error {
	return nil
}
func LogCommands(bootstrapper.LogOptions) map[string]string {
	return nil
}
func SetupCerts(cfg config.KubernetesConfig) error {
	return nil
}
func GetKubeletStatus() (string, error) {
	return "", nil
}
func GetAPIServerStatus(net.IP, int) (string, error) {
	return "", nil
}
