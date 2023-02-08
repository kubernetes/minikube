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

package lock

import (
	"crypto/sha1"
	"fmt"
	"os"
	"time"

	"github.com/juju/clock"
	"github.com/juju/mutex/v2"
	"github.com/pkg/errors"

	"k8s.io/klog/v2"
)

// WriteFile decorates os.WriteFile with a file lock and retry
func WriteFile(filename string, data []byte, perm os.FileMode) error {
	spec := PathMutexSpec(filename)
	klog.Infof("WriteFile acquiring %s: %+v", filename, spec)
	releaser, err := mutex.Acquire(spec)
	if err != nil {
		return errors.Wrapf(err, "failed to acquire lock for %s: %+v", filename, spec)
	}

	defer releaser.Release()

	return os.WriteFile(filename, data, perm)
}

// AppendToFile appends DATA bytes to the specified FILENAME in a mutually exclusive way.
// The file is created if it does not exist, using the specified PERM (before umask)
func AppendToFile(filename string, data []byte, perm os.FileMode) error {
	spec := PathMutexSpec(filename)
	klog.Infof("WriteFile acquiring %s: %+v", filename, spec)
	releaser, err := mutex.Acquire(spec)
	if err != nil {
		return errors.Wrapf(err, "failed to acquire lock for %s: %+v", filename, spec)
	}

	defer releaser.Release()

	fd, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, perm)
	if err != nil {
		return errors.Wrapf(err, "failed to open %s: %+v", filename, spec)
	}

	_, err = fd.Write(data)
	return err
}

// PathMutexSpec returns a mutex spec for a path
func PathMutexSpec(path string) mutex.Spec {
	s := mutex.Spec{
		Name:  fmt.Sprintf("mk%x", sha1.Sum([]byte(path)))[0:40],
		Clock: clock.WallClock,
		// Poll the lock twice a second
		Delay: 500 * time.Millisecond,
		// panic after a minute instead of locking infinitely
		Timeout: 60 * time.Second,
	}
	return s
}
