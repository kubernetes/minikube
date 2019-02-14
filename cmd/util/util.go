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

// package util is a hodge-podge of utility functions that should be moved elsewhere.
package util

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
)

type ServiceContext struct {
	Service string `json:"service"`
	Version string `json:"version"`
}

type LookPath func(filename string) (string, error)

var lookPath LookPath

func init() {
	lookPath = exec.LookPath
}

func MaybePrintKubectlDownloadMsg(goos string, out io.Writer) {
	if !viper.GetBool(config.WantKubectlDownloadMsg) {
		return
	}

	verb := "run"
	installInstructions := "curl -Lo kubectl https://storage.googleapis.com/kubernetes-release/release/%s/bin/%s/%s/kubectl && chmod +x kubectl && sudo cp kubectl /usr/local/bin/ && rm kubectl"
	if goos == "windows" {
		verb = "do"
		installInstructions = `download kubectl from:
https://storage.googleapis.com/kubernetes-release/release/%s/bin/%s/%s/kubectl.exe
Add kubectl to your system PATH`
	}

	_, err := lookPath("kubectl")
	if err != nil && goos == "windows" {
		_, err = lookPath("kubectl.exe")
	}
	if err != nil {
		fmt.Fprintf(out,
			`========================================
kubectl could not be found on your path. kubectl is a requirement for using minikube
To install kubectl, please %s the following:

%s

To disable this message, run the following:

minikube config set WantKubectlDownloadMsg false
========================================
`,
			verb, fmt.Sprintf(installInstructions, constants.DefaultKubernetesVersion, goos, runtime.GOARCH))
	}
}

// Ask the kernel for a free open port that is ready to use
func GetPort() (string, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return "", errors.Errorf("Error accessing port %d", addr.Port)
	}
	defer l.Close()
	return strconv.Itoa(l.Addr().(*net.TCPAddr).Port), nil
}

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

func GetKubeConfigPath() string {
	kubeConfigEnv := os.Getenv(constants.KubeconfigEnvVar)
	if kubeConfigEnv == "" {
		return constants.KubeconfigPath
	}
	return filepath.SplitList(kubeConfigEnv)[0]
}
