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
	"io/ioutil"
	"os"
	"runtime"
	"testing"
)

func TestCacheBinary(t *testing.T) {
	oldMinikubeHome := os.Getenv("MINIKUBE_HOME")
	defer os.Setenv("MINIKUBE_HOME", oldMinikubeHome)

	minikubeHome, err := ioutil.TempDir("/tmp", "")
	if err != nil {
		t.Fatalf("error during creating tmp dir: %v", err)
	}
	defer os.RemoveAll(minikubeHome)
	noWritePermDir, err := ioutil.TempDir("/tmp", "")
	if err != nil {
		t.Fatalf("error during creating tmp dir: %v", err)
	}
	defer os.RemoveAll(noWritePermDir)
	err = os.Chmod(noWritePermDir, 0000)
	if err != nil {
		t.Fatalf("error (%v) during changing permissions of dir %v", err, noWritePermDir)
	}

	var tc = []struct {
		desc, version, osName, archName   string
		minikubeHome, binary, description string
		err                               bool
	}{
		{
			desc:         "ok kubeadm",
			version:      "v1.16.0",
			osName:       "linux",
			archName:     runtime.GOARCH,
			binary:       "kubeadm",
			err:          false,
			minikubeHome: minikubeHome,
		},
		{
			desc:         "minikube home in dir without perms and arm runtime",
			version:      "v1.16.0",
			osName:       runtime.GOOS,
			archName:     "arm",
			binary:       "kubectl",
			err:          true,
			minikubeHome: noWritePermDir,
		},
		{
			desc:         "minikube home in dir without perms",
			version:      "v1.16.0",
			osName:       runtime.GOOS,
			archName:     runtime.GOARCH,
			binary:       "kubectl",
			err:          true,
			minikubeHome: noWritePermDir,
		},
		{
			desc:         "binary foo",
			version:      "v1.16.0",
			osName:       runtime.GOOS,
			archName:     runtime.GOARCH,
			binary:       "foo",
			err:          true,
			minikubeHome: minikubeHome,
		},
		{
			desc:         "version 9000",
			version:      "v9000",
			osName:       runtime.GOOS,
			archName:     runtime.GOARCH,
			binary:       "foo",
			err:          true,
			minikubeHome: minikubeHome,
		},
		{
			desc:         "bad os",
			version:      "v1.16.0",
			osName:       "no-such-os",
			archName:     runtime.GOARCH,
			binary:       "kubectl",
			err:          true,
			minikubeHome: minikubeHome,
		},
	}
	for _, test := range tc {
		t.Run(test.desc, func(t *testing.T) {
			os.Setenv("MINIKUBE_HOME", test.minikubeHome)
			_, err := Binary(test.binary, test.version, test.osName, test.archName)
			if err != nil && !test.err {
				t.Fatalf("Got unexpected error %v", err)
			}
			if err == nil && test.err {
				t.Fatalf("Expected error but got %v", err)
			}
		})
	}
}
