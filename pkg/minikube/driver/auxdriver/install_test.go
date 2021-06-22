/*
Copyright 2019 The Kubernetes Authors All rights reserved.

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

package auxdriver

import (
	"testing"
)

func TestExtractDriverVersion(t *testing.T) {
	v := extractDriverVersion("")
	if len(v) != 0 {
		t.Error("Expected empty string")
	}

	v = extractDriverVersion("random text")
	if len(v) != 0 {
		t.Error("Expected empty string")
	}

	expectedVersion := "1.2.3"

	v = extractDriverVersion("version: v1.2.3")
	if expectedVersion != v {
		t.Errorf("Expected version: %s, got: %s", expectedVersion, v)
	}

	v = extractDriverVersion("version: 1.2.3")
	if expectedVersion != v {
		t.Errorf("Expected version: %s, got: %s", expectedVersion, v)
	}
}
