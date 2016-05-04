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
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	apiserver "k8s.io/kubernetes/cmd/kube-apiserver/app"
	"k8s.io/kubernetes/cmd/kube-apiserver/app/options"
	etcdstorage "k8s.io/kubernetes/pkg/storage/etcd"
)

const (
	APIServerName = "apiserver"
	APIServerHost = "0.0.0.0"
	APIServerPort = 8080
)

var (
	APIServerURL   string
	ServiceIPRange = "10.1.30.0/24"
	APIServerStop  chan struct{}
)

func init() {
	APIServerURL = fmt.Sprintf("http://%s:%d", APIServerHost, APIServerPort)
	if ipRange := os.Getenv("SERVICE_IP_RANGE"); len(ipRange) != 0 {
		ServiceIPRange = ipRange
	}
}

func NewAPIServer() Server {
	return &SimpleServer{
		ComponentName: APIServerName,
		StartupFn:     StartAPIServer,
		ShutdownFn: func() {
			close(APIServerStop)
		},
	}
}

func StartAPIServer() {
	APIServerStop = make(chan struct{})
	config := options.NewAPIServer()

	// use host/port from vars
	config.InsecureBindAddress = net.ParseIP(APIServerHost)
	config.InsecurePort = APIServerPort

	// use localkube etcd
	config.EtcdConfig = etcdstorage.EtcdConfig{
		ServerList: KubeEtcdClientURLs,
	}

	// set Service IP range
	_, ipnet, err := net.ParseCIDR(ServiceIPRange)
	if err != nil {
		panic(err)
	}
	config.ServiceClusterIPRange = *ipnet

	// defaults from apiserver command
	config.EnableProfiling = true
	config.EnableWatchCache = true
	config.MinRequestTimeout = 1800

	fn := func() error {
		return apiserver.Run(config)
	}

	// start API server in it's own goroutine
	go until(fn, os.Stdout, APIServerName, 200*time.Millisecond, SchedulerStop)
}

// notFoundErr returns true if the passed error is an API server object not found error
func notFoundErr(err error) bool {
	if err == nil {
		return false
	}
	return strings.HasSuffix(err.Error(), "not found")
}
