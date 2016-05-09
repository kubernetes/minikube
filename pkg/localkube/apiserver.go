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
	"net"
	"path/filepath"
	"strings"

	apiserver "k8s.io/kubernetes/cmd/kube-apiserver/app"
	"k8s.io/kubernetes/cmd/kube-apiserver/app/options"
	etcdstorage "k8s.io/kubernetes/pkg/storage/etcd"

	kuberest "k8s.io/kubernetes/pkg/client/restclient"
	kubeclient "k8s.io/kubernetes/pkg/client/unversioned"
)

const (
	certPath = "/srv/kubernetes/certs/"
)

func (lk LocalkubeServer) NewAPIServer() Server {
	return NewSimpleServer("apiserver", serverInterval, StartAPIServer(lk))
}

func StartAPIServer(lk LocalkubeServer) func() error {
	config := options.NewAPIServer()

	config.BindAddress = net.ParseIP(lk.APIServerAddress)
	config.SecurePort = lk.APIServerPort
	config.InsecureBindAddress = net.ParseIP(lk.APIServerInsecureAddress)
	config.InsecurePort = lk.APIServerInsecurePort

	config.ClientCAFile = filepath.Join(certPath, "ca.crt")
	config.TLSCertFile = filepath.Join(certPath, "kubernetes-master.crt")
	config.TLSPrivateKeyFile = filepath.Join(certPath, "kubernetes-master.key")
	config.AdmissionControl = "NamespaceLifecycle,LimitRanger,SecurityContextDeny,ServiceAccount,ResourceQuota"

	// use localkube etcd
	config.EtcdConfig = etcdstorage.EtcdConfig{
		ServerList: KubeEtcdClientURLs,
	}

	// set Service IP range
	_, ipnet, err := net.ParseCIDR(lk.ServiceClusterIPRange)
	if err != nil {
		panic(err)
	}
	config.ServiceClusterIPRange = *ipnet

	// defaults from apiserver command
	config.EnableProfiling = true
	config.EnableWatchCache = true
	config.MinRequestTimeout = 1800

	return func() error {
		return apiserver.Run(config)
	}
}

// notFoundErr returns true if the passed error is an API server object not found error
func notFoundErr(err error) bool {
	if err == nil {
		return false
	}
	return strings.HasSuffix(err.Error(), "not found")
}

func kubeClient() *kubeclient.Client {
	config := &kuberest.Config{
		Host: "http://localhost:8080", // TODO: Make configurable
	}
	client, err := kubeclient.New(config)
	if err != nil {
		panic(err)
	}
	return client
}
