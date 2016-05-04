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

	kubeproxy "k8s.io/kubernetes/cmd/kube-proxy/app"
	"k8s.io/kubernetes/cmd/kube-proxy/app/options"
	"k8s.io/kubernetes/pkg/apis/componentconfig"
	"k8s.io/kubernetes/pkg/kubelet/qos"
)

const (
	ProxyName = "proxy"
)

var (
	MasqueradeBit = 14
	ProxyStop     chan struct{}
)

func NewProxyServer() Server {
	return &SimpleServer{
		ComponentName: ProxyName,
		StartupFn:     StartProxyServer,
		ShutdownFn: func() {
			close(ProxyStop)
		},
	}
}

func StartProxyServer() {
	ProxyStop = make(chan struct{})
	config := options.NewProxyConfig()

	// master details
	config.Master = APIServerURL

	// TODO: investigate why IP tables is not working
	config.Mode = componentconfig.ProxyModeUserspace

	// defaults
	oom := qos.KubeProxyOOMScoreAdj
	config.OOMScoreAdj = &oom
	config.IPTablesMasqueradeBit = &MasqueradeBit

	server, err := kubeproxy.NewProxyServerDefault(config)
	if err != nil {
		panic(err)
	}

	go until(server.Run, os.Stdout, ProxyName, 200*time.Millisecond, ProxyStop)
}
