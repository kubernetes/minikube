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
	"strings"
	"testing"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/out/register"
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

	type validateJSONOutputFunc func(context.Context, *testing.T, *RunResult)
	t.Run("serial", func(t *testing.T) {
		serialTests := []struct {
			name      string
			validator validateJSONOutputFunc
		}{
			{"CloudEvents", validateCloudEvents},
		}
		for _, stc := range serialTests {
			t.Run(stc.name, func(t *testing.T) {
				stc.validator(ctx, t, rr)
			})
		}
	})

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
	last := newCloudEvent(t, ces[len(ces)-1])
	if last.Type() != register.NewError("").Type() {
		t.Fatalf("last cloud event is not of type error: %v", last)
	}
	last.validateData(t, "exitcode", fmt.Sprintf("%v", exit.Unavailable))
	last.validateData(t, "message", fmt.Sprintf("The driver 'fail' is not supported on %s\n", runtime.GOOS))
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
		t.Fatalf("values in cloud events do not match:\nActual:\n%v\nExpected:\n%v\n", v, value)
	}
}

func cloudEvents(t *testing.T, rr *RunResult) ([]cloudevents.Event, error) {
	ces := []cloudevents.Event{}
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
		ces = append(ces, event)
	}
	return ces, nil
}

//  make sure all output can be marshaled as a cloud event
func validateCloudEvents(ctx context.Context, t *testing.T, rr *RunResult) {
	_, err := cloudEvents(t, rr)
	if err != nil {
		t.Fatalf("converting to cloud events: %v\n", err)
	}
}
