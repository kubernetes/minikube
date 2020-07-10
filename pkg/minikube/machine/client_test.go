/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package machine

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"testing"

	"github.com/docker/machine/libmachine/drivers/plugin/localbinary"

	"k8s.io/minikube/pkg/minikube/driver"
	_ "k8s.io/minikube/pkg/minikube/registry/drvs/virtualbox"
	testutil "k8s.io/minikube/pkg/minikube/tests"
)

const vboxConfig = `
{
        "IPAddress": "192.168.99.101",
        "MachineName": "minikube",
        "SSHUser": "docker",
        "SSHPort": 33627,
        "SSHKeyPath": "/home/sundarp/.minikube/machines/minikube/id_rsa",
        "StorePath": "/home/sundarp/.minikube",
        "SwarmMaster": false,
        "SwarmHost": "",
        "SwarmDiscovery": "",
        "VBoxManager": {},
        "HostInterfaces": {},
        "CPU": 4,
        "Memory": 16384,
        "DiskSize": 20000,
        "NatNicType": "82540EM",
        "Boot2DockerURL": "file:///home/sundarp/.minikube/cache/iso/minikube-v1.0.6.iso",
        "Boot2DockerImportVM": "",
        "HostDNSResolver": false,
        "HostOnlyCIDR": "192.168.99.1/24",
        "HostOnlyNicType": "82540EM",
        "HostOnlyPromiscMode": "deny",
        "UIType": "headless",
        "HostOnlyNoDHCP": false,
        "NoShare": false,
        "DNSProxy": true,
        "NoVTXCheck": false
}
`

func TestLocalClientNewHost(t *testing.T) {
	c, err := NewAPIClient()
	if err != nil {
		t.Fatal(err)
	}

	var tests = []struct {
		description string
		driver      string
		rawDriver   []byte
		err         bool
	}{
		{
			description: "host vbox correct",
			driver:      driver.VirtualBox,
			rawDriver:   []byte(vboxConfig),
		},
		{
			description: "host vbox incorrect",
			driver:      driver.VirtualBox,
			rawDriver:   []byte("?"),
			err:         true,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.description, func(t *testing.T) {
			t.Parallel()
			host, err := c.NewHost(test.driver, test.rawDriver)
			// A few sanity checks that we can do on the host
			if host != nil {
				if host.DriverName != test.driver {
					t.Errorf("Host driver name is not correct.  Expected: %s, got: %s", test.driver, host.DriverName)
				}
				if host.Name != host.Driver.GetMachineName() {
					t.Errorf("Host name is not correct.  Expected :%s, got: %s", host.Driver.GetMachineName(), host.Name)
				}
			}
			if err != nil && !test.err {
				t.Errorf("Unexpected error: %v", err)
			}
			if err == nil && test.err {
				t.Errorf("No error returned, but expected err")
			}
		})
	}
}

func TestRunNotDriver(t *testing.T) {
	tempDir := testutil.MakeTempDir()
	defer testutil.RemoveTempDir(tempDir)
	StartDriver()
	if !localbinary.CurrentBinaryIsDockerMachine {
		t.Fatal("CurrentBinaryIsDockerMachine not set. This will break driver initialization.")
	}
}

func TestRunDriver(t *testing.T) {
	// This test is a bit complicated. It verifies that when the root command is
	// called with the proper environment variables, we setup the libmachine driver.
	tempDir := testutil.MakeTempDir()
	defer testutil.RemoveTempDir(tempDir)

	os.Setenv(localbinary.PluginEnvKey, localbinary.PluginEnvVal)
	os.Setenv(localbinary.PluginEnvDriverName, driver.VirtualBox)

	// Capture stdout and reset it later.
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() {
		os.Stdout = old
	}()

	// Run the command asynchronously. It should listen on a port for connections.
	go StartDriver()

	// The command will write out what port it's listening on over stdout.
	reader := bufio.NewReader(r)
	addr, _, err := reader.ReadLine()
	if err != nil {
		t.Fatal("Failed to read address over stdout.")
	}
	os.Stdout = old

	fmt.Println(string(addr))

	// Now that we got the port, make sure we can connect.
	if _, err := net.Dial("tcp", string(addr)); err != nil {
		t.Fatal("Driver not listening.")
	}
}
