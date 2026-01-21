/*
Copyright 2022 The Kubernetes Authors All rights reserved.

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

package host

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/minikube/pkg/libmachine/auth"
	"k8s.io/minikube/pkg/libmachine/drivers/nodriver"
)

func TestMigrateHost(t *testing.T) {
	mustMarshal := func(v any) []byte {
		t.Helper()
		b, err := json.Marshal(v)
		if err != nil {
			t.Fatalf("failed to marshal expected JSON: %v", err)
		}
		return b
	}

	v2MigratedStorePath := filepath.Clean("/Users/nathanleclaire/.docker/machine")
	v2MigratedDriverData := mustMarshal(map[string]string{
		"MachineName": "default",
		"StorePath":   v2MigratedStorePath,
	})

	testCases := []struct {
		description                string
		hostBefore                 *Host
		rawData                    []byte
		expectedHostAfter          *Host
		expectedMigrationPerformed bool
		expectedMigrationError     error
	}{
		{
			// Point of this test is largely that no matter what was in RawDriver
			// before, it should load into the Host struct based on what is actually
			// in the Driver field.
			//
			// Note that we don't check for the presence of RawDriver's literal "on
			// disk" here.  It's intentional.
			description: "Config version 3 load with existing RawDriver on disk",
			hostBefore: &Host{
				Name: "default",
			},
			rawData: []byte(`{
    "ConfigVersion": 3,
    "Driver": {"MachineName": "default"},
    "DriverName": "virtualbox",
    "HostOptions": {
        "Driver": "",
        "Memory": 0,
        "Disk": 0,
        "AuthOptions": {
            "StorePath": "/Users/nathanleclaire/.docker/machine/machines/default"
        }
    },
    "Name": "default",
    "RawDriver": "eyJWQm94TWFuYWdlciI6e30sIklQQWRkcmVzcyI6IjE5Mi4xNjguOTkuMTAwIiwiTWFjaGluZU5hbWUiOiJkZWZhdWx0IiwiU1NIVXNlciI6ImRvY2tlciIsIlNTSFBvcnQiOjU4MTQ1LCJTU0hLZXlQYXRoIjoiL1VzZXJzL25hdGhhbmxlY2xhaXJlLy5kb2NrZXIvbWFjaGluZS9tYWNoaW5lcy9kZWZhdWx0L2lkX3JzYSIsIlN0b3JlUGF0aCI6Ii9Vc2Vycy9uYXRoYW5sZWNsYWlyZS8uZG9ja2VyL21hY2hpbmUiLCJTd2FybU1hc3RlciI6ZmFsc2UsIlN3YXJtSG9zdCI6InRjcDovLzAuMC4wLjA6MzM3NiIsIlN3YXJtRGlzY292ZXJ5IjoiIiwiQ1BVIjoxLCJNZW1vcnkiOjEwMjQsIkRpc2tTaXplIjoyMDAwMCwiQm9vdDJEb2NrZXJVUkwiOiIiLCJCb290MkRvY2tlckltcG9ydFZNIjoiIiwiSG9zdE9ubHlDSURSIjoiMTkyLjE2OC45OS4xLzI0IiwiSG9zdE9ubHlOaWNUeXBlIjoiODI1NDBFTSIsIkhvc3RPbmx5UHJvbWlzY01vZGUiOiJkZW55IiwiTm9TaGFyZSI6ZmFsc2V9"
}`),
			expectedHostAfter: &Host{
				ConfigVersion: 3,
				HostOptions: &Options{
					AuthOptions: &auth.Options{
						StorePath: "/Users/nathanleclaire/.docker/machine/machines/default",
					},
				},
				Name:       "default",
				DriverName: "virtualbox",
				RawDriver:  []byte(`{"MachineName": "default"}`),
				Driver: &RawDataDriver{
					Data: []byte(`{"MachineName": "default"}`),

					// TODO (nathanleclaire): The "." argument here is a already existing
					// bug (or at least likely to cause them in the future) and most
					// likely should be "/Users/nathanleclaire/.docker/machine"
					//
					// These default StorePath settings get over-written when we
					// instantiate the plugin driver, but this seems entirely incidental.
					Driver: nodriver.NewDriver("default", "."),
				},
			},
			expectedMigrationPerformed: false,
			expectedMigrationError:     nil,
		},
		{
			description: "Config version 4 (from the FUTURE) on disk",
			hostBefore: &Host{
				Name: "default",
			},
			rawData: []byte(`{
    "ConfigVersion": 4,
    "Driver": {"MachineName": "default"},
    "DriverName": "virtualbox",
    "HostOptions": {
        "Driver": "",
        "Memory": 0,
        "Disk": 0,
        "AuthOptions": {
            "StorePath": "/Users/nathanleclaire/.docker/machine/machines/default"
        }
    },
    "Name": "default"
}`),
			expectedHostAfter:          nil,
			expectedMigrationPerformed: false,
			expectedMigrationError:     errConfigFromFuture,
		},
		{
			description: "Config version 3 load WITHOUT any existing RawDriver field on disk",
			hostBefore: &Host{
				Name: "default",
			},
			rawData: []byte(`{
    "ConfigVersion": 3,
    "Driver": {"MachineName": "default"},
    "DriverName": "virtualbox",
    "HostOptions": {
        "Driver": "",
        "Memory": 0,
        "Disk": 0,
        "AuthOptions": {
            "StorePath": "/Users/nathanleclaire/.docker/machine/machines/default"
        }
    },
    "Name": "default"
}`),
			expectedHostAfter: &Host{
				ConfigVersion: 3,
				HostOptions: &Options{
					AuthOptions: &auth.Options{
						StorePath: "/Users/nathanleclaire/.docker/machine/machines/default",
					},
				},
				Name:       "default",
				DriverName: "virtualbox",
				RawDriver:  []byte(`{"MachineName": "default"}`),
				Driver: &RawDataDriver{
					Data: []byte(`{"MachineName": "default"}`),

					// TODO: See note above.
					Driver: nodriver.NewDriver("default", "."),
				},
			},
			expectedMigrationPerformed: false,
			expectedMigrationError:     nil,
		},
		{
			description: "Config version 2 load and migrate.  Ensure StorePath gets set properly.",
			hostBefore: &Host{
				Name: "default",
			},
			rawData: []byte(`{
    "ConfigVersion": 2,
    "Driver": {"MachineName": "default"},
    "DriverName": "virtualbox",
    "HostOptions": {
        "Driver": "",
        "Memory": 0,
        "Disk": 0,
        "AuthOptions": {
            "StorePath": "/Users/nathanleclaire/.docker/machine/machines/default"
        }
    },
    "StorePath": "/Users/nathanleclaire/.docker/machine/machines/default",
    "Name": "default"
}`),
			expectedHostAfter: &Host{
				ConfigVersion: 3,
				HostOptions: &Options{
					AuthOptions: &auth.Options{
						StorePath: "/Users/nathanleclaire/.docker/machine/machines/default",
					},
				},
				Name:       "default",
				DriverName: "virtualbox",
				RawDriver:  v2MigratedDriverData,
				Driver: &RawDataDriver{
					Data:   v2MigratedDriverData,
					Driver: nodriver.NewDriver("default", v2MigratedStorePath),
				},
			},
			expectedMigrationPerformed: true,
			expectedMigrationError:     nil,
		},
	}

	for _, tc := range testCases {
		actualHostAfter, actualMigrationPerformed, actualMigrationError := MigrateHost(tc.hostBefore, tc.rawData)

		assert.Equal(t, tc.expectedHostAfter, actualHostAfter)
		assert.Equal(t, tc.expectedMigrationPerformed, actualMigrationPerformed)
		assert.Equal(t, tc.expectedMigrationError, actualMigrationError)
	}
}
