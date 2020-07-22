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
	profile := UniqueProfileName("json-output")
	ctx, cancel := context.WithTimeout(context.Background(), Minutes(40))
	defer Cleanup(t, profile, cancel)

	startArgs := []string{"start", "-p", profile, "--memory=2200", "--output=json", "--wait=true"}
	startArgs = append(startArgs, StartArgs()...)

	rr, err := Run(t, exec.CommandContext(ctx, Target(), startArgs...))
	if err != nil {
		t.Errorf("failed to clean up: args %q: %v", rr.Command(), err)
	}

	ces, err := cloudEvents(t, rr)
	if err != nil {
		t.Fatalf("converting to cloud events: %v\n", err)
	}

	type validateJSONOutputFunc func(context.Context, *testing.T, []*cloudEvent)
	t.Run("serial", func(t *testing.T) {
		serialTests := []struct {
			name      string
			validator validateJSONOutputFunc
		}{
			{"DistinctCurrentSteps", validateDistinctCurrentSteps},
		}
		for _, stc := range serialTests {
			t.Run(stc.name, func(t *testing.T) {
				stc.validator(ctx, t, ces)
			})
		}
	})
}

//  make sure each step has a distinct step number
func validateDistinctCurrentSteps(ctx context.Context, t *testing.T, ces []*cloudEvent) {
	steps := map[string]string{}
	for _, ce := range ces {
		currentStep, exists := ce.data["currentstep"]
		if !exists {
			continue
		}
		if msg, alreadySeen := steps[currentStep]; alreadySeen {
			t.Fatalf("step %v has already been assigned to another step:\n%v\nCannot use for:\n%v\n%v", currentStep, msg, ce.data["message"], ces)
		}
		steps[currentStep] = ce.data["message"]
	}
}

type cloudEvent struct {
	cloudevents.Event
	data map[string]string
}

func newCloudEvent(t *testing.T, ce cloudevents.Event) *cloudEvent {
	m := map[string]string{}
	data := ce.Data()
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("marshalling cloud event: %v", err)
	}
	return &cloudEvent{
		Event: ce,
		data:  m,
	}
}

func cloudEvents(t *testing.T, rr *RunResult) ([]*cloudEvent, error) {
	ces := []*cloudEvent{}
	stdout := strings.Split(rr.Stdout.String(), "\n")
	for _, s := range stdout {
		if s == "" {
			continue
		}
		event := cloudevents.NewEvent()
		if err := json.Unmarshal([]byte(s), &event); err != nil {
			t.Logf("unable to marshal output: %v", s)
			return nil, err
		}
		ces = append(ces, newCloudEvent(t, event))
	}
	return ces, nil
}
