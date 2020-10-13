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
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/host"
	"github.com/docker/machine/libmachine/state"
	"github.com/pkg/errors"

	"k8s.io/klog/v2"
)

// Status returns the status of a libmachine host
func Status(api libmachine.API, machineName string) (string, error) {
	exists, err := api.Exists(machineName)
	if err != nil {
		return "", errors.Wrapf(err, "%s exists", machineName)
	}
	if !exists {
		return state.None.String(), nil
	}

	host, err := api.Load(machineName)
	if err != nil {
		return "", errors.Wrapf(err, "load")
	}

	s, err := host.Driver.GetState()
	if err != nil {
		return "", errors.Wrap(err, "state")
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
		return nil, errors.Wrapf(err, "Error checking that machine exists: %s", machineName)
	}
	if !exists {
		return nil, errors.Errorf("machine %q does not exist", machineName)
	}

	h, err := api.Load(machineName)
	if err != nil {
		return nil, errors.Wrapf(err, "loading machine %q", machineName)
	}
	return h, nil
}
