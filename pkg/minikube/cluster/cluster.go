/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package cluster

import (
	"fmt"

	"k8s.io/minikube/pkg/libmachine"
	"k8s.io/minikube/pkg/libmachine/ssh"
	"github.com/pkg/errors"

	"k8s.io/minikube/pkg/minikube/bootstrapper"
	"k8s.io/minikube/pkg/minikube/bootstrapper/kubeadm"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/machine"
)

// This init function is used to set the logtostderr variable to false so that INFO level log info does not clutter the CLI
// INFO lvl logging is displayed due to the Kubernetes api calling flag.Set("logtostderr", "true") in its init()
// see: https://github.com/kubernetes/kubernetes/blob/master/pkg/kubectl/util/logs/logs.go#L32-L34
func init() {
	// Setting the default client to native gives much better performance.
	ssh.SetDefaultClient(ssh.Native)
}

// Bootstrapper returns a new bootstrapper for the cluster
func Bootstrapper(api libmachine.API, bootstrapperName string, cc config.ClusterConfig, r command.Runner) (bootstrapper.Bootstrapper, error) {
	var b bootstrapper.Bootstrapper
	var err error
	switch bootstrapperName {
	case bootstrapper.Kubeadm:
		b, err = kubeadm.NewBootstrapper(api, cc, r)
		if err != nil {
			return nil, errors.Wrap(err, "getting a new kubeadm bootstrapper")
		}
	default:
		return nil, fmt.Errorf("unknown bootstrapper: %s", bootstrapperName)
	}
	return b, nil
}

// ControlPlaneBootstrapper returns a bootstrapper for the first available cluster control-plane node.
func ControlPlaneBootstrapper(mAPI libmachine.API, cc *config.ClusterConfig, bootstrapperName string) (bootstrapper.Bootstrapper, error) {
	cp, err := config.ControlPlane(*cc)
	if err != nil {
		return nil, errors.Wrap(err, "get primary control-plane node")
	}
	h, err := machine.LoadHost(mAPI, config.MachineName(*cc, cp))
	if err != nil {
		return nil, errors.Wrap(err, "load primary control-plane host")
	}
	cpr, err := machine.CommandRunner(h)
	if err != nil {
		return nil, errors.Wrap(err, "get primary control-plane command runner")
	}

	bs, err := Bootstrapper(mAPI, bootstrapperName, *cc, cpr)
	return bs, err
}
