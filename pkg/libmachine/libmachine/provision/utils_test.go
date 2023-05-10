/*
Copyright 2023 The Kubernetes Authors All rights reserved.

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

package provision

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/minikube/pkg/libmachine/drivers/fakedriver"
	"k8s.io/minikube/pkg/libmachine/libmachine/auth"
	"k8s.io/minikube/pkg/libmachine/libmachine/engine"
	"k8s.io/minikube/pkg/libmachine/libmachine/provision/pkgaction"
	"k8s.io/minikube/pkg/libmachine/libmachine/provision/provisiontest"
	"k8s.io/minikube/pkg/libmachine/libmachine/provision/serviceaction"
	"k8s.io/minikube/pkg/libmachine/libmachine/swarm"
)

var (
	reDaemonListening = ":2376\\s+.*:.*"
)

func TestMatchNetstatOutMissing(t *testing.T) {
	nsOut := `Active Internet connections (servers and established)
Proto Recv-Q Send-Q Local Address           Foreign Address         State
tcp        0      0 0.0.0.0:22              0.0.0.0:*               LISTEN
tcp        0      0 0.0.0.0:237             0.0.0.0:*               LISTEN
tcp6       0      0 :::22                   :::*                    LISTEN
tcp6       0      0 :::23760                :::*                    LISTEN`
	if matchNetstatOut(reDaemonListening, nsOut) {
		t.Fatal("Expected not to match the netstat output as showing the daemon listening but got a match")
	}
}

func TestMatchNetstatOutPresent(t *testing.T) {
	nsOut := `Active Internet connections (servers and established)
Proto Recv-Q Send-Q Local Address           Foreign Address         State
tcp        0      0 0.0.0.0:22              0.0.0.0:*               LISTEN
tcp6       0      0 :::2376                 :::*                    LISTEN
tcp6       0      0 :::22                   :::*                    LISTEN`
	if !matchNetstatOut(reDaemonListening, nsOut) {
		t.Fatal("Expected to match the netstat output as showing the daemon listening but didn't")
	}
}

func TestMatchSsOutMissing(t *testing.T) {
	ssOut := `State      Recv-Q Send-Q Local Address:Port               Peer Address:Port              
LISTEN     0      128          *:22                       *:*                  
LISTEN     0      128         :::22                      :::*                  
LISTEN     0      128         :::23760                   :::*                  `
	if matchNetstatOut(reDaemonListening, ssOut) {
		t.Fatal("Expected not to match the ss output as showing the daemon listening but got a match")
	}
}

func TestMatchSsOutPresent(t *testing.T) {
	ssOut := `State      Recv-Q Send-Q Local Address:Port               Peer Address:Port              
LISTEN     0      128          *:22                       *:*                  
LISTEN     0      128         :::22                      :::*                  
LISTEN     0      128         :::2376                    :::*                  `
	if !matchNetstatOut(reDaemonListening, ssOut) {
		t.Fatal("Expected to match the ss output as showing the daemon listening but didn't")
	}
}

type fakeProvisioner struct {
	GenericProvisioner
}

func (provisioner *fakeProvisioner) Package(_ string, _ pkgaction.PackageAction) error {
	return nil
}

func (provisioner *fakeProvisioner) Provision(_ swarm.Options, _ auth.Options, _ engine.Options) error {
	return nil
}

func (provisioner *fakeProvisioner) Service(_ string, _ serviceaction.ServiceAction) error {
	return nil
}

func (provisioner *fakeProvisioner) String() string {
	return "fake"
}

func TestDecideStorageDriver(t *testing.T) {
	var tests = []struct {
		suppliedDriver       string
		defaultDriver        string
		remoteFilesystemType string
		expectedDriver       string
	}{
		{"", "aufs", "ext4", "aufs"},
		{"", "aufs", "btrfs", "btrfs"},
		{"", "overlay", "btrfs", "overlay"},
		{"devicemapper", "aufs", "ext4", "devicemapper"},
		{"devicemapper", "aufs", "btrfs", "devicemapper"},
	}

	p := &fakeProvisioner{GenericProvisioner{
		Driver: &fakedriver.Driver{},
	}}
	for _, test := range tests {
		p.Commander = provisiontest.NewFakeSSHCommander(
			provisiontest.FakeSSHCommanderOptions{
				FilesystemType: test.remoteFilesystemType,
			},
		)
		storageDriver, err := decideStorageDriver(p, test.defaultDriver, test.suppliedDriver)
		assert.NoError(t, err)
		assert.Equal(t, test.expectedDriver, storageDriver)
	}
}

func TestGetFilesystemType(t *testing.T) {
	p := &fakeProvisioner{GenericProvisioner{
		Driver: &fakedriver.Driver{},
	}}
	p.Commander = &provisiontest.FakeSSHCommander{
		Responses: map[string]string{
			"stat -f -c %T /var/lib": "btrfs\n",
		},
	}
	fsType, err := getFilesystemType(p, "/var/lib")
	assert.NoError(t, err)
	assert.Equal(t, "btrfs", fsType)
}

func TestDockerClientVersion(t *testing.T) {
	cases := []struct {
		output, want string
	}{
		{"Docker version 1.9.1, build a34a1d5\n", "1.9.1"},
		{"Docker version 1.9.1\n", "1.9.1"},
		{"Docker version 1.13.0-rc1, build deadbeef\n", "1.13.0-rc1"},
		{"Docker version 1.13.0-dev, build deadbeef\n", "1.13.0-dev"},
	}

	sshCmder := &provisiontest.FakeSSHCommander{
		Responses: make(map[string]string),
	}

	for _, tc := range cases {
		sshCmder.Responses["docker --version"] = tc.output
		got, err := DockerClientVersion(sshCmder)
		if err != nil {
			t.Fatal(err)
		}
		if got != tc.want {
			t.Errorf("Unexpected version string from %q; got %q, want %q", tc.output, tc.want, got)
		}
	}
}
