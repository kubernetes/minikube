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
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/vmpath"
	"k8s.io/minikube/pkg/util"
)

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

	config.NewUnversionedOption(Kubelet, "bootstrap-kubeconfig", "/etc/kubernetes/bootstrap-kubelet.conf"),
	config.NewUnversionedOption(Kubelet, "config", "/var/lib/kubelet/config.yaml"),
	config.NewUnversionedOption(Kubelet, "kubeconfig", "/etc/kubernetes/kubelet.conf"),
	{
		Option: config.ExtraOption{
			Component: Kubelet,
			Key:       "require-kubeconfig",
			Value:     "true",
		},
		LessThanOrEqual: semver.MustParse("1.9.10"),
	},

	{
		Option: config.ExtraOption{
			Component: Kubelet,
			Key:       "allow-privileged",
			Value:     "true",
		},
		LessThanOrEqual: semver.MustParse("1.15.0-alpha.3"),
	},

	// before 1.16.0-beta.2, kubeadm bug did not allow overriding this via config file, so this has
	// to be passed in as a kubelet flag. See https://github.com/kubernetes/kubernetes/pull/81903 for more details.
	{
		Option: config.ExtraOption{
			Component: Kubelet,
			Key:       "client-ca-file",
			Value:     path.Join(vmpath.GuestKubernetesCertsDir, "ca.crt"),
		},
		LessThanOrEqual: semver.MustParse("1.16.0-beta.1"),
	},

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
