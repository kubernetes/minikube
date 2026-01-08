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
	"text/template"
	"time"

	"github.com/spf13/viper"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/libmachine/auth"
	"k8s.io/minikube/pkg/libmachine/drivers"
	"k8s.io/minikube/pkg/libmachine/engine"
	"k8s.io/minikube/pkg/libmachine/provision"
	"k8s.io/minikube/pkg/libmachine/provision/pkgaction"
	"k8s.io/minikube/pkg/libmachine/swarm"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/util/retry"
)

// BuildrootProvisioner provisions the custom system based on Buildroot
type BuildrootProvisioner struct {
	provision.SystemdProvisioner
	clusterName string
}

// NewBuildrootProvisioner creates a new BuildrootProvisioner
func NewBuildrootProvisioner(d drivers.Driver) provision.Provisioner {
	return &BuildrootProvisioner{
		NewSystemdProvisioner("buildroot", d),
		viper.GetString(config.ProfileName),
	}
}

func (p *BuildrootProvisioner) String() string {
	return "buildroot"
}

// CompatibleWithHost checks if provisioner is compatible with host
func (p *BuildrootProvisioner) CompatibleWithHost() bool {
	return p.OsReleaseInfo.ID == "buildroot"
}

// GenerateDockerOptions generates the *provision.DockerOptions for this provisioner
func (p *BuildrootProvisioner) GenerateDockerOptions(dockerPort int) (*provision.DockerOptions, error) {
	var engineCfg bytes.Buffer

	drvLabel := fmt.Sprintf("provider=%s", p.Driver.DriverName())
	p.EngineOptions.Labels = append(p.EngineOptions.Labels, drvLabel)

	noPivot := true
	// Using pivot_root is not supported on fstype rootfs
	if fstype, err := rootFileSystemType(p); err == nil {
		klog.Infof("root file system type: %s", fstype)
		noPivot = fstype == "rootfs"
	}
	/* Template is copied from https://github.com/moby/moby/blob/44bca1adf3c806c4ed6688d67c6f995653b09b26/contrib/init/systemd/docker.service#L2 with the following changes:
	   After: minikube-automount.service is added
	   ExecStart: additional flags are added
	*/
	engineConfigTmpl := `[Unit]
Description=Docker Application Container Engine
Documentation=https://docs.docker.com
After=network-online.target minikube-automount.service nss-lookup.target docker.socket firewalld.service containerd.service time-set.target
Wants=network-online.target containerd.service
Requires=docker.socket
StartLimitBurst=3
StartLimitIntervalSec=60
[Service]
Type=notify
Restart=always
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

ExecStart=
ExecStart=/usr/bin/dockerd -H tcp://0.0.0.0:2376 \
	-H fd:// --containerd=/run/containerd/containerd.sock \
	-H unix:///var/run/docker.sock \
	--default-ulimit=nofile=1048576:1048576 \
	--tlsverify \
	--tlscacert {{.AuthOptions.CaCertRemotePath}} \
	--tlscert {{.AuthOptions.ServerCertRemotePath}} \
	--tlskey {{.AuthOptions.ServerKeyRemotePath}} {{ range .EngineOptions.Labels }}--label {{.}} {{ end }}{{ range .EngineOptions.InsecureRegistry }}--insecure-registry {{.}} {{ end }}{{ range .EngineOptions.RegistryMirror }}--registry-mirror {{.}} {{ end }}{{ range .EngineOptions.ArbitraryFlags }}--{{.}} {{ end }}
ExecReload=/bin/kill -s HUP \$MAINPID

# Having non-zero Limit*s causes performance problems due to accounting overhead
# in the kernel. We recommend using cgroups to do container-local accounting.
LimitNOFILE=infinity
LimitNPROC=infinity
LimitCORE=infinity

# Uncomment TasksMax if your systemd version supports it.
# Only systemd 226 and above support this version.
TasksMax=infinity

# set delegate yes so that systemd does not reset the cgroups of docker containers
Delegate=yes

# kill only the docker process, not all processes in the cgroup
KillMode=process
OOMScoreAdjust=-500

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
func (p *BuildrootProvisioner) Package(_ string, _ pkgaction.PackageAction) error {
	return nil
}

// Provision does the provisioning
func (p *BuildrootProvisioner) Provision(swarmOptions swarm.Options, authOptions auth.Options, engineOptions engine.Options) error {
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
