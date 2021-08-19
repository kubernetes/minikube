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

package register

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"k8s.io/minikube/pkg/minikube/tests"
)

func TestPrintStep(t *testing.T) {
	Reg.SetStep(InitialSetup)

	expected := `{"data":{"currentstep":"0","message":"message","name":"Initial Minikube Setup","totalsteps":"%v"},"datacontenttype":"application/json","id":"random-id","source":"https://minikube.sigs.k8s.io/","specversion":"1.0","type":"io.k8s.sigs.minikube.step"}`
	expected = fmt.Sprintf(expected, Reg.totalSteps())
	expected += "\n"

	buf := bytes.NewBuffer([]byte{})
	SetOutputFile(buf)
	defer func() { SetOutputFile(os.Stdout) }()

	GetUUID = func() string {
		return "random-id"
	}

	PrintStep("message")
	actual := buf.Bytes()

	tests.CompareJSON(t, actual, []byte(expected))
}

func TestPrintInfo(t *testing.T) {
	expected := `{"data":{"message":"info"},"datacontenttype":"application/json","id":"random-id","source":"https://minikube.sigs.k8s.io/","specversion":"1.0","type":"io.k8s.sigs.minikube.info"}`
	expected += "\n"

	buf := bytes.NewBuffer([]byte{})
	SetOutputFile(buf)
	defer func() { SetOutputFile(os.Stdout) }()

	GetUUID = func() string {
		return "random-id"
	}

	PrintInfo("info")
	actual := buf.Bytes()

	tests.CompareJSON(t, actual, []byte(expected))
}

func TestError(t *testing.T) {
	expected := `{"data":{"message":"error"},"datacontenttype":"application/json","id":"random-id","source":"https://minikube.sigs.k8s.io/","specversion":"1.0","type":"io.k8s.sigs.minikube.error"}`
	expected += "\n"

	buf := bytes.NewBuffer([]byte{})
	SetOutputFile(buf)
	defer func() { SetOutputFile(os.Stdout) }()

	GetUUID = func() string {
		return "random-id"
	}

	PrintError("error")
	actual := buf.Bytes()

	tests.CompareJSON(t, actual, []byte(expected))
}

func TestErrorExitCode(t *testing.T) {
	expected := `{"data":{"a":"b","c":"d","exitcode":"5","message":"error"},"datacontenttype":"application/json","id":"random-id","source":"https://minikube.sigs.k8s.io/","specversion":"1.0","type":"io.k8s.sigs.minikube.error"}`
	expected += "\n"

	buf := bytes.NewBuffer([]byte{})
	SetOutputFile(buf)
	defer func() { SetOutputFile(os.Stdout) }()

	GetUUID = func() string {
		return "random-id"
	}

	PrintErrorExitCode("error", 5, map[string]string{"a": "b"}, map[string]string{"c": "d"})
	actual := buf.Bytes()

	tests.CompareJSON(t, actual, []byte(expected))
}

func TestWarning(t *testing.T) {
	expected := `{"data":{"message":"warning"},"datacontenttype":"application/json","id":"random-id","source":"https://minikube.sigs.k8s.io/","specversion":"1.0","type":"io.k8s.sigs.minikube.warning"}`
	expected += "\n"

	buf := bytes.NewBuffer([]byte{})
	SetOutputFile(buf)
	defer func() { SetOutputFile(os.Stdout) }()

	GetUUID = func() string {
		return "random-id"
	}

	PrintWarning("warning")
	actual := buf.Bytes()

	tests.CompareJSON(t, actual, []byte(expected))
}
