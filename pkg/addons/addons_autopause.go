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
	"strconv"

	"github.com/pkg/errors"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/kubeconfig"
	"k8s.io/minikube/pkg/minikube/mustload"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/sysinit"
)

// enableOrDisableAutoPause enables the service after the config was copied by generic enble
func enableOrDisableAutoPause(cc *config.ClusterConfig, name string, val string) error {
	enable, err := strconv.ParseBool(val)
	if err != nil {
		return errors.Wrapf(err, "parsing bool: %s", name)
	}
	out.Infof("auto-pause addon is an alpha feature and still in early development. Please file issues to help us make it better.")
	out.Infof("https://github.com/kubernetes/minikube/labels/co/auto-pause")

	co := mustload.Running(cc.Name)
	if enable {
		if err := sysinit.New(co.CP.Runner).EnableNow("auto-pause"); err != nil {
			klog.ErrorS(err, "failed to enable", "service", "auto-pause")
		}
	}

	port := co.CP.Port // api server port
	if enable {        // if enable then need to calculate the forwarded port
		port = constants.AutoPauseProxyPort
		if driver.NeedsPortForward(cc.Driver) {
			port, err = oci.ForwardedPort(cc.Driver, cc.Name, port)
			if err != nil {
				klog.ErrorS(err, "failed to get forwarded port for", "auto-pause port", port)
			}
		}
	}

	updated, err := kubeconfig.UpdateEndpoint(cc.Name, co.CP.Hostname, port, kubeconfig.PathFromEnv(), kubeconfig.NewExtension())
	if err != nil {
		klog.ErrorS(err, "failed to update kubeconfig", "auto-pause proxy endpoint")
		return err
	}
	if updated {
		klog.Infof("%s context has been updated to point to auto-pause proxy %s:%s", cc.Name, co.CP.Hostname, co.CP.Port)
	} else {
		klog.Info("no need to update kube-context for auto-pause proxy")
	}

	return nil
}
