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
	"net"
	"os"
	"sync"

	"strings"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	cmdUtil "k8s.io/minikube/cmd/util"
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
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
				fmt.Println("Errors occurred deleting mount process: ", err)
				os.Exit(1)
			}
			os.Exit(0)
		}

		if len(args) != 1 {
			errText := `Please specify the directory to be mounted: 
	minikube mount HOST_MOUNT_DIRECTORY:VM_MOUNT_DIRECTORY(ex:"/host-home:/vm-home")
`
			fmt.Fprintln(os.Stderr, errText)
			os.Exit(1)
		}
		mountString := args[0]
		idx := strings.LastIndex(mountString, ":")
		if idx == -1 { // no ":" was present
			errText := `Mount directory must be in the form: 
	HOST_MOUNT_DIRECTORY:VM_MOUNT_DIRECTORY`
			fmt.Fprintln(os.Stderr, errText)
			os.Exit(1)
		}
		hostPath := mountString[:idx]
		vmPath := mountString[idx+1:]
		if _, err := os.Stat(hostPath); err != nil {
			if os.IsNotExist(err) {
				errText := fmt.Sprintf("Cannot find directory %s for mount", hostPath)
				fmt.Fprintln(os.Stderr, errText)
			} else {
				errText := fmt.Sprintf("Error accessing directory %s for mount", hostPath)
				fmt.Fprintln(os.Stderr, errText)
			}
			os.Exit(1)
		}
		if len(vmPath) == 0 || !strings.HasPrefix(vmPath, "/") {
			errText := fmt.Sprintf("The :VM_MOUNT_DIRECTORY must be an absolute path")
			fmt.Fprintln(os.Stderr, errText)
			os.Exit(1)
		}
		var debugVal int
		if glog.V(1) {
			debugVal = 1 // ufs.StartServer takes int debug param
		}
		api, err := machine.NewAPIClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting client: %v\n", err)
			os.Exit(1)
		}
		defer api.Close()
		host, err := api.Load(config.GetMachineName())
		if err != nil {
			glog.Errorln("Error loading api: ", err)
			os.Exit(1)
		}
		if host.Driver.DriverName() == "none" {
			fmt.Println(`'none' driver does not support 'minikube mount' command`)
			os.Exit(0)
		}
		var ip net.IP
		if mountIP == "" {
			ip, err = cluster.GetVMHostIP(host)
			if err != nil {
				glog.Errorln("Error getting the host IP address to use from within the VM: ", err)
				os.Exit(1)
			}
		} else {
			ip = net.ParseIP(mountIP)
			if ip == nil {
				glog.Errorln("error parsing the input ip address for mount")
				os.Exit(1)
			}
		}
		fmt.Printf("Mounting %s into %s on the minikube VM\n", hostPath, vmPath)
		fmt.Println("This daemon process needs to stay alive for the mount to still be accessible...")
		port, err := cmdUtil.GetPort()
		if err != nil {
			glog.Errorln("Error finding port for mount: ", err)
			os.Exit(1)
		}
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			ufs.StartServer(net.JoinHostPort(ip.String(), port), debugVal, hostPath)
			wg.Done()
		}()
		err = cluster.MountHost(api, ip, vmPath, port, mountVersion, uid, gid, msize)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
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
