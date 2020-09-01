/*
Copyright 2018 The Kubernetes Authors All rights reserved.

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

package tunnel

import (
	"errors"
	"fmt"
	"testing"
)

func TestReporter(t *testing.T) {
	testCases := []struct {
		name           string
		tunnelState    *Status
		expectedOutput string
	}{
		{
			name: "simple",
			tunnelState: &Status{
				TunnelID: ID{
					Route:       unsafeParseRoute("1.2.3.4", "10.96.0.0/12"),
					MachineName: "testmachine",
					Pid:         1234,
				},
				MinikubeState: Running,
				MinikubeError: nil,

				RouteError: nil,

				PatchedServices:           []string{"svc1", "svc2"},
				LoadBalancerEmulatorError: nil,
			},
			expectedOutput: `Status:	
	machine: testmachine
	pid: 1234
	route: 10.96.0.0/12 -> 1.2.3.4
	minikube: Running
	services: [svc1, svc2]
    errors: 
		minikube: no errors
		router: no errors
		loadbalancer emulator: no errors
`,
		},
		{
			name: "errors",
			tunnelState: &Status{
				TunnelID: ID{
					Route:       unsafeParseRoute("1.2.3.4", "10.96.0.0/12"),
					MachineName: "testmachine",
					Pid:         1234,
				},
				MinikubeState: Unknown,
				MinikubeError: errors.New("minikubeerror"),

				RouteError: errors.New("route error"),

				PatchedServices:           nil,
				LoadBalancerEmulatorError: errors.New("lberror"),
			},
			expectedOutput: `Status:	
	machine: testmachine
	pid: 1234
	route: 10.96.0.0/12 -> 1.2.3.4
	minikube: Unknown
	services: []
    errors: 
		minikube: minikubeerror
		router: route error
		loadbalancer emulator: lberror
`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			out := &recordingWriter{}
			reporter := newReporter(out)
			reporter.Report(tc.tunnelState)
			if tc.expectedOutput != out.output {
				t.Errorf(`%s [FAIL].
Expected: "%s" 
Got:	  "%s"`, tc.name, tc.expectedOutput, out.output)
			}
		})
	}

	// testing deduplication
	out := &recordingWriter{}
	reporter := newReporter(out)
	reporter.Report(testCases[0].tunnelState)
	reporter.Report(testCases[0].tunnelState)
	reporter.Report(testCases[1].tunnelState)
	reporter.Report(testCases[1].tunnelState)
	reporter.Report(testCases[0].tunnelState)

	expectedOutput := fmt.Sprintf("%s%s%s",
		testCases[0].expectedOutput,
		testCases[1].expectedOutput,
		testCases[0].expectedOutput)

	if out.output != expectedOutput {
		t.Errorf(`Deduplication test [FAIL].
Expected: "%s" 
Got:	  "%s"`, expectedOutput, out.output)
	}
}

type recordingWriter struct {
	output string
}

func (w *recordingWriter) Write(p []byte) (n int, err error) {
	w.output = fmt.Sprintf("%s%s", w.output, p)
	return 0, nil
}
