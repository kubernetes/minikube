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

	"github.com/docker/machine/commands"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/shell"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/constants"
)

type shellConfig struct {
	commands.ShellConfig
	comment       string
	commandFormat string
}

// envCmd represents the docker-env command
var dockerEnvCmd = &cobra.Command{
	Use:   "docker-env",
	Short: "sets up docker env variables; similar to '$(docker-machine env)'",
	Long:  `sets up docker env variables; similar to '$(docker-machine env)'`,
	Run: func(cmd *cobra.Command, args []string) {
		api := libmachine.NewClient(constants.Minipath, constants.MakeMiniPath("certs"))
		defer api.Close()

		envMap, err := cluster.GetHostDockerEnv(api)
		if err != nil {
			glog.Errorln("Error setting machine env variable(s):", err)
			os.Exit(1)
		}
		fmt.Fprintln(os.Stdout, buildDockerEnvShellOutput(envMap))
	},
}

func buildDockerEnvShellOutput(envMap map[string]string) string {
	shellConfig, err := getShellConfig()
	if err != nil {
		glog.Errorln("Error discovering user shell:", err)
		os.Exit(1)
	}
	output := ""
	for envName, envVal := range envMap {
		output += fmt.Sprintf("%s%s%s%s%s",
			shellConfig.Prefix,
			envName,
			shellConfig.Delimiter,
			envVal,
			shellConfig.Suffix,
		)
	}
	cmd := fmt.Sprintf(shellConfig.commandFormat, "minikube docker-env")
	howToRun := fmt.Sprintf("%s Run this command to configure your shell: \n%s %s\n", shellConfig.comment, shellConfig.comment, cmd)
	return output + howToRun
}

func getShellConfig() (shellConfig, error) {
	shellConfig := shellConfig{comment: "#"}
	userShell, err := shell.Detect()
	if err != nil {
		return shellConfig, err
	}
	switch userShell {
	case "fish":
		shellConfig.Prefix = "set -gx "
		shellConfig.Suffix = "\";\n"
		shellConfig.Delimiter = " \""
		shellConfig.commandFormat = "eval (%s)"
	case "powershell":
		shellConfig.Prefix = "$Env:"
		shellConfig.Suffix = "\"\n"
		shellConfig.Delimiter = " = \""
		shellConfig.commandFormat = "& %s | Invoke-Expression"
	case "cmd":
		shellConfig.Prefix = "SET "
		shellConfig.Suffix = "\n"
		shellConfig.Delimiter = "="
		shellConfig.comment = "REM"
		shellConfig.commandFormat = "\t@FOR /f \"tokens=*\" %%i IN ('%s') DO @%%i"
	case "tcsh":
		shellConfig.Prefix = "setenv "
		shellConfig.Suffix = "\";\n"
		shellConfig.Delimiter = " \""
		shellConfig.Delimiter = "\" \""
		shellConfig.comment = ":"
		shellConfig.commandFormat = "eval `%s`"
	case "emacs":
		shellConfig.Prefix = "(setenv \""
		shellConfig.Suffix = "\")\n"
		shellConfig.Delimiter = "\" \""
		shellConfig.comment = ";;"
		shellConfig.commandFormat = "(with-temp-buffer (shell-command \"%s\" (current-buffer)) (eval-buffer))"
	default:
		shellConfig.Prefix = "export "
		shellConfig.Suffix = "\"\n"
		shellConfig.Delimiter = "=\""
		shellConfig.commandFormat = "eval $(%s)"
	}
	return shellConfig, nil
}

func init() {
	RootCmd.AddCommand(dockerEnvCmd)
}
