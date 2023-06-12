/*
Copyright 2017 The Kubernetes Authors All rights reserved.

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

package vmware

import (
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/state"
)

var skip = !check(vmrunbin) || !check(vdiskmanbin)

func check(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		log.Printf("%q is missing", path)
		return false
	}

	return true
}

func TestSetConfigFromFlags(t *testing.T) {
	driver := NewDriver("default", "path")

	checkFlags := &drivers.CheckDriverOptions{
		FlagsValues: map[string]interface{}{},
		CreateFlags: driver.GetCreateFlags(),
	}

	err := driver.SetConfigFromFlags(checkFlags)
	if err != nil {
		t.Fatal(err)
	}

	if len(checkFlags.InvalidFlags) != 0 {
		t.Fatalf("expect len(checkFlags.InvalidFlags) == 0; got %d", len(checkFlags.InvalidFlags))
	}
}

func TestDriver(t *testing.T) {
	// skip driver tests
	if true {
		t.Skip()
	}

	path, err := ioutil.TempDir("", "vmware-driver-test")
	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(path)

	driver := NewDriver("default", path)

	checkFlags := &drivers.CheckDriverOptions{
		FlagsValues: map[string]interface{}{},
		CreateFlags: driver.GetCreateFlags(),
	}

	err = driver.SetConfigFromFlags(checkFlags)
	if err != nil {
		t.Fatal(err)
	}

	driver.(*Driver).Boot2DockerURL = "https://github.com/boot2docker/boot2docker/releases/download/v17.10.0-ce-rc2/boot2docker.iso"

	err = driver.Create()
	if err != nil {
		t.Fatal(err)
	}

	defer driver.Remove()

	st, err := driver.GetState()
	if err != nil {
		t.Fatal(err)
	}
	if st != state.Running {
		t.Fatalf("expect state == Running; got %s", st.String())
	}

	ip, err := driver.GetIP()
	if err != nil {
		t.Fatal(err)
	}
	if ip == "" {
		t.Fatal("expect ip non-zero; got ''")
	}

	username := driver.GetSSHUsername()
	if username == "" {
		t.Fatal("expect username non-zero; got ''")
	}

	key := driver.GetSSHKeyPath()
	if key == "" {
		t.Fatal("expect key non-zero; got ''")
	}

	port, err := driver.GetSSHPort()
	if err != nil {
		t.Fatal(err)
	}
	if port == 0 {
		t.Fatal("expect port not 0; got 0")
	}

	host, err := driver.GetSSHHostname()
	if err != nil {
		t.Fatal(err)
	}
	if host == "" {
		t.Fatal("expect host non-zero; got ''")
	}
}
