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

package registry

import (
	"testing"

	"k8s.io/minikube/pkg/minikube/config"
)

func TestDriverString(t *testing.T) {
	bar := DriverDef{
		Name:    "bar",
		Builtin: true,
		ConfigCreator: func(_ config.MachineConfig) interface{} {
			return nil
		},
	}
	s := bar.String()
	if s != "{name: bar, builtin: true}" {
		t.Fatalf("Driver bar.String() returned unexpected: %v", s)
	}
}

func TestRegistry(t *testing.T) {
	foo := DriverDef{
		Name:    "foo",
		Builtin: true,
		ConfigCreator: func(_ config.MachineConfig) interface{} {
			return nil
		},
	}
	bar := DriverDef{
		Name:    "bar",
		Builtin: true,
		ConfigCreator: func(_ config.MachineConfig) interface{} {
			return nil
		},
	}

	registry := createRegistry()

	if err := registry.Register(foo); err != nil {
		t.Fatal("expect nil")
	}

	if err := registry.Register(foo); err != ErrDriverNameExist {
		t.Fatal("expect ErrDriverNameExist")
	}

	if err := registry.Register(bar); err != nil {
		t.Fatal("expect nil")
	}

	list := registry.List()
	if len(list) != 2 {
		t.Fatalf("expect len(list) to be %d; got %d", 2, len(list))
	}

	if !(list[0].Name == "bar" && list[1].Name == "foo" || list[0].Name == "foo" && list[1].Name == "bar") {
		t.Fatalf("expect registry.List return %s; got %s", []string{"bar", "foo"}, list)
	}
	if drivers := ListDrivers(); len(list) == len(drivers) {
		t.Fatalf("Expectect ListDrivers and registry.List() to return same number of items")
	}

	driver, err := registry.Driver("foo")
	if err != nil {
		t.Fatal("expect nil")
	}
	if driver.Name != "foo" {
		t.Fatal("expect registry.Driver(foo) returns registered driver")
	}

	_, err = registry.Driver("foo2")
	if err != ErrDriverNotFound {
		t.Fatal("expect ErrDriverNotFound")
	}

	if _, err := Driver("no_such_driver"); err == nil {
		t.Fatal("expect to get error during not existing driver")
	}

	if _, err := Driver("foo"); err == nil {
		t.Fatal("expect to not get error during existing driver foo")
	}

	if err := Register(foo); err != nil {
		t.Fatal("expect to not get error during registering driver foo")
	}
}
