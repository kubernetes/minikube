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
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/pflag"
	"k8s.io/klog/v2"

	"k8s.io/minikube/pkg/minikube/localpath"

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

	dconfig "github.com/docker/cli/cli/config"
	ddocker "github.com/docker/cli/cli/context/docker"
	dstore "github.com/docker/cli/cli/context/store"
)

const minikubeEnableProfile = "MINIKUBE_ENABLE_PROFILING"

var (
	// This regex is intentionally very specific, it's supposed to surface
	// unexpected errors from libmachine to the user.
	machineLogErrorRe   = regexp.MustCompile(`VirtualizationException`)
	machineLogWarningRe = regexp.MustCompile(`(?i)warning`)
	// This regex is to filter out logs that contain environment variables which could contain sensitive information
	machineLogEnvironmentRe = regexp.MustCompile(`&exec\.Cmd`)
)

func main() {
	bridgeLogMessages()
	defer klog.Flush()

	propagateDockerContextToEnv()

	// Don't parse flags when running as kubectl
	_, callingCmd := filepath.Split(os.Args[0])
	callingCmd = strings.TrimSuffix(callingCmd, ".exe")
	parse := callingCmd != "kubectl"
	setFlags(parse)

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
	if machineLogEnvironmentRe.Match(b) {
		return len(b), nil
	} else if machineLogErrorRe.Match(b) {
		klog.Errorf("libmachine: %s", b)
	} else if machineLogWarningRe.Match(b) {
		klog.Warningf("libmachine: %s", b)
	} else {
		klog.Infof("libmachine: %s", b)
	}
	return len(b), nil
}

// checkLogFileMaxSize checks if a file's size is greater than or equal to max size in KB
func checkLogFileMaxSize(file string, maxSizeKB int64) bool {
	f, err := os.Stat(file)
	if err != nil {
		return false
	}
	kb := (f.Size() / 1024)
	return kb >= maxSizeKB
}

// logFileName generates a default logfile name in the form minikube_<argv[1]>_<hash>_<count>.log from args
func logFileName(dir string, logIdx int64) string {
	h := sha1.New()
	user, err := user.Current()
	if err != nil {
		klog.Warningf("Unable to get username to add to log filename hash: %v", err)
	} else {
		_, err := h.Write([]byte(user.Username))
		if err != nil {
			klog.Warningf("Unable to add username %s to log filename hash: %v", user.Username, err)
		}
	}
	for _, s := range pflag.Args() {
		if _, err := h.Write([]byte(s)); err != nil {
			klog.Warningf("Unable to add arg %s to log filename hash: %v", s, err)
		}
	}
	hs := hex.EncodeToString(h.Sum(nil))
	var logfilePath string
	// check if subcommand specified
	if len(pflag.Args()) < 1 {
		logfilePath = filepath.Join(dir, fmt.Sprintf("minikube_%s_%d.log", hs, logIdx))
	} else {
		logfilePath = filepath.Join(dir, fmt.Sprintf("minikube_%s_%s_%d.log", pflag.Arg(0), hs, logIdx))
	}
	// if log has reached max size 1M, generate new logfile name by incrementing count
	if checkLogFileMaxSize(logfilePath, 1024) {
		return logFileName(dir, logIdx+1)
	}
	return logfilePath
}

// setFlags sets the flags
func setFlags(parse bool) {
	// parse flags beyond subcommand - get around go flag 'limitations':
	// "Flag parsing stops just before the first non-flag argument" (ref: https://pkg.go.dev/flag#hdr-Command_line_flag_syntax)
	pflag.CommandLine.ParseErrorsWhitelist.UnknownFlags = true
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	// avoid 'pflag: help requested' error, as help will be defined later by cobra cmd.Execute()
	pflag.BoolP("help", "h", false, "")
	if parse {
		pflag.Parse()
	}

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
	setLastStartFlags()

	// set default log_file name but don't override user's preferences
	if !pflag.CommandLine.Changed("log_file") {
		// default log_file dir to $TMP
		dir := os.TempDir()
		// set log_dir to user input if specified
		if pflag.CommandLine.Changed("log_dir") && pflag.Lookup("log_dir").Value.String() != "" {
			dir = pflag.Lookup("log_dir").Value.String()
		}
		l := logFileName(dir, 0)
		if err := pflag.Set("log_file", l); err != nil {
			klog.Warningf("Unable to set default flag value for log_file: %v", err)
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

// setLastStartFlags sets the log_file flag to lastStart.txt if start command and user doesn't specify log_file or log_dir flags.
func setLastStartFlags() {
	if pflag.Arg(0) != "start" {
		return
	}
	if pflag.CommandLine.Changed("log_file") || pflag.CommandLine.Changed("log_dir") {
		return
	}
	fp := localpath.LastStartLog()
	dp := filepath.Dir(fp)
	if err := os.MkdirAll(dp, 0755); err != nil {
		klog.Warningf("Unable to make log dir %s: %v", dp, err)
	}
	if _, err := os.Create(fp); err != nil {
		klog.Warningf("Unable to create/truncate file %s: %v", fp, err)
	}
	if err := pflag.Set("log_file", fp); err != nil {
		klog.Warningf("Unable to set default flag value for log_file: %v", err)
	}
}

// propagateDockerContextToEnv propagates the current context in ~/.docker/config.json to $DOCKER_HOST,
// so that google/go-containerregistry can pick it up.
func propagateDockerContextToEnv() {
	if os.Getenv("DOCKER_HOST") != "" {
		// Already explicitly set
		return
	}
	currentContext := os.Getenv("DOCKER_CONTEXT")
	if currentContext == "" {
		dockerConfigDir := dconfig.Dir()
		if _, err := os.Stat(dockerConfigDir); err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				klog.Warning(err)
			}
			return
		}
		cf, err := dconfig.Load(dockerConfigDir)
		if err != nil {
			klog.Warningf("Unable to load the current Docker config from %q: %v", dockerConfigDir, err)
			return
		}
		currentContext = cf.CurrentContext
	}
	if currentContext == "" {
		return
	}
	storeConfig := dstore.NewConfig(
		func() interface{} { return &ddocker.EndpointMeta{} },
		dstore.EndpointTypeGetter(ddocker.DockerEndpoint, func() interface{} { return &ddocker.EndpointMeta{} }),
	)
	st := dstore.New(dconfig.ContextStoreDir(), storeConfig)
	md, err := st.GetMetadata(currentContext)
	if err != nil {
		klog.Warningf("Unable to resolve the current Docker CLI context %q: %v", currentContext, err)
		klog.Warningf("Try running `docker context use %s` to resolve the above error", currentContext)
		return
	}
	dockerEP, ok := md.Endpoints[ddocker.DockerEndpoint]
	if !ok {
		// No warning (the context is not for Docker)
		return
	}
	dockerEPMeta, ok := dockerEP.(ddocker.EndpointMeta)
	if !ok {
		klog.Warningf("expected docker.EndpointMeta, got %T", dockerEP)
		return
	}
	if dockerEPMeta.Host != "" {
		os.Setenv("DOCKER_HOST", dockerEPMeta.Host)
	}
}
