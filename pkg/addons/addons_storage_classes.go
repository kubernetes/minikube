/*
Copyright 2021 The Kubernetes Authors All rights reserved.

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

package addons

import (
	"fmt"
	"strconv"

	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/run"
	"k8s.io/minikube/pkg/minikube/storageclass"
)

const defaultStorageClassProvisioner = "standard"

// enableOrDisableStorageClasses enables or disables storage classes
func enableOrDisableStorageClasses(cc *config.ClusterConfig, name string, val string, options *run.CommandOptions) error {
	klog.Infof("enableOrDisableStorageClasses %s=%v on %q", name, val, cc.Name)
	enable, err := strconv.ParseBool(val)
	if err != nil {
		return fmt.Errorf("Error parsing boolean: %w", err)
	}

	class := defaultStorageClassProvisioner
	if name == "storage-provisioner-rancher" {
		class = "local-path"
	}

	api, err := machine.NewAPIClient(options)
	if err != nil {
		return fmt.Errorf("machine client: %w", err)
	}
	defer api.Close()

	pcp, err := config.ControlPlane(*cc)
	if err != nil || !config.IsPrimaryControlPlane(*cc, pcp) {
		return fmt.Errorf("get primary control-plane node: %w", err)
	}
	machineName := config.MachineName(*cc, pcp)
	if !machine.IsRunning(api, machineName) {
		klog.Warningf("%q is not running, writing %s=%v to disk and skipping enablement", machineName, name, val)
		return EnableOrDisableAddon(cc, name, val, options)
	}

	storagev1, err := storageclass.GetStoragev1(cc.Name)
	if err != nil {
		return fmt.Errorf("Error getting storagev1 interface %v : %w", err, err)
	}

	if enable {
		// Enable addon before marking it as default
		if err = EnableOrDisableAddon(cc, name, val, options); err != nil {
			return err
		}
		// Only StorageClass for 'name' should be marked as default
		err = storageclass.SetDefaultStorageClass(storagev1, class)
		if err != nil {
			return fmt.Errorf("Error making %s the default storage class: %w", class, err)
		}
	} else {
		// Unset the StorageClass as default
		err := storageclass.DisableDefaultStorageClass(storagev1, class)
		if err != nil {
			return fmt.Errorf("Error disabling %s as the default storage class: %w", class, err)
		}
		if err = EnableOrDisableAddon(cc, name, val, options); err != nil {
			return err
		}
	}

	return nil
}
