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
	"io/ioutil"
	"os"
	"testing"
)

func TestAuditReport(t *testing.T) {
	f, err := ioutil.TempFile("", "audit.json")
	if err != nil {
		t.Fatalf("failed creating temporary file: %v", err)
	}
	defer os.Remove(f.Name())

	s := `{"data":{"args":"-p mini1","command":"start","endTime":"Wed, 03 Feb 2021 15:33:05 MST","profile":"mini1","startTime":"Wed, 03 Feb 2021 15:30:33 MST","user":"user1"},"datacontenttype":"application/json","id":"9b7593cb-fbec-49e5-a3ce-bdc2d0bfb208","source":"https://minikube.sigs.k8s.io/","specversion":"1.0","type":"io.k8s.si  gs.minikube.audit"}
{"data":{"args":"-p mini1","command":"start","endTime":"Wed, 03 Feb 2021 15:33:05 MST","profile":"mini1","startTime":"Wed, 03 Feb 2021 15:30:33 MST","user":"user1"},"datacontenttype":"application/json","id":"9b7593cb-fbec-49e5-a3ce-bdc2d0bfb208","source":"https://minikube.sigs.k8s.io/","specversion":"1.0","type":"io.k8s.si  gs.minikube.audit"}
{"data":{"args":"--user user2","command":"logs","endTime":"Tue, 02 Feb 2021 16:46:20 MST","profile":"minikube","startTime":"Tue, 02 Feb 2021 16:46:00 MST","user":"user2"},"datacontenttype":"application/json","id":"fec03227-2484-48b6-880a-88fd010b5efd","source":"https://minikube.sigs.k8s.io/","specversion":"1.0","type":"io.k8s.sigs.minikube.audit"}`

	if _, err := f.WriteString(s); err != nil {
		t.Fatalf("failed writing to file: %v", err)
	}
	if _, err := f.Seek(0, 0); err != nil {
		t.Fatalf("failed seeking to start of file: %v", err)
	}

	oldLogFile := *currentLogFile
	defer func() { currentLogFile = &oldLogFile }()
	currentLogFile = f

	wantedLines := 2
	r, err := Report(wantedLines)
	if err != nil {
		t.Fatalf("failed to create report: %v", err)
	}

	if len(r.rows) != wantedLines {
		t.Fatalf("report has %d lines of logs, want %d", len(r.rows), wantedLines)
	}
}
