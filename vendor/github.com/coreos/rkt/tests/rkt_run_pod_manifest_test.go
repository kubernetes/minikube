// Copyright 2015 The rkt Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	"testing"

	"github.com/appc/spec/schema"
	"github.com/appc/spec/schema/types"
	"github.com/coreos/rkt/tests/testutils"
	"github.com/syndtr/gocapability/capability"

	"github.com/coreos/rkt/common/cgroup"
)

const baseAppName = "rkt-inspect"

func intP(i int) *int {
	return &i
}
func stringP(s string) *string {
	return &s
}

func generatePodManifestFile(t *testing.T, manifest *schema.PodManifest) string {
	tmpDir := testutils.GetValueFromEnvOrPanic("FUNCTIONAL_TMP")
	f, err := ioutil.TempFile(tmpDir, "rkt-test-manifest-")
	if err != nil {
		t.Fatalf("Cannot create tmp pod manifest: %v", err)
	}

	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("Cannot marshal pod manifest: %v", err)
	}
	if err := ioutil.WriteFile(f.Name(), data, 0600); err != nil {
		t.Fatalf("Cannot write pod manifest file: %v", err)
	}
	return f.Name()
}

func verifyHostFile(t *testing.T, tmpdir, filename string, i int, expectedResult string) {
	filePath := path.Join(tmpdir, filename)
	defer os.Remove(filePath)

	// Verify the file is written to host.
	if strings.Contains(expectedResult, "host:") {
		data, err := ioutil.ReadFile(filePath)
		if err != nil {
			t.Fatalf("%d: Cannot read the host file: %v", i, err)
		}
		if string(data) != expectedResult {
			t.Fatalf("%d: Expecting %q in the host file, but saw %q", i, expectedResult, data)
		}
	}
}

func rawValue(value string) *json.RawMessage {
	msg := json.RawMessage(value)
	return &msg
}

func rawRequestLimit(request, limit string) *json.RawMessage {
	if request == "" {
		return rawValue(fmt.Sprintf(`{"limit":%q}`, limit))
	}
	if limit == "" {
		return rawValue(fmt.Sprintf(`{"request":%q}`, request))
	}
	return rawValue(fmt.Sprintf(`{"request":%q,"limit":%q}`, request, limit))
}

type imagePatch struct {
	name    string
	patches []string
}

// Test running pod manifests that contains just one app.
// TODO(yifan): Figure out a way to test port mapping on single host.
func TestPodManifest(t *testing.T) {
	ctx := testutils.NewRktRunCtx()
	defer ctx.Cleanup()

	tmpdir := createTempDirOrPanic("rkt-tests.")
	defer os.RemoveAll(tmpdir)

	boolFalse, boolTrue := false, true

	tests := []struct {
		// [image name]:[image patches]
		images         []imagePatch
		podManifest    *schema.PodManifest
		shouldSucceed  bool
		expectedResult string
		cgroup         string
	}{
		{
			// Special characters
			[]imagePatch{
				{"rkt-test-run-pod-manifest-special-characters.aci", []string{}},
			},
			&schema.PodManifest{
				Apps: []schema.RuntimeApp{
					{
						Name: baseAppName,
						App: &types.App{
							Exec:  []string{"/inspect", "--print-msg=\n'\"$"},
							User:  "0",
							Group: "0",
						},
					},
				},
			},
			true,
			`'"[$]`,
			"",
		},
		{
			// Simple read.
			[]imagePatch{
				{"rkt-test-run-pod-manifest-read.aci", []string{}},
			},
			&schema.PodManifest{
				Apps: []schema.RuntimeApp{
					{
						Name: baseAppName,
						App: &types.App{
							Exec:  []string{"/inspect", "--read-file"},
							User:  "0",
							Group: "0",
							Environment: []types.EnvironmentVariable{
								{"FILE", "/dir1/file"},
							},
						},
					},
				},
			},
			true,
			"dir1",
			"",
		},
		{
			// Simple read after write with *empty* volume mounted.
			[]imagePatch{
				{"rkt-test-run-pod-manifest-empty-vol-rw.aci", []string{}},
			},
			&schema.PodManifest{
				Apps: []schema.RuntimeApp{
					{
						Name: baseAppName,
						App: &types.App{
							Exec:  []string{"/inspect", "--write-file", "--read-file"},
							User:  "0",
							Group: "0",
							Environment: []types.EnvironmentVariable{
								{"FILE", "/dir1/file"},
								{"CONTENT", "empty:foo"},
							},
							MountPoints: []types.MountPoint{
								{"dir1", "/dir1", false},
							},
						},
					},
				},
				Volumes: []types.Volume{
					{
						Name: "dir1",
						Kind: "empty",
						Mode: stringP("0755"),
						UID:  intP(0),
						GID:  intP(0),
					},
				},
			},
			true,
			"empty:foo",
			"",
		},
		{
			// Stat directory in a *empty* volume mounted.
			[]imagePatch{
				{"rkt-test-run-pod-manifest-empty-vol-stat.aci", []string{}},
			},
			&schema.PodManifest{
				Apps: []schema.RuntimeApp{
					{
						Name: baseAppName,
						App: &types.App{
							Exec:  []string{"/inspect", "--stat-file"},
							User:  "0",
							Group: "0",
							Environment: []types.EnvironmentVariable{
								{"FILE", "/dir1"},
							},
							MountPoints: []types.MountPoint{
								{"dir1", "/dir1", false},
							},
						},
					},
				},
				Volumes: []types.Volume{
					{
						Name: "dir1",
						Kind: "empty",
						Mode: stringP("0123"),
						UID:  intP(9991),
						GID:  intP(9992),
					},
				},
			},
			true,
			"(?s)/dir1: mode: d--x-w--wx.*" + "/dir1: user: 9991.*" + "/dir1: group: 9992",
			"",
		},
		{
			// Simple read after write with volume mounted.
			[]imagePatch{
				{"rkt-test-run-pod-manifest-vol-rw.aci", []string{}},
			},
			&schema.PodManifest{
				Apps: []schema.RuntimeApp{
					{
						Name: baseAppName,
						App: &types.App{
							Exec:  []string{"/inspect", "--write-file", "--read-file"},
							User:  "0",
							Group: "0",
							Environment: []types.EnvironmentVariable{
								{"FILE", "/dir1/file"},
								{"CONTENT", "host:foo"},
							},
							MountPoints: []types.MountPoint{
								{"dir1", "/dir1", false},
							},
						},
					},
				},
				Volumes: []types.Volume{
					{"dir1", "host", tmpdir, nil, nil, nil, nil},
				},
			},
			true,
			"host:foo",
			"",
		},
		{
			// Simple read after write with read-only mount point, should fail.
			[]imagePatch{
				{"rkt-test-run-pod-manifest-vol-ro.aci", []string{}},
			},
			&schema.PodManifest{
				Apps: []schema.RuntimeApp{
					{
						Name: baseAppName,
						App: &types.App{
							Exec:  []string{"/inspect", "--write-file", "--read-file"},
							User:  "0",
							Group: "0",
							Environment: []types.EnvironmentVariable{
								{"FILE", "/dir1/file"},
								{"CONTENT", "bar"},
							},
							MountPoints: []types.MountPoint{
								{"dir1", "/dir1", true},
							},
						},
					},
				},
				Volumes: []types.Volume{
					{"dir1", "host", tmpdir, nil, nil, nil, nil},
				},
			},
			false,
			`Cannot write to file "/dir1/file": open /dir1/file: read-only file system`,
			"",
		},
		{
			// Simple read after write with volume mounted.
			// Override the image's mount point spec. This should fail as the volume is
			// read-only in pod manifest, (which will override the mount point in both image/pod manifest).
			[]imagePatch{
				{"rkt-test-run-pod-manifest-vol-rw-override.aci", []string{"--mounts=dir1,path=/dir1,readOnly=false"}},
			},
			&schema.PodManifest{
				Apps: []schema.RuntimeApp{
					{
						Name: baseAppName,
						App: &types.App{
							Exec:  []string{"/inspect", "--write-file", "--read-file"},
							User:  "0",
							Group: "0",
							Environment: []types.EnvironmentVariable{
								{"FILE", "/dir1/file"},
								{"CONTENT", "bar"},
							},
							MountPoints: []types.MountPoint{
								{"dir1", "/dir1", false},
							},
						},
					},
				},
				Volumes: []types.Volume{
					{"dir1", "host", tmpdir, &boolTrue, nil, nil, nil},
				},
			},
			false,
			`Cannot write to file "/dir1/file": open /dir1/file: read-only file system`,
			"",
		},
		{
			// Simple read after write with volume mounted.
			// Override the image's mount point spec.
			[]imagePatch{
				{"rkt-test-run-pod-manifest-vol-rw-override.aci", []string{"--mounts=dir1,path=/dir1,readOnly=true"}},
			},
			&schema.PodManifest{
				Apps: []schema.RuntimeApp{
					{
						Name: baseAppName,
						App: &types.App{
							Exec:  []string{"/inspect", "--write-file", "--read-file"},
							User:  "0",
							Group: "0",
							Environment: []types.EnvironmentVariable{
								{"FILE", "/dir2/file"},
								{"CONTENT", "host:bar"},
							},
							MountPoints: []types.MountPoint{
								{"dir1", "/dir2", false},
							},
						},
					},
				},
				Volumes: []types.Volume{
					{"dir1", "host", tmpdir, nil, nil, nil, nil},
				},
			},
			true,
			"host:bar",
			"",
		},
		{
			// Simple read after write with volume mounted, no apps in pod manifest.
			[]imagePatch{
				{
					"rkt-test-run-pod-manifest-vol-rw-no-app.aci",
					[]string{
						"--exec=/inspect --write-file --read-file --file-name=/dir1/file --content=host:baz",
						"--mounts=dir1,path=/dir1,readOnly=false",
					},
				},
			},
			&schema.PodManifest{
				Apps: []schema.RuntimeApp{
					{Name: baseAppName},
				},
				Volumes: []types.Volume{
					{"dir1", "host", tmpdir, nil, nil, nil, nil},
				},
			},
			true,
			"host:baz",
			"",
		},
		{
			// Simple read after write with volume mounted, no apps in pod manifest.
			// This should succeed even the mount point in image manifest is readOnly,
			// because it is overrided by the volume's readOnly.
			[]imagePatch{
				{
					"rkt-test-run-pod-manifest-vol-ro-no-app.aci",
					[]string{
						"--exec=/inspect --write-file --read-file --file-name=/dir1/file --content=host:zaz",
						"--mounts=dir1,path=/dir1,readOnly=true",
					},
				},
			},
			&schema.PodManifest{
				Apps: []schema.RuntimeApp{
					{Name: baseAppName},
				},
				Volumes: []types.Volume{
					{"dir1", "host", tmpdir, &boolFalse, nil, nil, nil},
				},
			},
			true,
			"host:zaz",
			"",
		},
		{
			// Simple read after write with read-only volume mounted, no apps in pod manifest.
			// This should fail as the volume is read-only.
			[]imagePatch{
				{
					"rkt-test-run-pod-manifest-vol-ro-no-app.aci",
					[]string{
						"--exec=/inspect --write-file --read-file --file-name=/dir1/file --content=baz",
						"--mounts=dir1,path=/dir1,readOnly=false",
					},
				},
			},
			&schema.PodManifest{
				Apps: []schema.RuntimeApp{
					{Name: baseAppName},
				},
				Volumes: []types.Volume{
					{"dir1", "host", tmpdir, &boolTrue, nil, nil, nil},
				},
			},
			false,
			`Cannot write to file "/dir1/file": open /dir1/file: read-only file system`,
			"",
		},
		{
			// Print CPU quota, which should be overwritten by the pod manifest.
			[]imagePatch{
				{"rkt-test-run-pod-manifest-cpu-isolator.aci", []string{}},
			},
			&schema.PodManifest{
				Apps: []schema.RuntimeApp{
					{
						Name: baseAppName,
						App: &types.App{
							Exec:  []string{"/inspect", "--print-cpuquota"},
							User:  "0",
							Group: "0",
							Isolators: []types.Isolator{
								{
									Name:     "resource/cpu",
									ValueRaw: rawRequestLimit("100", "100"),
								},
							},
						},
					},
				},
			},
			true,
			`CPU Quota: 100`,
			"cpu",
		},
		{
			// Print memory limit, which should be overwritten by the pod manifest.
			[]imagePatch{
				{"rkt-test-run-pod-manifest-memory-isolator.aci", []string{}},
			},
			&schema.PodManifest{
				Apps: []schema.RuntimeApp{
					{
						Name: baseAppName,
						App: &types.App{
							Exec:  []string{"/inspect", "--print-memorylimit"},
							User:  "0",
							Group: "0",
							Isolators: []types.Isolator{
								{
									Name: "resource/memory",
									// 4MB.
									ValueRaw: rawRequestLimit("4194304", "4194304"),
								},
							},
						},
					},
				},
			},
			true,
			`Memory Limit: 4194304`,
			"memory",
		},
		{
			// Multiple apps (with same images) in the pod. The first app will read out the content
			// written by the second app.
			[]imagePatch{
				{"rkt-test-run-pod-manifest-app.aci", []string{"--name=aci1"}},
				{"rkt-test-run-pod-manifest-app.aci", []string{"--name=aci2"}},
			},
			&schema.PodManifest{
				Apps: []schema.RuntimeApp{
					{
						Name: "rkt-inspect-readapp",
						App: &types.App{
							Exec:  []string{"/inspect", "--pre-sleep=10", "--read-file"},
							User:  "0",
							Group: "0",
							Environment: []types.EnvironmentVariable{
								{"FILE", "/dir/file"},
							},
							MountPoints: []types.MountPoint{
								{"dir", "/dir", false},
							},
						},
					},
					{
						Name: "rkt-inspect-writeapp",
						App: &types.App{
							Exec:  []string{"/inspect", "--write-file"},
							User:  "0",
							Group: "0",
							Environment: []types.EnvironmentVariable{
								{"FILE", "/dir/file"},
								{"CONTENT", "host:foo"},
							},
							MountPoints: []types.MountPoint{
								{"dir", "/dir", false},
							},
						},
					},
				},
				Volumes: []types.Volume{
					{"dir", "host", tmpdir, nil, nil, nil, nil},
				},
			},
			true,
			"host:foo",
			"",
		},
		{
			// Pod manifest overwrites the image's capability.
			[]imagePatch{
				{"rkt-test-run-pod-manifest-cap.aci", []string{"--capability=CAP_NET_ADMIN"}},
			},
			&schema.PodManifest{
				Apps: []schema.RuntimeApp{
					{
						Name: baseAppName,
						App: &types.App{
							Exec:  []string{"/inspect", "--print-caps-pid=0"},
							User:  "0",
							Group: "0",
							Environment: []types.EnvironmentVariable{
								{"CAPABILITY", strconv.Itoa(int(capability.CAP_NET_ADMIN))},
							},
						},
					},
				},
			},
			true,
			fmt.Sprintf("%v=disabled", capability.CAP_NET_ADMIN.String()),
			"",
		},
		{
			// Pod manifest overwrites the image's capability.
			[]imagePatch{
				{"rkt-test-run-pod-manifest-cap.aci", []string{"--capability=CAP_NET_BIND_SERVICE"}},
			},
			&schema.PodManifest{
				Apps: []schema.RuntimeApp{
					{
						Name: baseAppName,
						App: &types.App{
							Exec:  []string{"/inspect", "--print-caps-pid=0"},
							User:  "0",
							Group: "0",
							Environment: []types.EnvironmentVariable{
								{"CAPABILITY", strconv.Itoa(int(capability.CAP_NET_ADMIN))},
							},
							Isolators: []types.Isolator{
								{
									Name:     "os/linux/capabilities-retain-set",
									ValueRaw: rawValue(fmt.Sprintf(`{"set":["CAP_NET_ADMIN"]}`)),
								},
							},
						},
					},
				},
			},
			true,
			fmt.Sprintf("%v=enabled", capability.CAP_NET_ADMIN.String()),
			"",
		},
	}

	for i, tt := range tests {
		if tt.cgroup != "" && !cgroup.IsIsolatorSupported(tt.cgroup) {
			t.Logf("Skip test #%v: cgroup %s not supported", i, tt.cgroup)
			continue
		}

		var hashesToRemove []string
		for j, v := range tt.images {
			hash := patchImportAndFetchHash(v.name, v.patches, t, ctx)
			hashesToRemove = append(hashesToRemove, hash)
			imgName := types.MustACIdentifier(v.name)
			imgID, err := types.NewHash(hash)
			if err != nil {
				t.Fatalf("Cannot generate types.Hash from %v: %v", hash, err)
			}

			ra := &tt.podManifest.Apps[j]
			ra.Image.Name = imgName
			ra.Image.ID = *imgID
		}

		tt.podManifest.ACKind = schema.PodManifestKind
		tt.podManifest.ACVersion = schema.AppContainerVersion

		manifestFile := generatePodManifestFile(t, tt.podManifest)
		defer os.Remove(manifestFile)

		// 1. Test 'rkt run'.
		runCmd := fmt.Sprintf("%s run --mds-register=false --pod-manifest=%s", ctx.Cmd(), manifestFile)
		t.Logf("Running 'run' test #%v", i)
		child := spawnOrFail(t, runCmd)

		if tt.expectedResult != "" {
			if _, out, err := expectRegexWithOutput(child, tt.expectedResult); err != nil {
				t.Fatalf("Expected %q but not found: %v\n%s", tt.expectedResult, err, out)
			}
		}
		if err := child.Wait(); err != nil {
			if tt.shouldSucceed {
				t.Fatalf("rkt didn't terminate correctly: %v", err)
			}
		}
		verifyHostFile(t, tmpdir, "file", i, tt.expectedResult)

		// 2. Test 'rkt prepare' + 'rkt run-prepared'.
		rktCmd := fmt.Sprintf("%s --insecure-options=image prepare --pod-manifest=%s",
			ctx.Cmd(), manifestFile)
		uuid := runRktAndGetUUID(t, rktCmd)

		runPreparedCmd := fmt.Sprintf("%s run-prepared --mds-register=false %s", ctx.Cmd(), uuid)
		t.Logf("Running 'run-prepared' test #%v", i)
		child = spawnOrFail(t, runPreparedCmd)

		if tt.expectedResult != "" {
			if _, out, err := expectRegexWithOutput(child, tt.expectedResult); err != nil {
				t.Fatalf("Expected %q but not found: %v\n%s", tt.expectedResult, err, out)
			}
		}
		if err := child.Wait(); err != nil {
			if tt.shouldSucceed {
				t.Fatalf("rkt didn't terminate correctly: %v", err)
			}
		}
		verifyHostFile(t, tmpdir, "file", i, tt.expectedResult)

		// we run the garbage collector and remove the imported images to save
		// space
		runGC(t, ctx)
		for _, h := range hashesToRemove {
			removeFromCas(t, ctx, h)
		}
	}
}
