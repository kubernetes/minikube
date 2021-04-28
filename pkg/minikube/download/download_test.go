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
	"fmt"
	"io/fs"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/constants"
)

type mockLogger struct {
	downloads int
	t         *testing.T
}

func (ml *mockLogger) Enabled() bool {
	return true
}
func (ml *mockLogger) Info(msg string, keysAndValues ...interface{}) {
	fmt.Println("hi - ", msg)
	if strings.Contains(msg, "Mock download") {
		// Make "downloads" take longer to increase lock time.
		dur, err := time.ParseDuration("1s")
		if err != nil {
			ml.t.Errorf("Could not parse 1 second duration - should never happen")
		}
		time.Sleep(dur)

		ml.downloads++
	}
}
func (ml *mockLogger) Error(err error, msg string, keysAndValues ...interface{}) {}
func (ml *mockLogger) V(level int) logr.Logger {
	return ml
}
func (ml *mockLogger) WithValues(keysAndValues ...interface{}) logr.Logger {
	return ml
}
func (ml *mockLogger) WithName(name string) logr.Logger {
	return ml
}

func TestBinaryDownloadPreventsMultipleDownload(t *testing.T) {
	EnableMock(true)
	defer EnableMock(false)
	tlog := &mockLogger{downloads: 0, t: t}

	klog.SetLogger(tlog)
	defer klog.SetLogger(nil)

	checkCache = func(file string) (fs.FileInfo, error) {
		if tlog.downloads == 0 {
			return nil, fmt.Errorf("some error")
		}
		return nil, nil
	}

	var group sync.WaitGroup
	group.Add(2)
	dlCall := func() {
		if _, err := Binary("kubectl", "v1.20.2", "linux", "amd64"); err != nil {
			t.Errorf("Failed to download binary: %+v", err)
		}
		group.Done()
	}

	go dlCall()
	go dlCall()

	group.Wait()

	if tlog.downloads != 1 {
		t.Errorf("Wrong number of downloads occurred. Actual: %v, Expected: 1", tlog.downloads)
	}
}

func TestPreloadDownloadPreventsMultipleDownload(t *testing.T) {
	EnableMock(true)
	defer EnableMock(false)
	tlog := &mockLogger{downloads: 0, t: t}

	klog.SetLogger(tlog)
	defer klog.SetLogger(nil)

	checkCache = func(file string) (fs.FileInfo, error) {
		if tlog.downloads == 0 {
			return nil, fmt.Errorf("some error")
		}
		return nil, nil
	}
	checkPreloadExists = func(k8sVersion, containerRuntime string, forcePreload ...bool) bool { return true }
	compareChecksum = func(k8sVersion, containerRuntime, path string) error { return nil }

	var group sync.WaitGroup
	group.Add(2)
	dlCall := func() {
		if err := Preload(constants.DefaultKubernetesVersion, constants.DefaultContainerRuntime); err != nil {
			t.Errorf("Failed to download preload: %+v", err)
		}
		group.Done()
	}

	go dlCall()
	go dlCall()

	group.Wait()

	if tlog.downloads != 1 {
		t.Errorf("Wrong number of downloads occurred. Actual: %v, Expected: 1", tlog.downloads)
	}
}
