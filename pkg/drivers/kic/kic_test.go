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

package kic

import (
	"testing"
)

func TestMachineOrder(t *testing.T) {
	testCases := []struct {
		Name        string
		MachineName string
		Want        int
	}{
		{
			Name:        "default",
			MachineName: "minikube",
			Want:        1},
		{
			Name:        "second-node",
			MachineName: "minikube-m02",
			Want:        2},
		{
			Name:        "dash-profile",
			MachineName: "my-dashy-minikube",
			Want:        1},

		{
			Name:        "dash-profile-second-node",
			MachineName: "my-dashy-minikube-m02",
			Want:        2},
		{
			Name:        "michivious-user",
			MachineName: "michivious-user-m02-m03",
			Want:        3},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			got := machineOrder(tc.MachineName)
			if got != tc.Want {
				t.Errorf("want order %q but got %q", tc.Want, got)

			}
		})

	}
}
