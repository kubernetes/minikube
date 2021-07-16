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

	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/mustload"
	"k8s.io/minikube/pkg/minikube/node"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/reason"
)

type remotePath struct {
	node string
	path string
}

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

		src := newRemotePath(args[0])
		dst := newRemotePath(args[1])
		validateArgs(src, dst)

		co := mustload.Running(ClusterFlagValue())
		var runner command.Runner

		if dst.node != "" {
			runner = remoteCommandRunner(&co, dst.node)
		} else if src.node == "" {
			// if node name not explicitly specfied in both of source and target,
			// consider target is controlpanel node for backward compatibility.
			runner = co.CP.Runner
		} else {
			runner = command.NewExecRunner(false)
		}

		fa := copyableFile(&co, src, dst)
		if err := runner.Copy(fa); err != nil {
			exit.Error(reason.InternalCommandRunner, fmt.Sprintf("Fail to copy file %s", fa.GetSourcePath()), err)
		}
	},
}

func init() {
}

// split path to node name and file path
func newRemotePath(path string) *remotePath {
	// if destination path is not a absolute path, trying to parse with <node>:<abs path> format
	sp := strings.SplitN(path, ":", 2)
	if len(sp) == 2 && len(sp[0]) > 0 && !strings.Contains(sp[0], "/") && strings.HasPrefix(sp[1], "/") {
		return &remotePath{node: sp[0], path: sp[1]}
	}

	return &remotePath{node: "", path: path}
}

func remoteCommandRunner(co *mustload.ClusterController, nodeName string) command.Runner {
	n, _, err := node.Retrieve(*co.Config, nodeName)
	if err != nil {
		exit.Message(reason.GuestNodeRetrieve, "Node {{.nodeName}} does not exist.", out.V{"nodeName": nodeName})
	}

	h, err := machine.GetHost(co.API, *co.Config, *n)
	if err != nil {
		out.ErrLn("%v", errors.Wrap(err, "getting host"))
		os.Exit(1)
	}

	runner, err := machine.CommandRunner(h)
	if err != nil {
		out.ErrLn("%v", errors.Wrap(err, "getting command runner"))
		os.Exit(1)
	}

	return runner
}

func copyableFile(co *mustload.ClusterController, src, dst *remotePath) assets.CopyableFile {
	// get assets.CopyableFile from minikube node
	if src.node != "" {
		runner := remoteCommandRunner(co, src.node)
		f, err := runner.ReadableFile(src.path)
		if err != nil {
			out.ErrLn("%v", errors.Wrapf(err, "getting file from %s node", src.node))
			os.Exit(1)
		}

		return assets.NewBaseCopyableFile(f, pt.Dir(dst.path), pt.Base(dst.path))
	}

	if _, err := os.Stat(src.path); err != nil {
		if os.IsNotExist(err) {
			exit.Message(reason.HostPathMissing, "Cannot find directory {{.path}} for copy", out.V{"path": src})
		} else {
			exit.Error(reason.HostPathStat, "stat failed", err)
		}
	}

	fa, err := assets.NewFileAsset(src.path, pt.Dir(dst.path), pt.Base(dst.path), "0644")
	if err != nil {
		out.ErrLn("%v", errors.Wrap(err, "getting file asset"))
		os.Exit(1)
	}

	return fa
}

func validateArgs(src, dst *remotePath) {
	if src.path == "" {
		exit.Message(reason.Usage, "Source {{.path}} can not be empty", out.V{"path": src.path})
	}

	if dst.path == "" {
		exit.Message(reason.Usage, "Target {{.path}} can not be empty", out.V{"path": dst.path})
	}

	// if node name not explicitly specfied in both of source and target,
	// consider target node is controlpanel for backward compatibility.
	if src.node == "" && dst.node == "" && !strings.HasPrefix(dst.path, "/") {
		exit.Message(reason.Usage, `Target <remote file path> must be an absolute Path. Relative Path is not allowed (example: "minikube:/home/docker/copied.txt")`)
	}
}
