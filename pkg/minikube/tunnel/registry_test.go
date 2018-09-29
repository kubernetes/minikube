/*
Copyright 2018 The Kubernetes Authors All rights reserved.

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

package tunnel

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

func TestPersistentRegistryWithNoKey(t *testing.T) {
	registry, cleanup := createTestRegistry(t)
	defer cleanup()

	route := &ID{}
	err := registry.Register(route)

	if err == nil {
		t.Errorf("attempting to register ID without key should throw error")
	}
}

func TestPersistentRegistryNullableMetadata(t *testing.T) {
	registry, cleanup := createTestRegistry(t)
	defer cleanup()

	route := &ID{
		Route: unsafeParseRoute("1.2.3.4", "10.96.0.0/12"),
	}
	err := registry.Register(route)

	if err != nil {
		t.Errorf("metadata should be nullable, expected no error, got %s", err)
	}
}

func TestPersistentRegistryOperations(t *testing.T) {
	tcs := []struct {
		name     string
		fileName string
		setup    func(*testing.T, *persistentRegistry)
		test     func(*testing.T, *persistentRegistry)
	}{
		{
			name:     "Calling List on non-existent registry file should return empty list",
			fileName: "nonexistent.txt",
			test: func(t *testing.T, reg *persistentRegistry) {
				info, e := reg.List()
				expectedInfo := []*ID{}
				if !reflect.DeepEqual(info, expectedInfo) || e != nil {
					t.Errorf("expected %s, nil error, got %s, %s", expectedInfo, info, e)
				}
			},
		},
		{
			name:     "Calling Remove on non-existent registry file should return error",
			fileName: "nonexistent.txt",
			test: func(t *testing.T, reg *persistentRegistry) {
				e := reg.Remove(unsafeParseRoute("1.2.3.4", "1.2.3.4/5"))
				if e == nil {
					t.Errorf("expected error, got %s", e)
				}
			},
		},
		{
			name:     "Calling Register on non-existent registry file should create file",
			fileName: "nonexistent.txt",
			test: func(t *testing.T, reg *persistentRegistry) {
				e := reg.Register(&ID{Route: unsafeParseRoute("1.2.3.4", "1.2.3.4/5")})
				if e != nil {
					t.Errorf("expected no error, got %s", e)
				}
				fileName := "nonexistent.txt"
				f, e := os.Open(fileName)
				if e != nil {
					t.Errorf("expected file to exist, got: %s", e)
					return
				}
				f.Close()
				e = os.Remove("nonexistent.txt")
				if e != nil {
					t.Errorf("error removing nonexistent.txt: %s", e)
				}
			},
		},
		{
			name: "Calling Remove on non-existent tunnel should return error",
			test: func(t *testing.T, reg *persistentRegistry) {
				e := reg.Register(&ID{Route: unsafeParseRoute("1.2.3.4", "1.2.3.4/5")})
				if e != nil {
					t.Errorf("expected no error, got %s", e)
				}

				e = reg.Remove(unsafeParseRoute("5.6.7.8", "1.2.3.4/5"))
				if e == nil {
					t.Errorf("expected error, got nil")
				}
			},
		},
		{
			name: "Register + List should return tunnel info",
			test: func(t *testing.T, reg *persistentRegistry) {
				e := reg.Register(&ID{
					Route:       unsafeParseRoute("1.2.3.4", "1.2.3.4/5"),
					MachineName: "testmachine",
					Pid:         1234,
				})
				if e != nil {
					t.Errorf("failed to register: expected no error, got %s", e)
				}

				tunnelList, err := reg.List()
				if err != nil {
					t.Errorf("failed to list: expected no error, got %s", err)
				}

				expectedList := []*ID{
					{
						Route:       unsafeParseRoute("1.2.3.4", "1.2.3.4/5"),
						MachineName: "testmachine",
						Pid:         1234,
					},
				}

				if len(tunnelList) != 1 || !tunnelList[0].Equal(expectedList[0]) {
					t.Errorf("\nexpected %+v,\ngot      %+v", expectedList, tunnelList)
				}
			},
		},
		{
			name: "Register + Remove + List",
			test: func(t *testing.T, reg *persistentRegistry) {
				e := reg.Register(&ID{
					Route:       unsafeParseRoute("192.168.1.25", "10.96.0.0/12"),
					MachineName: "testmachine",
					Pid:         1234,
				})
				if e != nil {
					t.Errorf("failed to register: expected no error, got %s", e)
				}

				e = reg.Remove(unsafeParseRoute("192.168.1.25", "10.96.0.0/12"))

				if e != nil {
					t.Errorf("failed to remove: expected no error, got %s", e)
				}

				tunnelList, err := reg.List()
				if err != nil {
					t.Errorf("failed to list: expected no error, got %s", err)
				}

				expectedList := []*ID{}

				if len(tunnelList) != 0 {
					t.Errorf("\nexpected %+v,\ngot      %+v", expectedList, tunnelList)
				}
			},
		},
		{
			name: "Error on duplicate route with running process",
			test: func(t *testing.T, reg *persistentRegistry) {
				e := reg.Register(&ID{
					Route:       unsafeParseRoute("192.168.1.25", "10.96.0.0/12"),
					MachineName: "testmachine",
					Pid:         os.Getpid(),
				})
				if e != nil {
					t.Errorf("failed to register: expected no error, got %s", e)
				}
				e = reg.Register(&ID{
					Route:       unsafeParseRoute("192.168.1.25", "10.96.0.0/12"),
					MachineName: "testmachine",
					Pid:         5678,
				})
				if e == nil {
					t.Error("expected error on duplicate route, got nil")
				}
			},
		},
		{
			name: "Update duplicate route when process is not running",
			test: func(t *testing.T, reg *persistentRegistry) {
				e := reg.Register(&ID{
					Route:       unsafeParseRoute("192.168.1.25", "10.96.0.0/12"),
					MachineName: "testmachine",
					Pid:         12341234,
				})
				if e != nil {
					t.Errorf("failed to register: expected no error, got %s", e)
				}
				e = reg.Register(&ID{
					Route:       unsafeParseRoute("192.168.1.25", "10.96.0.0/12"),
					MachineName: "testmachine",
					Pid:         5678,
				})
				if e != nil {
					t.Errorf("failed to register: expected no error, got %s", e)
				}

				tunnelList, err := reg.List()
				if err != nil {
					t.Errorf("failed to list: expected no error, got %s", err)
				}

				expectedList := []*ID{
					{
						Route:       unsafeParseRoute("192.168.1.25", "10.96.0.0/12"),
						MachineName: "testmachine",
						Pid:         5678,
					},
				}

				if len(tunnelList) != 1 || !tunnelList[0].Equal(expectedList[0]) {
					t.Errorf("\nexpected %+v,\ngot      %+v", expectedList, tunnelList)
				}
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			fileName := tc.fileName

			if fileName == "" {
				f, err := ioutil.TempFile(os.TempDir(), "reg_")
				f.Close()
				if err != nil {
					t.Errorf("failed to create temp file %s", err)
				}
				fileName = f.Name()
				defer os.Remove(fileName)
			}

			registry := &persistentRegistry{
				fileName: fileName,
			}
			if tc.setup != nil {
				tc.setup(t, registry)
			}
			if tc.test != nil {
				tc.test(t, registry)
			}
		})
	}

}

func createTestRegistry(t *testing.T) (reg *persistentRegistry, cleanup func()) {
	f, err := ioutil.TempFile(os.TempDir(), "reg_")
	f.Close()
	if err != nil {
		t.Errorf("failed to create temp file %s", err)
	}

	registry := &persistentRegistry{
		fileName: f.Name(),
	}
	return registry, func() { os.Remove(f.Name()) }
}
