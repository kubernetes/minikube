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

func testDriver(name string) DriverDef {
	return DriverDef{
		Name:    name,
		Builtin: true,
		ConfigCreator: func(_ config.MachineConfig) interface{} {
			return nil
		},
	}
}

func TestRegistry1(t *testing.T) {
	foo := testDriver("foo")
	bar := testDriver("bar")

	registry := createRegistry()
	t.Run("registry.Register", func(t *testing.T) {
		t.Run("foo", func(t *testing.T) {
			if err := registry.Register(foo); err != nil {
				t.Fatalf("error not expected but got %v", err)
			}
		})
		t.Run("fooAlreadyExist", func(t *testing.T) {
			if err := registry.Register(foo); err != ErrDriverNameExist {
				t.Fatalf("expect ErrDriverNameExist but got: %v", err)
			}
		})
		t.Run("bar", func(t *testing.T) {
			if err := registry.Register(bar); err != nil {
				t.Fatalf("error not expect but got: %v", err)
			}
		})
	})
	t.Run("registry.List", func(t *testing.T) {
		list := registry.List()
		if !(list[0].Name == "bar" && list[1].Name == "foo" ||
			list[0].Name == "foo" && list[1].Name == "bar") {
			t.Fatalf("expect registry.List return %s; got %s", []string{"bar", "foo"}, list)
		}
		if drivers := ListDrivers(); len(list) == len(drivers) {
			t.Fatalf("Expectect ListDrivers and registry.List() to return same number of items, but got: drivers=%v and list=%v", drivers, list)
		} else if len(list) == len(drivers) {
			t.Fatalf("expect len(list) to be %d; got %d", 2, len(list))
		}
	})
}

func TestRegistry2(t *testing.T) {
	foo := testDriver("foo")
	bar := testDriver("bar")

	registry := createRegistry()
	if err := registry.Register(foo); err != nil {
		t.Skipf("error not expect but got: %v", err)
	}
	if err := registry.Register(bar); err != nil {
		t.Skipf("error not expect but got: %v", err)
	}
	t.Run("Driver", func(t *testing.T) {
		driverName := "foo"
		driver, err := registry.Driver(driverName)
		if err != nil {
			t.Fatalf("expect nil for registering foo driver, but got: %v", err)
		}
		if driver.Name != driverName {
			t.Fatalf("expect registry.Driver(%s) returns registered driver, but got: %s", driverName, driver.Name)
		}
	})
	t.Run("NotExistingDriver", func(t *testing.T) {
		_, err := registry.Driver("foo2")
		if err != ErrDriverNotFound {
			t.Fatalf("expect ErrDriverNotFound bug got: %v", err)
		}
	})
	t.Run("Driver", func(t *testing.T) {
		if _, err := Driver("no_such_driver"); err == nil {
			t.Fatal("expect to get error for not existing driver")
		}
	})
	if _, err := Driver("foo"); err == nil {
		t.Fatal("expect to not get error during existing driver foo")
	}
	t.Run("Register", func(t *testing.T) {
		if err := Register(foo); err != nil {
			t.Fatalf("expect to not get error during registering driver foo, but got: %v", err)
		}
	})
}
