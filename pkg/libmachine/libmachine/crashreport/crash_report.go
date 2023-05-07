/*
Copyright 2023 The Kubernetes Authors All rights reserved.

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

package crashreport

import (
	"fmt"
	"os"
	"runtime"

	"bytes"

	"os/exec"

	"path/filepath"

	"errors"

	"io/ioutil"

	"github.com/bugsnag/bugsnag-go"
	"k8s.io/minikube/pkg/libmachine/libmachine/log"
	"k8s.io/minikube/pkg/libmachine/libmachine/shell"
	"k8s.io/minikube/pkg/libmachine/version"
)

const (
	defaultAPIKey  = "a9697f9a010c33ee218a65e5b1f3b0c1"
	noreportAPIKey = "no-report"
)

type CrashReporter interface {
	Send(err CrashError) error
}

// CrashError describes an error that should be reported to bugsnag
type CrashError struct {
	Cause       error
	Command     string
	Context     string
	DriverName  string
	LogFilePath string
}

func (e CrashError) Error() string {
	return e.Cause.Error()
}

type BugsnagCrashReporter struct {
	baseDir string
	apiKey  string
}

// NewCrashReporter creates a new bugsnag based CrashReporter. Needs an apiKey.
var NewCrashReporter = func(baseDir string, apiKey string) CrashReporter {
	if apiKey == "" {
		apiKey = defaultAPIKey
	}

	return &BugsnagCrashReporter{
		baseDir: baseDir,
		apiKey:  apiKey,
	}
}

// Send sends a crash report to bugsnag via an http call.
func (r *BugsnagCrashReporter) Send(err CrashError) error {
	if r.noReportFileExist() || r.apiKey == noreportAPIKey {
		log.Debug("Opting out of crash reporting.")
		return nil
	}

	if r.apiKey == "" {
		return errors.New("no api key has been set")
	}

	bugsnag.Configure(bugsnag.Configuration{
		APIKey: r.apiKey,
		// XXX we need to abuse bugsnag metrics to get the OS/ARCH information as a usable filter
		// Can do that with either "stage" or "hostname"
		ReleaseStage:    fmt.Sprintf("%s (%s)", runtime.GOOS, runtime.GOARCH),
		ProjectPackages: []string{"k8s.io/minikube/pkg/libmachine/[^v]*"},
		AppVersion:      version.FullVersion(),
		Synchronous:     true,
		PanicHandler:    func() {},
		Logger:          new(logger),
	})

	metaData := bugsnag.MetaData{}

	metaData.Add("app", "compiler", fmt.Sprintf("%s (%s)", runtime.Compiler, runtime.Version()))
	metaData.Add("device", "os", runtime.GOOS)
	metaData.Add("device", "arch", runtime.GOARCH)

	detectRunningShell(&metaData)
	detectUname(&metaData)
	detectOSVersion(&metaData)
	addFile(err.LogFilePath, &metaData)

	var buffer bytes.Buffer
	for _, message := range log.History() {
		buffer.WriteString(message + "\n")
	}
	metaData.Add("history", "trace", buffer.String())

	return bugsnag.Notify(err.Cause, metaData, bugsnag.SeverityError, bugsnag.Context{String: err.Context}, bugsnag.ErrorClass{Name: fmt.Sprintf("%s/%s", err.DriverName, err.Command)})
}

func (r *BugsnagCrashReporter) noReportFileExist() bool {
	optOutFilePath := filepath.Join(r.baseDir, "no-error-report")
	if _, err := os.Stat(optOutFilePath); os.IsNotExist(err) {
		return false
	}
	return true
}

func addFile(path string, metaData *bugsnag.MetaData) {
	if path == "" {
		return
	}
	file, err := os.Open(path)
	if err != nil {
		log.Debug(err)
		return
	}
	data, err := ioutil.ReadAll(file)
	if err != nil {
		log.Debug(err)
		return
	}
	metaData.Add("logfile", filepath.Base(path), string(data))
}

func detectRunningShell(metaData *bugsnag.MetaData) {
	shell, err := shell.Detect()
	if err == nil {
		metaData.Add("device", "shell", shell)
	}
}

func detectUname(metaData *bugsnag.MetaData) {
	cmd := exec.Command("uname", "-s")
	output, err := cmd.Output()
	if err != nil {
		return
	}
	metaData.Add("device", "uname", string(output))
}

func detectOSVersion(metaData *bugsnag.MetaData) {
	metaData.Add("device", "os version", localOSVersion())
}
