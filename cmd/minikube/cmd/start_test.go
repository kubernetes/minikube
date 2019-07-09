/*
Copyright 2019 The Kubernetes Authors All rights reserved.

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

package cmd

import (
	"testing"

	"k8s.io/minikube/pkg/minikube/constants"
)

func Test_extractVMDriverVersion(t *testing.T) {
	v := extractVMDriverVersion("")
	if len(v) != 0 {
		t.Error("Expected empty string")
	}

	v = extractVMDriverVersion("random text")
	if len(v) != 0 {
		t.Error("Expected empty string")
	}

	expectedVersion := "1.2.3"

	v = extractVMDriverVersion("version: v1.2.3")
	if expectedVersion != v {
		t.Errorf("Expected version: %s, got: %s", expectedVersion, v)
	}

	v = extractVMDriverVersion("version: 1.2.3")
	if expectedVersion != v {
		t.Errorf("Expected version: %s, got: %s", expectedVersion, v)
	}
}

func Test_validateOSSupportVMDriver(t *testing.T) {
	tests := []struct {
		driver   string
		OS       string
		supports bool
	}{
		{constants.DriverNone, constants.Darwin, false},
		{constants.DriverNone, constants.Windows, false},
		{constants.DriverNone, constants.Linux, true},
		{constants.DriverVirtualbox, constants.Darwin, true},
		{constants.DriverVirtualbox, constants.Windows, true},
		{constants.DriverVirtualbox, constants.Linux, true},
	}

	for _, test := range tests {
		s := validateOSSupportVMDriver(test.OS, test.driver)
		if test.supports != s {
			t.Errorf("Expected: %t got: %t for driver %s support on %s", test.supports, s, test.driver, test.OS)
		}
	}
}
