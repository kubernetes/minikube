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

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	cmdUtil "k8s.io/minikube/cmd/util"
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/machine"
)

var (
	follow bool
)

// logsCmd represents the logs command
var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Gets the logs of the running localkube instance, used for debugging minikube, not user code",
	Long:  `Gets the logs of the running localkube instance, used for debugging minikube, not user code.`,
	Run: func(cmd *cobra.Command, args []string) {
		api, err := machine.NewAPIClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting client: %s\n", err)
			os.Exit(1)
		}
		defer api.Close()
		h, err := api.Load(config.GetMachineName())
		if err != nil {
			glog.Errorln("Error getting host")
		}
		cmdRunner, err := machine.GetCommandRunner(h)
		if err != nil {
			glog.Errorln("Error getting command runner interface")
		}
		s, err := cluster.GetHostLogs(cmdRunner, follow)
		if err != nil {
			log.Println("Error getting machine logs:", err)
			cmdUtil.MaybeReportErrorAndExit(err)
		}
		fmt.Fprintln(os.Stdout, s)
	},
}

func init() {
	logsCmd.Flags().BoolVarP(&follow, "follow", "f", false, "Show only the most recent journal entries, and continuously print new entries as they are appended to the journal.")
	RootCmd.AddCommand(logsCmd)
}
