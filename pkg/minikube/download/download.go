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

	"github.com/golang/glog"
	"github.com/hashicorp/go-getter"
	"github.com/pkg/errors"
)

var (
	// Mock allows tests to toggle making actual HTTP downloads
	Mock = false
)

// download is a well-configured atomic download function
func download(src string, dst string) error {
	tmpDst := dst + ".download"
	client := &getter.Client{
		Src:     src,
		Dst:     tmpDst,
		Dir:     false,
		Mode:    getter.ClientModeFile,
		Options: []getter.ClientOption{getter.WithProgress(DefaultProgressBar)},
		Getters: map[string]getter.Getter{
			"file":  &getter.FileGetter{Copy: false},
			"http":  &getter.HttpGetter{Netrc: false},
			"https": &getter.HttpGetter{Netrc: false},
		},
	}

	// Don't bother with getter.MockGetter, as we don't provide a way to inspect the outcome
	if Mock {
		glog.Infof("Mock download: %s -> %s", src, dst)
		return nil
	}

	// Politely prevent tests from shooting themselves in the foot
	if underTest() {
		return fmt.Errorf("unmocked download under test, set download.Mock=true")
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0750); err != nil {
		return errors.Wrap(err, "mkdir")
	}

	glog.Infof("Downloading: %s -> %s", src, dst)
	if err := client.Get(); err != nil {
		return errors.Wrapf(err, "getter: %+v", client)
	}
	return os.Rename(tmpDst, dst)
}

// detect if we are under test
func underTest() bool {
	return flag.Lookup("test.v") != nil || strings.HasSuffix(os.Args[0], "test")
}
