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

package addons

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/golang/glog"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
)

var kubectlPruneWhitelist = []string{
	"core/v1/ConfigMap",
	"core/v1/Endpoints",
	"core/v1/Namespace",
	"core/v1/PersistentVolumeClaim",
	"core/v1/PersistentVolume",
	"core/v1/Pod",
	"core/v1/ReplicationController",
	"core/v1/Secret",
	"core/v1/Service",
	"batch/v1/Job",
	"batch/v1beta1/CronJob",
	"apps/v1/DaemonSet",
	"apps/v1/Deployment",
	"apps/v1/ReplicaSet",
	"apps/v1/StatefulSet",
	"extensions/v1beta1/Ingress",
}

// ReconcileAddons runs kubectl apply -f on the addons directory
// to reconcile addons state
func ReconcileAddons(cmd command.Runner) error {
	reconcileCmd, err := kubectlCommand()
	if err != nil {
		return err
	}
	fmt.Println("running", reconcileCmd)
	rr, err := cmd.RunCmd(reconcileCmd)
	if err != nil {
		glog.Warningf("reconciling addons failed: %v", err)
		return err
	}
	fmt.Println(rr.Stdout.String())
	return nil
}

func kubectlCommand() (*exec.Cmd, error) {
	kubectlBinary, err := kubectlPath()
	if err != nil {
		return nil, err
	}
	args := []string{"KUBECONFIG=/var/lib/minikube/kubeconfig", kubectlBinary, "apply", "-f", "/etc/kubernetes/addons", "-l", "kubernetes.io/cluster-service!=true,addonmanager.kubernetes.io/mode=Reconcile", "--prune=true"}
	for _, k := range kubectlPruneWhitelist {
		args = append(args, []string{"--prune-whitelist", k}...)
	}
	args = append(args, "--recursive")
	cmd := exec.Command("sudo", args...)
	cmd.Env = append(os.Environ(), "KUBECONFIG=/var/lib/minikube/kubeconfig")
	return cmd, nil
}

func kubectlPath() (string, error) {
	cc, err := config.Load()
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}
	version := constants.DefaultKubernetesVersion
	if cc != nil {
		version = cc.KubernetesConfig.KubernetesVersion
	}
	path := fmt.Sprintf("/var/lib/minikube/binaries/%s/kubectl", version)
	return path, nil
}
