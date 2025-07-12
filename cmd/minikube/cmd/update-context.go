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
	"net/url"
	"strconv"

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

// updateContextCmd represents the update-context command
var updateContextCmd = &cobra.Command{
	Use:   "update-context",
	Short: "Update kubeconfig in case of an IP or port change",
	Long: `Retrieves the IP address of the running cluster, checks it
			with IP in kubeconfig, and corrects kubeconfig if incorrect.`,
	Run: func(_ *cobra.Command, _ []string) {
		cname := ClusterFlagValue()
		co := mustload.Running(cname)

		// Determine the endpoint to use (with tunneling or remote host)
		hostname := co.CP.Hostname
		port := co.CP.Port

		// Handle remote Docker contexts
		if oci.IsRemoteDockerContext() {
			if oci.IsSSHDockerContext() {
				klog.Infof("Remote SSH Docker context detected, setting up API server tunnel")

				// Set up SSH tunnel for API server access
				tunnelEndpoint, cleanup, err := oci.SetupAPIServerTunnel(co.CP.Port)
				if err != nil {
					klog.Warningf("Failed to setup SSH tunnel, falling back to direct connection: %v", err)
				} else if tunnelEndpoint != "" {
					// Parse the tunnel endpoint to get localhost and tunnel port
					hostname = "localhost"
					// Extract port from tunnelEndpoint (format: https://localhost:PORT)
					if len(tunnelEndpoint) > 19 { // len("https://localhost:")
						portStr := tunnelEndpoint[19:] // Skip "https://localhost:"
						if tunneledPort, parseErr := strconv.Atoi(portStr); parseErr == nil {
							port = tunneledPort
							klog.Infof("Using SSH tunnel: %s -> %s:%d", tunnelEndpoint, co.CP.Hostname, co.CP.Port)

							// Set up cleanup when the process exits
							defer func() {
								klog.Infof("Cleaning up SSH tunnel")
								cleanup()
							}()
						}
					}
				}
			} else {
				// TLS context - use the actual remote host (Docker TLS doesn't provide port forwarding)
				klog.Infof("Remote TLS Docker context detected, using direct connection to remote host")

				ctx, err := oci.GetCurrentContext()
				if err == nil && ctx.Host != "" {
					if u, parseErr := url.Parse(ctx.Host); parseErr == nil && u.Hostname() != "" {
						hostname = u.Hostname()
						klog.Infof("Using remote host for TLS context: %s (port %d)", hostname, port)
					} else {
						klog.Warningf("Failed to parse remote host from context %q, using default", ctx.Host)
					}
				} else {
					klog.Warningf("Failed to get Docker context info for TLS endpoint: %v", err)
				}
			}
		}

		updated, err := kubeconfig.UpdateEndpoint(cname, hostname, port, kubeconfig.PathFromEnv(), kubeconfig.NewExtension())
		if err != nil {
			exit.Error(reason.HostKubeconfigUpdate, "update config", err)
		}

		if updated {
			if hostname == "localhost" && oci.IsRemoteDockerContext() && oci.IsSSHDockerContext() {
				out.Step(style.Celebrate, `"{{.context}}" context has been updated to point to {{.hostname}}:{{.port}} (SSH tunnel to {{.original}})`,
					out.V{"context": cname, "hostname": hostname, "port": port, "original": co.CP.Hostname + ":" + strconv.Itoa(co.CP.Port)})
			} else if oci.IsRemoteDockerContext() && !oci.IsSSHDockerContext() {
				out.Step(style.Celebrate, `"{{.context}}" context has been updated to point to {{.hostname}}:{{.port}} (TLS remote connection)`,
					out.V{"context": cname, "hostname": hostname, "port": port})
			} else {
				out.Step(style.Celebrate, `"{{.context}}" context has been updated to point to {{.hostname}}:{{.port}}`,
					out.V{"context": cname, "hostname": hostname, "port": port})
			}
		} else {
			out.Styled(style.Meh, `No changes required for the "{{.context}}" context`, out.V{"context": cname})
		}

		if err := kubeconfig.SetCurrentContext(cname, kubeconfig.PathFromEnv()); err != nil {
			out.ErrT(style.Sad, `Error while setting kubectl current context:  {{.error}}`, out.V{"error": err})
		} else {
			out.Styled(style.Kubectl, `Current context is "{{.context}}"`, out.V{"context": cname})
		}
	},
}
