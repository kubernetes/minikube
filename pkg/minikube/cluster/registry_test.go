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

package cluster

import (
	"sort"
	"testing"
)

func TestRegistry(t *testing.T) {
	dummy := func(_ MachineConfig) RawDriver {
		return nil
	}

	registry := createRegistry()

	err := registry.Register("foo", dummy)
	if err != nil {
		t.Fatal("expect nil")
	}

	err = registry.Register("foo", dummy)
	if err != ErrDriverNameExist {
		t.Fatal("expect ErrDriverNameExist")
	}

	err = registry.Register("bar", dummy)
	if err != nil {
		t.Fatal("expect nil")
	}

	list := registry.List()
	if len(list) != 2 {
		t.Fatalf("expect len(list) to be %d; got %d", 2, len(list))
	}

	sort.Strings(list)
	if list[0] != "bar" || list[1] != "foo" {
		t.Fatalf("expect registry.List return %s; got %s", []string{"bar", "foo"}, list)
	}

	driver, err := registry.Driver("foo")
	if err != nil {
		t.Fatal("expect nil")
	}
	if driver == nil {
		t.Fatal("expect registry.Driver(foo) returns registered driver")
	}

	driver, err = registry.Driver("foo2")
	if err != ErrDriverNotFound {
		t.Fatal("expect ErrDriverNotFound")
	}
}
