/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package assets

import "testing"

// mapsEqual returns true if and only if `a` contains all the same pairs as `b`.
func mapsEqual(a, b map[string]string) bool {
	for aKey, aValue := range a {
		if bValue, ok := b[aKey]; !ok || aValue != bValue {
			return false
		}
	}

	for bKey := range b {
		if _, ok := a[bKey]; !ok {
			return false
		}
	}
	return true
}

func TestParseMapString(t *testing.T) {
	cases := map[string]map[string]string{
		"Ardvark=1,B=2,Cantaloupe=3":         {"Ardvark": "1", "B": "2", "Cantaloupe": "3"},
		"A=,B=2,C=":                          {"A": "", "B": "2", "C": ""},
		"":                                   {},
		"malformed,good=howdy,manyequals==,": {"good": "howdy"},
	}
	for actual, expected := range cases {
		if parsedMap := parseMapString(actual); !mapsEqual(parsedMap, expected) {
			t.Errorf("Parsed map from string \"%s\" differs from expected map: Actual: %v Expected: %v", actual, parsedMap, expected)
		}
	}
}

func TestMergeMaps(t *testing.T) {
	type TestCase struct {
		sourceMap   map[string]string
		overrideMap map[string]string
		expectedMap map[string]string
	}
	cases := []TestCase{
		{
			sourceMap:   map[string]string{"A": "1", "B": "2"},
			overrideMap: map[string]string{"B": "7", "C": "3"},
			expectedMap: map[string]string{"A": "1", "B": "7", "C": "3"},
		},
		{
			sourceMap:   map[string]string{"B": "7", "C": "3"},
			overrideMap: map[string]string{"A": "1", "B": "2"},
			expectedMap: map[string]string{"A": "1", "B": "2", "C": "3"},
		},
		{
			sourceMap:   map[string]string{"B": "7", "C": "3"},
			overrideMap: map[string]string{},
			expectedMap: map[string]string{"B": "7", "C": "3"},
		},
		{
			sourceMap:   map[string]string{},
			overrideMap: map[string]string{"B": "7", "C": "3"},
			expectedMap: map[string]string{"B": "7", "C": "3"},
		},
	}
	for _, test := range cases {
		if actualMap := mergeMaps(test.sourceMap, test.overrideMap); !mapsEqual(actualMap, test.expectedMap) {
			t.Errorf("Merging maps (source=%v, override=%v) differs from expected map: Actual: %v Expected: %v", test.sourceMap, test.overrideMap, actualMap, test.expectedMap)
		}
	}
}

func TestFilterKeySpace(t *testing.T) {
	type TestCase struct {
		keySpace    map[string]string
		targetMap   map[string]string
		expectedMap map[string]string
	}
	cases := []TestCase{
		{
			keySpace:    map[string]string{"A": "0", "B": ""},
			targetMap:   map[string]string{"B": "1", "C": "2", "D": "3"},
			expectedMap: map[string]string{"B": "1"},
		},
		{
			keySpace:    map[string]string{},
			targetMap:   map[string]string{"B": "1", "C": "2", "D": "3"},
			expectedMap: map[string]string{},
		},
		{
			keySpace:    map[string]string{"B": "1", "C": "2", "D": "3"},
			targetMap:   map[string]string{},
			expectedMap: map[string]string{},
		},
	}
	for _, test := range cases {
		if actualMap := filterKeySpace(test.keySpace, test.targetMap); !mapsEqual(actualMap, test.expectedMap) {
			t.Errorf("Filtering keyspace of map (keyspace=%v, target=%v) differs from expected map: Actual: %v Expected: %v", test.keySpace, test.targetMap, actualMap, test.expectedMap)
		}
	}
}

func TestOverrideDefautls(t *testing.T) {
	type TestCase struct {
		defaultMap  map[string]string
		overrideMap map[string]string
		expectedMap map[string]string
	}
	cases := []TestCase{
		{
			defaultMap:  map[string]string{"A": "1", "B": "2", "C": "3"},
			overrideMap: map[string]string{"B": "7", "C": "8"},
			expectedMap: map[string]string{"A": "1", "B": "7", "C": "8"},
		},
		{
			defaultMap:  map[string]string{"A": "1", "B": "2", "C": "3"},
			overrideMap: map[string]string{"B": "7", "D": "8", "E": "9"},
			expectedMap: map[string]string{"A": "1", "B": "7", "C": "3"},
		},
		{
			defaultMap:  map[string]string{"A": "1", "B": "2", "C": "3"},
			overrideMap: map[string]string{"B": "7", "D": "8", "E": "9"},
			expectedMap: map[string]string{"A": "1", "B": "7", "C": "3"},
		},
	}
	for _, test := range cases {
		if actualMap := overrideDefaults(test.defaultMap, test.overrideMap); !mapsEqual(actualMap, test.expectedMap) {
			t.Errorf("Override defaults (defaults=%v, overrides=%v) differs from expected map: Actual: %v Expected: %v", test.defaultMap, test.overrideMap, actualMap, test.expectedMap)
		}
	}
}
