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
	"k8s.io/kubernetes/pkg/apis/componentconfig"
	scheduler "k8s.io/kubernetes/plugin/cmd/kube-scheduler/app"
	"k8s.io/minikube/pkg/util"
)

func (lk LocalkubeServer) NewSchedulerServer() Server {
	return NewSimpleServer("scheduler", serverInterval, StartSchedulerServer(lk), noop)
}

func StartSchedulerServer(lk LocalkubeServer) func() error {
	config := &componentconfig.KubeSchedulerConfiguration{}
	opts, err := scheduler.NewOptions()
	if err != nil {
		panic(err)
	}
	config, err = opts.ApplyDefaults(config)
	if err != nil {
		panic(err)
	}

	// master details
	config.ClientConnection.KubeConfigFile = util.DefaultKubeConfigPath

	// defaults from command
	config.EnableProfiling = true

	lk.SetExtraConfigForComponent("scheduler", &config)

	return func() error {
		s, err := scheduler.NewSchedulerServer(config, "")
		if err != nil {
			return err
		}
		return s.Run(nil)
	}
}
