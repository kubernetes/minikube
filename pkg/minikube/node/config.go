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
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/out/register"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/minikube/style"
	"k8s.io/minikube/pkg/util/lock"
)

func showVersionInfo(k8sVersion string, cr cruntime.Manager) {
	version, _ := cr.Version()
	register.Reg.SetStep(register.PreparingKubernetes)
	out.Step(cr.Style(),"Preparing Kubernetes {{.k8sVersion}} on {{.runtime}} {{.runtimeVersion}} ...", out.V{"k8sVersion": k8sVersion, "runtime": cr.Name(), "runtimeVersion": version})
	for _, v := range config.DockerOpt {
		out.Infof("opt {{.docker_option}}", out.V{"docker_option": v})
	}
	for _, v := range config.DockerEnv {
		out.Infof("env {{.docker_env}}", out.V{"docker_env": v})
	}
}

// configureMounts configures any requested filesystem mounts
func configureMounts(wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	if !viper.GetBool(createMount) {
		return
	}

	out.Step(style.Mounting, "Creating mount {{.name}} ...", out.V{"name": viper.GetString(mountString)})
	path := os.Args[0]
	mountDebugVal := 0
	if klog.V(8).Enabled() {
		mountDebugVal = 1
	}
	mountCmd := exec.Command(path, "mount", fmt.Sprintf("--v=%d", mountDebugVal), viper.GetString(mountString))
	mountCmd.Env = append(os.Environ(), constants.IsMinikubeChildProcess+"=true")
	if klog.V(8).Enabled() {
		mountCmd.Stdout = os.Stdout
		mountCmd.Stderr = os.Stderr
	}
	if err := mountCmd.Start(); err != nil {
		exit.Error(reason.GuestMount, "Error starting mount", err)
	}
	if err := lock.WriteFile(filepath.Join(localpath.MiniPath(), constants.MountProcessFileName), []byte(strconv.Itoa(mountCmd.Process.Pid)), 0o644); err != nil {
		exit.Error(reason.HostMountPid, "Error writing mount pid", err)
	}
}
