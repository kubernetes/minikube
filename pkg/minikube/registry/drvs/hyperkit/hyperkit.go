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

var (
	// minimumVersion is used by hyperkit versionCheck(whether user's hyperkit is older than this)
	minimumVersion = "0.20190201"
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

	// Split version from v0.YYYYMMDD-HH-xxxxxxx or 0.YYYYMMDD to YYYYMMDD
	currentVersion := convertVersionToDate(string(out))
	specificVersion := splitHyperKitVersion(minimumVersion)
	// If current hyperkit is not newer than minimumVersion, suggest upgrade information
	isNew, err := isNewerVersion(currentVersion, specificVersion)
	if err != nil {
		return registry.State{Installed: true, Running: true, Healthy: true, Error: fmt.Errorf("hyperkit version check failed:\n%v", err), Doc: docURL}
	}
	if !isNew {
		return registry.State{Installed: true, Running: true, Healthy: true, Error: fmt.Errorf("the installed hyperkit version (0.%s) is older than the minimum recommended version (%s)", currentVersion, minimumVersion), Fix: "Run 'brew upgrade hyperkit'", Doc: docURL}
	}

	return registry.State{Installed: true, Running: true, Healthy: true}
}

// isNewerVersion checks whether current hyperkit is newer than specific version
func isNewerVersion(currentVersion string, specificVersion string) (bool, error) {
	// Convert hyperkit version to time.Date to compare whether hyperkit version is not old.
	layout := "20060102"
	currentVersionDate, err := time.Parse(layout, currentVersion)
	if err != nil {
		return false, errors.Wrap(err, "parse date")
	}
	specificVersionDate, err := time.Parse(layout, specificVersion)
	if err != nil {
		return false, errors.Wrap(err, "parse date")
	}
	// If currentVersionDate is equal to specificVersionDate, no need to upgrade hyperkit
	if currentVersionDate.Equal(specificVersionDate) {
		return true, nil
	}
	// If currentVersionDate is after specificVersionDate, return true
	return currentVersionDate.After(specificVersionDate), nil
}

// convertVersionToDate returns current hyperkit version
// hyperkit returns version info with two patterns(v0.YYYYMMDD-HH-xxxxxxx or 0.YYYYMMDD)
// convertVersionToDate splits version to YYYYMMDD format
func convertVersionToDate(hyperKitVersionOutput string) string {
	// `hyperkit -v` returns version info with multi line.
	// Cut off like "hyperkit: v0.20190201-11-gc0dd46" or "hyperkit: 0.20190201"
	versionLine := strings.Split(hyperKitVersionOutput, "\n")[0]
	// Cut off like "v0.20190201-11-gc0dd46" or "0.20190201"
	version := strings.Split(versionLine, ": ")[1]
	// Format to "YYYYMMDD"
	return splitHyperKitVersion(version)
}

// splitHyperKitVersion splits version from v0.YYYYMMDD-HH-xxxxxxx or 0.YYYYMMDD to YYYYMMDD
func splitHyperKitVersion(version string) string {
	if strings.Contains(version, ".") {
		// Cut off like "20190201-11-gc0dd46" or "20190201"
		version = strings.Split(version, ".")[1]
	}
	if len(version) >= 8 {
		// Format to "20190201"
		version = version[:8]
	}
	return version
}
