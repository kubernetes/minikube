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
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/out"
)

var (
	longOutFlag  bool
	rootOnlyFlag bool
)

var decimalUnitPrefixMap = map[int]string{
	0: "",
	1: "K",
	2: "M",
	3: "G",
	4: "T",
	5: "P",
	6: "E",
	7: "Z",
	8: "Y",
}

// runDiskUsage represents the disk-usage command
var diskUsageCmd = &cobra.Command{
	Use:   "disk-usage",
	Short: "Shows current disk usage of minikube",
	Long:  "Shows current disk usage of minikube",
	Run:   runDiskUsage,
}

func runDiskUsage(cmd *cobra.Command, args []string) {
	minikubeDirs := []string{dirs[0]}
	if !rootOnlyFlag {
		minikubeSubDirs, err := os.ReadDir(dirs[0])
		if err != nil {
			klog.Errorf("Error reading minikube directories:\n%s", err)
		}
		for _, file := range minikubeSubDirs {
			if file.IsDir() {
				minikubeDirs = append(minikubeDirs, fmt.Sprintf("%s/%s", dirs[0], file.Name()))
			}
		}
	}
	for _, dir := range minikubeDirs {
		dirNameList := strings.Split(dir, "/")
		dirName := dirNameList[len(dirNameList)-1]
		info, err := os.Lstat(dir)
		if err != nil {
			klog.Errorf("Error reading info about configured minikube path:\n%s", err)
			os.Exit(1)
		}
		totalFileSize := diskUsage(dir, info)
		totalFileSizeOut := strconv.FormatInt(diskUsage(dir, info), 10)

		var decimalUnitPrefixMapKey int
		if !longOutFlag {
			totalFileSizeOut, decimalUnitPrefixMapKey = int64ToFloat32FormattedString(totalFileSize)
		}

		out.Infof("{{.directory}} - {{.diskSize}}{{.decimalUnitPrefix}}B", out.V{"directory": dirName, "diskSize": totalFileSizeOut, "decimalUnitPrefix": decimalUnitPrefixMap[decimalUnitPrefixMapKey]})
	}
}

func int64ToFloat32FormattedString(in_int int64) (string, int) {
	decimalUnitPrefixMapKey := 0
	totalFileSizeFloat := float32(in_int)
	for totalFileSizeFloat > float32(1000) {
		totalFileSizeFloat = totalFileSizeFloat / 1024
		decimalUnitPrefixMapKey++
	}
	return fmt.Sprintf("%.2f", totalFileSizeFloat), decimalUnitPrefixMapKey
}

func diskUsage(currPath string, info os.FileInfo) int64 {
	var size int64

	dir, err := os.Open(currPath)
	if err != nil {
		klog.Errorf("Error opening configured minikube path:\n%s", err)
	}
	defer dir.Close()

	files, err := dir.Readdir(-1)
	if err != nil {
		klog.Errorf("Error reading files minikube path:\n%s", err)
		os.Exit(1)
	}

	for _, file := range files {
		if file.IsDir() {
			size += diskUsage(fmt.Sprintf("%s/%s", currPath, file.Name()), file)
		} else {
			size += file.Size()
		}
	}

	return size
}

func init() {
	diskUsageCmd.Flags().BoolVarP(&longOutFlag, "long", "l", false, "Whether to return a human readable formatted number, e.g. 1GB")
	diskUsageCmd.Flags().BoolVarP(&rootOnlyFlag, "root-only", "r", false, "Only show total of root minikube file instead of breaking out all sub directories")
}
