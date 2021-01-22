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
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"testing"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"k8s.io/minikube/pkg/minikube/out/register"
	"k8s.io/minikube/pkg/minikube/reason"
)

func TestJSONOutput(t *testing.T) {
	profile := UniqueProfileName("json-output")
	ctx, cancel := context.WithTimeout(context.Background(), Minutes(40))
	defer Cleanup(t, profile, cancel)

	tests := []struct {
		command string
		args    []string
	}{
		{
			command: "start",
			args:    append([]string{"--memory=2200", "--wait=true"}, StartArgs()...),
		}, {
			command: "pause",
		}, {
			command: "unpause",
		}, {
			command: "stop",
		},
	}

	for _, test := range tests {
		t.Run(test.command, func(t *testing.T) {
			args := []string{test.command, "-p", profile, "--output=json", "--user=testUser"}
			args = append(args, test.args...)

			rr, err := Run(t, exec.CommandContext(ctx, Target(), args...))
			if err != nil {
				t.Errorf("failed to clean up: args %q: %v", rr.Command(), err)
			}

			ces, err := cloudEvents(t, rr)
			if err != nil {
				t.Fatalf("converting to cloud events: %v\n", err)
			}

			t.Run("Audit", func(t *testing.T) {
				got, err := auditContains("testUser")
				if err != nil {
					t.Fatalf("failed to check audit log: %v", err)
				}
				if !got {
					t.Errorf("audit.json does not contain the user testUser")
				}
			})

			type validateJSONOutputFunc func(context.Context, *testing.T, []*cloudEvent)
			t.Run("parallel", func(t *testing.T) {
				parallelTests := []struct {
					name      string
					validator validateJSONOutputFunc
				}{
					{"DistinctCurrentSteps", validateDistinctCurrentSteps},
					{"IncreasingCurrentSteps", validateIncreasingCurrentSteps},
				}
				for _, stc := range parallelTests {
					stc := stc
					t.Run(stc.name, func(t *testing.T) {
						MaybeParallel(t)
						stc.validator(ctx, t, ces)
					})
				}
			})
		})
	}
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

// for successful minikube start, 'current step' should be increasing
func validateIncreasingCurrentSteps(ctx context.Context, t *testing.T, ces []*cloudEvent) {
	step := -1
	for _, ce := range ces {
		currentStep, exists := ce.data["currentstep"]
		if !exists {
			continue
		}
		cs, err := strconv.Atoi(currentStep)
		if err != nil {
			t.Fatalf("current step is not an integer: %v\n%v", currentStep, ce)
		}
		if cs <= step {
			t.Fatalf("current step is not in increasing order: %v", ces)
		}
		step = cs
	}
}

func TestJSONOutputError(t *testing.T) {
	profile := UniqueProfileName("json-output-error")
	ctx, cancel := context.WithTimeout(context.Background(), Minutes(2))
	defer Cleanup(t, profile, cancel)

	// force a failure via --driver=fail so that we can make sure errors
	// are printed as expected
	startArgs := []string{"start", "-p", profile, "--memory=2200", "--output=json", "--wait=true", "--driver=fail"}

	rr, err := Run(t, exec.CommandContext(ctx, Target(), startArgs...))
	if err == nil {
		t.Errorf("expected failure: args %q: %v", rr.Command(), err)
	}
	ces, err := cloudEvents(t, rr)
	if err != nil {
		t.Fatal(err)
	}
	// we want the last cloud event to be of type error and have the expected exit code and message
	last := ces[len(ces)-1]
	if last.Type() != register.NewError("").Type() {
		t.Fatalf("last cloud event is not of type error: %v", last)
	}
	last.validateData(t, "exitcode", fmt.Sprintf("%v", reason.ExDriverUnsupported))
	last.validateData(t, "message", fmt.Sprintf("The driver 'fail' is not supported on %s", runtime.GOOS))
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

func (c *cloudEvent) validateData(t *testing.T, key, value string) {
	v, ok := c.data[key]
	if !ok {
		t.Fatalf("expected key %s does not exist in cloud event", key)
	}
	if v != value {
		t.Fatalf("values in cloud events do not match:\nActual:\n'%v'\nExpected:\n'%v'\n", v, value)
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
