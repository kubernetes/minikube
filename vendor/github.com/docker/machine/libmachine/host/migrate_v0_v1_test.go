package host

import (
	"reflect"
	"testing"

	"github.com/docker/machine/commands/mcndirs"
	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/swarm"
)

func TestMigrateHostV0ToV1(t *testing.T) {
	mcndirs.BaseDir = "/tmp/migration"
	originalHost := &V0{
		HostOptions:    nil,
		SwarmDiscovery: "token://foobar",
		SwarmHost:      "1.2.3.4:2376",
		SwarmMaster:    true,
		CaCertPath:     "/tmp/migration/certs/ca.pem",
		PrivateKeyPath: "/tmp/migration/certs/ca-key.pem",
		ClientCertPath: "/tmp/migration/certs/cert.pem",
		ClientKeyPath:  "/tmp/migration/certs/key.pem",
		ServerCertPath: "/tmp/migration/certs/server.pem",
		ServerKeyPath:  "/tmp/migration/certs/server-key.pem",
	}
	hostOptions := &OptionsV1{
		SwarmOptions: &swarm.Options{
			Master:    true,
			Discovery: "token://foobar",
			Host:      "1.2.3.4:2376",
		},
		AuthOptions: &AuthOptionsV1{
			CaCertPath:     "/tmp/migration/certs/ca.pem",
			PrivateKeyPath: "/tmp/migration/certs/ca-key.pem",
			ClientCertPath: "/tmp/migration/certs/cert.pem",
			ClientKeyPath:  "/tmp/migration/certs/key.pem",
			ServerCertPath: "/tmp/migration/certs/server.pem",
			ServerKeyPath:  "/tmp/migration/certs/server-key.pem",
		},
		EngineOptions: &engine.Options{
			InstallURL: "https://get.docker.com",
			TLSVerify:  true,
		},
	}

	expectedHost := &V1{
		HostOptions: hostOptions,
	}

	host := MigrateHostV0ToHostV1(originalHost)

	if !reflect.DeepEqual(host, expectedHost) {
		t.Logf("\n%+v\n%+v", host, expectedHost)
		t.Logf("\n%+v\n%+v", host.HostOptions, expectedHost.HostOptions)
		t.Fatal("Expected these structs to be equal, they were different")
	}
}

func TestMigrateHostMetadataV0ToV1(t *testing.T) {
	metadata := &MetadataV0{
		HostOptions: Options{
			EngineOptions: nil,
			AuthOptions:   nil,
		},
		StorePath:      "/tmp/store",
		CaCertPath:     "/tmp/store/certs/ca.pem",
		ServerCertPath: "/tmp/store/certs/server.pem",
	}
	expectedAuthOptions := &auth.Options{
		StorePath:      "/tmp/store",
		CaCertPath:     "/tmp/store/certs/ca.pem",
		ServerCertPath: "/tmp/store/certs/server.pem",
	}

	expectedMetadata := &Metadata{
		HostOptions: Options{
			EngineOptions: &engine.Options{},
			AuthOptions:   expectedAuthOptions,
		},
	}

	m := MigrateHostMetadataV0ToHostMetadataV1(metadata)

	if !reflect.DeepEqual(m, expectedMetadata) {
		t.Logf("\n%+v\n%+v", m, expectedMetadata)
		t.Fatal("Expected these structs to be equal, they were different")
	}
}
