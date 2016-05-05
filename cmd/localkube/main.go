package main

import (
	"fmt"
	"os"
	"os/signal"

	"k8s.io/minikube/pkg/localkube"
)

var (
	LK *localkube.LocalKube

	DNSDomain     = "cluster.local"
	ClusterDNSIP  = "10.1.30.3"
	DNSServerAddr = "172.17.0.1:1970"
)

func init() {
	if name := os.Getenv("DNS_DOMAIN"); len(name) != 0 {
		DNSDomain = name
	}

	if addr := os.Getenv("DNS_SERVER"); len(addr) != 0 {
		DNSServerAddr = addr
	}
}

func load() {
	LK = new(localkube.LocalKube)

	// setup etc
	etcd, err := localkube.NewEtcd(localkube.KubeEtcdClientURLs, localkube.KubeEtcdPeerURLs, "kubeetcd", localkube.KubeEtcdDataDirectory)
	if err != nil {
		panic(err)
	}
	LK.Add(etcd)

	// setup apiserver
	apiserver := localkube.NewAPIServer()
	LK.Add(apiserver)

	// setup controller-manager
	controllerManager := localkube.NewControllerManagerServer()
	LK.Add(controllerManager)

	// setup scheduler
	scheduler := localkube.NewSchedulerServer()
	LK.Add(scheduler)

	// setup kubelet (configured for weave proxy)
	kubelet := localkube.NewKubeletServer(DNSDomain, ClusterDNSIP)
	LK.Add(kubelet)

	// proxy
	proxy := localkube.NewProxyServer()
	LK.Add(proxy)

	dns, err := localkube.NewDNSServer(DNSDomain, ClusterDNSIP, DNSServerAddr, localkube.APIServerURL)
	if err != nil {
		panic(err)
	}
	LK.Add(dns)
}

func main() {
	// check for network

	// if first
	load()
	err := LK.Run(os.Args, os.Stderr)
	if err != nil {
		fmt.Printf("localkube errored: %v\n", err)
		os.Exit(1)
	}
	defer LK.StopAll()

	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt)

	<-interruptChan
	fmt.Printf("\nShutting down...\n")
}
