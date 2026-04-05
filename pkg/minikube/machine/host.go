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

package machine

import (
	"fmt"

	"k8s.io/minikube/pkg/libmachine"
	"k8s.io/minikube/pkg/libmachine/host"
	"k8s.io/minikube/pkg/libmachine/state"

	"k8s.io/klog/v2"
)

// Status returns the status of a libmachine host
func Status(api libmachine.API, machineName string) (string, error) {
	exists, err := api.Exists(machineName)
	if err != nil {
		return "", fmt.Errorf("%s exists: %w", machineName, err)
	}
	if !exists {
		return state.None.String(), nil
	}

	hostInfo, err := api.Load(machineName)
	if err != nil {
		return "", fmt.Errorf("load: %w", err)
	}

	s, err := hostInfo.Driver.GetState()
	if err != nil {
		return "", fmt.Errorf("state: %w", err)
	}
	return s.String(), nil
}

// IsRunning asserts that a libmachine host is in state "Running"
func IsRunning(api libmachine.API, name string) bool {
	s, err := Status(api, name)
	if err != nil {
		klog.Warningf("host status for %q returned error: %v", name, err)
		return false
	}
	if s != state.Running.String() {
		klog.Warningf("%q host status: %s", name, s)
		return false
	}
	return true
}

// LoadHost returns a libmachine host by name
func LoadHost(api libmachine.API, machineName string) (*host.Host, error) {
	klog.Infof("Checking if %q exists ...", machineName)
	exists, err := api.Exists(machineName)
	if err != nil {
		return nil, fmt.Errorf("Error checking that machine exists: %s: %w", machineName, err)
	}
	if !exists {
		return nil, fmt.Errorf("machine %q does not exist", machineName)
	}

	h, err := api.Load(machineName)
	if err != nil {
		return nil, fmt.Errorf("loading machine %q: %w", machineName, err)
	}
	return h, nil
}
