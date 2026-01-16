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

// Package retry implements wrappers to retry function calls
package retry

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v5"

	"k8s.io/klog/v2"
)

const (
	defaultMaxRetries = 113
	// logRoundPrecision is the precision used for rounding duration in logs
	logRoundPrecision = 100 * time.Millisecond
)

var (
	// logDedupWindow is the minimum time between identical log messages
	logDedupWindow = 3 * time.Second
	// logStuckThreshold is the time after which a persistent error is flagged as "maybe stuck"
	logStuckThreshold = 30 * time.Second
	// maxDuplicateLogEntries is the maximum number of times a specific error is logged before displaying a shorter message
	maxDuplicateLogEntries = 5
	// timeNow is the func used for testing
	timeNow = time.Now
)

var (
	firstLogTime time.Time
	lastLogTime  time.Time
	lastLogErr   string
	logCount     int
	logMu        sync.Mutex
)

func notify(err error, d time.Duration) {
	logMu.Lock()
	if err.Error() != lastLogErr {
		firstLogTime = timeNow()
		lastLogErr = err.Error()
		logCount = 0
	}

	now := timeNow()
	if now.Sub(lastLogTime) < logDedupWindow {
		lastLogTime = now
		logMu.Unlock()
		return
	}
	lastLogTime = now
	logCount++

	if logCount > maxDuplicateLogEntries { // do not repeat the error message after maxDuplicateLogEntries
		logMu.Unlock()
		klog.Infof("will retry after %s: stuck on same error as above for %s...", d.Round(logRoundPrecision), timeNow().Sub(firstLogTime).Round(logRoundPrecision))
		return
	}

	msg := fmt.Sprintf("will retry after %s: %v", d.Round(logRoundPrecision), err)
	if timeNow().Sub(firstLogTime) > logStuckThreshold { // let user know if the error is repeated within logStuckThreshold
		msg += fmt.Sprintf(" (duplicate log for %s)", timeNow().Sub(firstLogTime).Round(logRoundPrecision))
	}
	logMu.Unlock()

	klog.Info(msg)
}

// Local is back-off retry for local connections
func Local(callback func() error, maxTime time.Duration) error {
	b := backoff.NewExponentialBackOff()
	b.InitialInterval = 250 * time.Millisecond
	b.RandomizationFactor = 0.25
	b.Multiplier = 1.25

	op := func() (struct{}, error) {
		return struct{}{}, callback()
	}
	_, err := backoff.Retry(context.Background(), op, backoff.WithBackOff(b), backoff.WithMaxElapsedTime(maxTime), backoff.WithNotify(notify))
	return err
}

// Expo is exponential backoff retry.
// initInterval is the initial waiting time to start with.
// maxTime is the max time allowed to spend on the all the retries.
// maxRetries is the optional max number of retries allowed with default of 113.
func Expo(callback func() error, initInterval time.Duration, maxTime time.Duration, maxRetries ...uint64) error {
	maxRetry := uint64(defaultMaxRetries) // max number of times to retry
	if len(maxRetries) > 0 {
		maxRetry = maxRetries[0]
	}

	b := backoff.NewExponentialBackOff()
	b.InitialInterval = initInterval
	b.RandomizationFactor = 0.5
	b.Multiplier = 1.5

	op := func() (struct{}, error) {
		return struct{}{}, callback()
	}
	_, err := backoff.Retry(context.Background(), op,
		backoff.WithBackOff(b),
		backoff.WithMaxElapsedTime(maxTime),
		backoff.WithMaxTries(uint(maxRetry)),
		backoff.WithNotify(notify),
	)
	return err
}

// RetriableError is an error that can be tried again
type RetriableError struct {
	Err error
}

func (r RetriableError) Error() string { return "Temporary Error: " + r.Err.Error() }
