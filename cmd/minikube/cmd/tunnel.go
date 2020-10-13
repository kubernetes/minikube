/*
Copyright 2018 The Kubernetes Authors All rights reserved.

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
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"

	"github.com/spf13/cobra"

	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/kapi"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/mustload"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/minikube/tunnel"
	"k8s.io/minikube/pkg/minikube/tunnel/kic"
)

var cleanup bool

// tunnelCmd represents the tunnel command
var tunnelCmd = &cobra.Command{
	Use:   "tunnel",
	Short: "Connect to LoadBalancer services",
	Long:  `tunnel creates a route to services deployed with type LoadBalancer and sets their Ingress to their ClusterIP. for a detailed example see https://minikube.sigs.k8s.io/docs/tasks/loadbalancer`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		RootCmd.PersistentPreRun(cmd, args)
	},
	Run: func(cmd *cobra.Command, args []string) {
		manager := tunnel.NewManager()
		cname := ClusterFlagValue()
		co := mustload.Healthy(cname)

		if cleanup {
			klog.Info("Checking for tunnels to cleanup...")
			if err := manager.CleanupNotRunningTunnels(); err != nil {
				klog.Errorf("error cleaning up: %s", err)
			}
		}

		// Tunnel uses the k8s clientset to query the API server for services in the LoadBalancerEmulator.
		// We define the tunnel and minikube error free if the API server responds within a second.
		// This also contributes to better UX, the tunnel status check can happen every second and
		// doesn't hang on the API server call during startup and shutdown time or if there is a temporary error.
		clientset, err := kapi.Client(cname)
		if err != nil {
			exit.Error(reason.InternalKubernetesClient, "error creating clientset", err)
		}

		ctrlC := make(chan os.Signal, 1)
		signal.Notify(ctrlC, os.Interrupt)
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			<-ctrlC
			cancel()
		}()

		if driver.NeedsPortForward(co.Config.Driver) {

			port, err := oci.ForwardedPort(oci.Docker, cname, 22)
			if err != nil {
				exit.Error(reason.DrvPortForward, "error getting ssh port", err)
			}
			sshPort := strconv.Itoa(port)
			sshKey := filepath.Join(localpath.MiniPath(), "machines", cname, "id_rsa")

			kicSSHTunnel := kic.NewSSHTunnel(ctx, sshPort, sshKey, clientset.CoreV1())
			err = kicSSHTunnel.Start()
			if err != nil {
				exit.Error(reason.SvcTunnelStart, "error starting tunnel", err)
			}

			return
		}

		done, err := manager.StartTunnel(ctx, cname, co.API, config.DefaultLoader, clientset.CoreV1())
		if err != nil {
			exit.Error(reason.SvcTunnelStart, "error starting tunnel", err)
		}
		<-done
	},
}

func init() {
	tunnelCmd.Flags().BoolVarP(&cleanup, "cleanup", "c", true, "call with cleanup=true to remove old tunnels")
}
