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
	"os"

	"github.com/docker/machine/drivers/hyperv"
	"github.com/docker/machine/drivers/virtualbox"
	"github.com/docker/machine/drivers/vmwarefusion"
	"github.com/docker/machine/libmachine/drivers/plugin"
	"github.com/docker/machine/libmachine/drivers/plugin/localbinary"
	"github.com/golang/glog"
)

// StartDriver starts the desired machine driver if necessary.
func StartDriver() {
	if os.Getenv(localbinary.PluginEnvKey) == localbinary.PluginEnvVal {
		driverName := os.Getenv(localbinary.PluginEnvDriverName)
		switch driverName {
		case "virtualbox":
			plugin.RegisterDriver(virtualbox.NewDriver("", ""))
		case "vmwarefusion":
			plugin.RegisterDriver(vmwarefusion.NewDriver("", ""))
		case "hyperv":
			plugin.RegisterDriver(hyperv.NewDriver("", ""))
		default:
			glog.Exitf("Unsupported driver: %s\n", driverName)
		}
		return
	}
	localbinary.CurrentBinaryIsDockerMachine = true
}
