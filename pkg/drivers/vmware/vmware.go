// +build !darwin,!linux,!windows

/*
Copyright 2017 The Kubernetes Authors All rights reserved.

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

package vmware

import "github.com/docker/machine/libmachine/drivers"

func NewDriver(hostName, storePath string) drivers.Driver {
	return drivers.NewDriverNotSupported("vmware", hostName, storePath)
}

func DhcpConfigFiles() string {
	return ""
}

func DhcpLeaseFiles() string {
	return ""
}

func SetUmask() {
}

func setVmwareCmd(cmd string) string {
	return ""
}

func getShareDriveAndName() (string, string, string) {
	return "", "", ""
}
