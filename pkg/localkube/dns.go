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
	"time"

	"k8s.io/minikube/pkg/localkube/kube2sky"

	etcd "github.com/coreos/etcd/client"
	"github.com/golang/glog"
	backendetcd "github.com/skynetservices/skydns/backends/etcd"
	"github.com/skynetservices/skydns/metrics"
	skydns "github.com/skynetservices/skydns/server"
	kube "k8s.io/kubernetes/pkg/api"
	kubeclient "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/util/intstr"

	"golang.org/x/net/context"
	"k8s.io/minikube/pkg/util"
)

const (
	DNSName = "dns"

	DNSServiceName      = "kube-dns"
	DNSServiceNamespace = "kube-system"
)

type DNSServer struct {
	etcd          *EtcdServer
	sky           runner
	kube2sky      func() error
	dnsServerAddr *net.UDPAddr
	clusterIP     string
	done          chan struct{}
}

func (lk LocalkubeServer) NewDNSServer(rootDomain, clusterIP, kubeAPIServer string) (*DNSServer, error) {
	// setup backing etcd store
	peerURLs := []string{"http://localhost:9256"}
	DNSEtcdURLs := []string{"http://localhost:49090"}

	publicIP, err := lk.GetHostIP()
	if err != nil {
		return nil, err
	}

	serverAddress := fmt.Sprintf("%s:%d", publicIP.String(), 53)
	etcdServer, err := lk.NewEtcd(DNSEtcdURLs, peerURLs, DNSName, lk.GetDNSDataDirectory())
	if err != nil {
		return nil, err
	}

	// setup skydns
	cfg := etcd.Config{
		Endpoints: DNSEtcdURLs,
		Transport: etcd.DefaultTransport,
	}
	etcdClient, err := etcd.New(cfg)
	if err != nil {
		return nil, err
	}
	keysApi := etcd.NewKeysAPI(etcdClient)

	skyConfig := &skydns.Config{
		DnsAddr: serverAddress,
		Domain:  rootDomain,
	}

	dnsAddress, err := net.ResolveUDPAddr("udp", serverAddress)
	if err != nil {
		return nil, err
	}

	skydns.SetDefaults(skyConfig)

	backend := backendetcd.NewBackend(
		keysApi,
		context.Background(),
		&backendetcd.Config{
			Ttl:      skyConfig.Ttl,
			Priority: skyConfig.Priority,
		})
	skyServer := skydns.New(backend, skyConfig)

	// setup so prometheus doesn't run into nil
	metrics.Metrics()

	// setup kube2sky
	k2s := kube2sky.NewKube2Sky(rootDomain, DNSEtcdURLs[0], "", kubeAPIServer, 10*time.Second, 8081)

	return &DNSServer{
		etcd:          etcdServer,
		sky:           skyServer,
		kube2sky:      k2s,
		dnsServerAddr: dnsAddress,
		clusterIP:     clusterIP,
	}, nil
}

func (dns *DNSServer) Start() {
	if dns.done != nil {
		fmt.Fprint(os.Stderr, util.Pad("DNS server already started"))
		return
	}

	dns.done = make(chan struct{})

	dns.etcd.Start()
	go util.Until(dns.kube2sky, os.Stderr, "kube2sky", 2*time.Second, dns.done)
	go util.Until(dns.sky.Run, os.Stderr, "skydns", 1*time.Second, dns.done)

	go func() {
		var err error
		client := kubeClient()

		meta := kube.ObjectMeta{
			Name:      DNSServiceName,
			Namespace: DNSServiceNamespace,
			Labels: map[string]string{
				"k8s-app":            "kube-dns",
				"kubernetes.io/name": "KubeDNS",
			},
		}

		for {
			if err != nil {
				time.Sleep(2 * time.Second)
			}

			if _, err = client.Namespaces().Get(DNSServiceNamespace); notFoundErr(err) {
				err = createKubeSystemIfNotPresent(client)
				if err != nil {
					glog.Infof("Failed to create the kube-system namespace: %v\n", err)
					continue
				}
			} else if err != nil {
				glog.Infof("Failed to check for kube-system namespace existence: %v\n", err)
				continue
			}

			// setup service
			if _, err = client.Services(meta.Namespace).Get(meta.Name); notFoundErr(err) {
				// create service if doesn't exist
				err = createService(client, meta, dns.clusterIP, dns.dnsServerAddr.Port)
				if err != nil {
					glog.Infof("Failed to create Service for DNS: %v\n", err)
					continue
				}
			} else if err != nil {
				// error if cannot check for Service
				glog.Infof("Failed to check for DNS Service existence: %v\n", err)
				continue
			}

			// setup endpoint
			if _, err = client.Endpoints(meta.Namespace).Get(meta.Name); notFoundErr(err) {
				// create endpoint if doesn't exist
				err = createEndpoint(client, meta, dns.dnsServerAddr.IP.String(), int32(dns.dnsServerAddr.Port))
				if err != nil {
					glog.Infof("Failed to create Endpoint for DNS: %v\n", err)
					continue
				}
			} else if err != nil {
				// error if cannot check for Endpoint
				glog.Infof("Failed to check for DNS Endpoint existence: %v\n", err)
				continue
			}

			// setup successful
			glog.Infof("DNS started succesfully.")
			break
		}
	}()

}

func (dns *DNSServer) Stop() {
	teardownService()

	// closing chan will prevent servers from restarting but will not kill running server
	close(dns.done)

	dns.etcd.Stop()
}

// Name returns the servers unique name
func (DNSServer) Name() string {
	return DNSName
}

// runner starts a server returning an error if it stops.
type runner interface {
	Run() error
}

func createService(client *kubeclient.Client, meta kube.ObjectMeta, clusterIP string, dnsPort int) error {
	service := &kube.Service{
		ObjectMeta: meta,
		Spec: kube.ServiceSpec{
			ClusterIP: clusterIP,
			Ports: []kube.ServicePort{
				{
					Name:       "dns",
					Port:       53,
					TargetPort: intstr.FromInt(dnsPort),
					Protocol:   kube.ProtocolUDP,
				},
				{
					Name:       "dns-tcp",
					Port:       53,
					TargetPort: intstr.FromInt(dnsPort),
					Protocol:   kube.ProtocolTCP,
				},
			},
		},
	}

	_, err := client.Services(meta.Namespace).Create(service)
	if err != nil {
		return err
	}
	return nil
}

func createEndpoint(client *kubeclient.Client, meta kube.ObjectMeta, dnsIP string, dnsPort int32) error {
	endpoints := &kube.Endpoints{
		ObjectMeta: meta,
		Subsets: []kube.EndpointSubset{
			{
				Addresses: []kube.EndpointAddress{
					{IP: dnsIP},
				},
				Ports: []kube.EndpointPort{
					{
						Name: "dns",
						Port: dnsPort,
					},
					{
						Name: "dns-tcp",
						Port: dnsPort,
					},
				},
			},
		},
	}

	_, err := client.Endpoints(meta.Namespace).Create(endpoints)
	if err != nil {
		return err
	}
	return nil
}

func createKubeSystemIfNotPresent(client *kubeclient.Client) error {

	var ns = &kube.Namespace{
		ObjectMeta: kube.ObjectMeta{
			Name: DNSServiceNamespace,
		},
	}
	_, err := client.Namespaces().Create(ns)
	if err != nil {
		return err
	}
	return nil
}

func teardownService() {
	client := kubeClient()
	client.Services(DNSServiceNamespace).Delete(DNSServiceName)
	client.Endpoints(DNSServiceNamespace).Delete(DNSServiceName)
}
