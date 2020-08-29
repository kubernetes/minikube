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

func TestListOnEmptyRegistry(t *testing.T) {
	reg := &persistentRegistry{
		path: "nonexistent.txt",
	}

	info, err := reg.List()
	expectedInfo := []*ID{}
	if !reflect.DeepEqual(info, expectedInfo) || err != nil {
		t.Errorf("expected %s, nil error, got %s, %s", expectedInfo, info, err)
	}
}

func TestRemoveOnEmptyRegistry(t *testing.T) {
	reg := &persistentRegistry{
		path: "nonexistent.txt",
	}

	e := reg.Remove(unsafeParseRoute("1.2.3.4", "1.2.3.4/5"))
	if e == nil {
		t.Errorf("expected error, got %s", e)
	}
}

func TestRegisterOnEmptyRegistry(t *testing.T) {
	reg := &persistentRegistry{
		path: "nonexistent.txt",
	}

	err := reg.Register(&ID{Route: unsafeParseRoute("1.2.3.4", "1.2.3.4/5")})
	if err != nil {
		t.Errorf("expected no error, got %s", err)
	}
	f, err := os.Open("nonexistent.txt")
	if err != nil {
		t.Errorf("expected file to exist, got: %s", err)
		return
	}
	f.Close()
	err = os.Remove("nonexistent.txt")
	if err != nil {
		t.Errorf("error removing nonexistent.txt: %s", err)
	}
}

func TestRemoveOnNonExistentTunnel(t *testing.T) {
	file := tmpFile(t)
	reg := &persistentRegistry{
		path: file,
	}
	defer os.Remove(file)

	err := reg.Register(&ID{Route: unsafeParseRoute("1.2.3.4", "1.2.3.4/5")})
	if err != nil {
		t.Errorf("expected no error, got %s", err)
	}

	err = reg.Remove(unsafeParseRoute("5.6.7.8", "1.2.3.4/5"))
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}

func TestListAfterRegister(t *testing.T) {
	file := tmpFile(t)
	reg := &persistentRegistry{
		path: file,
	}
	defer os.Remove(file)
	err := reg.Register(&ID{
		Route:       unsafeParseRoute("1.2.3.4", "1.2.3.4/5"),
		MachineName: "testmachine",
		Pid:         1234,
	})
	if err != nil {
		t.Errorf("failed to register: expected no error, got %s", err)
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
}

func TestRegisterRemoveList(t *testing.T) {
	file := tmpFile(t)
	reg := &persistentRegistry{
		path: file,
	}
	defer os.Remove(file)

	err := reg.Register(&ID{
		Route:       unsafeParseRoute("192.168.1.25", "10.96.0.0/12"),
		MachineName: "testmachine",
		Pid:         1234,
	})
	if err != nil {
		t.Errorf("failed to register: expected no error, got %s", err)
	}

	err = reg.Remove(unsafeParseRoute("192.168.1.25", "10.96.0.0/12"))

	if err != nil {
		t.Errorf("failed to remove: expected no error, got %s", err)
	}

	tunnelList, err := reg.List()
	if err != nil {
		t.Errorf("failed to list: expected no error, got %s", err)
	}

	expectedList := []*ID{}

	if len(tunnelList) != 0 {
		t.Errorf("\nexpected %+v,\ngot      %+v", expectedList, tunnelList)
	}
}

func TestDuplicateRouteError(t *testing.T) {
	file := tmpFile(t)
	reg := &persistentRegistry{
		path: file,
	}
	defer os.Remove(file)

	err := reg.Register(&ID{
		Route:       unsafeParseRoute("192.168.1.25", "10.96.0.0/12"),
		MachineName: "testmachine",
		Pid:         os.Getpid(),
	})
	if err != nil {
		t.Errorf("failed to register: expected no error, got %s", err)
	}
	err = reg.Register(&ID{
		Route:       unsafeParseRoute("192.168.1.25", "10.96.0.0/12"),
		MachineName: "testmachine",
		Pid:         5678,
	})
	if err == nil {
		t.Error("expected error on duplicate route, got nil")
	}
}

func TestTunnelTakeoverFromNonRunningProcess(t *testing.T) {
	file := tmpFile(t)
	reg := &persistentRegistry{
		path: file,
	}
	defer os.Remove(file)

	err := reg.Register(&ID{
		Route:       unsafeParseRoute("192.168.1.25", "10.96.0.0/12"),
		MachineName: "testmachine",
		Pid:         12341234,
	})
	if err != nil {
		t.Errorf("failed to register: expected no error, got %s", err)
	}
	err = reg.Register(&ID{
		Route:       unsafeParseRoute("192.168.1.25", "10.96.0.0/12"),
		MachineName: "testmachine",
		Pid:         5678,
	})
	if err != nil {
		t.Errorf("failed to register: expected no error, got %s", err)
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
}

func tmpFile(t *testing.T) string {
	t.Helper()
	f, err := ioutil.TempFile(os.TempDir(), "reg_")
	f.Close()
	if err != nil {
		t.Errorf("failed to create temp file %s", err)
	}
	return f.Name()
}

func createTestRegistry(t *testing.T) (reg *persistentRegistry, cleanup func()) {
	f, err := ioutil.TempFile(os.TempDir(), "reg_")
	f.Close()
	if err != nil {
		t.Errorf("failed to create temp file %s", err)
	}

	registry := &persistentRegistry{
		path: f.Name(),
	}
	return registry, func() { os.Remove(f.Name()) }
}
