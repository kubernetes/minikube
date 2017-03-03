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

package machine

import (
	"github.com/docker/machine/drivers/virtualbox"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/drivers/plugin"
	"github.com/golang/glog"
	"github.com/pkg/errors"
)

var driverMap = map[string]driverGetter{
	"kvm":        getKVMDriver,
	"virtualbox": getVirtualboxDriver,
}

func getKVMDriver(rawDriver []byte) (drivers.Driver, error) {
	return nil, errors.New(`
The KVM driver is not included in minikube yet.  Please follow the direction at
https://github.com/kubernetes/minikube/blob/master/DRIVERS.md#kvm-driver
`)
}

// StartDriver starts the desired machine driver if necessary.
func registerDriver(driverName string) {
	switch driverName {
	case "virtualbox":
		plugin.RegisterDriver(virtualbox.NewDriver("", ""))
	default:
		glog.Exitf("Unsupported driver: %s\n", driverName)
	}
}
