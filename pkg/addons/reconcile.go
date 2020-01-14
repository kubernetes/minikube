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
	"os"
	"os/exec"
	"path"

	"github.com/blang/semver"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/bootstrapper/bsutil"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
)

// taken from https://github.com/kubernetes/kubernetes/blob/master/cluster/addons/addon-manager/kube-addons.sh
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

// reconcile runs kubectl apply -f on the addons directory
// to reconcile addons state in all running profiles
func reconcile(cmd command.Runner, profile string) error {
	c, err := kubectlCommand(profile)
	if err != nil {
		return err
	}
	if _, err := cmd.RunCmd(c); err != nil {
		glog.Warningf("reconciling addons failed: %v", err)
		return err
	}
	return nil
}

func kubectlCommand(profile string) (*exec.Cmd, error) {
	kubectlBinary, err := kubectlBinaryPath(profile)
	if err != nil {
		return nil, err
	}
	v, err := kubernetesVersion(profile)
	if err != nil {
		return nil, err
	}
	// prune will delete any existing objects with the label specified by "-l" which don't appear in /etc/kubernetes/addons
	// this is how we delete disabled addons
	args := []string{"KUBECONFIG=/var/lib/minikube/kubeconfig", kubectlBinary, "apply", "-f", "/etc/kubernetes/addons", "-l", "kubernetes.io/cluster-service!=true,addonmanager.kubernetes.io/mode=Reconcile", "--prune=true"}
	for _, k := range kubectlPruneWhitelist {
		args = append(args, []string{"--prune-whitelist", k}...)
	}
	args = append(args, "--recursive")

	ok, err := shouldAppendNamespaceFlag(v)
	if err != nil {
		return nil, errors.Wrap(err, "appending namespace flag")
	}
	if ok {
		args = append(args, "--namespace=kube-system")
	}

	cmd := exec.Command("sudo", args...)
	return cmd, nil
}

func kubernetesVersion(profile string) (string, error) {
	cc, err := config.Load(profile)
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}
	version := constants.DefaultKubernetesVersion
	if cc != nil {
		version = cc.KubernetesConfig.KubernetesVersion
	}
	return version, nil
}

// We need to append --namespace=kube-system for Kubernetes versions >=1.17
// so that prune works as expected. See https://github.com/kubernetes/kubernetes/pull/83084/
func shouldAppendNamespaceFlag(version string) (bool, error) {
	v, err := bsutil.ParseKubernetesVersion(version)
	if err != nil {
		return false, err
	}
	return v.GTE(semver.MustParse("1.17.0")), nil
}

func kubectlBinaryPath(profile string) (string, error) {
	// TODO: get this for all profiles and run in each one
	v, err := kubernetesVersion(profile)
	if err != nil {
		return "", err
	}
	p := path.Join("/var/lib/minikube/binaries", v, "kubectl")
	return p, nil
}
