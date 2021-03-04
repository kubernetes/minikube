/*
Copyright 2020 The Kubernetes Authors All rights reserved.

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

package audit

import (
	"os"
	"os/user"
	"strings"
	"time"

	"github.com/spf13/viper"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/version"
)

// userName pulls the user flag, if empty gets the os username.
func userName() string {
	u := viper.GetString(config.UserFlag)
	if u != "" {
		return u
	}
	osUser, err := user.Current()
	if err != nil {
		return "UNKNOWN"
	}
	return osUser.Username
}

// args concats the args into space delimited string.
func args() string {
	// first arg is binary and second is command, anything beyond is a minikube arg
	if len(os.Args) < 3 {
		return ""
	}
	return strings.Join(os.Args[2:], " ")
}

// Log details about the executed command.
func Log(startTime time.Time) {
	if len(os.Args) < 2 || !shouldLog() {
		return
	}
	r := newRow(os.Args[1], args(), userName(), version.GetVersion(), startTime, time.Now())
	if err := appendToLog(r); err != nil {
		klog.Warning(err)
	}
}

// shouldLog returns if the command should be logged.
func shouldLog() bool {
	// in rare chance we get here without a command, don't log
	if len(os.Args) < 2 {
		return false
	}

	if isDeletePurge() {
		return false
	}

	// commands that should not be logged.
	no := []string{"status", "version"}
	a := os.Args[1]
	for _, c := range no {
		if a == c {
			return false
		}
	}
	return true
}

// isDeletePurge return true if command is delete with purge flag.
func isDeletePurge() bool {
	args := os.Args
	if len(args) < 2 {
		return false
	}
	if args[1] != "delete" {
		return false
	}
	for _, a := range args {
		if a == "--purge" {
			return true
		}
	}
	return false
}
