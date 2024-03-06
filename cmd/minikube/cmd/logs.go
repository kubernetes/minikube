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

	"github.com/docker/machine/libmachine/state"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/klog/v2"
	cmdcfg "k8s.io/minikube/cmd/minikube/cmd/config"
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/cruntime"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/logs"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/mustload"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/minikube/style"
)

const (
	// number of problems per log to output
	numberOfProblems = 10
)

var (
	nodeName string
	// followLogs triggers tail -f mode
	followLogs bool
	// numberOfLines is how many lines to output, set via -n
	numberOfLines int
	// showProblems only shows lines that match known issues
	showProblems bool
	// fileOutput is where to write logs to. If omitted, writes to stdout.
	fileOutput string
	// auditLogs only shows the audit logs
	auditLogs bool
	// lastStartOnly shows logs from last start
	lastStartOnly bool
)

// logsCmd represents the logs command
var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Returns logs to debug a local Kubernetes cluster",
	Long:  `Gets the logs of the running instance, used for debugging minikube, not user code.`,
	Run: func(_ *cobra.Command, _ []string) {
		var logOutput *os.File = os.Stdout
		var err error

		if fileOutput != "" {
			logOutput, err = os.Create(fileOutput)
			defer func() {
				err := logOutput.Close()
				if err != nil {
					klog.Warningf("Failed to close file: %v", err)
				}
			}()
			if err != nil {
				exit.Error(reason.Usage, "Failed to create file", err)
			}
		}
		if lastStartOnly {
			err := logs.OutputLastStart()
			if err != nil {
				klog.Errorf("failed to output last start logs: %v", err)
			}
			return
		}
		if auditLogs {
			err := logs.OutputAudit(numberOfLines)
			if err != nil {
				klog.Errorf("failed to output audit logs: %v", err)
			}
			return
		}
		logs.OutputOffline(numberOfLines, logOutput)

		if shouldSilentFail() {
			return
		}

		co := mustload.Running(ClusterFlagValue())

		bs, err := cluster.Bootstrapper(co.API, viper.GetString(cmdcfg.Bootstrapper), *co.Config, co.CP.Runner)
		if err != nil {
			exit.Error(reason.InternalBootstrapper, "Error getting cluster bootstrapper", err)
		}

		cr, err := cruntime.New(cruntime.Config{Type: co.Config.KubernetesConfig.ContainerRuntime, Runner: co.CP.Runner})
		if err != nil {
			exit.Error(reason.InternalNewRuntime, "Unable to get runtime", err)
		}
		if followLogs {
			err := logs.Follow(cr, bs, *co.Config, co.CP.Runner, logOutput)
			if err != nil {
				exit.Error(reason.InternalLogFollow, "Follow", err)
			}
			return
		}
		if showProblems {
			problems := logs.FindProblems(cr, bs, *co.Config, co.CP.Runner)
			logs.OutputProblems(problems, numberOfProblems, logOutput)
			return
		}
		logs.Output(cr, bs, *co.Config, co.CP.Runner, numberOfLines, logOutput)
		if fileOutput != "" {
			out.Styled(style.Success, "Logs file created ({{.logPath}}), remember to include it when reporting issues!", out.V{"logPath": fileOutput})
		}
	},
}

// shouldSilentFail returns true if the user specifies the --file flag and the host isn't running
// This is to prevent outputting the message 'The control plane node must be running for this command' which confuses
// many users while gathering logs to report their issue as the message makes them think the log file wasn't generated
func shouldSilentFail() bool {
	if fileOutput == "" {
		return false
	}

	api, cc := mustload.Partial(ClusterFlagValue())

	cp, err := config.ControlPlane(*cc)
	if err != nil {
		return false
	}

	machineName := config.MachineName(*cc, cp)
	hs, err := machine.Status(api, machineName)
	if err != nil {
		return false
	}

	return hs != state.Running.String()
}

func init() {
	logsCmd.Flags().BoolVarP(&followLogs, "follow", "f", false, "Show only the most recent journal entries, and continuously print new entries as they are appended to the journal.")
	logsCmd.Flags().BoolVar(&showProblems, "problems", false, "Show only log entries which point to known problems")
	logsCmd.Flags().IntVarP(&numberOfLines, "length", "n", 60, "Number of lines back to go within the log")
	logsCmd.Flags().StringVar(&nodeName, "node", "", "The node to get logs from. Defaults to the primary control plane.")
	logsCmd.Flags().StringVar(&fileOutput, "file", "", "If present, writes to the provided file instead of stdout.")
	logsCmd.Flags().BoolVar(&auditLogs, "audit", false, "Show only the audit logs")
	logsCmd.Flags().BoolVar(&lastStartOnly, "last-start-only", false, "Show only the last start logs.")
}
