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

package util

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"strconv"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh/terminal"
	minikubeConfig "k8s.io/minikube/cmd/minikube/cmd/config"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/version"
)

type ServiceContext struct {
	Service string `json:"service"`
	Version string `json:"version"`
}

type Message struct {
	Message        string `json:"message"`
	ServiceContext `json:"serviceContext"`
}

type LookPath func(filename string) (string, error)

var lookPath LookPath

func init() {
	lookPath = exec.LookPath
}

func ReportError(err error, url string) error {
	errMsg, err := FormatError(err)
	if err != nil {
		return errors.Wrap(err, "Error formatting error message")
	}
	jsonErrorMsg, err := MarshallError(errMsg, "default", version.GetVersion())
	if err != nil {
		return errors.Wrap(err, "Error marshalling error message to JSON")
	}
	err = UploadError(jsonErrorMsg, url)
	if err != nil {
		return errors.Wrap(err, "Error uploading error message")
	}
	return nil
}

func FormatError(err error) (string, error) {
	if err == nil {
		return "", errors.New("Error: ReportError was called with nil error value")
	}

	type stackTracer interface {
		StackTrace() errors.StackTrace
	}

	errOutput := []string{}
	errOutput = append(errOutput, err.Error())

	if err, ok := err.(stackTracer); ok {
		for _, f := range err.StackTrace() {
			errOutput = append(errOutput, fmt.Sprintf("\tat %n(%v)", f, f))
		}
	} else {
		return "", errors.New("Error msg with no stack trace cannot be reported")
	}
	return strings.Join(errOutput, "\n"), nil
}

func MarshallError(errMsg, service, version string) ([]byte, error) {
	m := Message{errMsg, ServiceContext{service, version}}
	b, err := json.Marshal(m)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}
	return b, nil
}

func UploadError(b []byte, url string) error {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(b))
	if err != nil {
		return errors.Wrap(err, "")
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, "")
	} else if resp.StatusCode != 200 {
		return errors.Errorf("Error sending error report to %s, got response code %d", url, resp.StatusCode)
	}
	return nil
}

func MaybeReportErrorAndExit(errToReport error) {
	MaybeReportErrorAndExitWithCode(errToReport, 1)
}

func MaybeReportErrorAndExitWithCode(errToReport error, returnCode int) {
	var err error
	if viper.GetBool(config.WantReportError) {
		err = ReportError(errToReport, constants.ReportingURL)
	} else if viper.GetBool(config.WantReportErrorPrompt) {
		fmt.Println(
			`================================================================================
An error has occurred. Would you like to opt in to sending anonymized crash
information to minikube to help prevent future errors?
To opt out of these messages, run the command:
	minikube config set WantReportErrorPrompt false
================================================================================`)
		if PromptUserForAccept(os.Stdin) {
			minikubeConfig.Set(config.WantReportError, "true")
			err = ReportError(errToReport, constants.ReportingURL)
		}
	}
	if err != nil {
		glog.Errorf(err.Error())
	}
	os.Exit(returnCode)
}

func getInput(input chan string, r io.Reader) {
	reader := bufio.NewReader(r)
	fmt.Print("Please enter your response [Y/n]: \n")
	response, err := reader.ReadString('\n')
	if err != nil {
		glog.Errorf(err.Error())
	}
	input <- response
}

func PromptUserForAccept(r io.Reader) bool {
	if !terminal.IsTerminal(int(os.Stdout.Fd())) {
		return false
	}
	input := make(chan string, 1)
	go getInput(input, r)
	select {
	case response := <-input:
		response = strings.ToLower(strings.TrimSpace(response))
		if response == "y" || response == "yes" || response == "" {
			return true
		} else if response == "n" || response == "no" {
			return false
		} else {
			fmt.Println("Invalid response, error reporting remains disabled. Must be in form [Y/n]")
			return false
		}
	case <-time.After(30 * time.Second):
		return false
	}
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
