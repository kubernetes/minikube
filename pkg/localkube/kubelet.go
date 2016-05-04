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

	kubelet "k8s.io/kubernetes/cmd/kubelet/app"
	"k8s.io/kubernetes/cmd/kubelet/app/options"
)

const (
	KubeletName = "kubelet"
)

var (
	WeaveProxySock = "unix:///var/run/weave/weave.sock"
	KubeletStop    chan struct{}
)

func NewKubeletServer(clusterDomain, clusterDNS string, containerized bool) Server {
	return &SimpleServer{
		ComponentName: KubeletName,
		StartupFn:     StartKubeletServer(clusterDomain, clusterDNS, containerized),
		ShutdownFn: func() {
			close(KubeletStop)
		},
	}
}

func StartKubeletServer(clusterDomain, clusterDNS string, containerized bool) func() {
	KubeletStop = make(chan struct{})
	config := options.NewKubeletServer()

	// master details
	config.APIServerList = []string{APIServerURL}

	// Docker
	config.Containerized = containerized
	config.DockerEndpoint = WeaveProxySock

	// Networking
	config.ClusterDomain = clusterDomain
	config.ClusterDNS = clusterDNS

	// use hosts resolver config
	if containerized {
		config.ResolverConfig = "/rootfs/etc/resolv.conf"
	} else {
		config.ResolverConfig = "/etc/resolv.conf"
	}

	schedFn := func() error {
		return kubelet.Run(config, nil)
	}

	return func() {
		go until(schedFn, os.Stdout, KubeletName, 200*time.Millisecond, KubeletStop)
	}
}
