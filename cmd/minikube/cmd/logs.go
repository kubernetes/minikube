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
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	cmdcfg "k8s.io/minikube/cmd/minikube/cmd/config"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/cruntime"
	"k8s.io/minikube/pkg/minikube/exit"
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
	Short: "Gets the logs of the running instance, used for debugging minikube, not user code.",
	Long:  `Gets the logs of the running instance, used for debugging minikube, not user code.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			exit.WithError("Error getting config", err)
		}

		api, err := machine.NewAPIClient()
		if err != nil {
			exit.WithError("Error getting client", err)
		}
		defer api.Close()

		h, err := api.Load(cfg.Name)
		if err != nil {
			exit.WithError("api load", err)
		}
		runner, err := machine.CommandRunner(h)
		if err != nil {
			exit.WithError("command runner", err)
		}
		bs, err := getClusterBootstrapper(api, viper.GetString(cmdcfg.Bootstrapper))
		if err != nil {
			exit.WithError("Error getting cluster bootstrapper", err)
		}

		cr, err := cruntime.New(cruntime.Config{Type: cfg.KubernetesConfig.ContainerRuntime, Runner: runner})
		if err != nil {
			exit.WithError("Unable to get runtime", err)
		}
		if followLogs {
			err := logs.Follow(cr, bs, runner)
			if err != nil {
				exit.WithError("Follow", err)
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
			exit.WithError("Error getting machine logs", err)
		}
	},
}

func init() {
	logsCmd.Flags().BoolVarP(&followLogs, "follow", "f", false, "Show only the most recent journal entries, and continuously print new entries as they are appended to the journal.")
	logsCmd.Flags().BoolVar(&showProblems, "problems", false, "Show only log entries which point to known problems")
	logsCmd.Flags().IntVarP(&numberOfLines, "length", "n", 60, "Number of lines back to go within the log")
}
