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

package config

import (
	"testing"
)

func TestGetNotFound(t *testing.T) {
	createTestConfig(t)
	_, err := Get("nonexistent")
	if err == nil || err.Error() != "specified key could not be found in config" {
		t.Fatalf("Get did not return error for unknown property")
	}
}

func TestGetOK(t *testing.T) {
	createTestConfig(t)
	name := "driver"
	err := Set(name, "ssh")
	if err != nil {
		t.Fatalf("Set returned error for property %s, %+v", name, err)
	}
	val, err := Get(name)
	if err != nil {
		t.Fatalf("Get returned error for property %s, %+v", name, err)
	}
	if val != "ssh" {
		t.Fatalf("Get returned %s, expected ssh", val)
	}
}
