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

package cmd

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/spf13/cobra"

	"k8s.io/minikube/pkg/localkube"
)

var Server *localkube.LocalkubeServer

var StartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the localkube server.",
	Long:  `Start the localkube server.`,
	Run: func(command *cobra.Command, args []string) {

		// TODO: Require root

		SetupServer(Server)
		Server.StartAll()

		defer Server.StopAll()

		interruptChan := make(chan os.Signal, 1)
		signal.Notify(interruptChan, os.Interrupt)

		<-interruptChan
		fmt.Println("Shutting down...")
	},
}

func init() {
	Server = NewLocalkubeServer()
	RootCmd.AddCommand(StartCmd)
}

func SetupServer(s *localkube.LocalkubeServer) {
	// setup etcd
	etcd, err := s.NewEtcd(localkube.KubeEtcdClientURLs, localkube.KubeEtcdPeerURLs, "kubeetcd", s.GetEtcdDataDirectory())
	if err != nil {
		panic(err)
	}
	s.AddServer(etcd)

	// setup apiserver
	apiserver := s.NewAPIServer()
	s.AddServer(apiserver)

	// setup controller-manager
	controllerManager := s.NewControllerManagerServer()
	s.AddServer(controllerManager)

	// setup scheduler
	scheduler := s.NewSchedulerServer()
	s.AddServer(scheduler)

	// setup kubelet
	kubelet := s.NewKubeletServer()
	s.AddServer(kubelet)

	// setup proxy
	proxy := s.NewProxyServer()
	s.AddServer(proxy)

	// setup dns if we should
	if s.EnableDNS {

		dns, err := s.NewDNSServer(s.DNSDomain, s.DNSIP, s.GetAPIServerInsecureURL())
		if err != nil {
			panic(err)
		}
		s.AddServer(dns)
	}
}
