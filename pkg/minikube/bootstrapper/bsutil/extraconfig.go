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

// Package bsutil will eventually be renamed to kubeadm package after getting rid of older one
package bsutil

import (
	"fmt"
	"sort"
	"strings"

	"github.com/blang/semver"
	"github.com/pkg/errors"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/config"
)

// enum to differentiate kubeadm command line parameters from kubeadm config file parameters (see the
// KubeadmExtraArgsAllowed variable for more info)
const (
	// KubeadmCmdParam is command parameters for kubeadm
	KubeadmCmdParam = iota
	// KubeadmConfigParam is config parameters for kubeadm
	KubeadmConfigParam = iota
)

// componentOptions holds extra args for a component
type componentOptions struct {
	Component string
	ExtraArgs map[string]string
	Pairs     map[string]string
}

// mapping of component to the section name in kubeadm.
var componentToKubeadmConfigKey = map[string]string{
	Apiserver:         "apiServer",
	ControllerManager: "controllerManager",
	Scheduler:         "scheduler",
	Etcd:              "etcd",
	Kubeadm:           "kubeadm",
	// The KubeProxy is handled in different config block
	Kubeproxy: "",
	// The Kubelet is not configured in kubeadm, only in systemd.
	Kubelet: "",
}

// KubeadmExtraArgsAllowed is a list of supported kubeadm params that can be supplied to kubeadm through
// minikube's ExtraArgs parameter. The list is split into two parts - params that can be supplied as flags on the
// command line and params that have to be inserted into the kubeadm config file. This is because of a kubeadm
// constraint which allows only certain params to be provided from the command line when the --config parameter
// is specified
var KubeadmExtraArgsAllowed = map[int][]string{
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
		"skip-phases",
	},
	KubeadmConfigParam: {
		"pod-network-cidr",
	},
}

// CreateFlagsFromExtraArgs converts kubeadm extra args into flags to be supplied from the command linne
func CreateFlagsFromExtraArgs(extraOptions config.ExtraOptionSlice) string {
	kubeadmExtraOpts := extraOptions.AsMap().Get(Kubeadm)

	// kubeadm allows only a small set of parameters to be supplied from the command line when the --config param
	// is specified, here we remove those that are not allowed
	for opt := range kubeadmExtraOpts {
		if !config.ContainsParam(KubeadmExtraArgsAllowed[KubeadmCmdParam], opt) {
			// kubeadmExtraOpts is a copy so safe to delete
			delete(kubeadmExtraOpts, opt)
		}
	}
	return convertToFlags(kubeadmExtraOpts)
}

// FindInvalidExtraConfigFlags returns all invalid 'extra-config' options
func FindInvalidExtraConfigFlags(opts config.ExtraOptionSlice) []string {
	invalidOptsMap := make(map[string]struct{})
	var invalidOpts []string
	for _, extraOpt := range opts {
		if _, ok := componentToKubeadmConfigKey[extraOpt.Component]; !ok {
			if _, ok := invalidOptsMap[extraOpt.Component]; !ok {
				invalidOpts = append(invalidOpts, extraOpt.Component)
				invalidOptsMap[extraOpt.Component] = struct{}{}
			}
		}
	}
	return invalidOpts
}

// extraConfigForComponent generates a map of flagname-value pairs for a k8s
// component.
func extraConfigForComponent(component string, opts config.ExtraOptionSlice, version semver.Version) (map[string]string, error) {
	versionedOpts, err := defaultOptionsForComponentAndVersion(component, version)
	if err != nil {
		return nil, errors.Wrapf(err, "setting version specific options for %s", component)
	}

	for _, opt := range opts {
		if opt.Component == component {
			if val, ok := versionedOpts[opt.Key]; ok {
				klog.Infof("Overwriting default %s=%s with user provided %s=%s for component %s", opt.Key, val, opt.Key, opt.Value, component)
			}
			versionedOpts[opt.Key] = opt.Value
		}
	}

	return versionedOpts, nil
}

// defaultOptionsForComponentAndVersion returns the default option for a component and version
func defaultOptionsForComponentAndVersion(component string, version semver.Version) (map[string]string, error) {
	versionedOpts := map[string]string{}
	for _, opts := range versionSpecificOpts {
		if opts.Option.Component == component {
			if versionIsBetween(version, opts.GreaterThanOrEqual, opts.LessThanOrEqual) {
				if val, ok := versionedOpts[opts.Option.Key]; ok {
					return nil, fmt.Errorf("flag %s=%q already set %s=%q", opts.Option.Key, opts.Option.Value, opts.Option.Key, val)
				}
				versionedOpts[opts.Option.Key] = opts.Option.Value
			}
		}
	}
	return versionedOpts, nil
}

// newComponentOptions creates a new componentOptions
func newComponentOptions(opts config.ExtraOptionSlice, version semver.Version, featureGates string, cp config.Node) ([]componentOptions, error) {
	if invalidOpts := FindInvalidExtraConfigFlags(opts); len(invalidOpts) > 0 {
		return nil, fmt.Errorf("unknown components %v. valid components are: %v", invalidOpts, KubeadmExtraConfigOpts)
	}

	var kubeadmExtraArgs []componentOptions
	for _, component := range KubeadmExtraConfigOpts {
		kubeadmComponentKey := componentToKubeadmConfigKey[component]
		if kubeadmComponentKey == "" {
			continue
		}
		extraConfig, err := extraConfigForComponent(component, opts, version)
		if err != nil {
			return nil, errors.Wrapf(err, "getting kubeadm extra args for %s", component)
		}
		if featureGates != "" {
			extraConfig["feature-gates"] = featureGates
		}
		if len(extraConfig) > 0 {
			kubeadmExtraArgs = append(kubeadmExtraArgs, componentOptions{
				Component: kubeadmComponentKey,
				ExtraArgs: extraConfig,
				Pairs:     optionPairsForComponent(component, version, cp),
			})
		}
	}

	return kubeadmExtraArgs, nil
}

// optionPairsForComponent generates a map of value pairs for a k8s component
func optionPairsForComponent(component string, version semver.Version, cp config.Node) map[string]string {
	// For the ktmpl.V1Beta1 users
	if component == Apiserver && version.GTE(semver.MustParse("1.14.0-alpha.0")) {
		return map[string]string{
			"certSANs": fmt.Sprintf(`["127.0.0.1", "localhost", "%s"]`, cp.IP),
		}
	}
	return nil
}

// kubeadm extra args should not be included in the kubeadm config in the extra args section (instead, they must
// be inserted explicitly in the appropriate places or supplied from the command line); here we remove all of the
// kubeadm extra args from the slice
// etcd must also not be included in that section, as those extra args exist in the `etcd` section
// createExtraComponentConfig generates a map of component to extra args for all of the components except kubeadm
func createExtraComponentConfig(extraOptions config.ExtraOptionSlice, version semver.Version, componentFeatureArgs string, cp config.Node) ([]componentOptions, error) {
	extraArgsSlice, err := newComponentOptions(extraOptions, version, componentFeatureArgs, cp)
	if err != nil {
		return nil, err
	}
	validComponents := []componentOptions{}
	for _, extraArgs := range extraArgsSlice {
		if extraArgs.Component == Kubeadm || extraArgs.Component == Etcd {
			continue
		}
		validComponents = append(validComponents, extraArgs)
	}
	return validComponents, nil
}

// createKubeProxyOptions generates a map of extra config for kube-proxy
func createKubeProxyOptions(extraOptions config.ExtraOptionSlice) map[string]string {
	kubeProxyOptions := extraOptions.AsMap().Get(Kubeproxy)
	return kubeProxyOptions
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
