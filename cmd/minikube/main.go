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
	"regexp"
	"strconv"

	// initflag must be imported before any other minikube pkg.
	// Fix for https://github.com/kubernetes/minikube/issues/4866

	"k8s.io/klog/v2"
	_ "k8s.io/minikube/pkg/initflag"

	// Register drivers
	_ "k8s.io/minikube/pkg/minikube/registry/drvs"

	// Force exp dependency
	_ "golang.org/x/exp/ebnf"

	mlog "github.com/docker/machine/libmachine/log"

	"github.com/google/slowjam/pkg/stacklog"
	"github.com/pkg/profile"

	"k8s.io/minikube/cmd/minikube/cmd"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/out"
	_ "k8s.io/minikube/pkg/provision"
)

const minikubeEnableProfile = "MINIKUBE_ENABLE_PROFILING"

var (
	// This regex is intentionally very specific, it's supposed to surface
	// unexpected errors from libmachine to the user.
	machineLogErrorRe   = regexp.MustCompile(`VirtualizationException`)
	machineLogWarningRe = regexp.MustCompile(`(?i)warning`)
)

func main() {
	bridgeLogMessages()
	defer klog.Flush()

	s := stacklog.MustStartFromEnv("STACKLOG_PATH")
	defer s.Stop()

	if os.Getenv(minikubeEnableProfile) == "1" {
		defer profile.Start(profile.TraceProfile).Stop()
	}
	if os.Getenv(constants.IsMinikubeChildProcess) == "" {
		machine.StartDriver()
	}
	out.SetOutFile(os.Stdout)
	out.SetErrFile(os.Stderr)
	cmd.Execute()
}

// bridgeLogMessages bridges non-glog logs into glog
func bridgeLogMessages() {
	log.SetFlags(log.Lshortfile)
	log.SetOutput(stdLogBridge{})
	mlog.SetErrWriter(machineLogBridge{})
	mlog.SetOutWriter(machineLogBridge{})
	mlog.SetDebug(true)
}

type stdLogBridge struct{}

// Write parses the standard logging line and passes its components to glog
func (lb stdLogBridge) Write(b []byte) (n int, err error) {
	// Split "d.go:23: message" into "d.go", "23", and "message".
	parts := bytes.SplitN(b, []byte{':'}, 3)
	if len(parts) != 3 || len(parts[0]) < 1 || len(parts[2]) < 1 {
		klog.Errorf("bad log format: %s", b)
		return
	}

	file := string(parts[0])
	text := string(parts[2][1:]) // skip leading space
	line, err := strconv.Atoi(string(parts[1]))
	if err != nil {
		text = fmt.Sprintf("bad line number: %s", b)
		line = 0
	}
	klog.Infof("stdlog: %s:%d %s", file, line, text)
	return len(b), nil
}

// libmachine log bridge
type machineLogBridge struct{}

// Write passes machine driver logs to glog
func (lb machineLogBridge) Write(b []byte) (n int, err error) {
	if machineLogErrorRe.Match(b) {
		klog.Errorf("libmachine: %s", b)
	} else if machineLogWarningRe.Match(b) {
		klog.Warningf("libmachine: %s", b)
	} else {
		klog.Infof("libmachine: %s", b)
	}
	return len(b), nil
}
