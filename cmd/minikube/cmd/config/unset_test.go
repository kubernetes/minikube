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

package config

import (
	"testing"

	"k8s.io/minikube/pkg/minikube/config"
)

func TestUnsetConfig(t *testing.T) {
	createTestConfig(t)
	propName := "cpus"
	propValue := "1"
	err := Set(propName, propValue)
	if err != nil {
		t.Errorf("Failed to set the property %q", propName)
	}

	cpus, err := config.Get("cpus")
	if err != nil {
		t.Errorf("Failed to read config %q", err)
	}

	if cpus != propValue {
		t.Errorf("Expected cpus to be %s but got %s", propValue, cpus)
	}

	err = Unset(propName)
	if err != nil {
		t.Errorf("Failed to unset property %q", err)
	}

	_, err = config.Get("cpus")
	if err != config.ErrKeyNotFound {
		t.Errorf("Expected error %q but got %q", config.ErrKeyNotFound, err)
	}
}
