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

	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/out/register"
)

var (
	// currentLogFile the file that's used to store audit logs
	currentLogFile *os.File

	// auditOverrideFilename overrides the default audit log filename, used for testing purposes
	auditOverrideFilename string
)

// openAuditLog opens the audit log file or creates it if it doesn't exist.
func openAuditLog() error {
	// this is so we can manually set the log file for tests
	if currentLogFile != nil {
		return nil
	}
	f, err := os.OpenFile(auditPath(), os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("failed to open the audit log: %v", err)
	}
	currentLogFile = f
	return nil
}

// closeAuditLog closes the audit log file
func closeAuditLog() {
	if err := currentLogFile.Close(); err != nil {
		klog.Errorf("failed to close the audit log: %v", err)
	}
	currentLogFile = nil
}

// appendToLog appends the row to the log file.
func appendToLog(row *row) error {
	ce := register.CloudEvent(row, row.toMap())
	bs, err := ce.MarshalJSON()
	if err != nil {
		return fmt.Errorf("error marshalling event: %v", err)
	}
	if err := openAuditLog(); err != nil {
		return err
	}
	defer closeAuditLog()
	if _, err := currentLogFile.WriteString(string(bs) + "\n"); err != nil {
		return fmt.Errorf("unable to write to audit log: %v", err)
	}
	return nil
}

// truncateAuditLog truncates the audit log file
func truncateAuditLog() error {
	if err := os.Truncate(auditPath(), 0); err != nil {
		return fmt.Errorf("failed to truncate audit log: %v", err)
	}
	return nil
}

func auditPath() string {
	if auditOverrideFilename != "" {
		return auditOverrideFilename
	}
	return localpath.AuditLog()
}
