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

package kubeconfig

import (
	"github.com/pkg/errors"
	"k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/klog/v2"
)

// UnsetCurrentContext unsets the current-context from minikube to "" on minikube stop
func UnsetCurrentContext(machineName string, configPath ...string) error {
	fPath := PathFromEnv()
	if configPath != nil {
		fPath = configPath[0]
	}
	kCfg, err := readOrNew(fPath)
	if err != nil {
		return errors.Wrap(err, "Error getting kubeconfig status")
	}

	// Unset current-context only if profile is the current-context
	if kCfg.CurrentContext == machineName {
		kCfg.CurrentContext = ""
		if err := writeToFile(kCfg, fPath); err != nil {
			return errors.Wrap(err, "writing kubeconfig")
		}
		return nil
	}

	return nil
}

// SetCurrentContext sets the kubectl's current-context
func SetCurrentContext(name string, configPath ...string) error {
	fPath := PathFromEnv()
	if configPath != nil {
		fPath = configPath[0]
	}
	kcfg, err := readOrNew(fPath)
	if err != nil {
		return errors.Wrap(err, "Error getting kubeconfig status")
	}
	kcfg.CurrentContext = name
	return writeToFile(kcfg, fPath)
}

// DeleteContext deletes the specified machine's kubeconfig context
func DeleteContext(machineName string, configPath ...string) error {
	fPath := PathFromEnv()
	if configPath != nil {
		fPath = configPath[0]
	}
	kcfg, err := readOrNew(fPath)
	if err != nil {
		return errors.Wrap(err, "Error getting kubeconfig status")
	}

	if kcfg == nil || api.IsConfigEmpty(kcfg) {
		klog.V(2).Info("kubeconfig is empty")
		return nil
	}

	delete(kcfg.Clusters, machineName)
	delete(kcfg.AuthInfos, machineName)
	delete(kcfg.Contexts, machineName)

	if kcfg.CurrentContext == machineName {
		kcfg.CurrentContext = ""
	}

	if err := writeToFile(kcfg, fPath); err != nil {
		return errors.Wrap(err, "writing kubeconfig")
	}
	return nil
}
