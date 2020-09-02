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
	"io/ioutil"
	"os"
	"time"

	"github.com/golang/glog"
	"github.com/juju/clock"
	"github.com/juju/mutex"
	"github.com/pkg/errors"
)

// WriteFile decorates ioutil.WriteFile with a file lock and retry
func WriteFile(filename string, data []byte, perm os.FileMode) error {
	spec := PathMutexSpec(filename)
	glog.Infof("WriteFile acquiring %s: %+v", filename, spec)
	releaser, err := mutex.Acquire(spec)
	if err != nil {
		return errors.Wrapf(err, "failed to acquire lock for %s: %+v", filename, spec)
	}

	defer releaser.Release()

	return ioutil.WriteFile(filename, data, perm)
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
