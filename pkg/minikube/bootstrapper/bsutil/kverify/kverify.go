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

// Package kverify verifies a running Kubernetes cluster is healthy
package kverify

import (
	"time"
)

// minLogCheckTime how long to wait before spamming error logs to console
const minLogCheckTime = 60 * time.Second

const (
	// APIServerWaitKey is the name used in the flags for k8s api server
	APIServerWaitKey = "apiserver"
	// SystemPodsWaitKey is the name used in the flags for pods in the kube system
	SystemPodsWaitKey = "system_pods"
	// DefaultSAWaitKey is the name used in the flags for default service account
	DefaultSAWaitKey = "default_sa"
	// AppsRunningKey is the name used in the flags for waiting for k8s-apps to be running
	AppsRunningKey = "apps_running"
	// NodeReadyKey is the name used in the flags for waiting for the node status to be ready
	NodeReadyKey = "node_ready"
	// KubeletKey is the name used in the flags for waiting for the kubelet status to be ready
	KubeletKey = "kubelet"
	// ExtraKey is the name used for extra waiting for pods in CorePodsList to be Ready
	ExtraKey = "extra"
)

//  vars related to the --wait flag
var (
	// DefaultComponents is map of the the default components to wait for
	DefaultComponents = map[string]bool{APIServerWaitKey: true, SystemPodsWaitKey: true}
	// NoWaitComponents is map of componets to wait for if specified 'none' or 'false'
	NoComponents = map[string]bool{APIServerWaitKey: false, SystemPodsWaitKey: false, DefaultSAWaitKey: false, AppsRunningKey: false, NodeReadyKey: false, KubeletKey: false, ExtraKey: false}
	// AllComponents is map for waiting for all components.
	AllComponents = map[string]bool{APIServerWaitKey: true, SystemPodsWaitKey: true, DefaultSAWaitKey: true, AppsRunningKey: true, NodeReadyKey: true, KubeletKey: true, ExtraKey: true}
	// DefaultWaitList is list of all default components to wait for. only names to be used for start flags.
	DefaultWaitList = []string{APIServerWaitKey, SystemPodsWaitKey}
	// AllComponentsList list of all valid components keys to wait for. only names to be used used for start flags.
	AllComponentsList = []string{APIServerWaitKey, SystemPodsWaitKey, DefaultSAWaitKey, AppsRunningKey, NodeReadyKey, KubeletKey}
	// AppsRunningList running list are valid k8s-app components to wait for them to be running
	AppsRunningList = []string{
		"kube-dns", // coredns
		"etcd",
		"kube-apiserver",
		"kube-controller-manager",
		"kube-proxy",
		"kube-scheduler",
	}
	// CorePodsList is a list of essential pods for running kurnetes to extra wait for them to be Ready
	CorePodsList = []string{
		"kube-dns", // coredns
		"etcd",
		"kube-apiserver",
		"kube-controller-manager",
		"kube-proxy",
		"kube-scheduler",
	}
)

// ShouldWait will return true if the config says need to wait
func ShouldWait(wcs map[string]bool) bool {
	for _, c := range AllComponentsList {
		if wcs[c] {
			return true
		}
	}
	return false
}
