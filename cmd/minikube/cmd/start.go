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
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/constants"
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
	localkubeURL string
)

func runStart(cmd *cobra.Command, args []string) {

	fmt.Println("Starting local Kubernetes cluster...")
	api := libmachine.NewClient(constants.Minipath, constants.MakeMiniPath("certs"))
	defer api.Close()
	host, err := cluster.StartHost(api)
	if err != nil {
		log.Println("Error starting host: ", err)
		os.Exit(1)
	}

	config := cluster.KubernetesConfig{
		LocalkubeURL: localkubeURL,
	}

	if err := cluster.StartCluster(host, config); err != nil {
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
	fmt.Println("Run this command to use the cluster: ")
	fmt.Printf("kubectl config set-cluster minikube --server=%s --certificate-authority=$HOME/.minikube/ca.crt\n", kubeHost)
	fmt.Println("kubectl config set-credentials minikube --client-certificate=$HOME/.minikube/kubecfg.crt --client-key=$HOME/.minikube/kubecfg.key")
	fmt.Println("kubectl config set-context minikube --cluster=minikube --user=minikube")
	fmt.Println("kubectl config use-context minikube")

	if err := cluster.GetCreds(host); err != nil {
		log.Println("Error configuring authentication: ", err)
		os.Exit(1)
	}
}

func init() {
	startCmd.Flags().StringVarP(&localkubeURL, "localkube-url", "", "https://storage.googleapis.com/tinykube/localkube", "Location of the localkube binary")
	startCmd.Flags().MarkHidden("localkube-url")
	RootCmd.AddCommand(startCmd)
}
