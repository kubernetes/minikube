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
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"k8s.io/minikube/pkg/minikube/localpath"
)

func TestLogFile(t *testing.T) {
	t.Run("OpenAuditLog", func(t *testing.T) {
		// make sure logs directory exists
		if err := os.MkdirAll(filepath.Dir(localpath.AuditLog()), 0755); err != nil {
			t.Fatalf("Error creating logs directory: %v", err)
		}
		if err := openAuditLog(); err != nil {
			t.Fatal(err)
		}
		closeAuditLog()
	})

	t.Run("AppendToLog", func(t *testing.T) {
		f, err := os.CreateTemp("", "audit.json")
		if err != nil {
			t.Fatalf("Error creating temporary file: %v", err)
		}
		defer os.Remove(f.Name())

		currentLogFile = f
		defer closeAuditLog()

		r := newRow("start", "-v", "user1", "v0.17.1", time.Now(), uuid.New().String())
		if err := appendToLog(r); err != nil {
			t.Fatalf("Error appendingToLog: %v", err)
		}

		currentLogFile, err = os.Open(f.Name())
		if err != nil {
			t.Fatal(err)
		}
		b := make([]byte, 100)
		if _, err := currentLogFile.Read(b); err != nil && err != io.EOF {
			t.Errorf("Log was not appended to file: %v", err)
		}
	})
}
