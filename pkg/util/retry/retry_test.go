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

package retry

import (
	"errors"
	"testing"
)

// Returns a function that will return n errors, then return successfully forever.
func errorGenerator(n int, retryable bool) func() error {
	errorCount := 0
	return func() (err error) {
		if errorCount < n {
			errorCount++
			e := errors.New("Error")
			if retryable {
				return &RetriableError{Err: e}
			}
			return e

		}

		return nil
	}
}

func TestErrorGenerator(t *testing.T) {
	errors := 3
	f := errorGenerator(errors, false)
	for i := 0; i < errors-1; i++ {
		if err := f(); err == nil {
			t.Fatalf("Error should have been thrown at iteration %v", i)
		}
	}
	if err := f(); err == nil {
		t.Fatalf("Error should not have been thrown this call!")
	}
}
