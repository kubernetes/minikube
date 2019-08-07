/*
Copyright 2019 The Kubernetes Authors All rights reserved.

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

package kubeconfig

import "sync/atomic"

// Setup is the kubeconfig setup
type Setup struct {
	// The name of the cluster for this context
	ClusterName string

	// ClusterServerAddress is the address of the kubernetes cluster
	ClusterServerAddress string

	// ClientCertificate is the path to a client cert file for TLS.
	ClientCertificate string

	// CertificateAuthority is the path to a cert file for the certificate authority.
	CertificateAuthority string

	// ClientKey is the path to a client key file for TLS.
	ClientKey string

	// Should the current context be kept when setting up this one
	KeepContext bool

	// Should the certificate files be embedded instead of referenced by path
	EmbedCerts bool

	// kubeConfigFile is the path where the kube config is stored
	// Only access this with atomic ops
	kubeConfigFile atomic.Value
}

// SetKubeConfigFile sets the kubeconfig file
func (k *Setup) setPath(kubeConfigFile string) {
	k.kubeConfigFile.Store(kubeConfigFile)
}

// fileContent gets the kubeconfig file
func (k *Setup) fileContent() string {
	return k.kubeConfigFile.Load().(string)
}
