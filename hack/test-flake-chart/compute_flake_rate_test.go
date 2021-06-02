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
	"fmt"
	"strings"
	"testing"
	"time"
)

func simpleDate(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func compareEntrySlices(t *testing.T, actualData, expectedData []TestEntry, extra string) {
	if extra != "" {
		extra = fmt.Sprintf(" (%s)", extra)
	}
	for i, actual := range actualData {
		if len(expectedData) <= i {
			t.Errorf("Received unmatched actual element at index %d%s. Actual: %v", i, extra, actual)
			continue
		}
		expected := expectedData[i]
		if actual != expected {
			t.Errorf("Elements differ at index %d%s. Expected: %v, Actual: %v", i, extra, expected, actual)
		}
	}

	if len(actualData) < len(expectedData) {
		for i := len(actualData); i < len(expectedData); i++ {
			t.Errorf("Missing unmatched expected element at index %d%s. Expected: %v", i, extra, expectedData[i])
		}
	}
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

	compareEntrySlices(t, actualData, expectedData, "")
}

func compareSplitData(t *testing.T, actual, expected map[string]map[string][]TestEntry) {
	for environment, actualTests := range actual {
		expectedTests, environmentOk := expected[environment]
		if !environmentOk {
			t.Errorf("Unexpected environment %s in actual", environment)
			continue
		}

		for test, actualEntries := range actualTests {
			expectedEntries, testOk := expectedTests[test]
			if !testOk {
				t.Errorf("Unexpected test %s (in environment %s) in actual", test, environment)
				continue
			}

			compareEntrySlices(t, actualEntries, expectedEntries, fmt.Sprintf("environment %s, test %s", environment, test))
		}

		for test := range expectedTests {
			_, testOk := actualTests[test]
			if !testOk {
				t.Errorf("Missing expected test %s (in environment %s) in actual", test, environment)
			}
		}
	}

	for environment := range expected {
		_, environmentOk := actual[environment]
		if !environmentOk {
			t.Errorf("Missing expected environment %s in actual", environment)
		}
	}
}

func TestSplitData(t *testing.T) {
	entry_e1_t1_1, entry_e1_t1_2 := TestEntry{
		name:        "test1",
		environment: "env1",
		date:        simpleDate(2000, time.January, 1),
		status:      "Passed",
	}, TestEntry{
		name:        "test1",
		environment: "env1",
		date:        simpleDate(2000, time.January, 2),
		status:      "Passed",
	}
	entry_e1_t2 := TestEntry{
		name:        "test2",
		environment: "env1",
		date:        simpleDate(2000, time.January, 1),
		status:      "Passed",
	}
	entry_e2_t1 := TestEntry{
		name:        "test1",
		environment: "env2",
		date:        simpleDate(2000, time.January, 1),
		status:      "Passed",
	}
	entry_e2_t2 := TestEntry{
		name:        "test2",
		environment: "env2",
		date:        simpleDate(2000, time.January, 1),
		status:      "Passed",
	}
	actual := SplitData([]TestEntry{entry_e1_t1_1, entry_e1_t1_2, entry_e1_t2, entry_e2_t1, entry_e2_t2})
	expected := map[string]map[string][]TestEntry{
		"env1": {
			"test1": {entry_e1_t1_1, entry_e1_t1_2},
			"test2": {entry_e1_t2},
		},
		"env2": {
			"test1": {entry_e2_t1},
			"test2": {entry_e2_t2},
		},
	}

	compareSplitData(t, actual, expected)
}

func TestFilterRecentEntries(t *testing.T) {
	entry_e1_t1_r1, entry_e1_t1_r2, entry_e1_t1_r3, entry_e1_t1_o1, entry_e1_t1_o2 := TestEntry{
		name:        "test1",
		environment: "env1",
		date:        simpleDate(2000, time.January, 4),
		status:      "Passed",
	}, TestEntry{
		name:        "test1",
		environment: "env1",
		date:        simpleDate(2000, time.January, 3),
		status:      "Passed",
	}, TestEntry{
		name:        "test1",
		environment: "env1",
		date:        simpleDate(2000, time.January, 3),
		status:      "Passed",
	}, TestEntry{
		name:        "test1",
		environment: "env1",
		date:        simpleDate(2000, time.January, 2),
		status:      "Passed",
	}, TestEntry{
		name:        "test1",
		environment: "env1",
		date:        simpleDate(2000, time.January, 1),
		status:      "Passed",
	}
	entry_e1_t2_r1, entry_e1_t2_r2, entry_e1_t2_o1 := TestEntry{
		name:        "test2",
		environment: "env1",
		date:        simpleDate(2001, time.January, 3),
		status:      "Passed",
	}, TestEntry{
		name:        "test2",
		environment: "env1",
		date:        simpleDate(2001, time.January, 2),
		status:      "Passed",
	}, TestEntry{
		name:        "test2",
		environment: "env1",
		date:        simpleDate(2001, time.January, 1),
		status:      "Passed",
	}
	entry_e2_t2_r1, entry_e2_t2_r2, entry_e2_t2_o1 := TestEntry{
		name:        "test2",
		environment: "env2",
		date:        simpleDate(2003, time.January, 3),
		status:      "Passed",
	}, TestEntry{
		name:        "test2",
		environment: "env2",
		date:        simpleDate(2003, time.January, 2),
		status:      "Passed",
	}, TestEntry{
		name:        "test2",
		environment: "env2",
		date:        simpleDate(2003, time.January, 1),
		status:      "Passed",
	}

	actualData := FilterRecentEntries(map[string]map[string][]TestEntry{
		"env1": {
			"test1": {
				entry_e1_t1_r1,
				entry_e1_t1_r2,
				entry_e1_t1_r3,
				entry_e1_t1_o1,
				entry_e1_t1_o2,
			},
			"test2": {
				entry_e1_t2_r1,
				entry_e1_t2_r2,
				entry_e1_t2_o1,
			},
		},
		"env2": {
			"test2": {
				entry_e2_t2_r1,
				entry_e2_t2_r2,
				entry_e2_t2_o1,
			},
		},
	}, 2)

	expectedData := map[string]map[string][]TestEntry{
		"env1": {
			"test1": {
				entry_e1_t1_r1,
				entry_e1_t1_r2,
				entry_e1_t1_r3,
			},
			"test2": {
				entry_e1_t2_r1,
				entry_e1_t2_r2,
			},
		},
		"env2": {
			"test2": {
				entry_e2_t2_r1,
				entry_e2_t2_r2,
			},
		},
	}

	compareSplitData(t, actualData, expectedData)
}
