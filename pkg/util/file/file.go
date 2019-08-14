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

package file

import (
	"io/ioutil"
	"os"
	"time"

	"github.com/golang/glog"
	"github.com/juju/fslock"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/util/retry"
)

// WriteWithLock wraps ioutil.WriteFile with a file lock and retry
func WriteWithLock(filename string, data []byte, perm os.FileMode) error {
	lock := fslock.New(filename)

	getLock := func() error {
		lockErr := lock.TryLock()
		if lockErr != nil {
			glog.Infof("temproary error : %v", lockErr.Error())
			return errors.Wrapf(lockErr, "falied to acquire file lock for %s > ", filename)
		}
		return nil
	}
	err := retry.Expo(getLock, 1*time.Second, 13*time.Second)
	if err != nil {
		return errors.Wrapf(err, "acquiring file lock for %s", filename)
	}

	if err := ioutil.WriteFile(filename, data, perm); err != nil {
		return errors.Wrapf(err, "error writing file %s", filename)
	}

	// release the lock
	err = lock.Unlock()
	if err != nil {
		return errors.Wrapf(err, "error releasing lock for file: %s", filename)
	}
	return nil
}
