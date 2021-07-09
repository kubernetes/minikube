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

package hyperkit

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestExtractFile(t *testing.T) {
	testDir, err := ioutil.TempDir(os.TempDir(), "")
	if nil != err {
		return
	}
	defer func() { // clean up tempdir
		err := os.RemoveAll(testDir)
		if err != nil {
			t.Errorf("failed to clean up temp folder  %q", testDir)
		}
	}()

	tests := []struct {
		name          string
		isoPath       string
		srcPath       string
		destPath      string
		expectedError bool
	}{
		{
			name:          "all is right",
			isoPath:       "iso_test.iso",
			srcPath:       "/test1.txt",
			destPath:      testDir + "/test1.txt",
			expectedError: false,
		},
		{
			name:          "isoPath is error",
			isoPath:       "tests.iso",
			srcPath:       "/test1.txt",
			destPath:      testDir + "/test1.txt",
			expectedError: true,
		},
		{
			name:          "srcPath is empty",
			isoPath:       "iso_tests.iso",
			srcPath:       "",
			destPath:      testDir + "/test1.txt",
			expectedError: true,
		},
		{
			name:          "srcPath is error",
			isoPath:       "iso_tests.iso",
			srcPath:       "/t1.txt",
			destPath:      testDir + "/test1.txt",
			expectedError: true,
		},
		{
			name:          "destPath is empty",
			isoPath:       "iso_test.iso",
			srcPath:       "/test1.txt",
			destPath:      "",
			expectedError: true,
		},
		{
			name:          "find files in a folder",
			isoPath:       "./iso_test.iso",
			srcPath:       "/test2/test2.txt",
			destPath:      testDir + "/test2.txt",
			expectedError: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ExtractFile(tt.isoPath, tt.srcPath, tt.destPath)
			if (nil != err) != tt.expectedError {
				t.Errorf("expectedError = %v, get = %v", tt.expectedError, err)
				return
			}
		})
	}
}
