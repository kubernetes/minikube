/*
Copyright 2021 The Kubernetes Authors All rights reserved.

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

func simpleDate(year int, day int) time.Time {
	return time.Date(year, time.January, day, 0, 0, 0, 0, time.UTC)
}

func compareEntrySlices(t *testing.T, actualData, expectedData []testEntry, extra string) {
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
	actualData := readData(strings.NewReader(
		`A,B,C,D,E,F
		hash,2000-01-01,env1,test1,Passed,1
		hash,2001-01-01,env2,test2,Failed,0.5
		hash,,,test1,,0.6
		hash,2002-01-01,,,Passed,0.9
		hash,2003-01-01,env3,test3,Passed,2`,
	))
	expectedData := []testEntry{
		{
			name:        "test1",
			environment: "env1",
			date:        simpleDate(2000, 1),
			status:      "Passed",
			duration:    1,
		},
		{
			name:        "test2",
			environment: "env2",
			date:        simpleDate(2001, 1),
			status:      "Failed",
			duration:    0.5,
		},
		{
			name:        "test1",
			environment: "env2",
			date:        simpleDate(2001, 1),
			status:      "Failed",
			duration:    0.6,
		},
		{
			name:        "test1",
			environment: "env2",
			date:        simpleDate(2002, 1),
			status:      "Passed",
			duration:    0.9,
		},
		{
			name:        "test3",
			environment: "env3",
			date:        simpleDate(2003, 1),
			status:      "Passed",
			duration:    2,
		},
	}

	compareEntrySlices(t, actualData, expectedData, "")
}

func compareSplitData(t *testing.T, actual, expected splitEntryMap) {
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
	entryE1T1_1, entryE1T1_2 := testEntry{
		name:        "test1",
		environment: "env1",
		date:        simpleDate(2000, 1),
		status:      "Passed",
	}, testEntry{
		name:        "test1",
		environment: "env1",
		date:        simpleDate(2000, 2),
		status:      "Passed",
	}
	entryE1T2 := testEntry{
		name:        "test2",
		environment: "env1",
		date:        simpleDate(2000, 1),
		status:      "Passed",
	}
	entryE2T1 := testEntry{
		name:        "test1",
		environment: "env2",
		date:        simpleDate(2000, 1),
		status:      "Passed",
	}
	entryE2T2 := testEntry{
		name:        "test2",
		environment: "env2",
		date:        simpleDate(2000, 1),
		status:      "Passed",
	}
	actual := splitData([]testEntry{entryE1T1_1, entryE1T1_2, entryE1T2, entryE2T1, entryE2T2})
	expected := splitEntryMap{
		"env1": {
			"test1": {entryE1T1_1, entryE1T1_2},
			"test2": {entryE1T2},
		},
		"env2": {
			"test1": {entryE2T1},
			"test2": {entryE2T2},
		},
	}

	compareSplitData(t, actual, expected)
}

func TestFilterRecentEntries(t *testing.T) {
	entryE1T1R1, entryE1T1R2, entryE1T1R3, entryE1T1O1, entryE1T1O2 := testEntry{
		name:        "test1",
		environment: "env1",
		date:        simpleDate(2000, 4),
		status:      "Passed",
	}, testEntry{
		name:        "test1",
		environment: "env1",
		date:        simpleDate(2000, 3),
		status:      "Passed",
	}, testEntry{
		name:        "test1",
		environment: "env1",
		date:        simpleDate(2000, 3),
		status:      "Passed",
	}, testEntry{
		name:        "test1",
		environment: "env1",
		date:        simpleDate(2000, 2),
		status:      "Passed",
	}, testEntry{
		name:        "test1",
		environment: "env1",
		date:        simpleDate(2000, 1),
		status:      "Passed",
	}
	entryE1T2R1, entryE1T2R2, entryE1T2O1 := testEntry{
		name:        "test2",
		environment: "env1",
		date:        simpleDate(2001, 3),
		status:      "Passed",
	}, testEntry{
		name:        "test2",
		environment: "env1",
		date:        simpleDate(2001, 2),
		status:      "Passed",
	}, testEntry{
		name:        "test2",
		environment: "env1",
		date:        simpleDate(2001, 1),
		status:      "Passed",
	}
	entryE2T2R1, entryE2T2R2, entryE2T2O1 := testEntry{
		name:        "test2",
		environment: "env2",
		date:        simpleDate(2003, 3),
		status:      "Passed",
	}, testEntry{
		name:        "test2",
		environment: "env2",
		date:        simpleDate(2003, 2),
		status:      "Passed",
	}, testEntry{
		name:        "test2",
		environment: "env2",
		date:        simpleDate(2003, 1),
		status:      "Passed",
	}

	actualData := filterRecentEntries(splitEntryMap{
		"env1": {
			"test1": {
				entryE1T1R1,
				entryE1T1R2,
				entryE1T1R3,
				entryE1T1O1,
				entryE1T1O2,
			},
			"test2": {
				entryE1T2R1,
				entryE1T2R2,
				entryE1T2O1,
			},
		},
		"env2": {
			"test2": {
				entryE2T2R1,
				entryE2T2R2,
				entryE2T2O1,
			},
		},
	}, 2)

	expectedData := splitEntryMap{
		"env1": {
			"test1": {
				entryE1T1R1,
				entryE1T1R2,
				entryE1T1R3,
			},
			"test2": {
				entryE1T2R1,
				entryE1T2R2,
			},
		},
		"env2": {
			"test2": {
				entryE2T2R1,
				entryE2T2R2,
			},
		},
	}

	compareSplitData(t, actualData, expectedData)
}

func compareValues(t *testing.T, actualValues, expectedValues map[string]map[string]float32) {
	for environment, actualTests := range actualValues {
		expectedTests, environmentOk := expectedValues[environment]
		if !environmentOk {
			t.Errorf("Unexpected environment %s in actual", environment)
			continue
		}

		for test, actualValue := range actualTests {
			expectedValue, testOk := expectedTests[test]
			if !testOk {
				t.Errorf("Unexpected test %s (in environment %s) in actual", test, environment)
				continue
			}

			if actualValue != expectedValue {
				t.Errorf("Wrong value at environment %s and test %s. Expected: %v, Actual: %v", environment, test, expectedValue, actualValue)
			}
		}

		for test := range expectedTests {
			_, testOk := actualTests[test]
			if !testOk {
				t.Errorf("Missing expected test %s (in environment %s) in actual", test, environment)
			}
		}
	}

	for environment := range expectedValues {
		_, environmentOk := actualValues[environment]
		if !environmentOk {
			t.Errorf("Missing expected environment %s in actual", environment)
		}
	}
}

func TestComputeFlakeRates(t *testing.T) {
	actualData := computeFlakeRates(splitEntryMap{
		"env1": {
			"test1": {
				{
					name:        "test1",
					environment: "env1",
					date:        simpleDate(2000, 4),
					status:      "Passed",
				}, {
					name:        "test1",
					environment: "env1",
					date:        simpleDate(2000, 3),
					status:      "Passed",
				}, {
					name:        "test1",
					environment: "env1",
					date:        simpleDate(2000, 3),
					status:      "Passed",
				}, {
					name:        "test1",
					environment: "env1",
					date:        simpleDate(2000, 2),
					status:      "Passed",
				}, {
					name:        "test1",
					environment: "env1",
					date:        simpleDate(2000, 1),
					status:      "Failed",
				},
			},
			"test2": {
				{
					name:        "test2",
					environment: "env1",
					date:        simpleDate(2001, 3),
					status:      "Failed",
				}, {
					name:        "test2",
					environment: "env1",
					date:        simpleDate(2001, 2),
					status:      "Failed",
				}, {
					name:        "test2",
					environment: "env1",
					date:        simpleDate(2001, 1),
					status:      "Failed",
				},
			},
		},
		"env2": {
			"test2": {
				{
					name:        "test2",
					environment: "env2",
					date:        simpleDate(2003, 3),
					status:      "Passed",
				}, testEntry{
					name:        "test2",
					environment: "env2",
					date:        simpleDate(2003, 2),
					status:      "Failed",
				},
			},
		},
	})

	expectedData := map[string]map[string]float32{
		"env1": {
			"test1": 0.2,
			"test2": 1,
		},
		"env2": {
			"test2": 0.5,
		},
	}

	compareValues(t, actualData, expectedData)
}

func TestComputeAverageDurations(t *testing.T) {
	actualData := computeAverageDurations(splitEntryMap{
		"env1": {
			"test1": {
				{
					name:        "test1",
					environment: "env1",
					date:        simpleDate(2000, 4),
					status:      "Passed",
					duration:    1,
				}, {
					name:        "test1",
					environment: "env1",
					date:        simpleDate(2000, 3),
					status:      "Passed",
					duration:    2,
				}, {
					name:        "test1",
					environment: "env1",
					date:        simpleDate(2000, 3),
					status:      "Passed",
					duration:    3,
				}, {
					name:        "test1",
					environment: "env1",
					date:        simpleDate(2000, 2),
					status:      "Passed",
					duration:    3,
				}, {
					name:        "test1",
					environment: "env1",
					date:        simpleDate(2000, 1),
					status:      "Failed",
					duration:    3,
				},
			},
			"test2": {
				{
					name:        "test2",
					environment: "env1",
					date:        simpleDate(2001, 3),
					status:      "Failed",
					duration:    1,
				}, {
					name:        "test2",
					environment: "env1",
					date:        simpleDate(2001, 2),
					status:      "Failed",
					duration:    3,
				}, {
					name:        "test2",
					environment: "env1",
					date:        simpleDate(2001, 1),
					status:      "Failed",
					duration:    3,
				},
			},
		},
		"env2": {
			"test2": {
				{
					name:        "test2",
					environment: "env2",
					date:        simpleDate(2003, 3),
					status:      "Passed",
					duration:    0.5,
				}, testEntry{
					name:        "test2",
					environment: "env2",
					date:        simpleDate(2003, 2),
					status:      "Failed",
					duration:    1.5,
				},
			},
		},
	})

	expectedData := map[string]map[string]float32{
		"env1": {
			"test1": float32(12) / float32(5),
			"test2": float32(7) / float32(3),
		},
		"env2": {
			"test2": 1,
		},
	}

	compareValues(t, actualData, expectedData)
}
