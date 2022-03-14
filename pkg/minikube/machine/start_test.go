/*
Copyright 2021 The Kubernetes Authors All rights reserved.

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
	"fmt"
	"net"
	"os"
	"testing"

	"k8s.io/minikube/pkg/minikube/command"
)

const initialEtcHostsContent string = `127.0.0.1	localhost
127.0.1.1	mymachinename

# The following lines are desirable for IPv6 capable hosts
::1     ip6-localhost ip6-loopback
fe00::0 ip6-localnet
127.0.0.1	host.minikube.internal
192.168.42.139	control-plane.minikube.internal
10.8.0.22	control-plane.minikube.internal
`

const expectedEtcHostsContent string = `127.0.0.1	localhost
127.0.1.1	mymachinename

# The following lines are desirable for IPv6 capable hosts
::1     ip6-localhost ip6-loopback
fe00::0 ip6-localnet
127.0.0.1	host.minikube.internal
10.1.2.3	control-plane.minikube.internal
`

func TestAddHostAliasInner(t *testing.T) {
	// Arrange
	tempFilePath, err := writeContentToTempFile(initialEtcHostsContent)
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(tempFilePath)
	hostname := "control-plane.minikube.internal"
	ip := net.ParseIP("10.1.2.3")
	hostsFileLine := fmt.Sprintf("%s\t%s", ip, hostname)

	// Act
	cmd := addHostAliasCommand(hostname, hostsFileLine, false, tempFilePath)
	if err = cmd.Run(); err != nil {
		t.Error(err)
	}

	// Assert
	if err != nil {
		t.Error(err)
	}

	buff, err := os.ReadFile(tempFilePath)
	if err != nil {
		t.Error(err)
	}

	actualEtcHostsContent := string(buff)
	if expectedEtcHostsContent != actualEtcHostsContent {
		t.Errorf("\n=== EXPECTED ===\n%s\n=== ACTUAL ===\n%s\n",
			expectedEtcHostsContent,
			actualEtcHostsContent)
	}
}

func writeContentToTempFile(content string) (string, error) {
	etcHosts, err := os.CreateTemp("", "hosts")
	if err != nil {
		return "", err
	}

	path := etcHosts.Name()
	if _, err = etcHosts.WriteString(content); err != nil {
		return "", err
	}

	if err = etcHosts.Close(); err != nil {
		return "", err
	}

	return path, nil
}

func TestDiskUsed(t *testing.T) {
	ex := command.NewFakeCommandRunner()
	ex.SetCommandToOutput(map[string]string{
		"sh -c \"df -h /var | awk 'NR==2{print $5}'\"": "20%",
	})
	nonex := command.NewFakeCommandRunner()
	nonex.SetCommandToOutput(map[string]string{
		"sh -c \"df -h /nonexistent | awk 'NR==2{print $5}'\"": "df: /nonexistent: No such file or directory",
	})

	type args struct {
		cr  command.Runner
		dir string
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{
			name:    "existent",
			args:    args{ex, "/var"},
			want:    20,
			wantErr: false,
		},
		{
			name:    "existent",
			args:    args{nonex, "/nonexistent"},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Logf("starting %v", tt.name)
		t.Run(tt.name, func(t *testing.T) {
			got, err := DiskUsed(tt.args.cr, tt.args.dir)
			t.Logf("err: %v\n", err)
			t.Logf("got: %v\n", got)
			if (err != nil) != tt.wantErr {
				t.Fatalf("DiskUsed() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("DiskUsed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDiskAvailable(t *testing.T) {
	ex := command.NewFakeCommandRunner()
	ex.SetCommandToOutput(map[string]string{
		"sh -c \"df -BG /var | awk 'NR==2{print $4}'\"": "20",
	})
	nonex := command.NewFakeCommandRunner()
	nonex.SetCommandToOutput(map[string]string{
		"sh -c \"df -BG /nonexistent | awk 'NR==2{print $4}'\"": "df: /nonexistent: No such file or directory",
	})

	type args struct {
		cr  command.Runner
		dir string
	}
	tests := []struct {
		name    string
		args    args
		want    int
		wantErr bool
	}{
		{
			name:    "existent",
			args:    args{ex, "/var"},
			want:    20,
			wantErr: false,
		},
		{
			name:    "existent",
			args:    args{nonex, "/nonexistent"},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Logf("starting %v", tt.name)
		t.Run(tt.name, func(t *testing.T) {
			got, err := DiskAvailable(tt.args.cr, tt.args.dir)
			t.Logf("err: %v\n", err)
			if (err != nil) != tt.wantErr {
				t.Fatalf("DiskAvailable() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("DiskAvailable() = %v, want %v", got, tt.want)
			}
		})
	}
}
