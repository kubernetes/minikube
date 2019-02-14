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
	"os"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	cmdcfg "k8s.io/minikube/cmd/minikube/cmd/config"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/console"
	"k8s.io/minikube/pkg/minikube/cruntime"
	"k8s.io/minikube/pkg/minikube/logs"
	"k8s.io/minikube/pkg/minikube/machine"
)

const (
	// number of problems per log to output
	numberOfProblems = 5
)

var (
	// followLogs triggers tail -f mode
	followLogs bool
	// numberOfLines is how many lines to output, set via -n
	numberOfLines int
	// showProblems only shows lines that match known issues
	showProblems bool
)

// logsCmd represents the logs command
var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Gets the logs of the running instance, used for debugging minikube, not user code",
	Long:  `Gets the logs of the running instance, used for debugging minikube, not user code.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			console.Fatal("Error getting config: %v", err)
		}

		api, err := machine.NewAPIClient()
		if err != nil {
			console.Fatal("Error getting client: %v", err)
			os.Exit(1)
		}
		defer api.Close()

		h, err := api.Load(config.GetMachineName())
		if err != nil {
			glog.Exitf("api load: %v", err)
		}
		runner, err := machine.CommandRunner(h)
		if err != nil {
			glog.Exitf("command runner: %v", err)
		}
		bs, err := GetClusterBootstrapper(api, viper.GetString(cmdcfg.Bootstrapper))
		if err != nil {
			glog.Exitf("Error getting cluster bootstrapper: %v", err)
		}

		cr, err := cruntime.New(cruntime.Config{Type: cfg.KubernetesConfig.ContainerRuntime, Runner: runner})
		if err != nil {
			glog.Exitf("Unable to get runtime: %v", err)
		}
		if followLogs {
			err := logs.Follow(cr, bs, runner)
			if err != nil {
				console.Failure("output: %v", err)
			}
			return
		}
		if showProblems {
			problems := logs.FindProblems(cr, bs, runner)
			logs.OutputProblems(problems, numberOfProblems)
			return
		}
		err = logs.Output(cr, bs, runner, numberOfLines)
		if err != nil {
			console.Failure("output: %v", err)
		}
	},
}

func init() {
	logsCmd.Flags().BoolVarP(&followLogs, "follow", "f", false, "Show only the most recent journal entries, and continuously print new entries as they are appended to the journal.")
	logsCmd.Flags().BoolVar(&showProblems, "problems", false, "Show only log entries which point to known problems")
	logsCmd.Flags().IntVarP(&numberOfLines, "length", "n", 50, "Number of lines back to go within the log")
	RootCmd.AddCommand(logsCmd)
}
