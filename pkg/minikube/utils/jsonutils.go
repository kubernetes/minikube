package utils

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
