/*
Copyright 2022 The Kubernetes Authors All rights reserved.

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

package persisttest

import (
	"fmt"
	"reflect"
	"testing"

	"k8s.io/minikube/pkg/libmachine/host"
)

func TestExists(t *testing.T) {
	store := FakeStore{
		Hosts: []*host.Host{{Name: "my-host"}},
	}
	exist, err := store.Exists("my-host")
	if err != nil {
		t.Fatal(err)
	}
	if exist == false {
		t.Fatal("Expected host 'my-host' to exist")
	}
	exist, err = store.Exists("not-found")
	if err != nil {
		t.Fatal(err)
	}
	if exist == true {
		t.Fatal("Expected host 'not-found' to no exist")
	}
	store.ExistsErr = fmt.Errorf("error checking host")
	_, err = store.Exists("my-host")
	if err != store.ExistsErr {
		t.Fatalf("Expected err %s.", store.ExistsErr)
	}
}

func TestList(t *testing.T) {
	store := FakeStore{
		Hosts: []*host.Host{{Name: "my-host"}, {Name: "my-host-2"}},
	}
	list, err := store.List()
	if err != nil {
		t.Fatal(err)
	}
	expected := []string{"my-host", "my-host-2"}
	if !reflect.DeepEqual(list, expected) {
		t.Fatalf("Expected hosts to be %s. Got %s.", expected, list)
	}
	store.ListErr = fmt.Errorf("error listing")
	_, err = store.List()
	if err != store.ListErr {
		t.Fatal(err)
	}
}

func TestLoad(t *testing.T) {
	expectedHost := &host.Host{Name: "my-host"}
	store := FakeStore{
		Hosts: []*host.Host{expectedHost},
	}
	h, err := store.Load("my-host")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(expectedHost, h) {
		t.Fatalf("Wrong host. Expected %v. Got %v.", expectedHost, h)
	}
	h, err = store.Load("not-found")
	if err != nil {
		t.Fatal(err)
	}
	if h != nil {
		t.Fatalf("Expected nil host. Got %v.", h)
	}
	store.LoadErr = fmt.Errorf("error loading")
	h, err = store.Load("my-host")
	if err != store.LoadErr {
		t.Fatalf("Wrong error. Expected %s. Got %s.", store.LoadErr, err)
	}
	if h != nil {
		t.Fatalf("Expected nil host. Got %v.", h)
	}
}

func TestRemove(t *testing.T) {
	store := FakeStore{
		Hosts: []*host.Host{{Name: "my-host"}},
	}
	err := store.Remove("not-found")
	if err != nil {
		t.Fatal(err)
	}
	err = store.Remove("my-host")
	if err != nil {
		t.Fatal(err)
	}
	if len(store.Hosts) != 0 {
		t.Fatalf("Expected hosts length to be zero. Got %d", len(store.Hosts))
	}
	store.RemoveErr = fmt.Errorf("error removing")
	err = store.Remove("my-host")
	if err != store.RemoveErr {
		t.Fatal(err)
	}
}

func TestSave(t *testing.T) {
	store := FakeStore{}
	err := store.Save(&host.Host{Name: "my-host"})
	if err != nil {
		t.Fatal(err)
	}
	expectedHosts := []*host.Host{{Name: "my-host"}}
	if !reflect.DeepEqual(store.Hosts, expectedHosts) {
		t.Fatalf("Expected hosts to be %v. Got %v.", expectedHosts, store.Hosts)
	}
	store.SaveErr = fmt.Errorf("error saving")
	err = store.Save(&host.Host{Name: "new-host"})
	if err != store.SaveErr {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(store.Hosts, expectedHosts) {
		t.Fatalf("Expected hosts to be %v. Got %v.", expectedHosts, store.Hosts)
	}
}
