/*
Copyright 2017 The Kubernetes Authors All rights reserved.

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
	"bytes"
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func mockCollectTimes(times [][]float64) func(ctx context.Context, binaries []*Binary) ([][]float64, error) {
	return func(ctx context.Context, binaries []*Binary) ([][]float64, error) {
		return times, nil
	}
}

func TestCompareMinikubeStartOutput(t *testing.T) {
	binaries := []*Binary{
		{
			path: "minikube1",
		}, {
			path: "minikube2",
		},
	}
	tests := []struct {
		description string
		times       [][]float64
		expected    string
	}{
		{
			description: "standard run",
			times:       [][]float64{{4.5, 6}, {1, 2}},
			expected: `Results for minikube1:
Times: [4.5 6]
Average Time: 5.250000

Results for minikube2:
Times: [1 2]
Average Time: 1.500000

`,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			originalCollectTimes := collectTimes
			collectTimeMinikubeStart = mockCollectTimes(test.times)
			defer func() { collectTimeMinikubeStart = originalCollectTimes }()

			buf := bytes.NewBuffer([]byte{})
			err := CompareMinikubeStart(context.Background(), buf, binaries)
			if err != nil {
				t.Fatalf("error comparing minikube start: %v", err)
			}

			actual := buf.String()
			if diff := cmp.Diff(test.expected, actual); diff != "" {
				t.Errorf("machines mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
func TestAverage(t *testing.T) {
	tests := []struct {
		description string
		nums        []float64
		expected    float64
	}{
		{
			description: "one number",
			nums:        []float64{4},
			expected:    4,
		}, {
			description: "multiple numbers",
			nums:        []float64{1, 4},
			expected:    2.5,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			actual := average(test.nums)
			if actual != test.expected {
				t.Fatalf("actual output does not match expected output\nActual: %v\nExpected: %v", actual, test.expected)
			}
		})
	}
}
