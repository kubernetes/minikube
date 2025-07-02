/*
Copyright 2024 The Kubernetes Authors All rights reserved.

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
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/kubeconfig"
	"k8s.io/minikube/pkg/minikube/mustload"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/minikube/style"
)

// tunnelSSHCmd represents the tunnel-ssh command for persistent SSH tunneling
var tunnelSSHCmd = &cobra.Command{
	Use:   "tunnel-ssh",
	Short: "Create persistent SSH tunnel for remote Docker contexts",
	Long: `Creates and maintains an SSH tunnel for API server access when using
remote Docker contexts. The tunnel runs in the foreground and can be stopped with Ctrl+C.`,
	Run: func(_ *cobra.Command, _ []string) {
		// Check if we're using a remote SSH Docker context
		if !oci.IsRemoteDockerContext() {
			out.Styled(style.Meh, "No remote Docker context detected - tunnel not needed")
			return
		}

		if !oci.IsSSHDockerContext() {
			out.ErrT(style.Sad, "SSH tunnel only supported for SSH-based Docker contexts")
			exit.Error(reason.Usage, "unsupported context type", fmt.Errorf("not an SSH context"))
		}

		cname := ClusterFlagValue()
		co := mustload.Running(cname)

		out.Step(style.Launch, "Starting SSH tunnel for API server access...")
		klog.Infof("Setting up persistent SSH tunnel for cluster %s", cname)

		// Set up SSH tunnel for API server access
		tunnelEndpoint, cleanup, err := oci.SetupAPIServerTunnel(co.CP.Port)
		if err != nil {
			exit.Error(reason.HostKubeconfigUpdate, "setting up SSH tunnel", err)
		}

		if tunnelEndpoint == "" {
			out.Styled(style.Meh, "No tunnel needed for this context")
			return
		}

		// Parse tunnel endpoint to get port
		var tunnelPort int
		if len(tunnelEndpoint) > 19 { // len("https://localhost:")
			portStr := tunnelEndpoint[19:] // Skip "https://localhost:"
			if port, parseErr := strconv.Atoi(portStr); parseErr == nil {
				tunnelPort = port
			}
		}

		// Update kubeconfig to use the tunnel
		updated, err := kubeconfig.UpdateEndpoint(cname, "localhost", tunnelPort, kubeconfig.PathFromEnv(), kubeconfig.NewExtension())
		if err != nil {
			cleanup()
			exit.Error(reason.HostKubeconfigUpdate, "updating kubeconfig", err)
		}

		if updated {
			out.Step(style.Celebrate, `"{{.context}}" context updated to use SSH tunnel {{.endpoint}}`,
				out.V{"context": cname, "endpoint": tunnelEndpoint})
		}

		out.Step(style.Running, "SSH tunnel active on {{.endpoint}} ({{.original}} -> localhost:{{.port}})",
			out.V{
				"endpoint": tunnelEndpoint,
				"original": fmt.Sprintf("%s:%d", co.CP.Hostname, co.CP.Port),
				"port": tunnelPort,
			})
		out.Styled(style.Tip, "Press Ctrl+C to stop the tunnel")

		// Set up signal handling for graceful shutdown
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)

		// Wait for interrupt signal
		<-c
		out.Step(style.Shutdown, "Stopping SSH tunnel...")
		cleanup()
		out.Styled(style.Celebrate, "SSH tunnel stopped")
	},
}

func init() {
	RootCmd.AddCommand(tunnelSSHCmd)
}