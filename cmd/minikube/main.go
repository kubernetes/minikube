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
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"

	"github.com/spf13/pflag"
	"k8s.io/klog/v2"

	// Register drivers
	"k8s.io/minikube/pkg/minikube/localpath"
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

	setFlags()

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

// bridgeLogMessages bridges non-glog logs into klog
func bridgeLogMessages() {
	log.SetFlags(log.Lshortfile)
	log.SetOutput(stdLogBridge{})
	mlog.SetErrWriter(machineLogBridge{})
	mlog.SetOutWriter(machineLogBridge{})
	mlog.SetDebug(true)
}

type stdLogBridge struct{}

// Write parses the standard logging line and passes its components to klog
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

// Write passes machine driver logs to klog
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

// setFlags sets the flags
func setFlags() {
	// parse flags beyond subcommand - get aroung go flag 'limitations':
	// "Flag parsing stops just before the first non-flag argument" (ref: https://pkg.go.dev/flag#hdr-Command_line_flag_syntax)
	pflag.CommandLine.ParseErrorsWhitelist.UnknownFlags = true
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	// avoid 'pflag: help requested' error, as help will be defined later by cobra cmd.Execute()
	pflag.BoolP("help", "h", false, "")
	pflag.Parse()

	// set default flag value for logtostderr and alsologtostderr but don't override user's preferences
	if !pflag.CommandLine.Changed("logtostderr") {
		if err := pflag.Set("logtostderr", "false"); err != nil {
			klog.Warningf("Unable to set default flag value for logtostderr: %v", err)
		}
	}
	if !pflag.CommandLine.Changed("alsologtostderr") {
		if err := pflag.Set("alsologtostderr", "false"); err != nil {
			klog.Warningf("Unable to set default flag value for alsologtostderr: %v", err)
		}
	}
	if os.Args[1] == "start" {
		fp := localpath.LastStartLog()
		if err := os.Remove(fp); err != nil {
			klog.Warningf("Unable to delete file %s: %v", err)
		}
		if !pflag.CommandLine.Changed("log_file") {
			if err := pflag.Set("log_file", fp); err != nil {
				klog.Warningf("Unable to set default flag value for log_file: %v", err)
			}
		}
	}

	// make sure log_dir exists if log_file is not also set - the log_dir is mutually exclusive with the log_file option
	// ref: https://github.com/kubernetes/klog/blob/52c62e3b70a9a46101f33ebaf0b100ec55099975/klog.go#L491
	if pflag.Lookup("log_file") != nil && pflag.Lookup("log_file").Value.String() == "" &&
		pflag.Lookup("log_dir") != nil && pflag.Lookup("log_dir").Value.String() != "" {
		if err := os.MkdirAll(pflag.Lookup("log_dir").Value.String(), 0755); err != nil {
			klog.Warningf("unable to create log directory: %v", err)
		}
	}
}
