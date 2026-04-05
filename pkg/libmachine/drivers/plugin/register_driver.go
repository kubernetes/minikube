/*
Copyright 2022 The Kubernetes Authors All rights reserved.

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

package plugin

import (
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"time"

	"k8s.io/minikube/pkg/libmachine/drivers"
	"k8s.io/minikube/pkg/libmachine/drivers/plugin/localbinary"
	"k8s.io/minikube/pkg/libmachine/drivers/rpcdriver"
	"k8s.io/minikube/pkg/libmachine/log"
	"k8s.io/minikube/pkg/libmachine/version"
)

var (
	heartbeatTimeout = 10 * time.Second
)

func RegisterDriver(d drivers.Driver) {
	if os.Getenv(localbinary.PluginEnvKey) != localbinary.PluginEnvVal {
		fmt.Fprintf(os.Stderr, `This is a Docker Machine plugin binary.
Plugin binaries are not intended to be invoked directly.
Please use this plugin through the main 'docker-machine' binary.
(API version: %d)
`, version.APIVersion)
		os.Exit(1)
	}

	log.SetDebug(true)
	_ = os.Setenv("MACHINE_DEBUG", "1")

	rpcd := rpcdriver.NewRPCServerDriver(d)
	if err := rpc.RegisterName(rpcdriver.RPCServiceNameV0, rpcd); err != nil {
		fmt.Fprintf(os.Stderr, "Error registering RPC name: %s\n", err)
	}
	if err := rpc.RegisterName(rpcdriver.RPCServiceNameV1, rpcd); err != nil {
		fmt.Fprintf(os.Stderr, "Error registering RPC name: %s\n", err)
	}
	rpc.HandleHTTP()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading RPC server: %s\n", err)
		os.Exit(1)
	}
	defer listener.Close()

	fmt.Println(listener.Addr())

	go func() {
		for {
			select {
			case <-rpcd.CloseCh:
				log.Debug("Closing plugin on server side")
				os.Exit(0)
			case <-rpcd.HeartbeatCh:
				continue
			case <-time.After(heartbeatTimeout):
				// TODO: Add heartbeat retry logic
				os.Exit(1)
			}
		}
	}()

	// http.Serve never returns
	_ = http.Serve(listener, nil)
}
