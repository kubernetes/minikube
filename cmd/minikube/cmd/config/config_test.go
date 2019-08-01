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

package config

import (
	"bytes"
	"fmt"
	"testing"
)

func TestHiddenPrint(t *testing.T) {
	testCases := []struct {
		TestString  string
		Verbose     bool
		ShouldError bool
	}{
		{
			TestString: "gabbagabbahey",
		},
		{
			TestString: "gabbagabbahey",
			Verbose:    true,
		},
	}
	for _, test := range testCases {
		b := new(bytes.Buffer)
		_, err := b.WriteString(fmt.Sprintf("%s\r\n", test.TestString)) // you need the \r!
		if err != nil {
			t.Errorf("Could not prepare bytestring")
		}
		result, err := concealableAskForStaticValue(b, "hello", false)
		if err != nil && !test.ShouldError {
			t.Errorf("Error asking for concealable static value: %v", err)
			continue
		}
		if result != test.TestString {
			t.Errorf("Result %s not match %s", result, test.TestString)
		}
	}
}
