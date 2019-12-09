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
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"regexp"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/juju/clock"
	"github.com/juju/mutex"
	"github.com/pkg/errors"
)

var (
	// nonString is characters to strip from lock names
	nonString = regexp.MustCompile(`[\W_]+`)
	// forceID is a user id for consistent testing
	forceID = ""
)

// WriteFile decorates ioutil.WriteFile with a file lock and retry
func WriteFile(filename string, data []byte, perm os.FileMode) error {
	spec := UserMutexSpec(filename)
	glog.Infof("acquiring lock for %s: %+v", filename, spec)
	releaser, err := mutex.Acquire(spec)
	if err != nil {
		return errors.Wrapf(err, "error acquiring lock for %s", filename)
	}

	defer releaser.Release()

	if err = ioutil.WriteFile(filename, data, perm); err != nil {
		return errors.Wrapf(err, "error writing file %s", filename)
	}
	return err
}

// UserMutexSpec returns a mutex spec that will not collide with other users
func UserMutexSpec(path string) mutex.Spec {
	id := forceID
	if forceID == "" {
		u, err := user.Current()
		if err == nil {
			id = u.Uid
		}
	}

	s := mutex.Spec{
		Name:  getMutexNameForPath(fmt.Sprintf("%s-%s", path, id)),
		Clock: clock.WallClock,
		// Poll the lock twice a second
		Delay: 500 * time.Millisecond,
		// panic after a minute instead of locking infinitely
		Timeout: 60 * time.Second,
	}
	return s
}

func getMutexNameForPath(path string) string {
	// juju requires that names match ^[a-zA-Z][a-zA-Z0-9-]*$", and be under 40 chars long.
	n := strings.Trim(nonString.ReplaceAllString(path, "-"), "-")
	// we need to always guarantee an alphanumeric prefix
	prefix := "m"

	// Prefer the last 40 chars, as paths tend get more specific toward the end
	if len(n) >= 40 {
		return prefix + n[len(n)-39:]
	}
	return prefix + n
}
