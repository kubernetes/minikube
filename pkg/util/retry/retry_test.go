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
	"time"
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

func TestNotify(t *testing.T) {
	// Mock time
	mockTime := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	timeNow = func() time.Time { return mockTime }
	defer func() { timeNow = time.Now }()

	// Overwrite configuration
	// We need to restore them too!
	origDedup := logDedupWindow
	origStuck := logStuckThreshold
	origMax := maxDuplicateLogEntries

	logDedupWindow = 5 * time.Second
	logStuckThreshold = 10 * time.Second
	maxDuplicateLogEntries = 3

	defer func() {
		logDedupWindow = origDedup
		logStuckThreshold = origStuck
		maxDuplicateLogEntries = origMax
	}()

	// Helper to reset state
	resetState := func() {
		logMu.Lock()
		firstLogTime = time.Time{}
		lastLogTime = time.Time{}
		lastLogErr = ""
		logCount = 0
		logMu.Unlock()
	}

	t.Run("DedupLogic", func(t *testing.T) {
		resetState()
		// First call
		notify(errors.New("foo"), time.Second)

		logMu.Lock()
		if logCount != 1 {
			t.Errorf("expected logCount 1 after first call, got %d", logCount)
		}
		logMu.Unlock()

		// Second call soon
		mockTime = mockTime.Add(1 * time.Second)
		notify(errors.New("foo"), time.Second)

		logMu.Lock()
		if logCount != 1 {
			t.Errorf("expected logCount 1 after rapid duplicate, got %d", logCount)
		}
		logMu.Unlock()

		// Third call later
		mockTime = mockTime.Add(6 * time.Second) // Total 7s from start, > 5s window
		notify(errors.New("foo"), time.Second)

		logMu.Lock()
		if logCount != 2 {
			t.Errorf("expected logCount 2 after window, got %d", logCount)
		}
		logMu.Unlock()
	})

	t.Run("MaxDuplicates", func(t *testing.T) {
		resetState()
		notify(errors.New("bar"), time.Second)

		// Exceed maxDuplicateLogEntries (3)
		for i := 0; i < 10; i++ {
			mockTime = mockTime.Add(6 * time.Second) // Always > dedup window
			notify(errors.New("bar"), time.Second)
		}

		logMu.Lock()
		if logCount != 11 {
			t.Errorf("expected logCount 11, got %d", logCount)
		}
		logMu.Unlock()
	})

	t.Run("ResetOnNewError", func(t *testing.T) {
		resetState()
		notify(errors.New("err1"), time.Second)

		mockTime = mockTime.Add(6 * time.Second)
		notify(errors.New("err1"), time.Second)

		logMu.Lock()
		if logCount != 2 {
			t.Errorf("expected logCount 2, got %d", logCount)
		}
		logMu.Unlock()

		// Different error
		mockTime = mockTime.Add(6 * time.Second)
		notify(errors.New("err2"), time.Second)

		logMu.Lock()
		if logCount != 1 {
			t.Errorf("expected logCount 1 after new error, got %d", logCount)
		}
		if lastLogErr != "err2" {
			t.Errorf("expected lastLogErr 'err2', got '%s'", lastLogErr)
		}
		logMu.Unlock()
	})
}
