/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package machine

import (
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"testing"

	"k8s.io/minikube/pkg/minikube/constants"
)

func TestGetSrcRef(t *testing.T) {
	for _, image := range constants.LocalkubeCachedImages {
		if _, err := getSrcRef(image); err != nil {
			t.Errorf("Error getting src ref for %s: %s", image, err)
		}
	}
}

func TestGetDstRef(t *testing.T) {
	paths := []struct {
		path, separator string
	}{
		{`/Users/foo/.minikube/cache/images`, `/`},
		{`/home/foo/.minikube/cache/images`, `/`},
		{`\\?\Volume{aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee}\Users\foo\.minikube\cache\images`, `\`},
		{`\\?\Volume{aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee}\minikube\.minikube\cache\images`, `\`},
	}

	cases := []struct {
		image, dst string
	}{}
	for _, tp := range paths {
		for _, image := range constants.LocalkubeCachedImages {
			dst := strings.Join([]string{tp.path, strings.Replace(image, ":", "_", -1)}, tp.separator)
			cases = append(cases, struct{ image, dst string }{image, dst})
		}
	}

	for _, tc := range cases {
		if _, err := _getDstRef(tc.image, tc.dst); err != nil {
			t.Errorf("Error getting dst ref for %s: %s", tc.dst, err)
		}
	}
}

func TestReplaceWinDriveLetterToVolumeName(t *testing.T) {
	path, err := ioutil.TempDir("", "repwindl2vn")
	if err != nil {
		t.Fatalf("Error make tmp directory: %s", err)
	}
	defer os.RemoveAll(path)

	if runtime.GOOS != "windows" {
		// Replace to fake func.
		getWindowsVolumeName = func(d string) (string, error) {
			return `/`, nil
		}
		// Add dummy Windows drive letter.
		path = `C:` + path
	}

	if _, err := replaceWinDriveLetterToVolumeName(path); err != nil {
		t.Errorf("Error replace a Windows drive letter to a volume name: %s", err)
	}
}

func TestHasWindowsDriveLetter(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		{`C:\Users\Foo\.minikube`, true},
		{`D:\minikube\.minikube`, true},
		{`C\Foo\Bar\.minikube`, false},
		{`/home/foo/.minikube`, false},
	}

	for _, tc := range cases {
		if hasWindowsDriveLetter(tc.path) != tc.want {
			t.Errorf("%s have a Windows drive letter: %t", tc.path, tc.want)
		}
	}
}
