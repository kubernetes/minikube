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
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/juju/clock"
	"github.com/juju/mutex"
	"github.com/pkg/errors"
)

// WriteFile decorates ioutil.WriteFile with a file lock and retry
func WriteFile(filename string, data []byte, perm os.FileMode) error {
	dir, name := filepath.Split(filename)
	// Make the mutex name the file name and its parent directory
	profile := strings.ReplaceAll(filepath.Base(dir), ".", "-")
	mutexName := fmt.Sprintf("%s-%s", profile, strings.ReplaceAll(name, ".", "-"))
	// There's an arbitrary hard max on mutex name at 40.
	if len(mutexName) > 40 {
		mutexName = mutexName[:40]
	}
	spec := mutex.Spec{
		Name:  strings.TrimPrefix(mutexName, "-"),
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
