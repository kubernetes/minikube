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
	"path"
	"strconv"

	"github.com/coreos/etcd/embed"

	apiserveroptions "k8s.io/apiserver/pkg/server/options"
	"k8s.io/apiserver/pkg/storage/storagebackend"

	apiserver "k8s.io/kubernetes/cmd/kube-apiserver/app"
	"k8s.io/kubernetes/cmd/kube-apiserver/app/options"
	kubeapioptions "k8s.io/kubernetes/pkg/kubeapiserver/options"
)

func (lk LocalkubeServer) NewAPIServer() Server {
	return NewSimpleServer("apiserver", serverInterval, StartAPIServer(lk), readyFunc(lk))
}

func StartAPIServer(lk LocalkubeServer) func() error {
	config := options.NewServerRunOptions()

	config.SecureServing.BindAddress = lk.APIServerAddress
	config.SecureServing.BindPort = lk.APIServerPort

	config.InsecureServing.BindAddress = lk.APIServerInsecureAddress
	config.InsecureServing.BindPort = lk.APIServerInsecurePort

	config.Authentication.ClientCert.ClientCA = lk.GetCAPublicKeyCertPath()

	config.SecureServing.ServerCert.CertKey.CertFile = lk.GetPublicKeyCertPath()
	config.SecureServing.ServerCert.CertKey.KeyFile = lk.GetPrivateKeyCertPath()
	config.Admission.PluginNames = []string{
		"NamespaceLifecycle",
		"LimitRanger",
		"ServiceAccount",
		"DefaultStorageClass",
		"ResourceQuota",
	}
	// use localkube etcd

	config.Etcd.StorageConfig.ServerList = []string{embed.DefaultListenClientURLs}
	config.Etcd.StorageConfig.Type = storagebackend.StorageTypeETCD3

	// set Service IP range
	config.ServiceClusterIPRange = lk.ServiceClusterIPRange
	config.Etcd.EnableWatchCache = true

	config.Features = &apiserveroptions.FeatureOptions{
		EnableProfiling: true,
	}

	// defaults from apiserver command
	config.GenericServerRunOptions.MinRequestTimeout = 1800

	config.AllowPrivileged = true

	config.APIEnablement = &kubeapioptions.APIEnablementOptions{
		RuntimeConfig: lk.RuntimeConfig,
	}

	config.ProxyClientCertFile = lk.GetProxyClientPublicKeyCertPath()
	config.ProxyClientKeyFile = lk.GetProxyClientPrivateKeyCertPath()
	config.Authentication.RequestHeader.AllowedNames =
		[]string{}
	config.Authentication.RequestHeader.UsernameHeaders =
		[]string{"X-Remote-User"}
	config.Authentication.RequestHeader.GroupHeaders =
		[]string{"X-Remote-Group"}
	config.Authentication.RequestHeader.ExtraHeaderPrefixes =
		[]string{"X-Remote-Extra-"}
	config.Authentication.RequestHeader.ClientCAFile =
		lk.GetProxyClientCAPublicKeyCertPath()

	lk.SetExtraConfigForComponent("apiserver", &config)

	return func() error {
		stop := make(chan struct{})
		return apiserver.Run(config, stop)
	}
}

func readyFunc(lk LocalkubeServer) HealthCheck {
	return healthCheck(lk.GetAPIServerProtocol()+path.Join(
		net.JoinHostPort("localhost", strconv.Itoa(lk.APIServerPort)), "healthz"), lk)
}
