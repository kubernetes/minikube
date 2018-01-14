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
	"k8s.io/minikube/pkg/util"

	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/kubelet/qos"
	"k8s.io/kubernetes/pkg/proxy/apis/kubeproxyconfig"
)

var (
	MasqueradeBit = int32(14)
	OOMScoreAdj   = int32(qos.KubeProxyOOMScoreAdj)
)

func (lk LocalkubeServer) NewProxyServer() Server {
	return NewSimpleServer("proxy", serverInterval, StartProxyServer(lk), noop)
}

func StartProxyServer(lk LocalkubeServer) func() error {
	bindaddress := lk.APIServerAddress.String()
	if lk.APIServerInsecurePort != 0 {
		bindaddress = lk.APIServerInsecureAddress.String()
	}

	opts := kubeproxy.NewOptions()
	config := &kubeproxyconfig.KubeProxyConfiguration{
		OOMScoreAdj: &OOMScoreAdj,
		ClientConnection: kubeproxyconfig.ClientConnectionConfiguration{
			Burst:          10,
			QPS:            5,
			KubeConfigFile: util.DefaultKubeConfigPath,
		},
		ConfigSyncPeriod: v1.Duration{Duration: 15 * time.Minute},
		IPTables: kubeproxyconfig.KubeProxyIPTablesConfiguration{
			MasqueradeBit: &MasqueradeBit,
			SyncPeriod:    v1.Duration{Duration: 30 * time.Second},
			MinSyncPeriod: v1.Duration{Duration: 5 * time.Second},
		},
		BindAddress:  bindaddress,
		Mode:         kubeproxyconfig.ProxyModeIPTables,
		FeatureGates: lk.FeatureGates,
		// Disable the healthz check
		HealthzBindAddress: "0",
	}
	_, err := opts.ApplyDefaults(config)
	if err != nil {
		panic(err)
	}
	opts.SetConfig(config)

	lk.SetExtraConfigForComponent("proxy", &config)

	return func() error {
		return opts.Run()
	}
}
