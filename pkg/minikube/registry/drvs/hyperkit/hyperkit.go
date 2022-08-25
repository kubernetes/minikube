//go:build darwin

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

package hyperkit

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/pborman/uuid"

	"k8s.io/minikube/pkg/drivers/hyperkit"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/download"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/registry"
)

const (
	docURL = "https://minikube.sigs.k8s.io/docs/reference/drivers/hyperkit/"
)

func init() {
	if err := registry.Register(registry.DriverDef{
		Name:     driver.HyperKit,
		Config:   configure,
		Status:   status,
		Default:  true,
		Priority: registry.Preferred,
	}); err != nil {
		panic(fmt.Sprintf("register: %v", err))
	}
}

func configure(cfg config.ClusterConfig, n config.Node) (interface{}, error) {
	u := cfg.UUID
	if u == "" {
		u = uuid.NewUUID().String()
	}

	return &hyperkit.Driver{
		BaseDriver: &drivers.BaseDriver{
			MachineName: config.MachineName(cfg, n),
			StorePath:   localpath.MiniPath(),
			SSHUser:     "docker",
		},
		Boot2DockerURL: download.LocalISOResource(cfg.MinikubeISO),
		DiskSize:       cfg.DiskSize,
		Memory:         cfg.Memory,
		CPU:            cfg.CPUs,
		NFSShares:      cfg.NFSShare,
		NFSSharesRoot:  cfg.NFSSharesRoot,
		UUID:           u,
		VpnKitSock:     cfg.HyperkitVpnKitSock,
		VSockPorts:     cfg.HyperkitVSockPorts,
		Cmdline:        "loglevel=3 console=ttyS0 console=tty0 noembed nomodeset norestore waitusb=10 systemd.legacy_systemd_cgroup_controller=yes random.trust_cpu=on hw_rng_model=virtio base host=" + cfg.Name,
		ExtraDisks:     cfg.ExtraDisks,
	}, nil
}

func status() registry.State {
	path, err := exec.LookPath("hyperkit")
	if err != nil {
		return registry.State{Error: err, Fix: "Run 'brew install hyperkit'", Doc: docURL}
	}

	// Allow no more than 2 seconds for querying state
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, path, "-v")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return registry.State{Installed: true, Running: false, Error: fmt.Errorf("%s failed:\n%s", strings.Join(cmd.Args, " "), out), Fix: "Run 'brew install hyperkit'", Doc: docURL}
	}

	return registry.State{Installed: true, Running: true, Healthy: true}
}
