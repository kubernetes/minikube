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
	"os"
	"sync"

	"strings"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/third_party/go9p/ufs"
)

// mountCmd represents the mount command
var mountCmd = &cobra.Command{
	Use:   "mount [flags] MOUNT_DIRECTORY(ex:\"/home\")",
	Short: "Mounts the specified directory into minikube.",
	Long:  `Mounts the specified directory into minikube.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			errText := `Please specify the directory to be mounted: 
\tminikube mount HOST_MOUNT_DIRECTORY:VM_MOUNT_DIRECTORY(ex:"/host-home:/vm-home")
`
			fmt.Fprintln(os.Stderr, errText)
			os.Exit(1)
		}
		mountString := args[0]
		idx := strings.LastIndex(mountString, ":")
		if idx == -1 { // no ":" was present
			errText := `Mount directory must be in the form: 
			\tHOST_MOUNT_DIRECTORY:VM_MOUNT_DIRECTORY`
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
				errText := fmt.Sprintf("Error accesssing directory %s for mount", hostPath)
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
		api, err := machine.NewAPIClient(clientType)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting client: %s\n", err)
			os.Exit(1)
		}
		defer api.Close()

		fmt.Printf("Mounting %s into %s on the minikubeVM\n", hostPath, vmPath)
		fmt.Println("This daemon process needs to stay alive for the mount to still be accessible...")
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			ufs.StartServer(constants.DefaultUfsAddress, debugVal, hostPath)
			wg.Done()
		}()
		err = cluster.MountHost(api, vmPath)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		wg.Wait()
	},
}

func init() {
	RootCmd.AddCommand(mountCmd)
}
