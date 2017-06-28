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
	kubeproxy "k8s.io/kubernetes/cmd/kube-proxy/app"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubernetes/pkg/apis/componentconfig"
	"k8s.io/kubernetes/pkg/kubelet/qos"
)

var (
	MasqueradeBit = int32(14)
	OOMScoreAdj   = int32(qos.KubeProxyOOMScoreAdj)
)

func (lk LocalkubeServer) NewProxyServer() Server {
	return NewSimpleServer("proxy", serverInterval, StartProxyServer(lk), noop)
}

func StartProxyServer(lk LocalkubeServer) func() error {
	config := &componentconfig.KubeProxyConfiguration{
		OOMScoreAdj: &OOMScoreAdj,
		IPTables: componentconfig.KubeProxyIPTablesConfiguration{
			MasqueradeBit: &MasqueradeBit,
		},
		BindAddress:  lk.APIServerInsecureAddress.String(),
		Mode:         componentconfig.ProxyModeIPTables,
		FeatureGates: lk.FeatureGates,
		// Disable the healthz check
		HealthzBindAddress: "0",
	}

	lk.SetExtraConfigForComponent("proxy", &config)

	return func() error {
		// Creating this config requires the API Server to be up, so do it in the start function itself.
		server, err := kubeproxy.NewProxyServer(config, false, runtime.NewScheme(), lk.GetAPIServerInsecureURL())
		if err != nil {
			panic(err)
		}
		return server.Run()
	}
}
