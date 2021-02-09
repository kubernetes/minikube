/*
Copyright 2019 The Kubernetes Authors All rights reserved.

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
	"text/template"
	"time"

	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/provision"
	"github.com/docker/machine/libmachine/provision/pkgaction"
	"github.com/docker/machine/libmachine/swarm"
	"github.com/spf13/viper"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/util/retry"
)

// UbuntuProvisioner provisions the ubuntu
type UbuntuProvisioner struct {
	BuildrootProvisioner
}

// NewUbuntuProvisioner creates a new UbuntuProvisioner
func NewUbuntuProvisioner(d drivers.Driver) provision.Provisioner {
	return &UbuntuProvisioner{
		BuildrootProvisioner{
			NewSystemdProvisioner("ubuntu", d),
			viper.GetString(config.ProfileName),
		},
	}
}

func (p *UbuntuProvisioner) String() string {
	return "ubuntu"
}

// CompatibleWithHost checks if provisioner is compatible with host
func (p *UbuntuProvisioner) CompatibleWithHost() bool {
	return p.OsReleaseInfo.ID == "ubuntu"
}

// GenerateDockerOptions generates the *provision.DockerOptions for this provisioner
func (p *UbuntuProvisioner) GenerateDockerOptions(dockerPort int) (*provision.DockerOptions, error) {
	var engineCfg bytes.Buffer

	drvLabel := fmt.Sprintf("provider=%s", p.Driver.DriverName())
	p.EngineOptions.Labels = append(p.EngineOptions.Labels, drvLabel)

	noPivot := true
	// Using pivot_root is not supported on fstype rootfs
	if fstype, err := rootFileSystemType(p); err == nil {
		klog.Infof("root file system type: %s", fstype)
		noPivot = fstype == "rootfs"
	}

	engineConfigTmpl := `[Unit]
Description=Docker Application Container Engine
Documentation=https://docs.docker.com
BindsTo=containerd.service
After=network-online.target firewalld.service containerd.service
Wants=network-online.target
Requires=docker.socket
StartLimitBurst=3
StartLimitIntervalSec=60

[Service]
Type=notify
Restart=on-failure
`
	if noPivot {
		klog.Warning("Using fundamentally insecure --no-pivot option")
		engineConfigTmpl += `
# DOCKER_RAMDISK disables pivot_root in Docker, using MS_MOVE instead.
Environment=DOCKER_RAMDISK=yes
`
	}
	engineConfigTmpl += `
{{range .EngineOptions.Env}}Environment={{.}}
{{end}}

# This file is a systemd drop-in unit that inherits from the base dockerd configuration.
# The base configuration already specifies an 'ExecStart=...' command. The first directive
# here is to clear out that command inherited from the base configuration. Without this,
# the command from the base configuration and the command specified here are treated as
# a sequence of commands, which is not the desired behavior, nor is it valid -- systemd
# will catch this invalid input and refuse to start the service with an error like:
#  Service has more than one ExecStart= setting, which is only allowed for Type=oneshot services.

# NOTE: default-ulimit=nofile is set to an arbitrary number for consistency with other
# container runtimes. If left unlimited, it may result in OOM issues with MySQL.
ExecStart=
ExecStart=/usr/bin/dockerd -H tcp://0.0.0.0:2376 -H unix:///var/run/docker.sock --default-ulimit=nofile=1048576:1048576 --tlsverify --tlscacert {{.AuthOptions.CaCertRemotePath}} --tlscert {{.AuthOptions.ServerCertRemotePath}} --tlskey {{.AuthOptions.ServerKeyRemotePath}} {{ range .EngineOptions.Labels }}--label {{.}} {{ end }}{{ range .EngineOptions.InsecureRegistry }}--insecure-registry {{.}} {{ end }}{{ range .EngineOptions.RegistryMirror }}--registry-mirror {{.}} {{ end }}{{ range .EngineOptions.ArbitraryFlags }}--{{.}} {{ end }}
ExecReload=/bin/kill -s HUP \$MAINPID

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

	do := &provision.DockerOptions{
		EngineOptions:     engineCfg.String(),
		EngineOptionsPath: "/lib/systemd/system/docker.service",
	}
	return do, updateUnit(p, "docker", do.EngineOptions, do.EngineOptionsPath)
}

// Package installs a package
func (p *UbuntuProvisioner) Package(name string, action pkgaction.PackageAction) error {
	return nil
}

// Provision does the provisioning
func (p *UbuntuProvisioner) Provision(swarmOptions swarm.Options, authOptions auth.Options, engineOptions engine.Options) error {
	p.SwarmOptions = swarmOptions
	p.AuthOptions = authOptions
	p.EngineOptions = engineOptions

	klog.Infof("provisioning hostname %q", p.Driver.GetMachineName())
	if err := p.SetHostname(p.Driver.GetMachineName()); err != nil {
		return err
	}

	p.AuthOptions = setRemoteAuthOptions(p)
	klog.Infof("set auth options %+v", p.AuthOptions)

	klog.Infof("setting up certificates")
	configAuth := func() error {
		if err := configureAuth(p); err != nil {
			klog.Warningf("configureAuth failed: %v", err)
			return &retry.RetriableError{Err: err}
		}
		return nil
	}

	err := retry.Expo(configAuth, 100*time.Microsecond, 2*time.Minute)

	if err != nil {
		klog.Infof("Error configuring auth during provisioning %v", err)
		return err
	}

	klog.Infof("setting minikube options for container-runtime")
	if err := setContainerRuntimeOptions(p.clusterName, p); err != nil {
		klog.Infof("Error setting container-runtime options during provisioning %v", err)
		return err
	}

	return nil
}
