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
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/juju/clock"
	"github.com/juju/mutex"
	"github.com/pkg/errors"
)

// WriteFile decorates ioutil.WriteFile with a file lock and retry
func WriteFile(filename string, data []byte, perm os.FileMode) error {
	spec := mutex.Spec{
		Name:  getMutexName(filename),
		Clock: clock.WallClock,
		Delay: 13 * time.Second,
	}
	glog.Infof("attempting to write to file %q with filemode %v", filename, perm)

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

func getMutexName(filename string) string {
	// Make the mutex name the file name and its parent directory
	dir, name := filepath.Split(filename)

	// Replace underscores and periods with dashes, the only valid punctuation for mutex name
	name = strings.ReplaceAll(name, ".", "-")
	name = strings.ReplaceAll(name, "_", "-")

	p := strings.ReplaceAll(filepath.Base(dir), ".", "-")
	p = strings.ReplaceAll(p, "_", "-")
	mutexName := fmt.Sprintf("%s-%s", p, strings.ReplaceAll(name, ".", "-"))

	// Check if name starts with an int and prepend a string instead
	if _, err := strconv.Atoi(mutexName[:1]); err == nil {
		mutexName = "m" + mutexName
	}
	// There's an arbitrary hard max on mutex name at 40.
	if len(mutexName) > 40 {
		mutexName = mutexName[:40]
	}

	// Make sure name doesn't start or end with punctuation
	mutexName = strings.TrimPrefix(mutexName, "-")
	mutexName = strings.TrimSuffix(mutexName, "-")

	return mutexName
}
