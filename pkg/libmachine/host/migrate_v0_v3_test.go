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

import "testing"

var (
	v0conf = []byte(`{"DriverName":"virtualbox","Driver":{"IPAddress":"192.168.99.100","SSHUser":"docker","SSHPort":53507,"MachineName":"dev","CaCertPath":"/Users/ljrittle/.docker/machine/certs/ca.pem","PrivateKeyPath":"/Users/ljrittle/.docker/machine/certs/ca-key.pem","SwarmMaster":false,"SwarmHost":"tcp://0.0.0.0:3376","SwarmDiscovery":"","CPU":-1,"Memory":1024,"DiskSize":20000,"Boot2DockerURL":"","Boot2DockerImportVM":"","HostOnlyCIDR":""},"StorePath":"/Users/ljrittle/.docker/machine/machines/dev","HostOptions":{"Driver":"","Memory":0,"Disk":0,"EngineOptions":{"ArbitraryFlags":null,"Dns":null,"GraphDir":"","Ipv6":false,"InsecureRegistry":null,"Labels":null,"LogLevel":"","StorageDriver":"","SelinuxEnabled":false,"TlsCaCert":"","TlsCert":"","TlsKey":"","TlsVerify":false,"RegistryMirror":null,"InstallURL":""},"SwarmOptions":{"IsSwarm":false,"Address":"","Discovery":"","Master":false,"Host":"tcp://0.0.0.0:3376","Image":"","Strategy":"","Heartbeat":0,"Overcommit":0,"TlsCaCert":"","TlsCert":"","TlsKey":"","TlsVerify":false,"ArbitraryFlags":null},"AuthOptions":{"StorePath":"/Users/ljrittle/.docker/machine/machines/dev","CaCertPath":"/Users/ljrittle/.docker/machine/certs/ca.pem","CaCertRemotePath":"","ServerCertPath":"/Users/ljrittle/.docker/machine/certs/server.pem","ServerKeyPath":"/Users/ljrittle/.docker/machine/certs/server-key.pem","ClientKeyPath":"/Users/ljrittle/.docker/machine/certs/key.pem","ServerCertRemotePath":"","ServerKeyRemotePath":"","PrivateKeyPath":"/Users/ljrittle/.docker/machine/certs/ca-key.pem","ClientCertPath":"/Users/ljrittle/.docker/machine/certs/cert.pem"}}}`)
)

func TestMigrateHostV0ToHostV3(t *testing.T) {
	h := &Host{}
	migratedHost, migrationPerformed, err := MigrateHost(h, v0conf)
	if err != nil {
		t.Fatalf("Error attempting to migrate host: %s", err)
	}

	if !migrationPerformed {
		t.Fatal("Expected a migration to be reported as performed but it was not")
	}

	if migratedHost.DriverName != "virtualbox" {
		t.Fatalf("Expected %q, got %q for the driver name", "virtualbox", migratedHost.DriverName)
	}
}
