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
	"context"
	"crypto/sha1"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gofrs/flock"
	"github.com/pkg/errors"

	"k8s.io/klog/v2"
)

// Releaser is an interface for releasing a lock
type Releaser interface {
	Release()
}

// Spec describes the lock to acquire
type Spec struct {
	Name    string
	Timeout time.Duration
	Delay   time.Duration
}

// flockReleaser adapts flock.Flock to Releaser interface
type flockReleaser struct {
	f *flock.Flock
}

func (r *flockReleaser) Release() {
	if err := r.f.Unlock(); err != nil {
		klog.Errorf("failed to release lock: %v", err)
	}
}

// WriteFile decorates os.WriteFile with a file lock and retry
func WriteFile(filename string, data []byte, perm os.FileMode) error {
	spec := PathMutexSpec(filename)
	klog.Infof("WriteFile acquiring %s: %+v", filename, spec)
	releaser, err := Acquire(spec)
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
	releaser, err := Acquire(spec)
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
func PathMutexSpec(path string) Spec {
	s := Spec{
		Name:    fmt.Sprintf("mk%x", sha1.Sum([]byte(path)))[0:40],
		Delay:   500 * time.Millisecond,
		Timeout: 60 * time.Second,
	}
	return s
}

// Acquire acquires the lock specified by spec
func Acquire(spec Spec) (Releaser, error) {
	tmpDir := os.TempDir()
	lockDir := filepath.Join(tmpDir, "minikube-locks")
	if err := os.MkdirAll(lockDir, 0755); err != nil {
		return nil, errors.Wrap(err, "creating lock dir")
	}

	lockPath := filepath.Join(lockDir, spec.Name+".lock")
	f := flock.New(lockPath)

	ctx, cancel := context.WithTimeout(context.Background(), spec.Timeout)
	defer cancel()

	// TryLockContext will retry every spec.Delay until success or context cancellation
	locked, err := f.TryLockContext(ctx, spec.Delay)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, errors.Errorf("timed out waiting for lock %s", spec.Name)
		}
		return nil, errors.Wrap(err, "acquiring lock")
	}

	if !locked {
		return nil, errors.Errorf("failed to acquire lock %s", spec.Name)
	}

	return &flockReleaser{f: f}, nil
}
