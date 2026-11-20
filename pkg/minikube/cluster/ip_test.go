/*
Copyright 2026 The Kubernetes Authors All rights reserved.

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

package cluster

import (
	"errors"
	"net"
	"testing"

	"k8s.io/minikube/pkg/libmachine/host"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/tests"
)

type mockDriver struct {
	tests.MockDriver
	ip  string
	err error
}

func (m *mockDriver) GetIP() (string, error) {
	return m.ip, m.err
}

func TestHostIP(t *testing.T) {
	tests := []struct {
		name       string
		driverName string
		driverIP   string
		driverErr  error
		expectedIP net.IP
		expectErr  bool
	}{
		{
			name:       "vfkit+nat",
			driverName: driver.VFKit,
			driverIP:   "192.168.64.2",
			expectedIP: net.ParseIP("192.168.64.1"),
		},
		{
			name:       "vfkit+vmnet-shared",
			driverName: driver.VFKit,
			driverIP:   "192.168.105.2",
			expectedIP: net.ParseIP("192.168.105.1"),
		},
		{
			name:       "krunkit+vmnet-shared",
			driverName: driver.Krunkit,
			driverIP:   "192.168.105.5",
			expectedIP: net.ParseIP("192.168.105.1"),
		},
		{
			name:       "hyperkit",
			driverName: driver.HyperKit,
			driverIP:   "192.168.64.10",
			expectedIP: net.ParseIP("192.168.64.1"),
		},
		{
			name:       "none driver",
			driverName: driver.None,
			expectedIP: net.ParseIP("127.0.0.1"),
		},
		{
			name:       "driver error",
			driverName: driver.VFKit,
			driverErr:  errors.New("failed to get IP"),
			expectErr:  true,
		},
		{
			name:       "invalid IP",
			driverName: driver.VFKit,
			driverIP:   "not-an-ip",
			expectErr:  true,
		},
		{
			name:       "empty IP",
			driverName: driver.Krunkit,
			driverIP:   "",
			expectErr:  true,
		},
		{
			name:       "unimplemented driver",
			driverName: "unknown-driver",
			expectErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := &host.Host{
				DriverName: tc.driverName,
				Driver: &mockDriver{
					ip:  tc.driverIP,
					err: tc.driverErr,
				},
			}

			ip, err := HostIP(h, "minikube")
			if tc.expectErr {
				if err == nil {
					t.Fatalf("expected error, got ip: %v", ip)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !ip.Equal(tc.expectedIP) {
				t.Errorf("expected IP %s, got %s", tc.expectedIP, ip)
			}
		})
	}
}
