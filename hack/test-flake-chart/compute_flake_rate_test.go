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

package main

import (
	"strings"
	"testing"
	"time"
)

func simpleDate(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func TestReadData(t *testing.T) {
	actualData := ReadData(strings.NewReader(
		`A,B,C,D,E,F
		hash,2000-01-01,env1,test1,Passed,1
		hash,2001-01-01,env2,test2,Failed,1
		hash,,,test1,,1
		hash,2002-01-01,,,Passed,1
		hash,2003-01-01,env3,test3,Passed,1`,
	))
	expectedData := []TestEntry{
		{
			name:        "test1",
			environment: "env1",
			date:        simpleDate(2000, time.January, 1),
			status:      "Passed",
		},
		{
			name:        "test2",
			environment: "env2",
			date:        simpleDate(2001, time.January, 1),
			status:      "Failed",
		},
		{
			name:        "test1",
			environment: "env2",
			date:        simpleDate(2001, time.January, 1),
			status:      "Failed",
		},
		{
			name:        "test1",
			environment: "env2",
			date:        simpleDate(2002, time.January, 1),
			status:      "Passed",
		},
		{
			name:        "test3",
			environment: "env3",
			date:        simpleDate(2003, time.January, 1),
			status:      "Passed",
		},
	}

	for i, actual := range actualData {
		if len(expectedData) <= i {
			t.Errorf("Received unmatched actual element at index %d. Actual: %v", i, actual)
			continue
		}
		expected := expectedData[i]
		if actual != expected {
			t.Errorf("Elements differ at index %d. Expected: %v, Actual: %v", i, expected, actual)
		}
	}

	if len(actualData) < len(expectedData) {
		for i := len(actualData); i < len(expectedData); i++ {
			t.Errorf("Missing unmatched expected element at index %d. Expected: %v", i, expectedData[i])
		}
	}
}
