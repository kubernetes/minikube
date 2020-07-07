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
	"encoding/json"
	"os/exec"
	"strings"
	"testing"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

func TestJSONOutput(t *testing.T) {
	if NoneDriver() || DockerDriver() {
		t.Skipf("skipping: test drivers once all JSON output is enabled")
	}
	profile := UniqueProfileName("json-output")
	ctx, cancel := context.WithTimeout(context.Background(), Minutes(40))
	defer Cleanup(t, profile, cancel)

	startArgs := []string{"start", "-p", profile, "--memory=2200", "--output=json", "--wait=true"}
	startArgs = append(startArgs, StartArgs()...)

	rr, err := Run(t, exec.CommandContext(ctx, Target(), startArgs...))
	if err != nil {
		t.Errorf("failed to clean up: args %q: %v", rr.Command(), err)
	}

	type validateJSONOutputFunc func(context.Context, *testing.T, *RunResult)
	t.Run("serial", func(t *testing.T) {
		serialTests := []struct {
			name      string
			validator validateJSONOutputFunc
		}{
			{"CloudEvents", validateCloudEvents},
			{"CurrentSteps", validateCurrentSteps},
		}
		for _, stc := range serialTests {
			t.Run(stc.name, func(t *testing.T) {
				stc.validator(ctx, t, rr)
			})
		}
	})

}

//  make sure all output can be marshaled as a cloud event
func validateCloudEvents(ctx context.Context, t *testing.T, rr *RunResult) {
	stdout := strings.Split(rr.Stdout.String(), "\n")
	for _, s := range stdout {
		if s == "" {
			continue
		}
		event := cloudevents.NewEvent()
		if err := json.Unmarshal([]byte(s), &event); err != nil {
			t.Fatalf("unable to unmarshal output: %v\n%s", err, s)
		}
	}
}

// make sure each step in a successful `minikube start` has a distict step number
func validateCurrentSteps(ctx context.Context, t *testing.T, rr *RunResult) {
	stdout := strings.Split(rr.Stdout.String(), "\n")
	data := map[string]string{}
	currentSteps := map[string]struct{}{}
	for _, s := range stdout {
		if s == "" {
			continue
		}
		event := cloudevents.NewEvent()
		if err := json.Unmarshal([]byte(s), &event); err != nil {
			t.Fatalf("unable to unmarshal output: %v\n%s", err, s)
		}
		if err := json.Unmarshal(event.Data(), &data); err != nil {
			t.Fatalf("unable to unmarshal output: %v\n%s", err, s)
		}
		cs := data["currentstep"]
		if _, ok := currentSteps[cs]; ok {
			t.Fatalf("The log \"%s\" has already been logged, please create a new log for this step: %v", data["name"], data["message"])
		} else {
			currentSteps[cs] = struct{}{}
		}
	}
}
