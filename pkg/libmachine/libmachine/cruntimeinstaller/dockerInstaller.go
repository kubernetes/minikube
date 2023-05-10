/*
Copyright 2023 The Kubernetes Authors All rights reserved.

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

package cruntimeinstaller

import (
	"bytes"
	"fmt"
	"html/template"
	"os/exec"

	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/libmachine/libmachine/auth"
	"k8s.io/minikube/pkg/libmachine/libmachine/engine"
	"k8s.io/minikube/pkg/libmachine/libmachine/log"
	"k8s.io/minikube/pkg/libmachine/libmachine/runner"
)

type dockerInstaller struct {
	CRuntimeName       string
	Options            *engine.Options
	AuthOptions        *auth.Options
	Commander          runner.Runner
	CRuntimeOptionsDir string
	Provider           string
}

func newDockerInstaller(opts *engine.Options, cmd runner.Runner, provider string, authOptions *auth.Options) *dockerInstaller {
	return &dockerInstaller{
		Options:      opts,
		CRuntimeName: "Docker",
		Commander:    cmd,
		Provider:     provider,
		AuthOptions:  authOptions,
	}
}

type DockerOptions struct {
	EngineOptions     string
	EngineOptionsPath string
}

// x7NOTE: Complete this
// InstallCRuntime checks if docker is installed in the machine;
// if not, it installs it from get.docker.com.
// It also configures the daemon.
func (di *dockerInstaller) InstallCRuntime() error {
	err := di.installDockerGeneric(di.Options.InstallURL)
	if err != nil {
		return errors.Wrap(err, "error while trying to install docker into machine")
	}

	if err := di.SetCRuntimeOptions(); err != nil {
		return errors.Wrapf(err, "error setting container-runtime (%s) options during provisioning", di.CRuntimeName)
	}

	return nil
}

// x7NOTE: incompatible with non-systemd systems
// SetCRuntimeOptions generates and updates the existing docker configuration
func (di *dockerInstaller) SetCRuntimeOptions() error {
	var engineCfg bytes.Buffer

	drvLabel := fmt.Sprintf("provider=%s", di.Provider)
	di.Options.Labels = append(di.Options.Labels, drvLabel)

	noPivot := true
	// Using pivot_root is not supported on fstype rootfs
	if fstype, err := rootFileSystemType(di.Commander); err == nil {
		log.Infof("root file system type: %s", fstype)
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
		log.Info("Using fundamentally insecure --no-pivot option")
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
		return err
	}

	engineConfigContext := engine.ConfigContext{
		DockerPort:    engine.DefaultPort,
		AuthOptions:   *di.AuthOptions,
		EngineOptions: *di.Options,
	}

	escapeSystemdDirectives(&engineConfigContext)

	if err := t.Execute(&engineCfg, engineConfigContext); err != nil {
		return err
	}

	do := &DockerOptions{
		EngineOptions:     engineCfg.String(),
		EngineOptionsPath: "/lib/systemd/system/docker.service",
	}

	return updateUnit(di.Commander, "docker", do.EngineOptions, do.EngineOptionsPath)
}

func (di *dockerInstaller) installDockerGeneric(baseURL string) error {
	// install docker - until cloudinit we use ubuntu everywhere so we
	// just install it using the docker repos
	if output, err := di.Commander.RunCmd(exec.Command("bash", "-c", fmt.Sprintf("if ! type docker; then curl -sSL %s | sh -; fi", baseURL))); err != nil {
		return fmt.Errorf("error installing docker: %s", output.Stdout.String())
	}

	return nil
}

// x7TODO: linter says it's unused.. figure out why
// func (di *dockerInstaller) makeDockerOptionsDir() error {
// 	if _, err := di.Commander.RunCmd(exec.Command("bash", "-c", fmt.Sprintf("sudo mkdir -p %s", di.CRuntimeOptionsDir))); err != nil {
// 		return err
// 	}

// 	return nil
// }
