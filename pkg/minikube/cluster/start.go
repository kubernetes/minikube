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
	"github.com/docker/machine/libmachine/state"
	"github.com/golang/glog"
	"github.com/juju/mutex"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/minikube/assets"
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
		glog.Infoln("Machine does not exist... provisioning new machine")
		glog.Infof("Provisioning machine with config: %+v", cfg)
		return createHost(api, cfg)
	}

	glog.Infoln("Skipping create...Using existing machine configuration")

	h, err := api.Load(cfg.Name)
	if err != nil {
		return nil, errors.Wrap(err, "Error loading existing host. Please try running [minikube delete], then run [minikube start] again.")
	}

	if exists && cfg.Name == constants.DefaultMachineName {
		out.T(out.Tip, "Tip: Use 'minikube start -p <name>' to create a new cluster, or 'minikube delete' to delete this one.")
	}

	s, err := h.Driver.GetState()
	glog.Infoln("Machine state: ", s)
	if err != nil {
		return nil, errors.Wrap(err, "Error getting state for host")
	}

	if s == state.Running {
		out.T(out.Running, `Using the running {{.driver_name}} "{{.profile_name}}" VM ...`, out.V{"driver_name": cfg.VMDriver, "profile_name": cfg.Name})
	} else {
		out.T(out.Restarting, `Starting existing {{.driver_name}} VM for "{{.profile_name}}" ...`, out.V{"driver_name": cfg.VMDriver, "profile_name": cfg.Name})
		if err := h.Driver.Start(); err != nil {
			return nil, errors.Wrap(err, "start")
		}
		if err := api.Save(h); err != nil {
			return nil, errors.Wrap(err, "save")
		}
	}

	e := engineOptions(cfg)
	glog.Infof("engine options: %+v", e)

	out.T(out.Waiting, "Waiting for the host to be provisioned ...")
	err = fixHost(h, e)
	if err != nil {
		return nil, err
	}
	return h, nil
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

	if err := api.Create(h); err != nil {
		// Wait for all the logs to reach the client
		time.Sleep(2 * time.Second)
		return nil, errors.Wrap(err, "create")
	}

	if err := createRequiredDirectories(h); err != nil {
		return h, errors.Wrap(err, "required directories")
	}

	for _, addon := range addon.Assets {
		var f assets.CopyableFile
		var err error
		if addon.IsTemplate() {
			f, err = addon.Evaluate(data)
			if err != nil {
				return errors.Wrapf(err, "evaluate bundled addon %s asset", addon.GetAssetName())
			}

		} else {
			f = addon
		}
		fPath := path.Join(f.GetTargetDir(), f.GetTargetName())

		if enable {
			glog.Infof("installing %s", fPath)
			if err := cmd.Copy(f); err != nil {
				return err
			}
		} else {
			glog.Infof("Removing %+v", fPath)
			defer func() {
				if err := cmd.Remove(f); err != nil {
					glog.Warningf("error removing %s; addon should still be disabled as expected", fPath)
				}
			}()
		}
		files = append(files, fPath)
	}

	if driver.BareMetal(cfg.VMDriver) {
		showLocalOsRelease()
	} else if !driver.BareMetal(cfg.VMDriver) && !driver.IsKIC(cfg.VMDriver) {
		showRemoteOsRelease(h.Driver)
		// Ensure that even new VM's have proper time synchronization up front
		// It's 2019, and I can't believe I am still dealing with time desync as a problem.
		if err := ensureSyncedGuestClock(h); err != nil {
			return h, err
		}
	} // TODO:medyagh add show-os release for kic

	if err := api.Save(h); err != nil {
		return nil, errors.Wrap(err, "save")
	}
	return h, nil
}

// createRequiredDirectories creates directories expected by minikube to exist
func createRequiredDirectories(h *host.Host) error {
	if h.DriverName == driver.Mock {
		glog.Infof("skipping createRequiredDirectories")
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
	return nil
}

// commandRunner returns best available command runner for this host
func commandRunner(h *host.Host) (command.Runner, error) {
	if h.DriverName == driver.Mock {
		glog.Errorf("commandRunner: returning unconfigured FakeCommandRunner, commands will fail!")
		return &command.FakeCommandRunner{}, nil
	}
	if driver.BareMetal(h.Driver.DriverName()) {
		return &command.ExecRunner{}, nil
	}
	if h.Driver.DriverName() == driver.Docker {
		return command.NewKICRunner(h.Name, "docker"), nil
	}
	client, err := sshutil.NewSSHClient(h.Driver)
	if err != nil {
		return nil, errors.Wrap(err, "getting ssh client for bootstrapper")
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
