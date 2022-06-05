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

package tests

import (
	"encoding/json"
	"reflect"
	"testing"
)

// TestEvent simulates a CloudEvent for our JSON output
type TestEvent struct {
	Data            map[string]string `json:"data"`
	Datacontenttype string            `json:"datacontenttype"`
	ID              string            `json:"id"`
	Source          string            `json:"source"`
	Specversion     string            `json:"specversion"`
	Eventtype       string            `json:"type"`
}

// CompareJSON takes two byte slices, unmarshals them to TestEvent
// and compares them, failing the test if they don't match
func CompareJSON(t *testing.T, actual, expected []byte) {
	var actualJSON, expectedJSON TestEvent

	err := json.Unmarshal(actual, &actualJSON)
	if err != nil {
		t.Fatalf("error unmarshalling json: %v", err)
	}

	err = json.Unmarshal(expected, &expectedJSON)
	if err != nil {
		t.Fatalf("error unmarshalling json: %v", err)
	}

	if !reflect.DeepEqual(actualJSON, expectedJSON) {
		t.Fatalf("expected didn't match actual:\nExpected:\n%v\n\nActual:\n%v", expected, actual)
	}
}
