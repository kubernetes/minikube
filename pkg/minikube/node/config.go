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

package node

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/spf13/viper"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/cruntime"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/out/register"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/minikube/style"
	"k8s.io/minikube/pkg/util"
	"k8s.io/minikube/pkg/util/lock"
)

func showVersionInfo(k8sVersion string, cr cruntime.Manager) {
	version, _ := cr.Version()
	register.Reg.SetStep(register.PreparingKubernetes)
	out.Step(cr.Style(), "Preparing Kubernetes {{.k8sVersion}} on {{.runtime}} {{.runtimeVersion}} ...", out.V{"k8sVersion": k8sVersion, "runtime": cr.Name(), "runtimeVersion": version})
	for _, v := range config.DockerOpt {
		v = util.MaskProxyPasswordWithKey(v)
		out.Infof("opt {{.docker_option}}", out.V{"docker_option": v})
	}
	for _, v := range config.DockerEnv {
		out.Infof("env {{.docker_env}}", out.V{"docker_env": v})
	}
}

func showNoK8sVersionInfo(cr cruntime.Manager) {
	err := cruntime.CheckCompatibility(cr)
	if err != nil {
		klog.Warningf("%s check compatibility failed: %v", cr.Name(), err)

	}

	version, err := cr.Version()
	if err != nil {
		klog.Warningf("%s get version failed: %v", cr.Name(), err)
	}

	out.Step(cr.Style(), "Preparing {{.runtime}} {{.runtimeVersion}} ...", out.V{"runtime": cr.Name(), "runtimeVersion": version})
}

// configureMounts configures any requested filesystem mounts
func configureMounts(wg *sync.WaitGroup, cc config.ClusterConfig) {
	wg.Add(1)
	defer wg.Done()

	if !cc.Mount || driver.IsKIC(cc.Driver) {
		return
	}

	out.Step(style.Mounting, "Creating mount {{.name}} ...", out.V{"name": cc.MountString})
	path := os.Args[0]
	profile := viper.GetString("profile")

	args := generateMountArgs(profile, cc)
	mountCmd := exec.Command(path, args...)
	mountCmd.Env = append(os.Environ(), constants.IsMinikubeChildProcess+"=true")
	if klog.V(8).Enabled() {
		mountCmd.Stdout = os.Stdout
		mountCmd.Stderr = os.Stderr
	}
	if err := mountCmd.Start(); err != nil {
		exit.Error(reason.GuestMount, "Error starting mount", err)
	}
	if err := lock.AppendToFile(filepath.Join(localpath.Profile(profile), constants.MountProcessFileName), []byte(fmt.Sprintf(" %s", strconv.Itoa(mountCmd.Process.Pid))), 0o644); err != nil {
		exit.Error(reason.HostMountPid, "Error writing mount pid", err)
	}
}

func generateMountArgs(profile string, cc config.ClusterConfig) []string {
	mountDebugVal := 0
	if klog.V(8).Enabled() {
		mountDebugVal = 1
	}

	args := []string{"mount", cc.MountString}
	flags := []struct {
		name  string
		value string
	}{
		{"profile", profile},
		{"v", fmt.Sprintf("%d", mountDebugVal)},
		{constants.Mount9PVersionFlag, cc.Mount9PVersion},
		{constants.MountGIDFlag, cc.MountGID},
		{constants.MountIPFlag, cc.MountIP},
		{constants.MountMSizeFlag, fmt.Sprintf("%d", cc.MountMSize)},
		{constants.MountPortFlag, fmt.Sprintf("%d", cc.MountPort)},
		{constants.MountTypeFlag, cc.MountType},
		{constants.MountUIDFlag, cc.MountUID},
	}
	for _, flag := range flags {
		args = append(args, fmt.Sprintf("--%s", flag.name), flag.value)
	}
	for _, option := range cc.MountOptions {
		args = append(args, fmt.Sprintf("--%s", constants.MountOptionsFlag), option)
	}
	return args
}
