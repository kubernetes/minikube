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
	"os/exec"
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

// NeedsSudo try to execute driver related commands to determine if sudo is required
func NeedsSudo(driverName string) bool {
	var cmd *exec.Cmd

	switch driverName {
	case "docker":
		cmd = exec.Command("docker", "info")
	case "podman":
		cmd = exec.Command("podman", "info")
	case "kvm2":
		cmd = exec.Command("virsh", "list", "--all")
	case "qemu2":
		cmd = exec.Command("qemu-system-x86_64", "--version")
	case "virtualbox":
		cmd = exec.Command("VBoxManage", "list", "vms")
	case "hyperkit":
		cmd = exec.Command("hyperkit", "-v")
	case "hyperv":
		cmd = exec.Command("powershell", "-Command", "Get-VM") // Windows Hyper-V
	case "vmware":
		cmd = exec.Command("vmrun", "list")
	case "parallels":
		cmd = exec.Command("prlctl", "list")
	case "vfkit":
		cmd = exec.Command("vfctl", "list") // macOS vfkit CLI
	case "ssh":
		cmd = exec.Command("ssh", "-V")
	case "krunkit":
		cmd = exec.Command("krun", "--version") // krunkit CLI
	case "none":
		return true // none driver almost always requires root
	default:
		return false // By default, no sudo is required
	}

	// Execute the command and check if it succeeds
	if err := cmd.Run(); err != nil {
		// If it fails, it may require sudo
		return true
	}
	return false
}

func TestNeedsSudo(t *testing.T) {
	drivers := []string{"docker", "podman", "kvm2", "none", "qemu2", "virtualbox", "hyperkit", "vmware", "parallels", "ssh", "krunkit", "hyperv", "vfkit"}
	for _, d := range drivers {
		t.Logf("%s NeedsSudo = %v", d, NeedsSudo(d))
	}
}
