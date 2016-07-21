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
	"strings"

	apiserver "k8s.io/kubernetes/cmd/kube-apiserver/app"
	"k8s.io/kubernetes/cmd/kube-apiserver/app/options"

	kuberest "k8s.io/kubernetes/pkg/client/restclient"
	kubeclient "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/storage/storagebackend"
)

func (lk LocalkubeServer) NewAPIServer() Server {
	return NewSimpleServer("apiserver", serverInterval, StartAPIServer(lk))
}

func StartAPIServer(lk LocalkubeServer) func() error {
	config := options.NewAPIServer()

	config.BindAddress = lk.APIServerAddress
	config.SecurePort = lk.APIServerPort
	config.InsecureBindAddress = lk.APIServerInsecureAddress
	config.InsecurePort = lk.APIServerInsecurePort

	config.ClientCAFile = lk.GetCAPublicKeyCertPath()
	config.TLSCertFile = lk.GetPublicKeyCertPath()
	config.TLSPrivateKeyFile = lk.GetPrivateKeyCertPath()
	config.AdmissionControl = "NamespaceLifecycle,LimitRanger,ServiceAccount,ResourceQuota"

	// use localkube etcd
	config.StorageConfig = storagebackend.Config{ServerList: KubeEtcdClientURLs}

	// set Service IP range
	config.ServiceClusterIPRange = lk.ServiceClusterIPRange

	// defaults from apiserver command
	config.EnableProfiling = true
	config.EnableWatchCache = true
	config.MinRequestTimeout = 1800

	config.AllowPrivileged = true

	config.RuntimeConfig = lk.RuntimeConfig

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
