/*
Copyright 2026 The Kubernetes Authors All rights reserved.

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

// Package doctor runs read-only diagnostics on a minikube profile and the
// host it depends on, and reports each problem with a suggested fix.
package doctor

import (
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/registry"
)

// Status is the outcome of a single check.
type Status string

const (
	// Pass means the check found nothing wrong.
	Pass Status = "PASS"
	// Warn means minikube can work, but something should be improved.
	Warn Status = "WARNING"
	// Fail means minikube will not work until this is fixed.
	Fail Status = "FAIL"
	// Skip means the check could not run because an earlier check failed.
	Skip Status = "SKIPPED"
)

// Section groups checks in the report.
type Section string

const (
	// SectionConfig covers the profile configuration.
	SectionConfig Section = "Configuration"
	// SectionHost covers tools and services on the host.
	SectionHost Section = "Host"
	// SectionCluster covers the running cluster.
	SectionCluster Section = "Cluster"
	// SectionResources covers resources allocated to the cluster.
	SectionResources Section = "Resources"
)

// Check is the result of one diagnostic.
type Check struct {
	Section Section `json:"section"`
	Name    string  `json:"name"`
	Status  Status  `json:"status"`
	Message string  `json:"message"`
	Details string  `json:"details,omitempty"`
	Fix     string  `json:"fix,omitempty"`
}

// Report is the result of running all checks for one profile.
type Report struct {
	Profile    string  `json:"profile"`
	Driver     string  `json:"driver,omitempty"`
	Runtime    string  `json:"containerRuntime,omitempty"`
	Kubernetes string  `json:"kubernetesVersion,omitempty"`
	Checks     []Check `json:"checks"`
}

// Count returns how many checks ended with status s.
func (r Report) Count(s Status) int {
	n := 0
	for _, c := range r.Checks {
		if c.Status == s {
			n++
		}
	}
	return n
}

// Host is everything the checks read from minikube and the host. The
// command wires it to the real implementations; tests replace it.
type Host struct {
	// LoadConfig loads the profile's cluster configuration.
	LoadConfig func(profile string) (*config.ClusterConfig, error)
	// ListProfiles returns the valid and invalid profiles.
	ListProfiles func() (valid, invalid []*config.Profile, err error)
	// DriverStatus reports whether a driver is installed and healthy.
	DriverStatus func(driver string) (registry.State, bool)
	// ClusterStatus reports the status of every node.
	ClusterStatus func(cc *config.ClusterConfig) ([]*cluster.Status, error)
	// LookPath finds an executable on PATH.
	LookPath func(name string) (string, error)
	// CurrentContext returns kubectl's current context.
	CurrentContext func() (string, error)
}

// Run checks the profile and returns the report. It never exits: a
// missing or broken profile is reported as a failed check.
func Run(profile string, h Host) Report {
	r := Report{Profile: profile}

	r.Checks = append(r.Checks, profileValidation(h))

	cc, check := loadProfile(profile, h)
	r.Checks = append(r.Checks, check)
	r.Checks = append(r.Checks, kubectl(h))
	if cc == nil {
		for _, name := range []string{"Driver", "Cluster status", "API server", "Nodes", "Kubeconfig"} {
			sec := SectionCluster
			if name == "Driver" {
				sec = SectionHost
			}
			r.Checks = append(r.Checks, skipped(sec, name, "profile not loaded"))
		}
		return r
	}

	r.Driver = cc.Driver
	r.Runtime = cc.KubernetesConfig.ContainerRuntime
	r.Kubernetes = cc.KubernetesConfig.KubernetesVersion

	drv := driver(cc.Driver, h)
	r.Checks = append(r.Checks, drv)
	if drv.Status == Fail {
		for _, name := range []string{"Cluster status", "API server", "Nodes", "Kubeconfig"} {
			r.Checks = append(r.Checks, skipped(SectionCluster, name, "driver "+cc.Driver+" is not usable"))
		}
	} else {
		r.Checks = append(r.Checks, clusterChecks(cc, h)...)
	}
	r.Checks = append(r.Checks, resources(cc)...)
	return r
}

func skipped(sec Section, name, why string) Check {
	return Check{Section: sec, Name: name, Status: Skip, Message: "skipped: " + why}
}
