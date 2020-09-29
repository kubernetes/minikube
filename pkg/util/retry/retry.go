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
	"time"

	"github.com/cenkalti/backoff"
	
	"k8s.io/klog/v2"
)

const defaultMaxRetries = 113

func notify(err error, d time.Duration) {
	klog.Infof("will retry after %s: %v", d, err)
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
