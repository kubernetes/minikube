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
	"reflect"
	"testing"
)

func mockCollectTimeMinikubeStart(durations []float64) func(ctx context.Context, binary string) (float64, error) {
	index := 0
	return func(context.Context, string) (float64, error) {
		duration := durations[index]
		index++
		return duration, nil
	}
}

func TestCompareMinikubeStartOutput(t *testing.T) {
	tests := []struct {
		description string
		durations   []float64
		expected    string
	}{
		{
			description: "standard run",
			durations:   []float64{4.5, 6},
			expected:    "Old binary: [4.5]\nNew binary: [6]\nAverage Old: 4.500000\nAverage New: 6.000000\n",
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			originalCollectTimes := collectTimeMinikubeStart
			collectTimeMinikubeStart = mockCollectTimeMinikubeStart(test.durations)
			defer func() { collectTimeMinikubeStart = originalCollectTimes }()

			buf := bytes.NewBuffer([]byte{})
			err := CompareMinikubeStart(context.Background(), buf, []string{"", ""})
			if err != nil {
				t.Fatalf("error comparing minikube start: %v", err)
			}

			actual := buf.String()
			if test.expected != actual {
				t.Fatalf("actual output does not match expected output\nActual: %v\nExpected: %v", actual, test.expected)
			}
		})
	}
}

func TestCollectTimes(t *testing.T) {
	tests := []struct {
		description string
		durations   []float64
		expected    [][]float64
	}{
		{
			description: "test collect time",
			durations:   []float64{1, 2},
			expected: [][]float64{
				{1},
				{2},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			originalCollectTimes := collectTimeMinikubeStart
			collectTimeMinikubeStart = mockCollectTimeMinikubeStart(test.durations)
			defer func() { collectTimeMinikubeStart = originalCollectTimes }()

			actual, err := collectTimes(context.Background(), []string{"", ""})
			if err != nil {
				t.Fatalf("error collecting times: %v", err)
			}

			if !reflect.DeepEqual(actual, test.expected) {
				t.Fatalf("actual output does not match expected output\nActual: %v\nExpected: %v", actual, test.expected)
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
