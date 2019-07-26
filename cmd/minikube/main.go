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

package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/golang/glog"
	"github.com/pkg/profile"
	"k8s.io/minikube/cmd/minikube/cmd"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/translate"
	_ "k8s.io/minikube/pkg/provision"
)

const minikubeEnableProfile = "MINIKUBE_ENABLE_PROFILING"

func main() {
	captureStdLogMessages()
	defer glog.Flush()

	if os.Getenv(minikubeEnableProfile) == "1" {
		defer profile.Start(profile.TraceProfile).Stop()
	}
	if os.Getenv(constants.IsMinikubeChildProcess) == "" {
		machine.StartDriver()
	}
	out.SetOutFile(os.Stdout)
	out.SetErrFile(os.Stderr)
	translate.DetermineLocale()
	cmd.Execute()
}

// captureStdLogMessages arranges for messages written to the Go "log" package's to appear in glog
func captureStdLogMessages() {
	log.SetFlags(log.Lshortfile)
	log.SetOutput(logBridge{})
}

type logBridge struct{}

// Write parses the standard logging line and passes its components to glog
func (lb logBridge) Write(b []byte) (n int, err error) {
	// Split "d.go:23: message" into "d.go", "23", and "message".
	parts := bytes.SplitN(b, []byte{':'}, 3)
	if len(parts) != 3 || len(parts[0]) < 1 || len(parts[2]) < 1 {
		glog.Errorf("bad log format: %s", b)
		return
	}

	file := string(parts[0])
	text := string(parts[2][1:]) // skip leading space
	line, err := strconv.Atoi(string(parts[1]))
	if err != nil {
		text = fmt.Sprintf("bad line number: %s", b)
		line = 0
	}
	glog.Infof("stdlog: %s:%d %s", file, line, text)
	return len(b), nil
}
