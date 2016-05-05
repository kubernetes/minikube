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
	"os"
	"time"

	controllerManager "k8s.io/kubernetes/cmd/kube-controller-manager/app"
	"k8s.io/kubernetes/cmd/kube-controller-manager/app/options"
)

const (
	ControllerManagerName = "controller-manager"
)

var (
	CMStop chan struct{}
)

func NewControllerManagerServer() Server {
	return &SimpleServer{
		ComponentName: ControllerManagerName,
		StartupFn:     StartControllerManagerServer,
		ShutdownFn: func() {
			close(CMStop)
		},
	}
}

func StartControllerManagerServer() {
	CMStop = make(chan struct{})
	config := options.NewCMServer()

	// defaults from command
	config.DeletingPodsQps = 0.1
	config.DeletingPodsBurst = 10
	config.EnableProfiling = true

	fn := func() error {
		return controllerManager.Run(config)
	}

	// start controller manager in it's own goroutine
	go until(fn, os.Stdout, ControllerManagerName, 200*time.Millisecond, SchedulerStop)
}
