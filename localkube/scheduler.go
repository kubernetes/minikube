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
	scheduler "k8s.io/kubernetes/plugin/cmd/kube-scheduler/app"
	"k8s.io/kubernetes/plugin/cmd/kube-scheduler/app/options"
	"os"
	"time"
)

const (
	SchedulerName = "scheduler"
)

var (
	SchedulerStop chan struct{}
)

func NewSchedulerServer() Server {
	return &SimpleServer{
		ComponentName: SchedulerName,
		StartupFn:     StartSchedulerServer,
		ShutdownFn: func() {
			close(SchedulerStop)
		},
	}
}

func StartSchedulerServer() {
	SchedulerStop = make(chan struct{})
	config := options.NewSchedulerServer()

	// master details
	config.Master = APIServerURL

	// defaults from command
	config.EnableProfiling = true

	schedFn := func() error {
		return scheduler.Run(config)
	}

	go until(schedFn, os.Stdout, SchedulerName, 200*time.Millisecond, SchedulerStop)
}
