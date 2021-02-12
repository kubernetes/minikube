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
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestLogFile(t *testing.T) {
	t.Run("SetLogFile", func(t *testing.T) {
		if err := setLogFile(); err != nil {
			t.Error(err)
		}
	})

	t.Run("AppendToLog", func(t *testing.T) {
		f, err := ioutil.TempFile("", "audit.json")
		if err != nil {
			t.Fatalf("Error creating temporary file: %v", err)
		}
		defer os.Remove(f.Name())

		oldLogFile := *currentLogFile
		defer func() { currentLogFile = &oldLogFile }()
		currentLogFile = f

		e := newEntry("start", "-v", "user1", "v0.17.1", time.Now(), time.Now())
		if err := appendToLog(e); err != nil {
			t.Fatalf("Error appendingToLog: %v", err)
		}

		b := make([]byte, 100)
		if _, err := f.Read(b); err != nil && err != io.EOF {
			t.Errorf("Log was not appended to file: %v", err)
		}
	})
}
