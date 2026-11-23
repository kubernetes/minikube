/*
Copyright 2026 The Kubernetes Authors All rights reserved.

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

package integration

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// Timeout returns a per-test deadline of 3x typical, rounded up to the next minute.
// A partial minute rounds up, so a positive typical shorter than a minute yields one minute.
// The deadline is then scaled by -timeout-multiplier, same as Minutes.
// A zero or negative typical is a programmer error and panics.
func Timeout(typical time.Duration) time.Duration {
	if typical <= 0 {
		panic(fmt.Sprintf("typical duration must be positive, got %s", typical))
	}
	scaled := 3 * typical
	minutes := int((scaled + time.Minute - 1) / time.Minute)
	return Minutes(minutes)
}

// ContextWithTimeout returns a context derived from t.Context() and canceled after Timeout(typical).
//
// Call it only after the last t.Parallel call has returned. t.Parallel blocks until this test
// has a slot, and that wait has no bound. A deadline started earlier expires while the test is
// still queued. Subtests created after ContextWithTimeout must not call t.Parallel: they would
// wait for a slot while this deadline is already running.
//
// Correct: the deadline starts on bar, after bar's t.Parallel has returned.
//
//	t.Run("foo", func(t *testing.T) {
//		t.Parallel()
//		t.Run("bar", func(t *testing.T) {
//			t.Parallel()
//			ctx, cancel := ContextWithTimeout(t, time.Second)
//			defer cancel()
//		})
//	})
//
// Incorrect: foo's deadline runs while bar waits for a slot.
//
//	t.Run("foo", func(t *testing.T) {
//		t.Parallel()
//		ctx, cancel := ContextWithTimeout(t, time.Second)
//		defer cancel()
//		t.Run("bar", func(t *testing.T) {
//			t.Parallel()
//		})
//	})
func ContextWithTimeout(t *testing.T, typical time.Duration) (context.Context, context.CancelFunc) {
	t.Helper()
	return context.WithTimeout(t.Context(), Timeout(typical))
}

func TestTimeout(t *testing.T) {
	t.Run("calculation", func(t *testing.T) {
		cases := []struct {
			typical time.Duration
			minutes int
		}{
			{340 * time.Millisecond, 1},
			{time.Second, 1},
			{20 * time.Second, 1}, // 3x is exactly 1 minute
			{21 * time.Second, 2}, // 3x is 63s
			{22 * time.Second, 2},
			{40 * time.Second, 2}, // 3x is exactly 2 minutes
			{41 * time.Second, 3},
		}
		for _, tc := range cases {
			if got, want := Timeout(tc.typical), Minutes(tc.minutes); got != want {
				t.Errorf("Timeout(%s) = %s, want %s", tc.typical, got, want)
			}
		}
	})

	t.Run("deadline", func(t *testing.T) {
		ctx, cancel := ContextWithTimeout(t, 22*time.Second)
		defer cancel()
		deadline, ok := ctx.Deadline()
		if !ok {
			t.Fatal("context has no deadline")
		}
		if got, want := time.Until(deadline), Timeout(22*time.Second); got < want-time.Second || got > want {
			t.Errorf("deadline in %s, want about %s", got, want)
		}
	})

	t.Run("panic", func(t *testing.T) {
		for _, typical := range []time.Duration{0, -time.Second} {
			t.Run(typical.String(), func(t *testing.T) {
				defer func() {
					if recover() == nil {
						t.Fatalf("Timeout(%s) did not panic", typical)
					}
				}()
				Timeout(typical)
			})
		}
	})
}
