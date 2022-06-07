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

package integration

import (
	"context"
	"encoding/json"
	"os/exec"
	"testing"

	"k8s.io/minikube/pkg/minikube/config"
)

// profileJSON is the output of `minikube profile list -ojson`
type profileJSON struct {
	Valid   []config.Profile `json:"valid"`
	Invalid []config.Profile `json:"invalid"`
}

func TestMinikubeProfile(t *testing.T) {
	// 1. Setup two minikube cluster profiles
	profiles := [2]string{UniqueProfileName("first"), UniqueProfileName("second")}
	ctx, cancel := context.WithTimeout(context.Background(), Minutes(10))
	// TODO(@vharsh): Figure out why go vet complains when this is moved into a loop
	defer CleanupWithLogs(t, profiles[0], cancel)
	defer CleanupWithLogs(t, profiles[1], cancel)
	for _, p := range profiles {
		c := []string{"start", "-p", p}
		c = append(c, StartArgs()...)
		rr, err := Run(t, exec.CommandContext(ctx, Target(), c...))
		if err != nil {
			t.Fatalf("test pre-condition failed. args %q: %v", rr.Command(), err)
		}
	}
	// 2. Change minikube profile
	for _, p := range profiles {
		rr, err := Run(t, exec.CommandContext(ctx, Target(), "profile", p))
		if err != nil {
			t.Fatalf("cmd: %s failed with error: %v\n", rr.Command(), err)
		}
		r, err := Run(t, exec.CommandContext(ctx, Target(), "profile", "list", "-ojson"))
		if err != nil {
			t.Fatalf("cmd: %s failed with error: %v\n", r.Command(), err)
		}
		var profile profileJSON
		err = json.NewDecoder(r.Stdout).Decode(&profile)
		if err != nil {
			t.Fatalf("unmarshalling %s cmd output failed with error: %v\n", r.Command(), err)
		}
		// 3. Assert minikube profile is set to the correct profile in JSON
		for _, s := range profile.Valid {
			if s.Name == p && !s.Active {
				t.Errorf("minikube profile %s is not active\n", p)
			} else if s.Name != p && s.Active {
				t.Errorf("minikube profile %s should not have been active but is active\n", p)
			}
		}
	}
}
