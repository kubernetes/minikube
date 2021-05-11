/*
Copyright 2020 The Kubernetes Authors All rights reserved.

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

package download

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/go-getter"
	"github.com/juju/mutex"
	"github.com/pkg/errors"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/style"
	"k8s.io/minikube/pkg/util/lock"
)

var (
	// DownloadMock is called instead of the download implementation if not nil.
	DownloadMock func(src, dst string) error = nil
	checkCache                               = os.Stat
)

// CreateDstDownloadMock is the default mock implementation of download.
func CreateDstDownloadMock(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return errors.Wrap(err, "mkdir")
	}

	_, err := os.Create(dst)
	return err
}

// download is a well-configured atomic download function
func download(src string, dst string) error {
	progress := getter.WithProgress(DefaultProgressBar)
	if out.JSON {
		progress = getter.WithProgress(DefaultJSONOutput)
	}
	tmpDst := dst + ".download"
	client := &getter.Client{
		Src:     src,
		Dst:     tmpDst,
		Dir:     false,
		Mode:    getter.ClientModeFile,
		Options: []getter.ClientOption{progress},
		Getters: map[string]getter.Getter{
			"file":  &getter.FileGetter{Copy: false},
			"http":  &getter.HttpGetter{Netrc: false},
			"https": &getter.HttpGetter{Netrc: false},
		},
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return errors.Wrap(err, "mkdir")
	}

	if DownloadMock != nil {
		klog.Infof("Mock download: %s -> %s", src, dst)
		return DownloadMock(src, dst)
	}

	// Politely prevent tests from shooting themselves in the foot
	if withinUnitTest() {
		return fmt.Errorf("unmocked download under test")
	}

	klog.Infof("Downloading: %s -> %s", src, dst)
	if err := client.Get(); err != nil {
		return errors.Wrapf(err, "getter: %+v", client)
	}
	return os.Rename(tmpDst, dst)
}

// withinUnitTset detects if we are in running within a unit-test
func withinUnitTest() bool {
	// Nope, it's the integration test
	if flag.Lookup("minikube-start-args") != nil || strings.HasPrefix(filepath.Base(os.Args[0]), "e2e-") {
		return false
	}

	return flag.Lookup("test.v") != nil || strings.HasSuffix(os.Args[0], "test")
}

// lockDownload locks `file` if possible and returns a releaser that must be called to release the lock.
func lockDownload(file string) (mutex.Releaser, error) {
	type retPair struct {
		mutex.Releaser
		error
	}
	lockChannel := make(chan retPair)

	go func() {
		spec := lock.PathMutexSpec(file)
		releaser, err := mutex.Acquire(spec)
		if err != nil {
			lockChannel <- retPair{nil, errors.Wrapf(err, "failed to acquire lock \"%s\": %+v", file, spec)}
			return
		}
		lockChannel <- retPair{releaser, err}
	}()

	select {
	case r := <-lockChannel:
		return r.Releaser, r.error
	case <-time.After(time.Millisecond * 100):
		out.Step(style.WaitingWithSpinner, "Another minikube instance is downloading dependencies... ")
	}

	// lock.PathMutexSpec returns a spec including a 60s timeout. Therefore, this
	// will not block indefinitely.
	r := <-lockChannel
	return r.Releaser, r.error
}
