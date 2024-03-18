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
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/blang/semver/v4"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/host"
	"github.com/juju/mutex/v2"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/out/register"
	"k8s.io/minikube/pkg/minikube/proxy"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/minikube/registry"
	"k8s.io/minikube/pkg/minikube/style"
	"k8s.io/minikube/pkg/minikube/vmpath"
	"k8s.io/minikube/pkg/util"
	"k8s.io/minikube/pkg/util/lock"
)

// requiredDirectories are directories to create on the host during setup
var requiredDirectories = []string{
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

// StartHost starts a host VM.
func StartHost(api libmachine.API, cfg *config.ClusterConfig, n *config.Node) (*host.Host, bool, error) {
	machineName := config.MachineName(*cfg, *n)

	// Prevent machine-driver boot races, as well as our own certificate race
	releaser, err := acquireMachinesLock(machineName, cfg.Driver)
	if err != nil {
		return nil, false, errors.Wrap(err, "boot lock")
	}
	start := time.Now()
	defer func() {
		klog.Infof("releasing machines lock for %q, held for %s", machineName, time.Since(start))
		releaser.Release()
	}()

	exists, err := api.Exists(machineName)
	if err != nil {
		return nil, false, errors.Wrapf(err, "exists: %s", machineName)
	}
	var h *host.Host
	if !exists {
		klog.Infof("Provisioning new machine with config: %+v %+v", cfg, n)
		h, err = createHost(api, cfg, n)
	} else {
		klog.Infoln("Skipping create...Using existing machine configuration")
		h, err = fixHost(api, cfg, n)
	}
	if err != nil {
		return h, exists, err
	}
	return h, exists, ensureSyncedGuestClock(h, cfg.Driver)
}

// engineOptions returns docker engine options for the dockerd running inside minikube
func engineOptions(cfg config.ClusterConfig) *engine.Options {
	// get docker env from user's proxy settings
	dockerEnv := proxy.SetDockerEnv()
	// get docker env from user specifiec config
	dockerEnv = append(dockerEnv, cfg.DockerEnv...)

	uniqueEnvs := util.RemoveDuplicateStrings(dockerEnv)

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
	klog.Infof("createHost starting for %q (driver=%q)", n.Name, cfg.Driver)
	start := time.Now()
	defer func() {
		klog.Infof("duration metric: took %s to createHost", time.Since(start))
	}()

	if cfg.Driver != driver.SSH {
		showHostInfo(nil, *cfg)
	}

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
	klog.Infof("libmachine.API.Create for %q (driver=%q)", cfg.Name, cfg.Driver)

	if cfg.StartHostTimeout == 0 {
		cfg.StartHostTimeout = 6 * time.Minute
	}
	if err := timedCreateHost(h, api, cfg.StartHostTimeout); err != nil {
		return nil, errors.Wrap(err, "creating host")
	}
	klog.Infof("duration metric: took %s to libmachine.API.Create %q", time.Since(cstart), cfg.Name)
	if cfg.Driver == driver.SSH {
		showHostInfo(h, *cfg)
	}

	if err := postStartSetup(h, *cfg); err != nil {
		return h, errors.Wrap(err, "post-start")
	}

	if err := saveHost(api, h, cfg, n); err != nil {
		return h, err
	}
	return h, nil
}

func timedCreateHost(h *host.Host, api libmachine.API, t time.Duration) error {
	create := make(chan error, 1)
	go func() {
		defer close(create)
		create <- api.Create(h)
	}()

	select {
	case err := <-create:
		if err != nil {
			// Wait for all the logs to reach the client
			time.Sleep(2 * time.Second)
			return errors.Wrap(err, "create")
		}
		return nil
	case <-time.After(t):
		return fmt.Errorf("create host timed out in %f seconds", t.Seconds())
	}
}

// postStartValidations are validations against the host after it is created
// TODO: Add validations for VM drivers as well, see issue #9035
func postStartValidations(h *host.Host, drvName string) {
	if !driver.IsKIC(drvName) {
		return
	}
	r, err := CommandRunner(h)
	if err != nil {
		klog.Warningf("error getting command runner: %v", err)
	}

	var kind reason.Kind
	var name string
	if drvName == oci.Docker {
		kind = reason.RsrcInsufficientDockerStorage
		name = "Docker"
	}
	if drvName == oci.Podman {
		kind = reason.RsrcInsufficientPodmanStorage
		name = "Podman"
	}
	if name == "" {
		klog.Warningf("unknown KIC driver: %v", drvName)
		return
	}

	if viper.GetBool("force") {
		return
	}

	// make sure /var isn't full,  as pod deployments will fail if it is
	percentageFull, err := DiskUsed(r, "/var")
	if err != nil {
		klog.Warningf("error getting percentage of /var that is free: %v", err)
	}

	availableGiB, err := DiskAvailable(r, "/var")
	if err != nil {
		klog.Warningf("error getting GiB of /var that is available: %v", err)
	}
	const thresholdGiB = 20

	if percentageFull >= 99 && availableGiB < thresholdGiB {
		exit.Message(
			kind,
			`{{.n}} is out of disk space! (/var is at {{.p}}% of capacity). You can pass '--force' to skip this check.`,
			out.V{"n": name, "p": percentageFull},
		)
	}

	if percentageFull >= 85 && availableGiB < thresholdGiB {
		out.WarnReason(
			kind,
			`{{.n}} is nearly out of disk space, which may cause deployments to fail! ({{.p}}% of capacity). You can pass '--force' to skip this check.`,
			out.V{"n": name, "p": percentageFull},
		)
	}
}

// DiskUsed returns the capacity of dir in the VM/container as a percentage
func DiskUsed(cr command.Runner, dir string) (int, error) {
	if s := os.Getenv(constants.TestDiskUsedEnv); s != "" {
		return strconv.Atoi(s)
	}
	output, err := cr.RunCmd(exec.Command("sh", "-c", fmt.Sprintf("df -h %s | awk 'NR==2{print $5}'", dir)))
	if err != nil {
		klog.Warningf("error running df -h /var: %v\n%v", err, output.Output())
		return 0, err
	}
	percentage := strings.TrimSpace(output.Stdout.String())
	percentage = strings.Trim(percentage, "%")
	return strconv.Atoi(percentage)
}

// DiskAvailable returns the available capacity of dir in the VM/container in GiB
func DiskAvailable(cr command.Runner, dir string) (int, error) {
	if s := os.Getenv(constants.TestDiskAvailableEnv); s != "" {
		return strconv.Atoi(s)
	}
	output, err := cr.RunCmd(exec.Command("sh", "-c", fmt.Sprintf("df -BG %s | awk 'NR==2{print $4}'", dir)))
	if err != nil {
		klog.Warningf("error running df -BG /var: %v\n%v", err, output.Output())
		return 0, err
	}
	gib := strings.TrimSpace(output.Stdout.String())
	gib = strings.Trim(gib, "G")
	return strconv.Atoi(gib)
}

// postStartSetup are functions shared between startHost and fixHost
func postStartSetup(h *host.Host, mc config.ClusterConfig) error {
	klog.Infof("postStartSetup for %q (driver=%q)", h.Name, h.DriverName)
	start := time.Now()
	defer func() {
		klog.Infof("duration metric: took %s for postStartSetup", time.Since(start))
	}()

	if driver.IsMock(h.DriverName) {
		return nil
	}

	k8sVer, err := semver.ParseTolerant(mc.KubernetesConfig.KubernetesVersion)
	if err != nil {
		klog.Errorf("unable to parse Kubernetes version: %s", mc.KubernetesConfig.KubernetesVersion)
		return err
	}
	if driver.IsNone(h.DriverName) && k8sVer.GTE(semver.Version{Major: 1, Minor: 24}) {
		if mc.KubernetesConfig.ContainerRuntime == constants.Docker {
			if _, err := exec.LookPath("cri-dockerd"); err != nil {
				exit.Message(reason.NotFoundCriDockerd, "\n\n")
			}
			if _, err := exec.LookPath("dockerd"); err != nil {
				exit.Message(reason.NotFoundDockerd, "\n\n")
			}
		}
		if _, err := os.Stat("/opt/cni/bin"); err != nil {
			exit.Message(reason.NotFoundCNIPlugins, "\n\n")
		}
	}

	klog.Infof("creating required directories: %v", requiredDirectories)

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

	if driver.IsVM(mc.Driver) || driver.IsKIC(mc.Driver) || driver.IsSSH(mc.Driver) {
		logRemoteOsRelease(r)
	}

	return syncLocalAssets(r)
}

// acquireMachinesLock protects against code that is not parallel-safe (libmachine, cert setup)
func acquireMachinesLock(name string, drv string) (mutex.Releaser, error) {
	lockPath := filepath.Join(localpath.MiniPath(), "machines", drv)
	// With KIC, it's safe to provision multiple hosts simultaneously
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

	klog.Infof("acquireMachinesLock for %s: %+v", name, spec)
	start := time.Now()
	r, err := mutex.Acquire(spec)
	if err == nil {
		klog.Infof("duration metric: took %s to acquireMachinesLock for %q", time.Since(start), name)
	}
	return r, err
}

// showHostInfo shows host information
func showHostInfo(h *host.Host, cfg config.ClusterConfig) {
	machineType := driver.MachineType(cfg.Driver)
	if driver.BareMetal(cfg.Driver) {
		info, cpuErr, memErr, DiskErr := LocalHostInfo()
		if cpuErr == nil && memErr == nil && DiskErr == nil {
			register.Reg.SetStep(register.RunningLocalhost)
			out.Step(style.StartingNone, "Running on localhost (CPUs={{.number_of_cpus}}, Memory={{.memory_size}}MB, Disk={{.disk_size}}MB) ...", out.V{"number_of_cpus": info.CPUs, "memory_size": info.Memory, "disk_size": info.DiskSize})
		}
		return
	}
	if driver.IsSSH(cfg.Driver) {
		r, err := CommandRunner(h)
		if err != nil {
			klog.Warningf("error getting command runner: %v", err)
			return
		}
		info, cpuErr, memErr, DiskErr := RemoteHostInfo(r)
		if cpuErr == nil && memErr == nil && DiskErr == nil {
			register.Reg.SetStep(register.RunningRemotely)
			out.Step(style.StartingSSH, "Running remotely (CPUs={{.number_of_cpus}}, Memory={{.memory_size}}MB, Disk={{.disk_size}}MB) ...", out.V{"number_of_cpus": info.CPUs, "memory_size": info.Memory, "disk_size": info.DiskSize})
		}
		return
	}
	if driver.IsKIC(cfg.Driver) { // TODO:medyagh add free disk space on docker machine
		register.Reg.SetStep(register.CreatingContainer)
		out.Step(style.StartingVM, "Creating {{.driver_name}} {{.machine_type}} (CPUs={{if not .number_of_cpus}}no-limit{{else}}{{.number_of_cpus}}{{end}}, Memory={{if not .memory_size}}no-limit{{else}}{{.memory_size}}MB{{end}}) ...", out.V{"driver_name": cfg.Driver, "number_of_cpus": cfg.CPUs, "memory_size": cfg.Memory, "machine_type": machineType})
		return
	}
	register.Reg.SetStep(register.CreatingVM)
	out.Step(style.StartingVM, "Creating {{.driver_name}} {{.machine_type}} (CPUs={{.number_of_cpus}}, Memory={{.memory_size}}MB, Disk={{.disk_size}}MB) ...", out.V{"driver_name": cfg.Driver, "number_of_cpus": cfg.CPUs, "memory_size": cfg.Memory, "disk_size": cfg.DiskSize, "machine_type": machineType})
}

// AddHostAlias makes fine adjustments to pod resources that aren't possible via kubeadm config.
func AddHostAlias(c command.Runner, name string, ip net.IP) error {
	record := fmt.Sprintf("%s\t%s", ip, name)
	if _, err := c.RunCmd(exec.Command("grep", record+"$", "/etc/hosts")); err == nil {
		return nil
	}

	if _, err := c.RunCmd(addHostAliasCommand(name, record, true, "/etc/hosts")); err != nil {
		return errors.Wrap(err, "hosts update")
	}
	return nil
}

func addHostAliasCommand(name string, record string, sudo bool, path string) *exec.Cmd {
	sudoCmd := "sudo"
	if !sudo { // for testing
		sudoCmd = ""
	}

	script := fmt.Sprintf(
		`{ grep -v $'\t%s$' "%s"; echo "%s"; } > /tmp/h.$$; %s cp /tmp/h.$$ "%s"`,
		name,
		path,
		record,
		sudoCmd,
		path)
	return exec.Command("/bin/bash", "-c", script)
}
