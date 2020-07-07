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

func TestCloudEvents(t *testing.T) {
	profile := UniqueProfileName("json-output")
	ctx, cancel := context.WithTimeout(context.Background(), Minutes(40))
	defer Cleanup(t, profile, cancel)

	startArgs := []string{"start", "-p", profile, "--memory=2200", "--output=json", "--wait=true", "-p", profile}
	startArgs = append(startArgs, StartArgs()...)

	rr, err := Run(t, exec.CommandContext(ctx, Target(), startArgs...))
	if err != nil {
		t.Errorf("failed to clean up: args %q: %v", rr.Command(), err)
	}

	//  make sure all output can be marshaled as a cloud event
	stdout := strings.Split(rr.Stdout.String(), "\n")
	for _, s := range stdout {
		if s == "" {
			continue
		}
		event := cloudevents.NewEvent()
		if err := json.Unmarshal([]byte(s), &event); err != nil {
			t.Fatalf("unable to marshal output: %v\n%s", err, s)
		}
	}
}
