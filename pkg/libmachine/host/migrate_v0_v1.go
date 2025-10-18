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
	"k8s.io/minikube/pkg/libmachine/auth"
	"k8s.io/minikube/pkg/libmachine/engine"
	"k8s.io/minikube/pkg/libmachine/swarm"
)

// In the 0.1.0 => 0.2.0 transition, the JSON representation of
// machines changed from a "flat" to a more "nested" structure
// for various options and configuration settings.  To preserve
// compatibility with existing machines, these migration functions
// have been introduced.  They preserve backwards compat at the expense
// of some duplicated information.

// MigrateHostV0ToHostV1 validates host config and modifies if needed
// this is used for configuration updates
func MigrateHostV0ToHostV1(hostV0 *V0) *V1 {
	hostV1 := &V1{
		Driver:     hostV0.Driver,
		DriverName: hostV0.DriverName,
	}

	hostV1.HostOptions = &OptionsV1{}
	hostV1.HostOptions.EngineOptions = &engine.Options{
		TLSVerify:  true,
		InstallURL: "https://get.docker.com",
	}
	hostV1.HostOptions.SwarmOptions = &swarm.Options{
		Address:   "",
		Discovery: hostV0.SwarmDiscovery,
		Host:      hostV0.SwarmHost,
		Master:    hostV0.SwarmMaster,
	}
	hostV1.HostOptions.AuthOptions = &AuthOptionsV1{
		StorePath:            hostV0.StorePath,
		CaCertPath:           hostV0.CaCertPath,
		CaCertRemotePath:     "",
		ServerCertPath:       hostV0.ServerCertPath,
		ServerKeyPath:        hostV0.ServerKeyPath,
		ClientKeyPath:        hostV0.ClientKeyPath,
		ServerCertRemotePath: "",
		ServerKeyRemotePath:  "",
		PrivateKeyPath:       hostV0.PrivateKeyPath,
		ClientCertPath:       hostV0.ClientCertPath,
	}

	return hostV1
}

// MigrateHostMetadataV0ToHostMetadataV1 fills nested host metadata and modifies if needed
// this is used for configuration updates
func MigrateHostMetadataV0ToHostMetadataV1(m *MetadataV0) *Metadata {
	hostMetadata := &Metadata{}
	hostMetadata.DriverName = m.DriverName
	hostMetadata.HostOptions.EngineOptions = &engine.Options{}
	hostMetadata.HostOptions.AuthOptions = &auth.Options{
		StorePath:            m.StorePath,
		CaCertPath:           m.CaCertPath,
		CaCertRemotePath:     "",
		ServerCertPath:       m.ServerCertPath,
		ServerKeyPath:        m.ServerKeyPath,
		ClientKeyPath:        "",
		ServerCertRemotePath: "",
		ServerKeyRemotePath:  "",
		CaPrivateKeyPath:     m.PrivateKeyPath,
		ClientCertPath:       m.ClientCertPath,
	}

	hostMetadata.ConfigVersion = m.ConfigVersion

	return hostMetadata
}
