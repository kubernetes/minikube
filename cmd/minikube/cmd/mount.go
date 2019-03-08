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
	"net"
	"os"
	"strings"
	"sync"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	cmdUtil "k8s.io/minikube/cmd/util"
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/console"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/third_party/go9p/ufs"
)

var mountIP string
var mountVersion string
var isKill bool
var uid int
var gid int
var msize int

// mountCmd represents the mount command
var mountCmd = &cobra.Command{
	Use:   "mount [flags] MOUNT_DIRECTORY(ex:\"/home\")",
	Short: "Mounts the specified directory into minikube",
	Long:  `Mounts the specified directory into minikube.`,
	Run: func(cmd *cobra.Command, args []string) {
		if isKill {
			if err := cmdUtil.KillMountProcess(); err != nil {
				exit.WithError("Error killing mount process", err)
			}
			os.Exit(0)
		}

		if len(args) != 1 {
			exit.Usage(`Please specify the directory to be mounted: 
	minikube mount HOST_MOUNT_DIRECTORY:VM_MOUNT_DIRECTORY(ex:"/host-home:/vm-home")`)
		}
		mountString := args[0]
		idx := strings.LastIndex(mountString, ":")
		if idx == -1 { // no ":" was present
			exit.Usage(`Mount directory must be in the form: 
	HOST_MOUNT_DIRECTORY:VM_MOUNT_DIRECTORY`)
		}
		hostPath := mountString[:idx]
		vmPath := mountString[idx+1:]
		if _, err := os.Stat(hostPath); err != nil {
			if os.IsNotExist(err) {
				exit.WithCode(exit.NoInput, "Cannot find directory %s for mount", hostPath)
			} else {
				exit.WithError("stat failed", err)
			}
		}
		if len(vmPath) == 0 || !strings.HasPrefix(vmPath, "/") {
			exit.Usage("The :VM_MOUNT_DIRECTORY must be an absolute path")
		}
		var debugVal int
		if glog.V(1) {
			debugVal = 1 // ufs.StartServer takes int debug param
		}
		api, err := machine.NewAPIClient()
		if err != nil {
			exit.WithError("Error getting client", err)
		}
		defer api.Close()
		host, err := api.Load(config.GetMachineName())
		if err != nil {
			exit.WithError("Error loading api", err)
		}
		if host.Driver.DriverName() == "none" {
			exit.Usage(`'none' driver does not support 'minikube mount' command`)
		}
		var ip net.IP
		if mountIP == "" {
			ip, err = cluster.GetVMHostIP(host)
			if err != nil {
				exit.WithError("Error getting the host IP address to use from within the VM", err)
			}
		} else {
			ip = net.ParseIP(mountIP)
			if ip == nil {
				exit.WithCode(exit.Data, "error parsing the input ip address for mount")
			}
		}
		console.OutStyle("mounting", "Mounting %s into %s on the minikube VM", hostPath, vmPath)
		console.OutStyle("notice", "This daemon process needs to stay alive for the mount to be accessible ...")
		port, err := cmdUtil.GetPort()
		if err != nil {
			exit.WithError("Error finding port for mount", err)
		}
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			ufs.StartServer(net.JoinHostPort(ip.String(), port), debugVal, hostPath)
			wg.Done()
		}()
		err = cluster.MountHost(api, ip, vmPath, port, mountVersion, uid, gid, msize)
		if err != nil {
			exit.WithError("failed to mount host", err)
		}
		wg.Wait()
	},
}

func init() {
	mountCmd.Flags().StringVar(&mountIP, "ip", "", "Specify the ip that the mount should be setup on")
	mountCmd.Flags().StringVar(&mountVersion, "9p-version", constants.DefaultMountVersion, "Specify the 9p version that the mount should use")
	mountCmd.Flags().BoolVar(&isKill, "kill", false, "Kill the mount process spawned by minikube start")
	mountCmd.Flags().IntVar(&uid, "uid", 1001, "Default user id used for the mount")
	mountCmd.Flags().IntVar(&gid, "gid", 1001, "Default group id used for the mount")
	mountCmd.Flags().IntVar(&msize, "msize", constants.DefaultMsize, "The number of bytes to use for 9p packet payload")
	RootCmd.AddCommand(mountCmd)
}
