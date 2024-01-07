/*
Copyright 2019 The Kubernetes Authors All rights reserved.

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

package vmpath

const (
	// GuestAddonsDir is the default path of the addons configuration
	GuestAddonsDir = "/etc/kubernetes/addons"
	// GuestManifestsDir is where the kubelet should look for static Pod manifests
	GuestManifestsDir = "/etc/kubernetes/manifests"
	// GuestEphemeralDir is the path where ephemeral data should be stored within the VM
	GuestEphemeralDir = "/var/tmp/minikube"
	// GuestPersistentDir is the path where persistent data should be stored within the VM (not tmpfs)
	GuestPersistentDir = "/var/lib/minikube"
	// GuestBackupDir is the path where persistent backup data should be stored within the VM (not tmpfs)
	GuestBackupDir = GuestPersistentDir + "/backup"
	// GuestKubernetesCertsDir are where Kubernetes certificates are stored
	GuestKubernetesCertsDir = GuestPersistentDir + "/certs"
	// GuestCertAuthDir is where system CA certificates are installed to
	GuestCertAuthDir = "/usr/share/ca-certificates"
	// GuestCertStoreDir is where system SSL certificates are installed
	GuestCertStoreDir = "/etc/ssl/certs"
	// GuestGvisorDir is where gvisor bootstraps from
	GuestGvisorDir = "/tmp/gvisor"
)
