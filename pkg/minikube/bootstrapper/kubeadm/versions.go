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

package kubeadm

import (
	"fmt"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/blang/semver"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"k8s.io/kubernetes/cmd/kubeadm/app/features"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/util"
)

// These are the components that can be configured
// through the "extra-config"
const (
	Kubelet           = "kubelet"
	Kubeadm           = "kubeadm"
	Apiserver         = "apiserver"
	Scheduler         = "scheduler"
	ControllerManager = "controller-manager"
)

// ExtraConfigForComponent generates a map of flagname-value pairs for a k8s
// component.
func ExtraConfigForComponent(component string, opts util.ExtraOptionSlice, version semver.Version) (map[string]string, error) {
	versionedOpts, err := DefaultOptionsForComponentAndVersion(component, version)
	if err != nil {
		return nil, errors.Wrapf(err, "setting version specific options for %s", component)
	}

	for _, opt := range opts {
		if opt.Component == component {
			if val, ok := versionedOpts[opt.Key]; ok {
				glog.Infof("Overwriting default %s=%s with user provided %s=%s for component %s", opt.Key, val, opt.Key, opt.Value, component)
			}
			versionedOpts[opt.Key] = opt.Value
		}
	}

	return versionedOpts, nil
}

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

// NewComponentExtraArgs creates a new ComponentExtraArgs
func NewComponentExtraArgs(opts util.ExtraOptionSlice, version semver.Version, featureGates string) ([]ComponentExtraArgs, error) {
	var kubeadmExtraArgs []ComponentExtraArgs
	for _, extraOpt := range opts {
		if _, ok := componentToKubeadmConfigKey[extraOpt.Component]; !ok {
			return nil, fmt.Errorf("Unknown component %s.  Valid components and kubeadm config are %v", componentToKubeadmConfigKey, componentToKubeadmConfigKey)
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

// ParseFeatureArgs parses feature args into extra args
func ParseFeatureArgs(featureGates string) (map[string]bool, string, error) {
	kubeadmFeatureArgs := map[string]bool{}
	componentFeatureArgs := ""
	for _, s := range strings.Split(featureGates, ",") {
		if len(s) == 0 {
			continue
		}

		fg := strings.SplitN(s, "=", 2)
		if len(fg) != 2 {
			return nil, "", fmt.Errorf("missing value for key \"%v\"", s)
		}

		k := strings.TrimSpace(fg[0])
		v := strings.TrimSpace(fg[1])

		if !Supports(k) {
			componentFeatureArgs = fmt.Sprintf("%s%s,", componentFeatureArgs, s)
			continue
		}

		boolValue, err := strconv.ParseBool(v)
		if err != nil {
			return nil, "", errors.Wrapf(err, "failed to convert bool value \"%v\"", v)
		}
		kubeadmFeatureArgs[k] = boolValue
	}
	componentFeatureArgs = strings.TrimRight(componentFeatureArgs, ",")
	return kubeadmFeatureArgs, componentFeatureArgs, nil
}

// Supports indicates whether a feature name is supported on the
// feature gates for kubeadm
func Supports(featureName string) bool {
	for k := range features.InitFeatureGates {
		if featureName == string(k) {
			return true
		}
	}
	return false
}

// ParseKubernetesVersion parses the kubernetes version
func ParseKubernetesVersion(version string) (semver.Version, error) {
	// Strip leading 'v' prefix from version for semver parsing
	v, err := semver.Make(version[1:])
	if err != nil {
		return semver.Version{}, errors.Wrap(err, "invalid version, must begin with 'v'")
	}

	return v, nil
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

// VersionedExtraOption holds information on flags to apply to a specific range
// of versions
type VersionedExtraOption struct {
	// Special Cases:
	//
	// If LessThanOrEqual and GreaterThanOrEqual are both nil, the flag will be applied
	// to all versions
	//
	// If LessThanOrEqual == GreaterThanOrEqual, the flag will only be applied to that
	// specific version

	// The flag and component that will be set
	Option util.ExtraOption

	// This flag will only be applied to versions before or equal to this version
	// If it is the default value, it will have no upper bound on versions the
	// flag is applied to
	LessThanOrEqual semver.Version

	// The flag will only be applied to versions after or equal to this version
	// If it is the default value, it will have no lower bound on versions the
	// flag is applied to
	GreaterThanOrEqual semver.Version
}

// NewUnversionedOption returns a VersionedExtraOption that applies to all versions.
func NewUnversionedOption(component, k, v string) VersionedExtraOption {
	return VersionedExtraOption{
		Option: util.ExtraOption{
			Component: component,
			Key:       k,
			Value:     v,
		},
	}
}

var versionSpecificOpts = []VersionedExtraOption{
	{
		Option: util.ExtraOption{
			Component: Kubelet,
			Key:       "fail-swap-on",
			Value:     "false",
		},
		GreaterThanOrEqual: semver.MustParse("1.8.0-alpha.0"),
	},
	// Kubeconfig args
	NewUnversionedOption(Kubelet, "kubeconfig", "/etc/kubernetes/kubelet.conf"),
	NewUnversionedOption(Kubelet, "bootstrap-kubeconfig", "/etc/kubernetes/bootstrap-kubelet.conf"),
	{
		Option: util.ExtraOption{
			Component: Kubelet,
			Key:       "require-kubeconfig",
			Value:     "true",
		},
		LessThanOrEqual: semver.MustParse("1.9.10"),
	},
	NewUnversionedOption(Kubelet, "hostname-override", constants.DefaultNodeName),

	// System pods args
	NewUnversionedOption(Kubelet, "pod-manifest-path", "/etc/kubernetes/manifests"),
	NewUnversionedOption(Kubelet, "allow-privileged", "true"),

	// Network args
	NewUnversionedOption(Kubelet, "cluster-dns", "10.96.0.10"),
	NewUnversionedOption(Kubelet, "cluster-domain", "cluster.local"),

	// Auth args
	NewUnversionedOption(Kubelet, "authorization-mode", "Webhook"),
	NewUnversionedOption(Kubelet, "client-ca-file", path.Join(util.DefaultCertPath, "ca.crt")),

	// Cgroup args
	NewUnversionedOption(Kubelet, "cgroup-driver", "cgroupfs"),
	{
		Option: util.ExtraOption{
			Component: Apiserver,
			Key:       "admission-control",
			Value:     strings.Join(util.DefaultLegacyAdmissionControllers, ","),
		},
		LessThanOrEqual:    semver.MustParse("1.10.1000"), // Semver doesn't support wildcards.
		GreaterThanOrEqual: semver.MustParse("1.9.0-alpha.0"),
	},
	{
		Option: util.ExtraOption{
			Component: Apiserver,
			Key:       "enable-admission-plugins",
			Value:     strings.Join(util.DefaultLegacyAdmissionControllers, ","),
		},
		GreaterThanOrEqual: semver.MustParse("1.11.0-alpha.0"),
		LessThanOrEqual:    semver.MustParse("1.13.1000"),
	},
	{
		Option: util.ExtraOption{
			Component: Apiserver,
			Key:       "enable-admission-plugins",
			Value:     strings.Join(util.DefaultV114AdmissionControllers, ","),
		},
		GreaterThanOrEqual: semver.MustParse("1.14.0-alpha.0"),
	},

	{
		Option: util.ExtraOption{
			Component: Kubelet,
			Key:       "cadvisor-port",
			Value:     "0",
		},
		LessThanOrEqual: semver.MustParse("1.11.1000"),
	},
}

// VersionIsBetween checks if a version is between (or including) two given versions
func VersionIsBetween(version, gte, lte semver.Version) bool {
	if gte.NE(semver.Version{}) && !version.GTE(gte) {
		return false
	}
	if lte.NE(semver.Version{}) && !version.LTE(lte) {
		return false
	}

	return true
}

// DefaultOptionsForComponentAndVersion returns the default option for a component and version
func DefaultOptionsForComponentAndVersion(component string, version semver.Version) (map[string]string, error) {
	versionedOpts := map[string]string{}
	for _, opts := range versionSpecificOpts {
		if opts.Option.Component == component {
			if VersionIsBetween(version, opts.GreaterThanOrEqual, opts.LessThanOrEqual) {
				if val, ok := versionedOpts[opts.Option.Key]; ok {
					return nil, fmt.Errorf("Flag %s=%s already set %s=%s", opts.Option.Key, opts.Option.Value, opts.Option.Key, val)
				}
				versionedOpts[opts.Option.Key] = opts.Option.Value
			}
		}
	}
	return versionedOpts, nil
}
