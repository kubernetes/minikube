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

package util

import (
	"fmt"
	"testing"
)

// Returns a function that will return n errors, then return successfully forever.
func errorGenerator(n int) func() error {
	errors := 0
	return func() (err error) {
		if errors < n {
			errors += 1
			return fmt.Errorf("Error!")
		}
		return nil
	}
}

func TestErrorGenerator(t *testing.T) {
	errors := 3
	f := errorGenerator(errors)
	for i := 0; i < errors-1; i++ {
		if err := f(); err == nil {
			t.Fatalf("Error should have been thrown at iteration %v", i)
		}
	}
	if err := f(); err == nil {
		t.Fatalf("Error should not have been thrown this call!")
	}
}

func TestRetry(t *testing.T) {

	f := errorGenerator(4)
	if err := Retry(5, f); err != nil {
		t.Fatalf("Error should not have been raised by retry.")
	}

	f = errorGenerator(5)
	if err := Retry(4, f); err == nil {
		t.Fatalf("Error should have been raised by retry.")
	}

}
