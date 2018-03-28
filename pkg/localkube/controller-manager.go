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

package localkube

import (
	controllerManager "k8s.io/kubernetes/cmd/kube-controller-manager/app"
	"k8s.io/kubernetes/cmd/kube-controller-manager/app/config"
	"k8s.io/kubernetes/cmd/kube-controller-manager/app/options"
	"k8s.io/minikube/pkg/util"
)

func (lk LocalkubeServer) NewControllerManagerServer() Server {
	return NewSimpleServer("controller-manager", serverInterval, StartControllerManagerServer(lk), noop)
}

func StartControllerManagerServer(lk LocalkubeServer) func() error {
	opts := options.NewKubeControllerManagerOptions()

	opts.Generic.Kubeconfig = util.DefaultKubeConfigPath

	// defaults from command
	opts.Generic.ComponentConfig.DeletingPodsQps = 0.1
	opts.Generic.ComponentConfig.DeletingPodsBurst = 10
	opts.Generic.ComponentConfig.NodeEvictionRate = 0.1

	opts.Generic.ComponentConfig.EnableProfiling = true
	opts.Generic.ComponentConfig.VolumeConfiguration.EnableHostPathProvisioning = true
	opts.Generic.ComponentConfig.VolumeConfiguration.EnableDynamicProvisioning = true
	opts.Generic.ComponentConfig.ServiceAccountKeyFile = lk.GetPrivateKeyCertPath()
	opts.Generic.ComponentConfig.RootCAFile = lk.GetCAPublicKeyCertPath()

	lk.SetExtraConfigForComponent("controller-manager", &opts)

	cfg := config.Config{}
	if err := opts.ApplyTo(&cfg); err != nil {
		panic(err)
	}
	return func() error {
		return controllerManager.Run(cfg.Complete())
	}
}
