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

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/docker/machine/libmachine/host"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	cfg "k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/exit"
)

// Update sets up kubeconfig to be used by kubectl
func Update(h *host.Host, c *cfg.Config) *Setup {
	addr, err := h.Driver.GetURL()
	if err != nil {
		exit.WithError("Failed to get driver URL", err)
	}
	addr = strings.Replace(addr, "tcp://", "https://", -1)
	addr = strings.Replace(addr, ":2376", ":"+strconv.Itoa(c.KubernetesConfig.NodePort), -1)
	if c.KubernetesConfig.APIServerName != constants.APIServerName {
		addr = strings.Replace(addr, c.KubernetesConfig.NodeIP, c.KubernetesConfig.APIServerName, -1)
	}

	kcs := &Setup{
		ClusterName:          cfg.GetMachineName(),
		ClusterServerAddress: addr,
		ClientCertificate:    constants.MakeMiniPath("client.crt"),
		ClientKey:            constants.MakeMiniPath("client.key"),
		CertificateAuthority: constants.MakeMiniPath("ca.crt"),
		KeepContext:          viper.GetBool("keep-context"),
		EmbedCerts:           viper.GetBool("embed-certs"),
	}
	kcs.setPath(Path())
	if err := update(kcs); err != nil {
		exit.WithError("Failed to setup kubeconfig", err)
	}
	return kcs
}

// update reads config from disk, adds the minikube settings, and writes it back.
// activeContext is true when minikube is the CurrentContext
// If no CurrentContext is set, the given name will be used.
func update(cfg *Setup) error {
	glog.Infoln("Using kubeconfig: ", cfg.fileContent())

	// read existing config or create new if does not exist
	config, err := readOrNew(cfg.fileContent())
	if err != nil {
		return err
	}

	err = PopulateKubeConfig(cfg, config)
	if err != nil {
		return err
	}

	// write back to disk
	if err := writeToFile(config, cfg.fileContent()); err != nil {
		return errors.Wrap(err, "writing kubeconfig")
	}
	return nil
}

// UpdateIP overwrites the IP stored in kubeconfig with the provided IP.
func UpdateIP(ip net.IP, filename string, machineName string) (bool, error) {
	if ip == nil {
		return false, fmt.Errorf("error, empty ip passed")
	}
	kip, err := extractIP(filename, machineName)
	if err != nil {
		return false, err
	}
	if kip.Equal(ip) {
		return false, nil
	}
	kport, err := Port(filename, machineName)
	if err != nil {
		return false, err
	}
	con, err := readOrNew(filename)
	if err != nil {
		return false, errors.Wrap(err, "Error getting kubeconfig status")
	}
	// Safe to lookup server because if field non-existent getIPFromKubeconfig would have given an error
	con.Clusters[machineName].Server = "https://" + ip.String() + ":" + strconv.Itoa(kport)
	err = writeToFile(con, filename)
	if err != nil {
		return false, err
	}
	// Kubeconfig IP reconfigured
	return true, nil
}
