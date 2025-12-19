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
	"fmt"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"

	"k8s.io/klog/v2"
)

const (
	defaultMaxRetries = 113
	// logDedupWindow is the minimum time between identical log messages
	logDedupWindow = 3 * time.Second
	// logStuckThreshold is the time after which a persistent error is flagged as "maybe stuck"
	logStuckThreshold = 30 * time.Second
	// maxDuplicateLogEntries is the maximum number of times a specific error is logged before suppressing it
	maxDuplicateLogEntries = 10
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
		firstLogTime = time.Now()
		lastLogErr = err.Error()
		logCount = 0
	}

	now := time.Now()
	if time.Since(lastLogTime) < logDedupWindow {
		lastLogTime = now
		logMu.Unlock()
		return
	}
	lastLogTime = now
	logCount++

	if logCount > maxDuplicateLogEntries {
		logMu.Unlock()
		klog.Infof("will retry after %s: stuck on same error as above...", d)
		return
	}

	msg := fmt.Sprintf("will retry after %s: %v", d, err)
	if time.Since(firstLogTime) > logStuckThreshold {
		msg += fmt.Sprintf(" - maybe stuck %s", time.Since(firstLogTime).Round(time.Second))
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
	b.MaxElapsedTime = maxTime
	return backoff.RetryNotify(callback, b, notify)
}

// Expo is exponential backoff retry.
// initInterval is the initial waiting time to start with.
// maxTime is the max time allowed to spend on the all the retries.
// maxRetries is the optional max number of retries allowed with default of 13.
func Expo(callback func() error, initInterval time.Duration, maxTime time.Duration, maxRetries ...uint64) error {
	maxRetry := uint64(defaultMaxRetries) // max number of times to retry
	if maxRetries != nil {
		maxRetry = maxRetries[0]
	}

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = maxTime
	b.InitialInterval = initInterval
	b.RandomizationFactor = 0.5
	b.Multiplier = 1.5
	bm := backoff.WithMaxRetries(b, maxRetry)
	return backoff.RetryNotify(callback, bm, notify)
}

// RetriableError is an error that can be tried again
type RetriableError struct {
	Err error
}

func (r RetriableError) Error() string { return "Temporary Error: " + r.Err.Error() }
