// +build integration

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

package integration

import (
	"context"
	"os/exec"
	"testing"
)

func TestKicCustomNetwork(t *testing.T) {
	if !KicDriver() {
		t.Skip("only runs with docker driver")
	}

	tests := []struct {
		description string
		flag        string
	}{
		{
			description: "create custom network",
		}, {
			description: "use default bridge network",
			flag:        "--docker-network=bridge",
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			profile := UniqueProfileName("docker-network")
			ctx, cancel := context.WithTimeout(context.Background(), Minutes(5))
			defer Cleanup(t, profile, cancel)

			startArgs := []string{"start", "-p", profile, test.flag}
			c := exec.CommandContext(ctx, Target(), startArgs...)
			rr, err := Run(t, c)
			if err != nil {
				t.Fatalf("%v failed: %v\n%v", rr.Command(), err, rr.Output())
			}
		})
	}
}
