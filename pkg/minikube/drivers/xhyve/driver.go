// +build darwin

/*
Copyright 2018 The Kubernetes Authors All rights reserved.

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

package xhyve

import (
	"fmt"
	"os"

	"github.com/docker/machine/libmachine/drivers"
	cfg "k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/registry"
)

const errMsg = `
The Xhyve driver is not included in minikube yet.  Please follow the directions at
https://github.com/kubernetes/minikube/blob/master/docs/drivers.md#xhyve-driver
`

func init() {
	registry.Register(registry.DriverDef{
		Name:          "xhyve",
		Builtin:       false,
		ConfigCreator: createXhyveHost,
		DriverCreator: func() drivers.Driver {
			fmt.Fprintln(os.Stderr, errMsg)
			os.Exit(1)
			return nil
		},
	})
}

type xhyveDriver struct {
	*drivers.BaseDriver

	Boot2DockerURL string
	BootCmd        string
	CPU            int
	CaCertPath     string
	DiskSize       int64
	MacAddr        string
	Memory         int
	PrivateKeyPath string
	UUID           string
	NFSShare       bool
	DiskNumber     int
	Virtio9p       bool
	Virtio9pFolder string
	QCow2          bool
	RawDisk        bool
}

func createXhyveHost(config cfg.MachineConfig) interface{} {
	useVirtio9p := !config.DisableDriverMounts
	return &xhyveDriver{
		BaseDriver: &drivers.BaseDriver{
			MachineName: cfg.GetMachineName(),
			StorePath:   constants.GetMinipath(),
		},
		Memory:         config.Memory,
		CPU:            config.CPUs,
		Boot2DockerURL: config.Downloader.GetISOFileURI(config.MinikubeISO),
		BootCmd:        "loglevel=3 user=docker console=ttyS0 console=tty0 noembed nomodeset norestore waitusb=10 systemd.legacy_systemd_cgroup_controller=yes base host=" + cfg.GetMachineName(),
		DiskSize:       int64(config.DiskSize),
		Virtio9p:       useVirtio9p,
		Virtio9pFolder: "/Users",
		QCow2:          false,
		RawDisk:        config.XhyveDiskDriver == "virtio-blk",
	}
}
