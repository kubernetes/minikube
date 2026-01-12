/*
Copyright 2022 The Kubernetes Authors All rights reserved.

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

package provision

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path"
	"path/filepath"
	"text/template"
	"time"

	"k8s.io/minikube/pkg/libmachine/auth"
	"k8s.io/minikube/pkg/libmachine/drivers"
	"k8s.io/minikube/pkg/libmachine/engine"
	"k8s.io/minikube/pkg/libmachine/log"
	"k8s.io/minikube/pkg/libmachine/mcnutils"
	"k8s.io/minikube/pkg/libmachine/provision/pkgaction"
	"k8s.io/minikube/pkg/libmachine/provision/serviceaction"
	"k8s.io/minikube/pkg/libmachine/state"
	"k8s.io/minikube/pkg/libmachine/swarm"
)

func init() {
	Register("boot2docker", &RegisteredProvisioner{
		New: NewBoot2DockerProvisioner,
	})
}

func NewBoot2DockerProvisioner(d drivers.Driver) Provisioner {
	return &Boot2DockerProvisioner{
		Driver: d,
	}
}

type Boot2DockerProvisioner struct {
	OsReleaseInfo *OsRelease
	Driver        drivers.Driver
	AuthOptions   auth.Options
	EngineOptions engine.Options
	SwarmOptions  swarm.Options
}

func (provisioner *Boot2DockerProvisioner) String() string {
	return "boot2docker"
}

func (provisioner *Boot2DockerProvisioner) Service(name string, action serviceaction.ServiceAction) error {
	_, err := provisioner.SSHCommand(fmt.Sprintf("sudo /etc/init.d/%s %s", name, action.String()))
	return err
}

var (
	BaseDir = os.Getenv("MACHINE_STORAGE_PATH")
)

func GetBaseDir() string {
	if BaseDir == "" {
		BaseDir = filepath.Join(mcnutils.GetHomeDir(), ".docker", "machine")
	}
	return BaseDir
}

func (provisioner *Boot2DockerProvisioner) upgradeIso() error {
	b2dutils := mcnutils.NewB2dUtils(GetBaseDir())

	// Check if the driver has specified a custom b2d url
	jsonDriver, err := json.Marshal(provisioner.GetDriver())
	if err != nil {
		return err
	}
	var d struct {
		Boot2DockerURL string
	}
	err = json.Unmarshal(jsonDriver, &d)
	if err != nil {
		return err
	}

	log.Info("Stopping machine to do the upgrade...")

	if err := provisioner.Driver.Stop(); err != nil {
		return err
	}

	if err := mcnutils.WaitFor(drivers.MachineInState(provisioner.Driver, state.Stopped)); err != nil {
		return err
	}

	machineName := provisioner.GetDriver().GetMachineName()

	log.Infof("Upgrading machine %q...", machineName)

	// Either download the latest version of the b2d url that was explicitly
	// specified when creating the VM or copy the (updated) default ISO
	if err := b2dutils.CopyIsoToMachineDir(d.Boot2DockerURL, machineName); err != nil {
		return err
	}

	log.Infof("Starting machine back up...")

	if err := provisioner.Driver.Start(); err != nil {
		return err
	}

	return mcnutils.WaitFor(drivers.MachineInState(provisioner.Driver, state.Running))
}

func (provisioner *Boot2DockerProvisioner) Package(name string, action pkgaction.PackageAction) error {
	if name == "docker" && action == pkgaction.Upgrade {
		if err := provisioner.upgradeIso(); err != nil {
			return err
		}
	}
	return nil
}

func (provisioner *Boot2DockerProvisioner) Hostname() (string, error) {
	return provisioner.SSHCommand("hostname")
}

func (provisioner *Boot2DockerProvisioner) SetHostname(hostname string) error {
	if _, err := provisioner.SSHCommand(fmt.Sprintf(
		"sudo /usr/bin/sethostname %s && echo %q | sudo tee /var/lib/boot2docker/etc/hostname",
		hostname,
		hostname,
	)); err != nil {
		return err
	}

	return nil
}

func (provisioner *Boot2DockerProvisioner) GetDockerOptionsDir() string {
	return "/var/lib/boot2docker"
}

func (provisioner *Boot2DockerProvisioner) GetAuthOptions() auth.Options {
	return provisioner.AuthOptions
}

func (provisioner *Boot2DockerProvisioner) GetSwarmOptions() swarm.Options {
	return provisioner.SwarmOptions
}

func (provisioner *Boot2DockerProvisioner) GenerateDockerOptions(dockerPort int) (*DockerOptions, error) {
	var (
		engineCfg bytes.Buffer
	)

	driverNameLabel := fmt.Sprintf("provider=%s", provisioner.Driver.DriverName())
	provisioner.EngineOptions.Labels = append(provisioner.EngineOptions.Labels, driverNameLabel)

	engineConfigTmpl := `
EXTRA_ARGS='
{{ range .EngineOptions.Labels }}--label {{.}}
{{ end }}{{ range .EngineOptions.InsecureRegistry }}--insecure-registry {{.}}
{{ end }}{{ range .EngineOptions.RegistryMirror }}--registry-mirror {{.}}
{{ end }}{{ range .EngineOptions.ArbitraryFlags }}--{{.}}
{{ end }}
'
CACERT={{.AuthOptions.CaCertRemotePath}}
DOCKER_HOST='-H tcp://0.0.0.0:{{.DockerPort}}'
DOCKER_STORAGE={{.EngineOptions.StorageDriver}}
DOCKER_TLS=auto
SERVERKEY={{.AuthOptions.ServerKeyRemotePath}}
SERVERCERT={{.AuthOptions.ServerCertRemotePath}}

{{range .EngineOptions.Env}}export \"{{ printf "%q" . }}\"
{{end}}
`
	t, err := template.New("engineConfig").Parse(engineConfigTmpl)
	if err != nil {
		return nil, err
	}

	engineConfigContext := EngineConfigContext{
		DockerPort:    dockerPort,
		AuthOptions:   provisioner.AuthOptions,
		EngineOptions: provisioner.EngineOptions,
	}

	err = t.Execute(&engineCfg, engineConfigContext)
	if err != nil {
		return nil, err
	}

	daemonOptsDir := path.Join(provisioner.GetDockerOptionsDir(), "profile")
	return &DockerOptions{
		EngineOptions:     engineCfg.String(),
		EngineOptionsPath: daemonOptsDir,
	}, nil
}

func (provisioner *Boot2DockerProvisioner) CompatibleWithHost() bool {
	return provisioner.OsReleaseInfo.ID == "boot2docker"
}

func (provisioner *Boot2DockerProvisioner) SetOsReleaseInfo(info *OsRelease) {
	provisioner.OsReleaseInfo = info
}

func (provisioner *Boot2DockerProvisioner) GetOsReleaseInfo() (*OsRelease, error) {
	return provisioner.OsReleaseInfo, nil
}

func (provisioner *Boot2DockerProvisioner) AttemptIPContact(dockerPort int) {
	ip, err := provisioner.Driver.GetIP()
	if err != nil {
		log.Warnf("Could not get IP address for created machine: %s", err)
		return
	}

	if conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, dockerPort), 5*time.Second); err != nil {
		log.Warnf(`
This machine has been allocated an IP address, but Docker Machine could not
reach it successfully.

SSH for the machine should still work, but connecting to exposed ports, such as
the Docker daemon port (usually <ip>:%d), may not work properly.

You may need to add the route manually, or use another related workaround.

This could be due to a VPN, proxy, or host file configuration issue.

You also might want to clear any VirtualBox host only interfaces you are not using.`, engine.DefaultPort)
	} else {
		_ = conn.Close()
	}
}

func (provisioner *Boot2DockerProvisioner) Provision(swarmOptions swarm.Options, authOptions auth.Options, engineOptions engine.Options) error {
	var (
		err error
	)

	defer func() {
		if err == nil {
			provisioner.AttemptIPContact(engine.DefaultPort)
		}
	}()

	provisioner.SwarmOptions = swarmOptions
	provisioner.AuthOptions = authOptions
	provisioner.EngineOptions = engineOptions
	swarmOptions.Env = engineOptions.Env

	if provisioner.EngineOptions.StorageDriver == "" {
		provisioner.EngineOptions.StorageDriver = "overlay2"
	}

	if err = provisioner.SetHostname(provisioner.Driver.GetMachineName()); err != nil {
		return err
	}

	// b2d hosts need to wait for the daemon to be up
	// before continuing with provisioning
	if err = WaitForDocker(provisioner, engine.DefaultPort); err != nil {
		return err
	}

	if err = makeDockerOptionsDir(provisioner); err != nil {
		return err
	}

	provisioner.AuthOptions = setRemoteAuthOptions(provisioner)

	if err = ConfigureAuth(provisioner); err != nil {
		return err
	}

	err = configureSwarm(provisioner, swarmOptions, provisioner.AuthOptions)
	return err
}

func (provisioner *Boot2DockerProvisioner) SSHCommand(args string) (string, error) {
	return drivers.RunSSHCommandFromDriver(provisioner.Driver, args)
}

func (provisioner *Boot2DockerProvisioner) GetDriver() drivers.Driver {
	return provisioner.Driver
}
