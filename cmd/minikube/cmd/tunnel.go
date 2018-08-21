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
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/service"
	"k8s.io/minikube/pkg/minikube/tunnel"
	"os"
	"os/signal"
)

var cleanup bool

// tunnelCmd represents the tunnel command
var tunnelCmd = &cobra.Command{
	Use:   "tunnel",
	Short: "tunnel makes services of type LoadBalancer accessible on localhost",
	Long:  `tunnel creates a route to services deployed with type LoadBalancer and sets their Ingress to their ClusterIP`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		RootCmd.PersistentPreRun(cmd, args)
	},
	Run: func(cmd *cobra.Command, args []string) {
		manager := tunnel.NewManager()

		if cleanup {
			logrus.Info("Checking for tunnels to cleanup...")
			manager.CleanupNotRunningTunnels()
			return
		}

		logrus.Infof("Creating docker machine client...")
		api, e := machine.NewAPIClient()
		if e != nil {
			logrus.Fatalf("error creating dockermachine client: %s", e)
		}
		machineName := config.GetMachineName()

		logrus.Infof("Creating k8s client...")
		clientset, e := service.K8s.GetClientset()
		if e != nil {
			logrus.Fatalf("error creating K8S clientset: %s", e)
		}
		v1 := clientset.CoreV1()

		ctrlC := make(chan os.Signal, 1)
		signal.Notify(ctrlC, os.Interrupt)
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			<-ctrlC
			cancel()
		}()

		done, e := manager.StartTunnel(ctx, machineName, api, config.Loader, v1)
		if e != nil {
			logrus.Fatalf("error starting tunnel: %s", e)
		}
		<-done
	},
}

func init() {
	tunnelCmd.Flags().BoolVarP(&cleanup, "cleanup", "c", false, "call with cleanup=true to remove old tunnels")
	RootCmd.AddCommand(tunnelCmd)
}
