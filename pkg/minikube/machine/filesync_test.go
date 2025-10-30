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

package machine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/localpath"
	testutil "k8s.io/minikube/pkg/minikube/tests"
	"k8s.io/minikube/pkg/minikube/vmpath"
)

func collectAssets(t *testing.T, root, dest string, flatten bool) []assets.CopyableFile {
	t.Helper()
	files, err := assetsFromDir(root, dest, flatten)

	t.Cleanup(func() {
		for _, f := range files {
			if cerr := f.Close(); cerr != nil {
				t.Logf("warning: closing asset %s failed: %v", f.GetSourcePath(), cerr)
			}
		}
	})

	if err != nil {
		t.Fatalf("assetsFromDir(%q, %q, flatten=%v) unexpected error: %v", root, dest, flatten, err)
	}
	return files
}

func TestAssetsFromDir(t *testing.T) {
	tests := []struct {
		description string
		baseDir     string
		vmPath      string
		flatten     bool
		files       []struct {
			relativePath string
			expectedPath string
		}
	}{
		{
			description: "relative path assets",
			baseDir:     "/addons",
			flatten:     true,
			files: []struct {
				relativePath string
				expectedPath string
			}{
				{
					relativePath: "/dir1/file1.txt",
					expectedPath: vmpath.GuestAddonsDir,
				},
				{
					relativePath: "/dir1/file2.txt",
					expectedPath: vmpath.GuestAddonsDir,
				},
				{
					relativePath: "/dir2/file1.txt",
					expectedPath: vmpath.GuestAddonsDir,
				},
			},
			vmPath: vmpath.GuestAddonsDir,
		},
		{
			description: "absolute path assets",
			baseDir:     "/files",
			flatten:     false,
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
			vmPath: "/",
		},
	}

	var testDirs []string
	defer func() {
		for _, testDir := range testDirs {
			if err := os.RemoveAll(testDir); err != nil {
				t.Logf("got unexpected error removing test dir: %v", err)
			}
		}
	}()

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			testDir := testutil.MakeTempDir(t)
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

			actualFiles := collectAssets(t, testFileBaseDir, test.vmPath, test.flatten)

			got := make(map[string]string)
			for _, actualFile := range actualFiles {
				got[actualFile.GetSourcePath()] = actualFile.GetTargetDir()
			}
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("files differ: (-want +got)\n%s", diff)
			}
		})
	}

}

func TestSyncDest(t *testing.T) {
	tests := []struct {
		description string
		localParts  []string
		destRoot    string
		flatten     bool
		want        string
	}{
		{"simple", []string{"etc", "hosts"}, "/", false, "/etc/hosts"},
		{"nested", []string{"etc", "nested", "hosts"}, "/", false, "/etc/nested/hosts"},
		{"flat", []string{"etc", "nested", "hosts"}, "/test", true, "/test/hosts"},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			// Generate paths using filepath to mimic OS-specific issues
			localRoot := localpath.MakeMiniPath("sync")
			localParts := append([]string{localRoot}, test.localParts...)
			localPath := filepath.Join(localParts...)
			got, err := syncDest(localRoot, localPath, test.destRoot, test.flatten)
			if err != nil {
				t.Fatalf("syncDest(%s, %s, %v) unexpected err: %v", localRoot, localPath, test.flatten, err)
			}
			if got != test.want {
				t.Errorf("syncDest(%s, %s, %v) = %s, want: %s", localRoot, localPath, test.flatten, got, test.want)
			}
		})
	}
}
