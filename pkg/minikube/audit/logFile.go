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
	"fmt"
	"os"

	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/out/register"
)

// currentLogFile the file that's used to store audit logs
var currentLogFile *os.File

// setLogFile sets the logPath and creates the log file if it doesn't exist.
func setLogFile() error {
	lp := localpath.AuditLog()
	f, err := os.OpenFile(lp, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("unable to open %s: %v", lp, err)
	}
	currentLogFile = f
	return nil
}

// appendToLog appends the audit entry to the log file.
func appendToLog(entry *entry) error {
	if currentLogFile == nil {
		if err := setLogFile(); err != nil {
			return err
		}
	}
	e := register.CloudEvent(entry, entry.Data)
	bs, err := e.MarshalJSON()
	if err != nil {
		return fmt.Errorf("error marshalling event: %v", err)
	}
	if _, err := currentLogFile.WriteString(string(bs) + "\n"); err != nil {
		return fmt.Errorf("unable to write to audit log: %v", err)
	}
	return nil
}
