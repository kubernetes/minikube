/*
Copyright 2022 The Kubernetes Authors All rights reserved.

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

package host

import (
	"path/filepath"
	"testing"
)

var (
	v1conf = []byte(`{
    "ConfigVersion": 1,
    "Driver": {
        "IPAddress": "192.168.99.100",
        "SSHUser": "docker",
        "SSHPort": 64477,
        "MachineName": "foobar",
        "CaCertPath": "/Users/catbug/.docker/machine/certs/ca.pem",
        "PrivateKeyPath": "/Users/catbug/.docker/machine/certs/ca-key.pem",
        "SwarmMaster": false,
        "SwarmHost": "tcp://0.0.0.0:3376",
        "SwarmDiscovery": "",
        "CPU": 1,
        "Memory": 1024,
        "DiskSize": 20000,
        "Boot2DockerURL": "",
        "Boot2DockerImportVM": "",
        "HostOnlyCIDR": "192.168.99.1/24"
    },
    "DriverName": "virtualbox",
    "HostOptions": {
        "Driver": "",
        "Memory": 0,
        "Disk": 0,
        "EngineOptions": {
            "ArbitraryFlags": [],
            "Dns": null,
            "GraphDir": "",
            "Env": [],
            "Ipv6": false,
            "InsecureRegistry": [],
            "Labels": [],
            "LogLevel": "",
            "StorageDriver": "",
            "SelinuxEnabled": false,
            "TlsCaCert": "",
            "TlsCert": "",
            "TlsKey": "",
            "TlsVerify": true,
            "RegistryMirror": [],
            "InstallURL": "https://get.docker.com"
        },
        "SwarmOptions": {
            "IsSwarm": false,
            "Address": "",
            "Discovery": "",
            "Master": false,
            "Host": "tcp://0.0.0.0:3376",
            "Image": "swarm:latest",
            "Strategy": "spread",
            "Heartbeat": 0,
            "Overcommit": 0,
            "TlsCaCert": "",
            "TlsCert": "",
            "TlsKey": "",
            "TlsVerify": false,
            "ArbitraryFlags": []
        },
        "AuthOptions": {
            "StorePath": "",
            "CaCertPath": "/Users/catbug/.docker/machine/certs/ca.pem",
            "CaCertRemotePath": "",
            "ServerCertPath": "/Users/catbug/.docker/machine/machines/foobar/server.pem",
            "ServerKeyPath": "/Users/catbug/.docker/machine/machines/foobar/server-key.pem",
            "ClientKeyPath": "/Users/catbug/.docker/machine/certs/key.pem",
            "ServerCertRemotePath": "",
            "ServerKeyRemotePath": "",
            "PrivateKeyPath": "/Users/catbug/.docker/machine/certs/ca-key.pem",
            "ClientCertPath": "/Users/catbug/.docker/machine/certs/cert.pem"
        }
    },
    "StorePath": "/Users/catbug/.docker/machine/machines/foobar"
}`)
)

func TestMigrateHostV1ToHostV2(t *testing.T) {
	h := &Host{}
	expectedGlobalStorePath := filepath.Clean("/Users/catbug/.docker/machine")
	expectedCaPrivateKeyPath := "/Users/catbug/.docker/machine/certs/ca-key.pem"
	expectedCertDir := filepath.Join(expectedGlobalStorePath, "certs")
	migratedHost, migrationPerformed, err := MigrateHost(h, v1conf)
	if err != nil {
		t.Fatalf("Error attempting to migrate host: %s", err)
	}

	if !migrationPerformed {
		t.Fatal("Expected a migration to be reported as performed but it was not")
	}

	if filepath.Clean(migratedHost.HostOptions.AuthOptions.StorePath) != expectedGlobalStorePath {
		t.Fatalf("Expected %q, got %q for the store path in AuthOptions", expectedGlobalStorePath, filepath.Clean(migratedHost.HostOptions.AuthOptions.StorePath))
	}

	if migratedHost.HostOptions.AuthOptions.CaPrivateKeyPath != expectedCaPrivateKeyPath {
		t.Fatalf("Expected %q, got %q for the private key path in AuthOptions", expectedCaPrivateKeyPath, migratedHost.HostOptions.AuthOptions.CaPrivateKeyPath)
	}

	if filepath.Clean(migratedHost.HostOptions.AuthOptions.CertDir) != expectedCertDir {
		t.Fatalf("Expected %q, got %q for the cert dir in AuthOptions", expectedCertDir, filepath.Clean(migratedHost.HostOptions.AuthOptions.CertDir))
	}
}
