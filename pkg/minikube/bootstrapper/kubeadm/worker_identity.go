/*
Copyright 2026 The Kubernetes Authors All rights reserved.

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

package kubeadm

import (
	"fmt"
	"net"
	"os/exec"
	"path"
	"strconv"

	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/bootstrapper/bsutil"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/vmpath"
)

const workerKubeconfigPath = "/etc/kubernetes/kubelet.conf"
const workerClientIdentityPath = "/var/lib/kubelet/pki/kubelet-client-current.pem"

// guestFilePresent distinguishes an absent guest file from a failed transport.
// It reads metadata only; certificate and key bytes remain inside the guest.
func (k *Bootstrapper) guestFilePresent(filename string) (bool, error) {
	result, err := k.c.RunCmd(exec.Command("sudo", "test", "-s", filename))
	if err == nil {
		return true, nil
	}
	if result != nil && result.ExitCode == 1 {
		return false, nil
	}
	return false, fmt.Errorf("inspect guest identity file: %w", err)
}

// restoreManagedKubeletConfig recreates a boot-local kubeconfig using the
// worker's original rotating client certificate. Existing configuration remains
// untouched; a new worker without a client certificate proceeds to normal join.
func (k *Bootstrapper) restoreManagedKubeletConfig(cfg config.ClusterConfig) error {
	present, err := k.guestFilePresent(workerKubeconfigPath)
	if err != nil || present {
		return err
	}
	present, err = k.guestFilePresent(workerClientIdentityPath)
	if err != nil || !present {
		return err
	}
	ca := path.Join(vmpath.GuestKubernetesCertsDir, "ca.crt")
	present, err = k.guestFilePresent(ca)
	if err != nil {
		return err
	}
	if !present {
		return fmt.Errorf("the retained worker client identity requires its original cluster CA")
	}
	// File references preserve automatic certificate rotation. No private key
	// leaves its guest and no replacement Node or bootstrap token is created.
	server := "https://" + net.JoinHostPort(constants.ControlPlaneAlias, strconv.Itoa(cfg.APIServerPort))
	document := clientcmdapi.Config{
		Clusters:       map[string]*clientcmdapi.Cluster{"minikube": {Server: server, CertificateAuthority: ca}},
		AuthInfos:      map[string]*clientcmdapi.AuthInfo{"kubelet": {ClientCertificate: workerClientIdentityPath, ClientKey: workerClientIdentityPath}},
		Contexts:       map[string]*clientcmdapi.Context{"kubelet": {Cluster: "minikube", AuthInfo: "kubelet"}},
		CurrentContext: "kubelet",
	}
	payload, err := clientcmd.Write(document)
	if err != nil {
		return fmt.Errorf("encode retained worker kubeconfig: %w", err)
	}
	if err := bsutil.CopyFiles(k.c, []assets.CopyableFile{assets.NewMemoryAssetTarget(payload, workerKubeconfigPath, "0600")}); err != nil {
		return fmt.Errorf("restore retained worker kubeconfig: %w", err)
	}
	klog.Info("Restored managed worker kubeconfig using its retained client certificate")
	return nil
}
