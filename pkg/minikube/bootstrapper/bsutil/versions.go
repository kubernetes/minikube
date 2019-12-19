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

package bsutil

import (
	"path"
	"strings"

	"github.com/blang/semver"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/vmpath"
	"k8s.io/minikube/pkg/util"
)

// ParseKubernetesVersion parses the kubernetes version
func ParseKubernetesVersion(version string) (semver.Version, error) {
	// Strip leading 'v' prefix from version for semver parsing
	v, err := semver.Make(version[1:])
	if err != nil {
		return semver.Version{}, errors.Wrap(err, "invalid version, must begin with 'v'")
	}

	return v, nil
}

// versionIsBetween checks if a version is between (or including) two given versions
func versionIsBetween(version, gte, lte semver.Version) bool {
	if gte.NE(semver.Version{}) && !version.GTE(gte) {
		return false
	}
	if lte.NE(semver.Version{}) && !version.LTE(lte) {
		return false
	}

	return true
}

var versionSpecificOpts = []config.VersionedExtraOption{
	{
		Option: config.ExtraOption{
			Component: Kubelet,
			Key:       "fail-swap-on",
			Value:     "false",
		},
		GreaterThanOrEqual: semver.MustParse("1.8.0-alpha.0"),
	},
	// Kubeconfig args
	config.NewUnversionedOption(Kubelet, "kubeconfig", "/etc/kubernetes/kubelet.conf"),
	config.NewUnversionedOption(Kubelet, "bootstrap-kubeconfig", "/etc/kubernetes/bootstrap-kubelet.conf"),
	{
		Option: config.ExtraOption{
			Component: Kubelet,
			Key:       "require-kubeconfig",
			Value:     "true",
		},
		LessThanOrEqual: semver.MustParse("1.9.10"),
	},
	config.NewUnversionedOption(Kubelet, "hostname-override", constants.DefaultNodeName),

	// System pods args
	config.NewUnversionedOption(Kubelet, "pod-manifest-path", vmpath.GuestManifestsDir),
	{
		Option: config.ExtraOption{
			Component: Kubelet,
			Key:       "allow-privileged",
			Value:     "true",
		},
		LessThanOrEqual: semver.MustParse("1.15.0-alpha.3"),
	},

	// Kubelet config file
	config.NewUnversionedOption(Kubelet, "config", "/var/lib/kubelet/config.yaml"),

	// Network args
	config.NewUnversionedOption(Kubelet, "cluster-dns", "10.96.0.10"),
	config.NewUnversionedOption(Kubelet, "cluster-domain", "cluster.local"),

	// Auth args
	config.NewUnversionedOption(Kubelet, "authorization-mode", "Webhook"),
	config.NewUnversionedOption(Kubelet, "client-ca-file", path.Join(vmpath.GuestCertsDir, "ca.crt")),

	// Cgroup args
	config.NewUnversionedOption(Kubelet, "cgroup-driver", "cgroupfs"),
	{
		Option: config.ExtraOption{
			Component: Apiserver,
			Key:       "enable-admission-plugins",
			Value:     strings.Join(util.DefaultLegacyAdmissionControllers, ","),
		},
		GreaterThanOrEqual: semver.MustParse("1.11.0-alpha.0"),
		LessThanOrEqual:    semver.MustParse("1.13.1000"),
	},
	{
		Option: config.ExtraOption{
			Component: Apiserver,
			Key:       "enable-admission-plugins",
			Value:     strings.Join(util.DefaultV114AdmissionControllers, ","),
		},
		GreaterThanOrEqual: semver.MustParse("1.14.0-alpha.0"),
	},

	{
		Option: config.ExtraOption{
			Component: Kubelet,
			Key:       "cadvisor-port",
			Value:     "0",
		},
		LessThanOrEqual: semver.MustParse("1.11.1000"),
	},
}
