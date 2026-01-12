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

	"k8s.io/minikube/pkg/libmachine/auth"
)

func MigrateHostV1ToHostV2(hostV1 *V1) *V2 {
	// Changed:  Put StorePath directly in AuthOptions (for provisioning),
	// and AuthOptions.PrivateKeyPath => AuthOptions.CaPrivateKeyPath
	// Also, CertDir has been added.

	globalStorePath := filepath.Dir(filepath.Dir(hostV1.StorePath))

	h := &V2{
		ConfigVersion: hostV1.ConfigVersion,
		Driver:        hostV1.Driver,
		Name:          hostV1.Driver.GetMachineName(),
		DriverName:    hostV1.DriverName,
		HostOptions: &Options{
			Driver:        hostV1.HostOptions.Driver,
			Memory:        hostV1.HostOptions.Memory,
			Disk:          hostV1.HostOptions.Disk,
			EngineOptions: hostV1.HostOptions.EngineOptions,
			SwarmOptions:  hostV1.HostOptions.SwarmOptions,
			AuthOptions: &auth.Options{
				CertDir:              filepath.Join(globalStorePath, "certs"),
				CaCertPath:           hostV1.HostOptions.AuthOptions.CaCertPath,
				CaPrivateKeyPath:     hostV1.HostOptions.AuthOptions.PrivateKeyPath,
				CaCertRemotePath:     hostV1.HostOptions.AuthOptions.CaCertRemotePath,
				ServerCertPath:       hostV1.HostOptions.AuthOptions.ServerCertPath,
				ServerKeyPath:        hostV1.HostOptions.AuthOptions.ServerKeyPath,
				ClientKeyPath:        hostV1.HostOptions.AuthOptions.ClientKeyPath,
				ServerCertRemotePath: hostV1.HostOptions.AuthOptions.ServerCertRemotePath,
				ServerKeyRemotePath:  hostV1.HostOptions.AuthOptions.ServerKeyRemotePath,
				ClientCertPath:       hostV1.HostOptions.AuthOptions.ClientCertPath,
				StorePath:            globalStorePath,
			},
		},
	}

	return h
}
