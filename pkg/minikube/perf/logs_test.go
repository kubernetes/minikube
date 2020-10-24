// +build linux darwin

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

package perf

import (
	"os/exec"
	"testing"
)

func TestTimeCommandLogs(t *testing.T) {
	cmd := exec.Command("sh", "-c", "echo hi && sleep 3 && echo hey && sleep 1")
	actual, err := timeCommandLogs(cmd)
	if err != nil {
		t.Fatalf("error timing logs: %v", err)
	}
	expected := newResult()
	expected.timedLogs["hi"] = 3.0
	expected.timedLogs["hey"] = 1.0

	for log, time := range expected.timedLogs {
		actualTime, ok := actual.timedLogs[log]
		if !ok {
			t.Fatalf("expected log %s but didn't find it", log)
		}
		// Let's give a little wiggle room so we don't fail if time is 3 and actualTime is 2.99...
		if actualTime < time && time-actualTime > 0.01 {
			t.Fatalf("expected log \"%s\" to take more time than it actually did. got %v, expected > %v", log, actualTime, time)
		}
	}
}
