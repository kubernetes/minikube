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
	"k8s.io/apiserver/pkg/util/flag"
	kubelet "k8s.io/kubernetes/cmd/kubelet/app"
	"k8s.io/kubernetes/cmd/kubelet/app/options"
	"k8s.io/minikube/pkg/util"
)

func (lk LocalkubeServer) NewKubeletServer() Server {
	return NewSimpleServer("kubelet", serverInterval, StartKubeletServer(lk), noop)
}

func StartKubeletServer(lk LocalkubeServer) func() error {
	config, err := options.NewKubeletServer()
	if err != nil {
		return func() error { return err }
	}
	dnsIP, err := util.GetDNSIP(lk.ServiceClusterIPRange.String())
	if err != nil {
		return func() error { return err }
	}

	// Master details
	config.KubeConfig = flag.NewStringFlag(util.DefaultKubeConfigPath)
	config.RequireKubeConfig = true

	// Set containerized based on the flag
	config.Containerized = lk.Containerized

	config.AllowPrivileged = true
	config.PodManifestPath = "/etc/kubernetes/manifests"

	// Networking
	config.ClusterDomain = lk.DNSDomain
	config.ClusterDNS = []string{dnsIP.String()}
	// For kubenet plugin.
	config.PodCIDR = "10.180.1.0/24"

	config.NodeIP = lk.NodeIP.String()

	if lk.NetworkPlugin != "" {
		config.NetworkPluginName = lk.NetworkPlugin
	}

	// Runtime
	if lk.ContainerRuntime != "" {
		config.ContainerRuntime = lk.ContainerRuntime
	}
	if lk.RemoteRuntimeEndpoint != "" {
		config.RemoteRuntimeEndpoint = lk.RemoteRuntimeEndpoint
	}
	if lk.RemoteImageEndpoint != "" {
		config.RemoteImageEndpoint = lk.RemoteImageEndpoint
	}
	lk.SetExtraConfigForComponent("kubelet", &config)

	// Use the host's resolver config
	if lk.Containerized {
		config.ResolverConfig = "/rootfs/etc/resolv.conf"
	} else {
		config.ResolverConfig = "/etc/resolv.conf"
	}

	return func() error {
		return kubelet.Run(config, nil)
	}
}
