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

package drivers

import (
	"os"
	"path/filepath"
	"testing"

	"k8s.io/minikube/pkg/minikube/tests"
)

func Test_createDiskImage(t *testing.T) {
	tmpdir := tests.MakeTempDir(t)

	sshPath := filepath.Join(tmpdir, "ssh")
	if err := os.WriteFile(sshPath, []byte("mysshkey"), 0644); err != nil {
		t.Fatalf("writefile: %v", err)
	}
	diskPath := filepath.Join(tmpdir, "disk")

	sizeInMb := 100
	sizeInBytes := int64(104857600)
	if err := createRawDiskImage(sshPath, diskPath, sizeInMb); err != nil {
		t.Errorf("createDiskImage() error = %v", err)
	}
	fi, err := os.Lstat(diskPath)
	if err != nil {
		t.Errorf("Lstat() error = %v", err)
	}
	if fi.Size() != sizeInBytes {
		t.Errorf("Disk size is %v, want %v", fi.Size(), sizeInBytes)
	}
}

var validLeases = []byte(`{
        name=foo
        ip_address=1.2.3.4
        hw_address=1,a1:b2:c3:d4:e5:f6
        identifier=1,a2:b3:c4:d5:e6:f7
        lease=0x597e1267
}
{
        name=bootp
        ip_address=192.168.105.2
        hw_address=1,8e:1:99:9c:54:b1
        identifier=1,8e:1:99:9c:54:b1
        lease=0x597e1268
}
{
        name=bar
        ip_address=192.168.64.3
        hw_address=1,a4:b5:c6:d7:e8:f9
        identifier=1,a0:b0:c0:d0:e0:f0
        lease=0x597e1267
}
{
        name=bar
        ip_address=192.168.64.4
        hw_address=1,a5:b6:c7:d8:e9:f1
        identifier=1,a5:b6:c7:d8:e9:f1
        lease=0x597e1268
}`)

func Test_getIpAddressFromFile(t *testing.T) {
	tmpdir := tests.MakeTempDir(t)

	dhcpFile := filepath.Join(tmpdir, "dhcp")
	if err := os.WriteFile(dhcpFile, validLeases, 0644); err != nil {
		t.Fatalf("writefile: %v", err)
	}

	invalidFile := filepath.Join(tmpdir, "invalid")
	if err := os.WriteFile(invalidFile, []byte("foo"), 0644); err != nil {
		t.Fatalf("writefile: %v", err)
	}

	type args struct {
		mac  string
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			"valid",
			args{"a1:b2:c3:d4:e5:f6", dhcpFile},
			"1.2.3.4",
			false,
		},
		{
			// bootp use "%x" format instead of "%02x"
			// https://openradar.appspot.com/FB15382970
			"valid-bootp",
			args{"8e:01:99:9c:54:b1", dhcpFile},
			"192.168.105.2",
			false,
		},
		{
			"duplicate",
			args{"a4:b5:c6:d7:e8:f9", dhcpFile},
			"192.168.64.3",
			false,
		},
		{
			"invalid",
			args{"a1:b2:c3:d4:e5:f6", invalidFile},
			"",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getIPAddressFromFile(tt.args.mac, tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("getIPAddressFromFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getIPAddressFromFile() = %v, want %v", got, tt.want)
			}
		})
	}
}
