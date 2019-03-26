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

// Package util is a hodge-podge of utility functions that should be moved elsewhere.
package util

import (
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strconv"

	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/constants"
)

// GetPort asks the kernel for a free open port that is ready to use
func GetPort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return -1, errors.Errorf("Error accessing port %d", addr.Port)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

// KillMountProcess kills the mount process, if it is running
func KillMountProcess() error {
	out, err := ioutil.ReadFile(filepath.Join(constants.GetMinipath(), constants.MountProcessFileName))
	if err != nil {
		return nil // no mount process to kill
	}
	pid, err := strconv.Atoi(string(out))
	if err != nil {
		return errors.Wrap(err, "error converting mount string to pid")
	}
	mountProc, err := os.FindProcess(pid)
	if err != nil {
		return errors.Wrap(err, "error converting mount string to pid")
	}
	return mountProc.Kill()
}

// GetKubeConfigPath gets the path to the first kubeconfig
func GetKubeConfigPath() string {
	kubeConfigEnv := os.Getenv(constants.KubeconfigEnvVar)
	if kubeConfigEnv == "" {
		return constants.KubeconfigPath
	}
	return filepath.SplitList(kubeConfigEnv)[0]
}
