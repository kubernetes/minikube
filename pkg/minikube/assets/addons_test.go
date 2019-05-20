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

package assets

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"k8s.io/minikube/pkg/minikube/constants"
)

func setupTestDir() (string, error) {
	path, err := ioutil.TempDir("", "minipath")
	if err != nil {
		return "", err
	}

	os.Setenv(constants.MinikubeHome, path)
	return path, err
}

func TestAddMinikubeDirAssets(t *testing.T) {

	tests := []struct {
		description string
		baseDir     string
		vmPath      string
		files       []struct {
			relativePath string
			expectedPath string
		}
	}{
		{
			description: "relative path assets",
			baseDir:     "/files",
			files: []struct {
				relativePath string
				expectedPath string
			}{
				{
					relativePath: "/dir1/file1.txt",
					expectedPath: constants.AddonsPath,
				},
				{
					relativePath: "/dir1/file2.txt",
					expectedPath: constants.AddonsPath,
				},
				{
					relativePath: "/dir2/file1.txt",
					expectedPath: constants.AddonsPath,
				},
			},
			vmPath: constants.AddonsPath,
		},
		{
			description: "absolute path assets",
			baseDir:     "/files",
			files: []struct {
				relativePath string
				expectedPath string
			}{
				{
					relativePath: "/dir1/file1.txt",
					expectedPath: "/dir1",
				},
				{
					relativePath: "/dir1/file2.txt",
					expectedPath: "/dir1",
				},
				{
					relativePath: "/dir2/file1.txt",
					expectedPath: "/dir2",
				},
			},
			vmPath: "",
		},
	}
	var testDirs = make([]string, 0)
	defer func() {
		for _, testDir := range testDirs {
			err := os.RemoveAll(testDir)
			if err != nil {
				t.Logf("got unexpected error removing test dir: %v", err)
			}
		}
	}()

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			testDir, err := setupTestDir()
			if err != nil {
				t.Errorf("got unexpected error creating test dir: %v", err)
				return
			}

			testDirs = append(testDirs, testDir)
			testFileBaseDir := filepath.Join(testDir, test.baseDir)
			want := make(map[string]string)
			for _, fileDef := range test.files {
				err := func() error {
					path := filepath.Join(testFileBaseDir, fileDef.relativePath)
					err := os.MkdirAll(filepath.Dir(path), 0755)
					want[path] = fileDef.expectedPath
					if err != nil {
						return err
					}

					file, err := os.Create(path)
					if err != nil {
						return err
					}

					defer file.Close()

					_, err = file.WriteString("test")
					return err
				}()
				if err != nil {
					t.Errorf("unable to create file on fs: %v", err)
					return
				}
			}

			var actualFiles []CopyableFile
			err = addMinikubeDirToAssets(testFileBaseDir, test.vmPath, &actualFiles)
			if err != nil {
				t.Errorf("got unexpected error adding minikube dir assets: %v", err)
				return
			}

			got := make(map[string]string)
			for _, actualFile := range actualFiles {
				got[actualFile.GetAssetName()] = actualFile.GetTargetDir()
			}
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("files differ: (-want +got)\n%s", diff)
			}
		})
	}

}
