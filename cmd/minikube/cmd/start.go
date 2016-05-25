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
	"log"
	"os"
	"strings"

	"github.com/docker/machine/libmachine"
	"github.com/spf13/cobra"
	cfg "k8s.io/kubernetes/pkg/client/unversioned/clientcmd/api"
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/kubeconfig"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts a local kubernetes cluster.",
	Long: `Starts a local kubernetes cluster using Virtualbox. This command
assumes you already have Virtualbox installed.`,
	Run: runStart,
}

var (
	minikubeISO string
)

func runStart(cmd *cobra.Command, args []string) {

	fmt.Println("Starting local Kubernetes cluster...")
	api := libmachine.NewClient(constants.Minipath, constants.MakeMiniPath("certs"))
	defer api.Close()

	config := cluster.MachineConfig{
		MinikubeISO: minikubeISO,
	}

	host, err := cluster.StartHost(api, config)
	if err != nil {
		log.Println("Error starting host: ", err)
		os.Exit(1)
	}

	if err := cluster.UpdateCluster(host.Driver); err != nil {
		log.Println("Error updating cluster: ", err)
		os.Exit(1)
	}

	if err := cluster.SetupCerts(host.Driver); err != nil {
		log.Println("Error configuring authentication: ", err)
		os.Exit(1)
	}

	if err := cluster.StartCluster(host); err != nil {
		log.Println("Error starting cluster: ", err)
		os.Exit(1)
	}

	kubeHost, err := host.Driver.GetURL()
	if err != nil {
		log.Println("Error connecting to cluster: ", err)
	}
	kubeHost = strings.Replace(kubeHost, "tcp://", "https://", -1)
	kubeHost = strings.Replace(kubeHost, ":2376", ":443", -1)
	fmt.Printf("Kubernetes is available at %s.\n", kubeHost)

	// setup kubeconfig
	name := constants.MinikubeContext
	certAuth := constants.MakeMiniPath("apiserver.crt")
	clientCert := constants.MakeMiniPath("apiserver.crt")
	clientKey := constants.MakeMiniPath("apiserver.key")
	if active, err := setupKubeconfig(name, kubeHost, certAuth, clientCert, clientKey); err != nil {
		log.Println("Error setting up kubeconfig: ", err)
		os.Exit(1)
	} else if !active {
		fmt.Println("Run this command to use the cluster: ")
		fmt.Printf("kubectl config use-context %s\n", name)
	}
}

// setupKubeconfig reads config from disk, adds the minikube settings, and writes it back.
// activeContext is true when minikube is the CurrentContext
// If no CurrentContext is set, the given name will be used.
func setupKubeconfig(name, server, certAuth, cliCert, cliKey string) (activeContext bool, err error) {
	configFile := constants.KubeconfigPath

	// read existing config or create new if does not exist
	config, err := kubeconfig.ReadConfigOrNew(configFile)
	if err != nil {
		return false, err
	}

	clusterName := name
	cluster := cfg.NewCluster()
	cluster.Server = server
	cluster.CertificateAuthority = certAuth
	config.Clusters[clusterName] = cluster

	// user
	userName := name
	user := cfg.NewAuthInfo()
	user.ClientCertificate = cliCert
	user.ClientKey = cliKey
	config.AuthInfos[userName] = user

	// context
	contextName := name
	context := cfg.NewContext()
	context.Cluster = clusterName
	context.AuthInfo = userName
	config.Contexts[contextName] = context

	// set current context to minikube if unset
	if len(config.CurrentContext) == 0 {
		config.CurrentContext = contextName
	}

	// write back to disk
	if err := kubeconfig.WriteConfig(config, configFile); err != nil {
		return false, err
	}

	// activeContext if current matches name
	return name == config.CurrentContext, nil
}

func init() {
	startCmd.Flags().StringVarP(&minikubeISO, "iso-url", "", constants.DefaultIsoUrl, "Location of the minikube iso")
	RootCmd.AddCommand(startCmd)
}
