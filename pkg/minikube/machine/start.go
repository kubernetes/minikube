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

package machine

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
	"k8s.io/minikube/pkg/drivers/kic/oci"
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
		vmpath.GuestKubernetesCertsDir,
		path.Join(vmpath.GuestPersistentDir, "images"),
		path.Join(vmpath.GuestPersistentDir, "binaries"),
		vmpath.GuestGvisorDir,
		vmpath.GuestCertAuthDir,
		vmpath.GuestCertStoreDir,
	}
)

// StartHost starts a host VM.
func StartHost(api libmachine.API, cfg config.ClusterConfig) (*host.Host, error) {
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

func engineOptions(cfg config.ClusterConfig) *engine.Options {
	o := engine.Options{
		Env:              cfg.DockerEnv,
		InsecureRegistry: append([]string{constants.DefaultServiceCIDR}, cfg.InsecureRegistry...),
		RegistryMirror:   cfg.RegistryMirror,
		ArbitraryFlags:   cfg.DockerOpt,
		InstallURL:       drivers.DefaultEngineInstallURL,
	}
	return &o
}

func createHost(api libmachine.API, cfg config.ClusterConfig) (*host.Host, error) {
	glog.Infof("createHost starting for %q (driver=%q)", cfg.Name, cfg.Driver)
	start := time.Now()
	defer func() {
		glog.Infof("createHost completed in %s", time.Since(start))
	}()

	if cfg.Driver == driver.VMwareFusion && viper.GetBool(config.ShowDriverDeprecationNotification) {
		out.WarningT(`The vmwarefusion driver is deprecated and support for it will be removed in a future release.
			Please consider switching to the new vmware unified driver, which is intended to replace the vmwarefusion driver.
			See https://minikube.sigs.k8s.io/docs/reference/drivers/vmware/ for more information.
			To disable this message, run [minikube config set ShowDriverDeprecationNotification false]`)
	}
	showHostInfo(cfg)
	def := registry.Driver(cfg.Driver)
	if def.Empty() {
		return nil, fmt.Errorf("unsupported/missing driver: %s", cfg.Driver)
	}
	dd, err := def.Config(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "config")
	}
	data, err := json.Marshal(dd)
	if err != nil {
		return nil, errors.Wrap(err, "marshal")
	}

	h, err := api.NewHost(cfg.Driver, data)
	if err != nil {
		return nil, errors.Wrap(err, "new host")
	}

	h.HostOptions.AuthOptions.CertDir = localpath.MiniPath()
	h.HostOptions.AuthOptions.StorePath = localpath.MiniPath()
	h.HostOptions.EngineOptions = engineOptions(cfg)

	cstart := time.Now()
	glog.Infof("libmachine.API.Create for %q (driver=%q)", cfg.Name, cfg.Driver)
	// Allow two minutes to create host before failing fast
	if err := timedCreateHost(h, api, 2*time.Minute); err != nil {
		return nil, errors.Wrap(err, "creating host")
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

func timedCreateHost(h *host.Host, api libmachine.API, t time.Duration) error {
	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(t)
		timeout <- true
	}()

	createFinished := make(chan bool, 1)
	var err error
	go func() {
		err = api.Create(h)
		createFinished <- true
	}()

	select {
	case <-createFinished:
		if err != nil {
			// Wait for all the logs to reach the client
			time.Sleep(2 * time.Second)
			return errors.Wrap(err, "create")
		}
		return nil
	case <-timeout:
		return fmt.Errorf("create host timed out in %f seconds", t.Seconds())
	}
}

// postStart are functions shared between startHost and fixHost
func postStartSetup(h *host.Host, mc config.ClusterConfig) error {
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

	if driver.BareMetal(mc.Driver) {
		showLocalOsRelease()
	}
	if driver.IsVM(mc.Driver) {
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
		return command.NewKICRunner(h.Name, d), nil
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

// showHostInfo shows host information
func showHostInfo(cfg config.ClusterConfig) {
	if driver.BareMetal(cfg.Driver) {
		info, err := getHostInfo()
		if err == nil {
			out.T(out.StartingNone, "Running on localhost (CPUs={{.number_of_cpus}}, Memory={{.memory_size}}MB, Disk={{.disk_size}}MB) ...", out.V{"number_of_cpus": info.CPUs, "memory_size": info.Memory, "disk_size": info.DiskSize})
		}
		return
	}
	if driver.IsKIC(cfg.Driver) { // TODO:medyagh add free disk space on docker machine
		s, err := oci.DaemonInfo(cfg.Driver)
		if err == nil {
			var info hostInfo
			info.CPUs = s.CPUs
			info.Memory = megs(uint64(s.TotalMemory))
			out.T(out.StartingVM, "Creating Kubernetes in {{.driver_name}} container with (CPUs={{.number_of_cpus}}) ({{.number_of_host_cpus}} available), Memory={{.memory_size}}MB ({{.host_memory_size}}MB available) ...", out.V{"driver_name": cfg.Driver, "number_of_cpus": cfg.CPUs, "number_of_host_cpus": info.CPUs, "memory_size": cfg.Memory, "host_memory_size": info.Memory})
		}
		return
	}
	out.T(out.StartingVM, "Creating {{.driver_name}} VM (CPUs={{.number_of_cpus}}, Memory={{.memory_size}}MB, Disk={{.disk_size}}MB) ...", out.V{"driver_name": cfg.Driver, "number_of_cpus": cfg.CPUs, "memory_size": cfg.Memory, "disk_size": cfg.DiskSize})
}
