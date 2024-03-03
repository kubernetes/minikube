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
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/juju/fslock"
	"github.com/spf13/cobra"

	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/kapi"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/mustload"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/minikube/style"
	"k8s.io/minikube/pkg/minikube/tunnel"
	"k8s.io/minikube/pkg/minikube/tunnel/kic"
	pkgnetwork "k8s.io/minikube/pkg/network"
)

var cleanup bool
var bindAddress string
var lockHandle *fslock.Lock

// tunnelCmd represents the tunnel command
var tunnelCmd = &cobra.Command{
	Use:   "tunnel",
	Short: "Connect to LoadBalancer services",
	Long:  `tunnel creates a route to services deployed with type LoadBalancer and sets their Ingress to their ClusterIP. for a detailed example see https://minikube.sigs.k8s.io/docs/tasks/loadbalancer`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		RootCmd.PersistentPreRun(cmd, args)
	},
	Run: func(_ *cobra.Command, _ []string) {
		manager := tunnel.NewManager()
		cname := ClusterFlagValue()
		co := mustload.Healthy(cname)

		if driver.IsQEMU(co.Config.Driver) && pkgnetwork.IsBuiltinQEMU(co.Config.Network) {
			msg := "minikube tunnel is not currently implemented with the builtin network on QEMU"
			if runtime.GOOS == "darwin" {
				msg += ", try starting minikube with '--network=socket_vmnet'"
			}
			exit.Message(reason.Unimplemented, msg)
		}

		if cleanup {
			klog.Info("Checking for tunnels to cleanup...")
			if err := manager.CleanupNotRunningTunnels(); err != nil {
				klog.Errorf("error cleaning up: %s", err)
			}
		}

		mustLockOrExit(cname)
		defer cleanupLock()

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

		if driver.NeedsPortForward(co.Config.Driver) || bindAddress != "" {
			port, err := oci.ForwardedPort(co.Config.Driver, cname, 22)
			if err != nil {
				exit.Error(reason.DrvPortForward, "error getting ssh port", err)
			}
			sshPort := strconv.Itoa(port)
			sshKey := filepath.Join(localpath.MiniPath(), "machines", cname, "id_rsa")

			outputTunnelStarted()
			kicSSHTunnel := kic.NewSSHTunnel(ctx, sshPort, sshKey, bindAddress, clientset.CoreV1(), clientset.NetworkingV1())
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

func cleanupLock() {
	if lockHandle != nil {
		err := lockHandle.Unlock()
		if err != nil {
			out.Styled(style.Warning, fmt.Sprintf("failed to release lock during cleanup: %v", err))
		}
	}
}

func mustLockOrExit(profile string) {
	tunnelLockPath := filepath.Join(localpath.Profile(profile), ".tunnel_lock")

	lockHandle = fslock.New(tunnelLockPath)
	err := lockHandle.TryLock()
	if err == fslock.ErrLocked {
		exit.Message(reason.SvcTunnelAlreadyRunning, "Another tunnel process is already running, terminate the existing instance to start a new one")
	}
	if err != nil {
		exit.Error(reason.SvcTunnelStart, "failed to acquire lock due to unexpected error", err)
	}
}

func outputTunnelStarted() {
	out.Styled(style.Success, "Tunnel successfully started")
	out.Ln("")
	out.Styled(style.Notice, "NOTE: Please do not close this terminal as this process must stay alive for the tunnel to be accessible ...")
	out.Ln("")
}

func init() {
	tunnelCmd.Flags().BoolVarP(&cleanup, "cleanup", "c", true, "call with cleanup=true to remove old tunnels")
	tunnelCmd.Flags().StringVar(&bindAddress, "bind-address", "", "set tunnel bind address, empty or '*' indicates the tunnel should be available for all interfaces")
}
