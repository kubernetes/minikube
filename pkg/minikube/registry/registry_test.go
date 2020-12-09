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

	"github.com/google/go-cmp/cmp"
)

func TestRegister(t *testing.T) {
	r := newRegistry()
	foo := DriverDef{Name: "foo"}
	if err := r.Register(foo); err != nil {
		t.Errorf("Register = %v, expected nil", err)
	}
	if err := r.Register(foo); err == nil {
		t.Errorf("Register = nil, expected duplicate err")
	}
}

func TestDriver(t *testing.T) {
	foo := DriverDef{Name: "foo"}
	r := newRegistry()

	if err := r.Register(foo); err != nil {
		t.Errorf("Register = %v, expected nil", err)
	}

	d := r.Driver("foo")
	if d.Empty() {
		t.Errorf("driver.Empty = true, expected false")
	}

	d = r.Driver("bar")
	if !d.Empty() {
		t.Errorf("driver.Empty = false, expected true")
	}
}

func TestList(t *testing.T) {
	foo := DriverDef{Name: "foo"}
	r := newRegistry()
	if err := r.Register(foo); err != nil {
		t.Errorf("register returned error: %v", err)
	}

	if diff := cmp.Diff(r.List(), []DriverDef{foo}); diff != "" {
		t.Errorf("list mismatch (-want +got):\n%s", diff)
	}
}

func TestDriverAlias(t *testing.T) {
	foo := DriverDef{Name: "foo", Alias: []string{"foo-alias"}}
	r := newRegistry()

	if err := r.Register(foo); err != nil {
		t.Errorf("Register = %v, expected nil", err)
	}

	d := r.Driver("foo")
	if d.Empty() {
		t.Errorf("driver.Empty = true, expected false")
	}

	d = r.Driver("foo-alias")
	if d.Empty() {
		t.Errorf("driver.Empty = true, expected false")
	}

	if diff := cmp.Diff(r.List(), []DriverDef{foo}); diff != "" {
		t.Errorf("list mismatch (-want +got):\n%s", diff)
	}

	d = r.Driver("bar")
	if !d.Empty() {
		t.Errorf("driver.Empty = false, expected true")
	}
}
