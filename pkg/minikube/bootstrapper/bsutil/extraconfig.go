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

// bsutil package will eventually be renamed to kubeadm package after getting rid of older one
package bsutil

import (
	"fmt"
	"sort"
	"strings"

	"github.com/blang/semver"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/config"
)

const (
	KubeadmCmdParam    = iota
	KubeadmConfigParam = iota
)

// ComponentExtraArgs holds extra args for a component
type ComponentExtraArgs struct {
	Component string
	Options   map[string]string
}

// mapping of component to the section name in kubeadm.
var componentToKubeadmConfigKey = map[string]string{
	Apiserver:         "apiServer",
	ControllerManager: "controllerManager",
	Scheduler:         "scheduler",
	Kubeadm:           "kubeadm",
	// The Kubelet is not configured in kubeadm, only in systemd.
	Kubelet: "",
}

// KubeadmExtraArgsWhitelist is a whitelist of supported kubeadm params that can be supplied to kubeadm through
// minikube's ExtraArgs parameter. The list is split into two parts - params that can be supplied as flags on the
// command line and params that have to be inserted into the kubeadm config file. This is because of a kubeadm
// constraint which allows only certain params to be provided from the command line when the --config parameter
// is specified
var KubeadmExtraArgsWhitelist = map[int][]string{
	KubeadmCmdParam: {
		"ignore-preflight-errors",
		"dry-run",
		"kubeconfig",
		"kubeconfig-dir",
		"node-name",
		"cri-socket",
		"experimental-upload-certs",
		"certificate-key",
		"rootfs",
	},
	KubeadmConfigParam: {
		"pod-network-cidr",
	},
}

// newComponentExtraArgs creates a new ComponentExtraArgs
func newComponentExtraArgs(opts config.ExtraOptionSlice, version semver.Version, featureGates string) ([]ComponentExtraArgs, error) {
	var kubeadmExtraArgs []ComponentExtraArgs
	for _, extraOpt := range opts {
		if _, ok := componentToKubeadmConfigKey[extraOpt.Component]; !ok {
			return nil, fmt.Errorf("unknown component %q. valid components are: %v", componentToKubeadmConfigKey, componentToKubeadmConfigKey)
		}
	}

	keys := []string{}
	for k := range componentToKubeadmConfigKey {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, component := range keys {
		kubeadmComponentKey := componentToKubeadmConfigKey[component]
		if kubeadmComponentKey == "" {
			continue
		}
		extraConfig, err := ExtraConfigForComponent(component, opts, version)
		if err != nil {
			return nil, errors.Wrapf(err, "getting kubeadm extra args for %s", component)
		}
		if featureGates != "" {
			extraConfig["feature-gates"] = featureGates
		}
		if len(extraConfig) > 0 {
			kubeadmExtraArgs = append(kubeadmExtraArgs, ComponentExtraArgs{
				Component: kubeadmComponentKey,
				Options:   extraConfig,
			})
		}
	}

	return kubeadmExtraArgs, nil
}

// CreateFlagsFromExtraArgs converts kubeadm extra args into flags to be supplied from the command linne
func CreateFlagsFromExtraArgs(extraOptions config.ExtraOptionSlice) string {
	kubeadmExtraOpts := extraOptions.AsMap().Get(Kubeadm)

	// kubeadm allows only a small set of parameters to be supplied from the command line when the --config param
	// is specified, here we remove those that are not allowed
	for opt := range kubeadmExtraOpts {
		if !config.ContainsParam(KubeadmExtraArgsWhitelist[KubeadmCmdParam], opt) {
			// kubeadmExtraOpts is a copy so safe to delete
			delete(kubeadmExtraOpts, opt)
		}
	}
	return convertToFlags(kubeadmExtraOpts)
}

// createExtraComponentConfig generates a map of component to extra args for all of the components except kubeadm
func createExtraComponentConfig(extraOptions config.ExtraOptionSlice, version semver.Version, componentFeatureArgs string) ([]ComponentExtraArgs, error) {
	extraArgsSlice, err := newComponentExtraArgs(extraOptions, version, componentFeatureArgs)
	if err != nil {
		return nil, err
	}

	// kubeadm extra args should not be included in the kubeadm config in the extra args section (instead, they must
	// be inserted explicitly in the appropriate places or supplied from the command line); here we remove all of the
	// kubeadm extra args from the slice
	for i, extraArgs := range extraArgsSlice {
		if extraArgs.Component == Kubeadm {
			extraArgsSlice = append(extraArgsSlice[:i], extraArgsSlice[i+1:]...)
			break
		}
	}
	return extraArgsSlice, nil
}

func convertToFlags(opts map[string]string) string {
	var flags []string
	var keys []string
	for k := range opts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		flags = append(flags, fmt.Sprintf("--%s=%s", k, opts[k]))
	}
	return strings.Join(flags, " ")
}
