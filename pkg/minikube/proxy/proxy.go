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

package proxy

import (
	"fmt"
	"io"
	"net"

	"github.com/golang/glog"
	kubeApi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/watch"
	"k8s.io/minikube/pkg/minikube/cluster"
)

// StartHost starts a host VM.
func StartProxy(ip string) error {
	//   kube api watch -- map ports being used in a set of exposed ports
	//     best way to unset ports? (can diff last two and then remove the diff set)
	//   when you see a port exposed/changed, map the port to localhost
	//see that port is used via kubectl watch

	watchInterface, err := cluster.WatchKubernetesServicesWithNamespace("")
	if err != nil {
		return err
	}
	portToQuit := make(map[int]chan bool)
	for {
		select {
		case event := <-watchInterface.ResultChan():
			//should I proxy everything or only after command?
			//I need to make a map of ports to go funcs so that I can kill them on delete
			if service, ok := event.Object.(*kubeApi.Service); ok {
				if event.Type == watch.Added {
					port, err := cluster.GetServicePortFromService(service)
					if err == nil {
						quit := make(chan bool)
						portToQuit[port] = quit
						go func(string, int, int, chan bool) {
							if err := proxyVMPortToLocalhost(ip, port, port, quit); err != nil {
								glog.Infof(err.Error())
							}
						}(ip, port, port, quit)
					} else {
						glog.Infof(err.Error())
					}
				}
				if event.Type == watch.Deleted {
					port, err := cluster.GetServicePortFromService(service)
					if err == nil {
						portToQuit[port] <- true //should check the key is in map
						delete(portToQuit, port)
					} else {
						glog.Infof(err.Error())
					}
				}

			}
		}
	}
	return nil
}

func forward(conn net.Conn, outputURL string) {
	client, err := net.Dial("tcp", outputURL)
	if err != nil {
		glog.Fatalf("Dial failed: %v", err)
	}
	glog.Infof("Connected to localhost %v\n", conn)
	go func() {
		defer client.Close()
		defer conn.Close()
		io.Copy(client, conn)
	}()
	go func() {
		defer client.Close()
		defer conn.Close()
		io.Copy(conn, client)
	}()
}

func proxyVMPortToLocalhost(ip string, inputPort, outputPort int, quit chan bool) error {

	fmt.Printf("Proxying Ports: localhost:%d -- %s:%d\n", inputPort, ip, outputPort)

	ln, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", inputPort))
	if err != nil {
		return err
	}

	// accept multiple connections on port
	for {
		select {
		case <-quit:
			fmt.Printf("Removing Proxy: localhost:%d -- %s:%d\n", inputPort, ip, outputPort)
			return nil
		default:
			conn, err := ln.Accept()
			if err != nil {
				glog.Errorf("Could not create connection... [MAKE THIS BETTER]")
			}
			// There is an issue here with this not quitting properly on delete/quit-chan
			// Not sure if I should pass in a quit chan here or what would be the best thing to do...
			go forward(conn, fmt.Sprintf("%s:%d", ip, outputPort))
		}
	}
}
