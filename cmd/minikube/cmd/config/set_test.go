/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package config

import (
	"io/ioutil"
	"os"
	"testing"

	"k8s.io/minikube/pkg/minikube/localpath"
)

func TestNotFound(t *testing.T) {
	createTestConfig(t)
	err := Set("nonexistent", "10")
	if err == nil || err.Error() != "find settings for \"nonexistent\" value of \"10\": property name \"nonexistent\" not found" {
		t.Fatalf("Set did not return error for unknown property: %+v", err)
	}
}

func TestSetNotAllowed(t *testing.T) {
	createTestConfig(t)
	err := Set("driver", "123456")
	if err == nil || err.Error() != "run validations for \"driver\" with value of \"123456\": [driver \"123456\" is not supported]" {
		t.Fatalf("Set did not return error for unallowed value: %+v", err)
	}
	err = Set("memory", "10a")
	if err == nil || err.Error() != "run validations for \"memory\" with value of \"10a\": [invalid memory size: invalid size: '10a']" {
		t.Fatalf("Set did not return error for unallowed value: %+v", err)
	}
}

func TestSetOK(t *testing.T) {
	createTestConfig(t)
	err := Set("driver", "virtualbox")
	defer func() {
		err = Unset("driver")
		if err != nil {
			t.Errorf("failed to unset driver: %+v", err)
		}
	}()
	if err != nil {
		t.Fatalf("Set returned error for valid property value: %+v", err)
	}
	val, err := Get("driver")
	if err != nil {
		t.Fatalf("Get returned error for valid property: %+v", err)
	}
	if val != "virtualbox" {
		t.Fatalf("Get returned %s, expected \"virtualbox\"", val)
	}
}

func createTestConfig(t *testing.T) {
	t.Helper()
	td, err := ioutil.TempDir("", "config")
	if err != nil {
		t.Fatalf("tempdir: %v", err)
	}

	err = os.Setenv(localpath.MinikubeHome, td)
	if err != nil {
		t.Fatalf("error setting up test environment. could not set %s due to %+v", localpath.MinikubeHome, err)
	}

	// Not necessary, but it is a handy random alphanumeric
	if err = os.MkdirAll(localpath.MakeMiniPath("config"), 0777); err != nil {
		t.Fatalf("error creating temporary directory: %+v", err)
	}

	if err = os.MkdirAll(localpath.MakeMiniPath("profiles"), 0777); err != nil {
		t.Fatalf("error creating temporary profiles directory: %+v", err)
	}

	t.Cleanup(func() {
		err := os.RemoveAll(td)
		if err != nil {
			t.Errorf("failed to clean up temp folder  %q", td)
		}
	})
}
