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
	"os/exec"

	"github.com/golang/glog"
	"k8s.io/minikube/pkg/minikube/command"
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
	if _, err := cmd.RunCmd(kubectlCommand()); err != nil {
		glog.Warningf("reconciling addons failed: %v", err)
	}
	return nil
}

func kubectlCommand() *exec.Cmd {
	args := []string{"apply", "-f", "/etc/kubernetes/addons", "-l", "kubernetes.io/cluster-service!=true,addonmanager.kubernetes.io/mode=Reconcile", "--prune=true"}
	for _, k := range kubectlPruneWhitelist {
		args = append(args, []string{"--prune-whitelist", k}...)
	}
	args = append(args, "--recursive")
	cmd := exec.Command("kubectl", args...)
	return cmd
}
