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
)

func TestSetCurrentStep(t *testing.T) {
	secondStep := Reg.steps[1]
	Reg.SetStep(secondStep)

	expected := `{"data":{"currentstep":"1","message":"message","name":"%s","totalsteps":"%v"},"datacontenttype":"application/json","id":"random-id","source":"https://minikube.sigs.k8s.io/","specversion":"1.0","type":"io.k8s.sigs.minikube.step"}`
	expected = fmt.Sprintf(expected, secondStep, Reg.totalSteps())
	expected += "\n"

	buf := bytes.NewBuffer([]byte{})
	OutputFile = buf
	defer func() { OutputFile = os.Stdout }()

	GetUUID = func() string {
		return "random-id"
	}

	PrintStep("message")
	actual := buf.String()

	if actual != expected {
		t.Fatalf("expected didn't match actual:\nExpected:\n%v\n\nActual:\n%v", expected, actual)
	}
}
