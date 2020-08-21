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
	"net"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
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
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/out/register"
	"k8s.io/minikube/pkg/minikube/proxy"
	"k8s.io/minikube/pkg/minikube/registry"
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
func StartHost(api libmachine.API, cfg *config.ClusterConfig, n *config.Node) (*host.Host, bool, error) {
	machineName := driver.MachineName(*cfg, *n)

	// Prevent machine-driver boot races, as well as our own certificate race
	releaser, err := acquireMachinesLock(machineName, cfg.Driver)
	if err != nil {
		return nil, false, errors.Wrap(err, "boot lock")
	}
	start := time.Now()
	defer func() {
		glog.Infof("releasing machines lock for %q, held for %s", machineName, time.Since(start))
		releaser.Release()
	}()

	exists, err := api.Exists(machineName)
	if err != nil {
		return nil, false, errors.Wrapf(err, "exists: %s", machineName)
	}
	if !exists {
		glog.Infof("Provisioning new machine with config: %+v %+v", cfg, n)
		h, err := createHost(api, cfg, n)
		return h, exists, err
	}
	glog.Infoln("Skipping create...Using existing machine configuration")
	h, err := fixHost(api, cfg, n)
	return h, exists, err
}

// engineOptions returns docker engine options for the dockerd running inside minikube
func engineOptions(cfg config.ClusterConfig) *engine.Options {
	// get docker env from user's proxy settings
	dockerEnv := proxy.SetDockerEnv()
	// get docker env from user specifiec config
	dockerEnv = append(dockerEnv, cfg.DockerEnv...)

	// remove duplicates
	seen := map[string]bool{}
	uniqueEnvs := []string{}
	for e := range dockerEnv {
		if !seen[dockerEnv[e]] {
			seen[dockerEnv[e]] = true
			uniqueEnvs = append(uniqueEnvs, dockerEnv[e])
		}
	}

	o := engine.Options{
		Env:              uniqueEnvs,
		InsecureRegistry: append([]string{constants.DefaultServiceCIDR}, cfg.InsecureRegistry...),
		RegistryMirror:   cfg.RegistryMirror,
		ArbitraryFlags:   cfg.DockerOpt,
		InstallURL:       drivers.DefaultEngineInstallURL,
	}
	return &o
}

func createHost(api libmachine.API, cfg *config.ClusterConfig, n *config.Node) (*host.Host, error) {
	glog.Infof("createHost starting for %q (driver=%q)", n.Name, cfg.Driver)
	start := time.Now()
	defer func() {
		glog.Infof("duration metric: createHost completed in %s", time.Since(start))
	}()

	if cfg.Driver == driver.VMwareFusion && viper.GetBool(config.ShowDriverDeprecationNotification) {
		out.WarningT(`The vmwarefusion driver is deprecated and support for it will be removed in a future release.
			Please consider switching to the new vmware unified driver, which is intended to replace the vmwarefusion driver.
			See https://minikube.sigs.k8s.io/docs/reference/drivers/vmware/ for more information.
			To disable this message, run [minikube config set ShowDriverDeprecationNotification false]`)
	}
	showHostInfo(*cfg)
	def := registry.Driver(cfg.Driver)
	if def.Empty() {
		return nil, fmt.Errorf("unsupported/missing driver: %s", cfg.Driver)
	}
	dd, err := def.Config(*cfg, *n)
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
	defer postStartValidations(h, cfg.Driver)

	h.HostOptions.AuthOptions.CertDir = localpath.MiniPath()
	h.HostOptions.AuthOptions.StorePath = localpath.MiniPath()
	h.HostOptions.EngineOptions = engineOptions(*cfg)

	cstart := time.Now()
	glog.Infof("libmachine.API.Create for %q (driver=%q)", cfg.Name, cfg.Driver)

	if cfg.StartHostTimeout == 0 {
		cfg.StartHostTimeout = 6 * time.Minute
	}
	if err := timedCreateHost(h, api, cfg.StartHostTimeout); err != nil {
		return nil, errors.Wrap(err, "creating host")
	}
	glog.Infof("duration metric: libmachine.API.Create for %q took %s", cfg.Name, time.Since(cstart))

	if err := postStartSetup(h, *cfg); err != nil {
		return h, errors.Wrap(err, "post-start")
	}

	if err := saveHost(api, h, cfg, n); err != nil {
		return h, err
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

// postStartValidations are validations against the host after it is created
// TODO: Add validations for VM drivers as well, see issue #9035
func postStartValidations(h *host.Host, drvName string) {
	if !driver.IsKIC(drvName) {
		return
	}
	// make sure /var isn't full, otherwise warn
	output, err := h.RunSSHCommand("df -h /var | awk 'NR==2{print $5}'")
	if err != nil {
		glog.Warningf("error running df -h /var: %v", err)
	}
	output = strings.TrimSpace(output)
	output = strings.Trim(output, "%")
	percentageFull, err := strconv.Atoi(output)
	if err != nil {
		glog.Warningf("error getting percentage of /var that is free: %v", err)
	}
	if percentageFull >= 99 {
		exit.WithError("", fmt.Errorf("docker daemon out of memory. No space left on device"))
	}

	if percentageFull > 80 {
		out.WarningT("The docker daemon is almost out of memory, run 'docker system prune' to free up space")
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

	r, err := CommandRunner(h)
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
	if driver.IsVM(mc.Driver) || driver.IsKIC(mc.Driver) {
		logRemoteOsRelease(r)
	}
	return syncLocalAssets(r)
}

// acquireMachinesLock protects against code that is not parallel-safe (libmachine, cert setup)
func acquireMachinesLock(name string, drv string) (mutex.Releaser, error) {
	lockPath := filepath.Join(localpath.MiniPath(), "machines", drv)
	// "With KIC, it's safe to provision multiple hosts simultaneously"
	if driver.IsKIC(drv) {
		lockPath = filepath.Join(localpath.MiniPath(), "machines", drv, name)
	}
	spec := lock.PathMutexSpec(lockPath)
	// NOTE: Provisioning generally completes within 60 seconds
	// however in parallel integration testing it might take longer
	spec.Timeout = 13 * time.Minute
	if driver.IsKIC(drv) {
		spec.Timeout = 10 * time.Minute

	}

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
	machineType := driver.MachineType(cfg.Driver)
	if driver.BareMetal(cfg.Driver) {
		info, cpuErr, memErr, DiskErr := CachedHostInfo()
		if cpuErr == nil && memErr == nil && DiskErr == nil {
			register.Reg.SetStep(register.RunningLocalhost)
			out.T(out.StartingNone, "Running on localhost (CPUs={{.number_of_cpus}}, Memory={{.memory_size}}MB, Disk={{.disk_size}}MB) ...", out.V{"number_of_cpus": info.CPUs, "memory_size": info.Memory, "disk_size": info.DiskSize})
		}
		return
	}
	if driver.IsKIC(cfg.Driver) { // TODO:medyagh add free disk space on docker machine
		register.Reg.SetStep(register.CreatingContainer)
		out.T(out.StartingVM, "Creating {{.driver_name}} {{.machine_type}} (CPUs={{.number_of_cpus}}, Memory={{.memory_size}}MB) ...", out.V{"driver_name": cfg.Driver, "number_of_cpus": cfg.CPUs, "memory_size": cfg.Memory, "machine_type": machineType})
		return
	}
	register.Reg.SetStep(register.CreatingVM)
	out.T(out.StartingVM, "Creating {{.driver_name}} {{.machine_type}} (CPUs={{.number_of_cpus}}, Memory={{.memory_size}}MB, Disk={{.disk_size}}MB) ...", out.V{"driver_name": cfg.Driver, "number_of_cpus": cfg.CPUs, "memory_size": cfg.Memory, "disk_size": cfg.DiskSize, "machine_type": machineType})
}

// AddHostAlias makes fine adjustments to pod resources that aren't possible via kubeadm config.
func AddHostAlias(c command.Runner, name string, ip net.IP) error {
	record := fmt.Sprintf("%s\t%s", ip, name)
	if _, err := c.RunCmd(exec.Command("grep", record+"$", "/etc/hosts")); err == nil {
		return nil
	}

	script := fmt.Sprintf(`{ grep -v '\t%s$' /etc/hosts; echo "%s"; } > /tmp/h.$$; sudo cp /tmp/h.$$ /etc/hosts`, name, record)
	if _, err := c.RunCmd(exec.Command("/bin/bash", "-c", script)); err != nil {
		return errors.Wrap(err, "hosts update")
	}
	return nil
}
