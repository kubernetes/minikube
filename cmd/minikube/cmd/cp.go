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

	"os"
	pt "path"
	"path/filepath"

	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/mustload"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/reason"
)

// placeholders for flag values
var (
	srcPath string
	dstPath string
)

// cpCmd represents the cp command, similar to docker cp
var cpCmd = &cobra.Command{
	Use:   "cp <source file path> <target file absolute path>",
	Short: "Copy the specified file into minikube",
	Long: "Copy the specified file into minikube, it will be saved at path <target file absolute path> in your minikube.\n" +
		"Example Command : \"minikube cp a.txt /home/docker/b.txt\"\n",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 2 {
			exit.Message(reason.Usage, `Please specify the path to copy: 
	minikube cp <source file path> <target file absolute path> (example: "minikube cp a/b.txt /copied.txt")`)
		}

		srcPath = args[0]
		dstPath = args[1]
		validateArgs(srcPath, dstPath)

		co := mustload.Running(ClusterFlagValue())
		fa, err := assets.NewFileAsset(srcPath, pt.Dir(dstPath), pt.Base(dstPath), "0644")
		if err != nil {
			out.ErrLn("%v", errors.Wrap(err, "getting file asset"))
			os.Exit(1)
		}

		if err = co.CP.Runner.Copy(fa); err != nil {
			out.ErrLn("%v", errors.Wrap(err, "copying file"))
			os.Exit(1)
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

	if !filepath.IsAbs(dstPath) {
		exit.Message(reason.Usage, `<target file absolute path> must be an absolute Path. Relative Path is not allowed (example: "/home/docker/copied.txt")`)
	}
}
