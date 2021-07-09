/*
Copyright 2021 The Kubernetes Authors All rights reserved.

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
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"fmt"
	"os"
	pt "path"
	"strings"

	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/mustload"
	"k8s.io/minikube/pkg/minikube/node"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/reason"
)

// placeholders for flag values
var (
	srcPath string
	dstPath string
	dstNode string
)

// cpCmd represents the cp command, similar to docker cp
var cpCmd = &cobra.Command{
	Use:   "cp <source file path> <target node name>:<target file absolute path>",
	Short: "Copy the specified file into minikube",
	Long: "Copy the specified file into minikube, it will be saved at path <target file absolute path> in your minikube.\n" +
		"Example Command : \"minikube cp a.txt /home/docker/b.txt\"\n" +
		"                  \"minikube cp a.txt minikube-m02:/home/docker/b.txt\"\n",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 2 {
			exit.Message(reason.Usage, `Please specify the path to copy: 
	minikube cp <source file path> <target file absolute path> (example: "minikube cp a/b.txt /copied.txt")`)
		}

		srcPath = args[0]
		dstPath = args[1]

		// if destination path is not a absolute path, trying to parse with <node>:<abs path> format
		if !strings.HasPrefix(dstPath, "/") {
			if sp := strings.SplitN(dstPath, ":", 2); len(sp) == 2 {
				dstNode = sp[0]
				dstPath = sp[1]
			}
		}

		validateArgs(srcPath, dstPath)

		fa, err := assets.NewFileAsset(srcPath, pt.Dir(dstPath), pt.Base(dstPath), "0644")
		if err != nil {
			out.ErrLn("%v", errors.Wrap(err, "getting file asset"))
			os.Exit(1)
		}
		defer func() {
			if err := fa.Close(); err != nil {
				klog.Warningf("error closing the file %s: %v", fa.GetSourcePath(), err)
			}
		}()

		co := mustload.Running(ClusterFlagValue())
		var runner command.Runner
		if dstNode == "" {
			runner = co.CP.Runner
		} else {
			n, _, err := node.Retrieve(*co.Config, dstNode)
			if err != nil {
				exit.Message(reason.GuestNodeRetrieve, "Node {{.nodeName}} does not exist.", out.V{"nodeName": dstNode})
			}

			h, err := machine.GetHost(co.API, *co.Config, *n)
			if err != nil {
				exit.Error(reason.GuestLoadHost, "Error getting host", err)
			}

			runner, err = machine.CommandRunner(h)
			if err != nil {
				exit.Error(reason.InternalCommandRunner, "Failed to get command runner", err)
			}
		}

		if err = runner.Copy(fa); err != nil {
			exit.Error(reason.InternalCommandRunner, fmt.Sprintf("Fail to copy file %s", fa.GetSourcePath()), err)
		}
	},
}

func init() {
}

func validateArgs(srcPath string, dstPath string) {
	if srcPath == "" {
		exit.Message(reason.Usage, "Source {{.path}} can not be empty", out.V{"path": srcPath})
	}

	if dstPath == "" {
		exit.Message(reason.Usage, "Target {{.path}} can not be empty", out.V{"path": dstPath})
	}

	if _, err := os.Stat(srcPath); err != nil {
		if os.IsNotExist(err) {
			exit.Message(reason.HostPathMissing, "Cannot find directory {{.path}} for copy", out.V{"path": srcPath})
		} else {
			exit.Error(reason.HostPathStat, "stat failed", err)
		}
	}

	if !strings.HasPrefix(dstPath, "/") {
		exit.Message(reason.Usage, `<target file absolute path> must be an absolute Path. Relative Path is not allowed (example: "/home/docker/copied.txt")`)
	}
}
