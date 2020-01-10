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

package addons

import (
	"fmt"
	"os"
	"testing"

	"gotest.tools/assert"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/config"
)

func TestIsAddonAlreadySet(t *testing.T) {
	testCases := []struct {
		addonName string
	}{
		{
			addonName: "ingress",
		},

		{
			addonName: "registry",
		},
	}

	for _, test := range testCases {
		addon := assets.Addons[test.addonName]
		addonStatus, _ := addon.IsEnabled()

		alreadySet, err := isAddonAlreadySet(addon, addonStatus)
		if !alreadySet {
			if addonStatus {
				t.Errorf("Did not get expected status, \n\n expected %+v already enabled", test.addonName)
			} else {
				t.Errorf("Did not get expected status, \n\n expected %+v already disabled", test.addonName)
			}
		}
		if err != nil {
			t.Errorf("Got unexpected error: %+v", err)
		}
	}
}

func TestDisableUnknownAddon(t *testing.T) {
	tmpProfile := "temp-minikube-profile"
	if err := Set("InvalidAddon", "false", tmpProfile); err == nil {
		t.Fatalf("Disable did not return error for unknown addon")
	}
}

func TestEnableUnknownAddon(t *testing.T) {
	tmpProfile := "temp-minikube-profile"
	if err := Set("InvalidAddon", "true", tmpProfile); err == nil {
		t.Fatalf("Enable did not return error for unknown addon")
	}
}

func TestEnableAndDisableAddon(t *testing.T) {
	tests := []struct {
		name   string
		enable bool
	}{
		{
			name:   "test enable",
			enable: true,
		}, {
			name:   "test disable",
			enable: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tmpProfile := "temp-minikube-profile"
			if err := os.MkdirAll(config.ProfileFolderPath(tmpProfile), 0777); err != nil {
				t.Fatalf("error creating temporary directory")
			}
			defer os.RemoveAll(config.ProfileFolderPath(tmpProfile))

			if err := config.DefaultLoader.WriteConfigToFile(tmpProfile, &config.MachineConfig{}); err != nil {
				t.Fatalf("error creating temporary profile config: %v", err)
			}
			if err := Set("dashboard", fmt.Sprintf("%t", test.enable), tmpProfile); err != nil {
				t.Fatalf("Disable returned unexpected error: " + err.Error())
			}
			c, err := config.DefaultLoader.LoadConfigFromFile(tmpProfile)
			if err != nil {
				t.Fatalf("error loading config: %v", err)
			}
			assert.Equal(t, c.Addons["dashboard"], test.enable)
		})
	}
}
