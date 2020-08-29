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

package driver

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"k8s.io/minikube/pkg/minikube/registry"
)

func TestSupportedDrivers(t *testing.T) {
	got := SupportedDrivers()
	found := false
	for _, s := range SupportedDrivers() {
		if s == VirtualBox {
			found = true
		}
	}

	if found == false {
		t.Errorf("%s not in supported drivers: %v", VirtualBox, got)
	}
}

func TestSupported(t *testing.T) {
	if !Supported(VirtualBox) {
		t.Errorf("Supported(%s) is false", VirtualBox)
	}
	if Supported("yabba?") {
		t.Errorf("Supported(yabba?) is true")
	}
}

func TestBareMetal(t *testing.T) {
	if !BareMetal(None) {
		t.Errorf("Supported(%s) is false", None)
	}
	if BareMetal(VirtualBox) {
		t.Errorf("Supported(%s) is true", VirtualBox)
	}
}

func TestMachineType(t *testing.T) {
	types := map[string]string{
		Podman:       "container",
		Docker:       "container",
		Mock:         "bare metal machine",
		None:         "bare metal machine",
		KVM2:         "VM",
		VirtualBox:   "VM",
		HyperKit:     "VM",
		VMware:       "VM",
		VMwareFusion: "VM",
		HyperV:       "VM",
		Parallels:    "VM",
	}

	drivers := SupportedDrivers()
	for _, driver := range drivers {
		want := types[driver]
		got := MachineType(driver)
		if want != got {
			t.Errorf("mismatched machine type for driver %s: want = %s got = %s", driver, want, got)
		}
	}
}

func TestFlagDefaults(t *testing.T) {
	expected := FlagHints{CacheImages: true}
	if diff := cmp.Diff(FlagDefaults(VirtualBox), expected); diff != "" {
		t.Errorf("defaults mismatch (-want +got):\n%s", diff)
	}

	tf, err := ioutil.TempFile("", "resolv.conf")
	if err != nil {
		t.Fatalf("tempfile: %v", err)
	}
	defer os.Remove(tf.Name()) // clean up

	expected = FlagHints{
		CacheImages:  false,
		ExtraOptions: []string{fmt.Sprintf("kubelet.resolv-conf=%s", tf.Name())},
	}
	systemdResolvConf = tf.Name()
	if diff := cmp.Diff(FlagDefaults(None), expected); diff != "" {
		t.Errorf("defaults mismatch (-want +got):\n%s", diff)
	}
}

func TestSuggest(t *testing.T) {
	tests := []struct {
		def     registry.DriverDef
		choices []string
		pick    string
		alts    []string
		rejects []string
	}{
		{
			def: registry.DriverDef{
				Name:     "unhealthy",
				Priority: registry.Default,
				Status:   func() registry.State { return registry.State{Installed: true, Healthy: false} },
			},
			choices: []string{"unhealthy"},
			pick:    "",
			alts:    []string{},
			rejects: []string{"unhealthy"},
		},
		{
			def: registry.DriverDef{
				Name:     "discouraged",
				Priority: registry.Discouraged,
				Status:   func() registry.State { return registry.State{Installed: true, Healthy: true} },
			},
			choices: []string{"discouraged", "unhealthy"},
			pick:    "",
			alts:    []string{"discouraged"},
			rejects: []string{"unhealthy"},
		},
		{
			def: registry.DriverDef{
				Name:     "default",
				Priority: registry.Default,
				Status:   func() registry.State { return registry.State{Installed: true, Healthy: true} },
			},
			choices: []string{"default", "discouraged", "unhealthy"},
			pick:    "default",
			alts:    []string{"discouraged"},
			rejects: []string{"unhealthy"},
		},
		{
			def: registry.DriverDef{
				Name:     "preferred",
				Priority: registry.Preferred,
				Status:   func() registry.State { return registry.State{Installed: true, Healthy: true} },
			},
			choices: []string{"preferred", "default", "discouraged", "unhealthy"},
			pick:    "preferred",
			alts:    []string{"default", "discouraged"},
			rejects: []string{"unhealthy"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.def.Name, func(t *testing.T) {
			if tc.def.Name != "" {
				if err := registry.Register(tc.def); err != nil {
					t.Errorf("register returned error: %v", err)
				}
			}

			got := Choices(false)
			gotNames := []string{}
			for _, c := range got {
				gotNames = append(gotNames, c.Name)
			}

			if diff := cmp.Diff(gotNames, tc.choices); diff != "" {
				t.Errorf("choices mismatch (-want +got):\n%s", diff)
			}

			pick, alts, rejects := Suggest(got)
			if pick.Name != tc.pick {
				t.Errorf("pick = %q, expected %q", pick.Name, tc.pick)
			}

			gotAlts := []string{}
			for _, a := range alts {
				gotAlts = append(gotAlts, a.Name)
			}
			if diff := cmp.Diff(gotAlts, tc.alts); diff != "" {
				t.Errorf("alts mismatch (-want +got):\n%s", diff)
			}

			gotRejects := []string{}
			for _, r := range rejects {
				gotRejects = append(gotRejects, r.Name)
			}
			if diff := cmp.Diff(gotRejects, tc.rejects); diff != "" {
				t.Errorf("rejects mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
