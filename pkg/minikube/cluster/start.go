/*
Copyright 2020 The Kubernetes Authors All rights reserved.

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

package cluster

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path"
	"path/filepath"
	"time"

	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/host"
	"github.com/golang/glog"
	"github.com/juju/mutex"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/registry"
	"k8s.io/minikube/pkg/minikube/sshutil"
	"k8s.io/minikube/pkg/minikube/vmpath"
	"k8s.io/minikube/pkg/util/lock"
)

var (
	// requiredDirectories are directories to create on the host during setup
	requiredDirectories = []string{
		vmpath.GuestAddonsDir,
		vmpath.GuestManifestsDir,
		vmpath.GuestEphemeralDir,
		vmpath.GuestPersistentDir,
		vmpath.GuestCertsDir,
		path.Join(vmpath.GuestPersistentDir, "images"),
		path.Join(vmpath.GuestPersistentDir, "binaries"),
		"/tmp/gvisor",
	}
)

// StartHost starts a host VM.
func StartHost(api libmachine.API, cfg config.MachineConfig) (*host.Host, error) {
	// Prevent machine-driver boot races, as well as our own certificate race
	releaser, err := acquireMachinesLock(cfg.Name)
	if err != nil {
		return nil, errors.Wrap(err, "boot lock")
	}
	start := time.Now()
	defer func() {
		glog.Infof("releasing machines lock for %q, held for %s", cfg.Name, time.Since(start))
		releaser.Release()
	}()

	exists, err := api.Exists(cfg.Name)
	if err != nil {
		return nil, errors.Wrapf(err, "exists: %s", cfg.Name)
	}
	if !exists {
		glog.Infof("Provisioning new machine with config: %+v", cfg)
		return createHost(api, cfg)
	}
	glog.Infoln("Skipping create...Using existing machine configuration")
	return fixHost(api, cfg)
}

func engineOptions(cfg config.MachineConfig) *engine.Options {
	o := engine.Options{
		Env:              cfg.DockerEnv,
		InsecureRegistry: append([]string{constants.DefaultServiceCIDR}, cfg.InsecureRegistry...),
		RegistryMirror:   cfg.RegistryMirror,
		ArbitraryFlags:   cfg.DockerOpt,
		InstallURL:       drivers.DefaultEngineInstallURL,
	}
	return &o
}

func createHost(api libmachine.API, cfg config.MachineConfig) (*host.Host, error) {
	glog.Infof("createHost starting for %q (driver=%q)", cfg.Name, cfg.VMDriver)
	start := time.Now()
	defer func() {
		glog.Infof("createHost completed in %s", time.Since(start))
	}()

	if cfg.VMDriver == driver.VMwareFusion && viper.GetBool(config.ShowDriverDeprecationNotification) {
		out.WarningT(`The vmwarefusion driver is deprecated and support for it will be removed in a future release.
			Please consider switching to the new vmware unified driver, which is intended to replace the vmwarefusion driver.
			See https://minikube.sigs.k8s.io/docs/reference/drivers/vmware/ for more information.
			To disable this message, run [minikube config set ShowDriverDeprecationNotification false]`)
	}
	showHostInfo(cfg)
	def := registry.Driver(cfg.VMDriver)
	if def.Empty() {
		return nil, fmt.Errorf("unsupported/missing driver: %s", cfg.VMDriver)
	}
	dd := def.Config(cfg)
	data, err := json.Marshal(dd)
	if err != nil {
		return nil, errors.Wrap(err, "marshal")
	}

	h, err := api.NewHost(cfg.VMDriver, data)
	if err != nil {
		return nil, errors.Wrap(err, "new host")
	}

	h.HostOptions.AuthOptions.CertDir = localpath.MiniPath()
	h.HostOptions.AuthOptions.StorePath = localpath.MiniPath()
	h.HostOptions.EngineOptions = engineOptions(cfg)

	cstart := time.Now()
	glog.Infof("libmachine.API.Create for %q (driver=%q)", cfg.Name, cfg.VMDriver)
	if err := api.Create(h); err != nil {
		// Wait for all the logs to reach the client
		time.Sleep(2 * time.Second)
		return nil, errors.Wrap(err, "create")
	}
	glog.Infof("libmachine.API.Create for %q took %s", cfg.Name, time.Since(cstart))

	if err := postStartSetup(h, cfg); err != nil {
		return h, errors.Wrap(err, "post-start")
	}

	if err := api.Save(h); err != nil {
		return nil, errors.Wrap(err, "save")
	}
	return h, nil
}

// postStart are functions shared between startHost and fixHost
func postStartSetup(h *host.Host, mc config.MachineConfig) error {
	glog.Infof("post-start starting for %q (driver=%q)", h.Name, h.DriverName)
	start := time.Now()
	defer func() {
		glog.Infof("post-start completed in %s", time.Since(start))
	}()

	if driver.IsMock(h.DriverName) {
		return nil
	}

	glog.Infof("creating required directories: %v", requiredDirectories)
	r, err := commandRunner(h)
	if err != nil {
		return errors.Wrap(err, "command runner")
	}

	args := append([]string{"mkdir", "-p"}, requiredDirectories...)
	if _, err := r.RunCmd(exec.Command("sudo", args...)); err != nil {
		return errors.Wrapf(err, "sudo mkdir (%s)", h.DriverName)
	}

	if driver.BareMetal(mc.VMDriver) {
		showLocalOsRelease()
	}
	if driver.IsVM(mc.VMDriver) {
		logRemoteOsRelease(h.Driver)
	}
	return syncLocalAssets(r)
}

// commandRunner returns best available command runner for this host
func commandRunner(h *host.Host) (command.Runner, error) {
	d := h.Driver.DriverName()
	glog.V(1).Infof("determining appropriate runner for %q", d)
	if driver.IsMock(d) {
		glog.Infof("returning FakeCommandRunner for %q driver", d)
		return &command.FakeCommandRunner{}, nil
	}

	if driver.BareMetal(h.Driver.DriverName()) {
		glog.Infof("returning ExecRunner for %q driver", d)
		return command.NewExecRunner(), nil
	}
	if driver.IsKIC(d) {
		glog.Infof("Returning KICRunner for %q driver", d)
		return command.NewKICRunner(h.Name, "docker"), nil
	}

	glog.Infof("Creating SSH client and returning SSHRunner for %q driver", d)
	client, err := sshutil.NewSSHClient(h.Driver)
	if err != nil {
		return nil, errors.Wrap(err, "ssh client")
	}
	return command.NewSSHRunner(client), nil
}

// acquireMachinesLock protects against code that is not parallel-safe (libmachine, cert setup)
func acquireMachinesLock(name string) (mutex.Releaser, error) {
	spec := lock.PathMutexSpec(filepath.Join(localpath.MiniPath(), "machines"))
	// NOTE: Provisioning generally completes within 60 seconds
	spec.Timeout = 15 * time.Minute

	glog.Infof("acquiring machines lock for %s: %+v", name, spec)
	start := time.Now()
	r, err := mutex.Acquire(spec)
	if err == nil {
		glog.Infof("acquired machines lock for %q in %s", name, time.Since(start))
	}
	return r, err
}
