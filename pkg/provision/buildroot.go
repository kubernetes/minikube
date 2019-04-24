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

package provision

import (
	"bytes"
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/cert"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/docker/machine/libmachine/provision"
	"github.com/docker/machine/libmachine/provision/pkgaction"
	"github.com/docker/machine/libmachine/provision/serviceaction"
	"github.com/docker/machine/libmachine/swarm"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/bootstrapper"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/sshutil"
	"k8s.io/minikube/pkg/util"
)

// BuildrootProvisioner provisions the custom system based on Buildroot
type BuildrootProvisioner struct {
	provision.SystemdProvisioner
}

// for escaping systemd template specifiers (e.g. '%i'), which are not supported by minikube
var systemdSpecifierEscaper = strings.NewReplacer("%", "%%")

func init() {
	provision.Register("Buildroot", &provision.RegisteredProvisioner{
		New: NewBuildrootProvisioner,
	})
}

// NewBuildrootProvisioner creates a new BuildrootProvisioner
func NewBuildrootProvisioner(d drivers.Driver) provision.Provisioner {
	return &BuildrootProvisioner{
		provision.NewSystemdProvisioner("buildroot", d),
	}
}

func (p *BuildrootProvisioner) String() string {
	return "buildroot"
}

// escapeSystemdDirectives escapes special characters in the input variables used to create the
// systemd unit file, which would otherwise be interpreted as systemd directives. An example
// are template specifiers (e.g. '%i') which are predefined variables that get evaluated dynamically
// (see systemd man pages for more info). This is not supported by minikube, thus needs to be escaped.
func escapeSystemdDirectives(engineConfigContext *provision.EngineConfigContext) {
	// escape '%' in Environment option so that it does not evaluate into a template specifier
	engineConfigContext.EngineOptions.Env = util.ReplaceChars(engineConfigContext.EngineOptions.Env, systemdSpecifierEscaper)
	// input might contain whitespaces, wrap it in quotes
	engineConfigContext.EngineOptions.Env = util.ConcatStrings(engineConfigContext.EngineOptions.Env, "\"", "\"")
}

// GenerateDockerOptions generates the *provision.DockerOptions for this provisioner
func (p *BuildrootProvisioner) GenerateDockerOptions(dockerPort int) (*provision.DockerOptions, error) {
	var engineCfg bytes.Buffer

	driverNameLabel := fmt.Sprintf("provider=%s", p.Driver.DriverName())
	p.EngineOptions.Labels = append(p.EngineOptions.Labels, driverNameLabel)

	engineConfigTmpl := `[Unit]
Description=Docker Application Container Engine
Documentation=https://docs.docker.com
After=network.target  minikube-automount.service docker.socket
Requires= minikube-automount.service docker.socket 

[Service]
Type=notify

# DOCKER_RAMDISK disables pivot_root in Docker, using MS_MOVE instead.
Environment=DOCKER_RAMDISK=yes
{{range .EngineOptions.Env}}Environment={{.}}
{{end}}

# This file is a systemd drop-in unit that inherits from the base dockerd configuration.
# The base configuration already specifies an 'ExecStart=...' command. The first directive
# here is to clear out that command inherited from the base configuration. Without this,
# the command from the base configuration and the command specified here are treated as
# a sequence of commands, which is not the desired behavior, nor is it valid -- systemd
# will catch this invalid input and refuse to start the service with an error like:
#  Service has more than one ExecStart= setting, which is only allowed for Type=oneshot services.
ExecStart=
ExecStart=/usr/bin/dockerd -H tcp://0.0.0.0:{{.DockerPort}} -H unix:///var/run/docker.sock --tlsverify --tlscacert {{.AuthOptions.CaCertRemotePath}} --tlscert {{.AuthOptions.ServerCertRemotePath}} --tlskey {{.AuthOptions.ServerKeyRemotePath}} {{ range .EngineOptions.Labels }}--label {{.}} {{ end }}{{ range .EngineOptions.InsecureRegistry }}--insecure-registry {{.}} {{ end }}{{ range .EngineOptions.RegistryMirror }}--registry-mirror {{.}} {{ end }}{{ range .EngineOptions.ArbitraryFlags }}--{{.}} {{ end }}
ExecReload=/bin/kill -s HUP $MAINPID

# Having non-zero Limit*s causes performance problems due to accounting overhead
# in the kernel. We recommend using cgroups to do container-local accounting.
LimitNOFILE=infinity
LimitNPROC=infinity
LimitCORE=infinity

# Uncomment TasksMax if your systemd version supports it.
# Only systemd 226 and above support this version.
TasksMax=infinity
TimeoutStartSec=0

# set delegate yes so that systemd does not reset the cgroups of docker containers
Delegate=yes

# kill only the docker process, not all processes in the cgroup
KillMode=process

[Install]
WantedBy=multi-user.target
`
	t, err := template.New("engineConfig").Parse(engineConfigTmpl)
	if err != nil {
		return nil, err
	}

	engineConfigContext := provision.EngineConfigContext{
		DockerPort:    dockerPort,
		AuthOptions:   p.AuthOptions,
		EngineOptions: p.EngineOptions,
	}

	escapeSystemdDirectives(&engineConfigContext)

	if err := t.Execute(&engineCfg, engineConfigContext); err != nil {
		return nil, err
	}

	return &provision.DockerOptions{
		EngineOptions:     engineCfg.String(),
		EngineOptionsPath: "/lib/systemd/system/docker.service",
	}, nil
}

// Package installs a package
func (p *BuildrootProvisioner) Package(name string, action pkgaction.PackageAction) error {
	return nil
}

// Provision does the provisioning
func (p *BuildrootProvisioner) Provision(swarmOptions swarm.Options, authOptions auth.Options, engineOptions engine.Options) error {
	p.SwarmOptions = swarmOptions
	p.AuthOptions = authOptions
	p.EngineOptions = engineOptions

	log.Debugf("setting hostname %q", p.Driver.GetMachineName())
	if err := p.SetHostname(p.Driver.GetMachineName()); err != nil {
		return err
	}

	p.AuthOptions = setRemoteAuthOptions(p)
	log.Debugf("set auth options %+v", p.AuthOptions)

	log.Debugf("setting up certificates")
	configureAuth := func() error {
		if err := configureAuth(p); err != nil {
			return &util.RetriableError{Err: err}
		}
		return nil
	}
	err := util.RetryAfter(5, configureAuth, time.Second*10)
	if err != nil {
		log.Debugf("Error configuring auth during provisioning %v", err)
		return err
	}

	log.Debugf("setting minikube options for container-runtime")
	if err := setMinikubeOptions(p); err != nil {
		log.Debugf("Error setting container-runtime options during provisioning %v", err)
		return err
	}

	return nil
}

func setRemoteAuthOptions(p provision.Provisioner) auth.Options {
	dockerDir := p.GetDockerOptionsDir()
	authOptions := p.GetAuthOptions()

	// due to windows clients, we cannot use filepath.Join as the paths
	// will be mucked on the linux hosts
	authOptions.CaCertRemotePath = path.Join(dockerDir, "ca.pem")
	authOptions.ServerCertRemotePath = path.Join(dockerDir, "server.pem")
	authOptions.ServerKeyRemotePath = path.Join(dockerDir, "server-key.pem")

	return authOptions
}

func setMinikubeOptions(p *BuildrootProvisioner) error {
	// pass through --insecure-registry
	var (
		crioOptsTmpl = `
CRIO_MINIKUBE_OPTIONS='{{ range .EngineOptions.InsecureRegistry }}--insecure-registry {{.}} {{ end }}'
`
		crioOptsPath = "/etc/sysconfig/crio.minikube"
	)
	t, err := template.New("crioOpts").Parse(crioOptsTmpl)
	if err != nil {
		return err
	}
	var crioOptsBuf bytes.Buffer
	if err := t.Execute(&crioOptsBuf, p); err != nil {
		return err
	}

	if _, err = p.SSHCommand(fmt.Sprintf("sudo mkdir -p %s && printf %%s \"%s\" | sudo tee %s", path.Dir(crioOptsPath), crioOptsBuf.String(), crioOptsPath)); err != nil {
		return err
	}

	// This is unlikely to cause issues unless the user has explicitly requested CRIO, so just log a warning.
	if err := p.Service("crio", serviceaction.Restart); err != nil {
		log.Warn("Unable to restart crio service. Error: %v", err)
	}

	return nil
}

func configureAuth(p *BuildrootProvisioner) error {
	driver := p.GetDriver()
	machineName := driver.GetMachineName()
	authOptions := p.GetAuthOptions()
	org := mcnutils.GetUsername() + "." + machineName
	bits := 2048

	ip, err := driver.GetIP()
	if err != nil {
		return errors.Wrap(err, "error getting ip during provisioning")
	}

	err = copyHostCerts(authOptions)
	if err != nil {
		return err
	}

	// The Host IP is always added to the certificate's SANs list
	hosts := append(authOptions.ServerCertSANs, ip, "localhost")
	log.Debugf("generating server cert: %s ca-key=%s private-key=%s org=%s san=%s",
		authOptions.ServerCertPath,
		authOptions.CaCertPath,
		authOptions.CaPrivateKeyPath,
		org,
		hosts,
	)

	err = cert.GenerateCert(&cert.Options{
		Hosts:     hosts,
		CertFile:  authOptions.ServerCertPath,
		KeyFile:   authOptions.ServerKeyPath,
		CAFile:    authOptions.CaCertPath,
		CAKeyFile: authOptions.CaPrivateKeyPath,
		Org:       org,
		Bits:      bits,
	})

	if err != nil {
		return fmt.Errorf("error generating server cert: %v", err)
	}

	err = copyRemoteCerts(authOptions, driver)
	if err != nil {
		return err
	}

	config, err := config.Load()
	if err != nil {
		return errors.Wrap(err, "getting cluster config")
	}

	dockerCfg, err := p.GenerateDockerOptions(engine.DefaultPort)
	if err != nil {
		return errors.Wrap(err, "generating docker options")
	}

	log.Info("Setting Docker configuration on the remote daemon...")

	if _, err = p.SSHCommand(fmt.Sprintf("sudo mkdir -p %s && printf %%s \"%s\" | sudo tee %s", path.Dir(dockerCfg.EngineOptionsPath), dockerCfg.EngineOptions, dockerCfg.EngineOptionsPath)); err != nil {
		return err
	}

	if config.MachineConfig.ContainerRuntime == "" {

		if err := p.Service("docker", serviceaction.Enable); err != nil {
			return err
		}

		if err := p.Service("docker", serviceaction.Restart); err != nil {
			return err
		}
	}

	return nil
}

func copyHostCerts(authOptions auth.Options) error {
	execRunner := &bootstrapper.ExecRunner{}
	hostCerts := map[string]string{
		authOptions.CaCertPath:     path.Join(authOptions.StorePath, "ca.pem"),
		authOptions.ClientCertPath: path.Join(authOptions.StorePath, "cert.pem"),
		authOptions.ClientKeyPath:  path.Join(authOptions.StorePath, "key.pem"),
	}

	for src, dst := range hostCerts {
		f, err := assets.NewFileAsset(src, path.Dir(dst), filepath.Base(dst), "0777")
		if err != nil {
			return errors.Wrapf(err, "open cert file: %s", src)
		}
		if err := execRunner.Copy(f); err != nil {
			return errors.Wrapf(err, "transferring file: %+v", f)
		}
	}

	return nil
}

func copyRemoteCerts(authOptions auth.Options, driver drivers.Driver) error {
	remoteCerts := map[string]string{
		authOptions.CaCertPath:     authOptions.CaCertRemotePath,
		authOptions.ServerCertPath: authOptions.ServerCertRemotePath,
		authOptions.ServerKeyPath:  authOptions.ServerKeyRemotePath,
	}

	sshClient, err := sshutil.NewSSHClient(driver)
	if err != nil {
		return errors.Wrap(err, "provisioning: error getting ssh client")
	}
	sshRunner := bootstrapper.NewSSHRunner(sshClient)
	for src, dst := range remoteCerts {
		f, err := assets.NewFileAsset(src, path.Dir(dst), filepath.Base(dst), "0640")
		if err != nil {
			return errors.Wrapf(err, "error copying %s to %s", src, dst)
		}
		if err := sshRunner.Copy(f); err != nil {
			return errors.Wrapf(err, "transferring file to machine %v", f)
		}
	}

	return nil
}
