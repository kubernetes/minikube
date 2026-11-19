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

package doctor

import (
	"fmt"
	"strings"

	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
)

// These mirror minUsableMem, minRecommendedMem and minimumCPUS in
// cmd/minikube/cmd/start_flags.go, which `minikube start` enforces.
const (
	minUsableMemMB      = 1800
	minRecommendedMemMB = 1900
	minCPUs             = 2
)

func profileValidation(h Host) Check {
	c := Check{Section: SectionConfig, Name: "Profiles"}
	_, invalid, err := h.ListProfiles()
	if err != nil {
		c.Status, c.Message, c.Details = Fail, "cannot list profiles", err.Error()
		return c
	}
	if len(invalid) == 0 {
		c.Status, c.Message = Pass, "all profile configurations are valid"
		return c
	}
	var names []string
	for _, p := range invalid {
		names = append(names, p.Name)
	}
	// Other profiles being broken does not stop this one from working.
	c.Status = Warn
	c.Message = fmt.Sprintf("invalid profile configuration: %s", strings.Join(names, ", "))
	c.Fix = "Delete the invalid profile with: minikube delete -p <profile>"
	return c
}

func loadProfile(profile string, h Host) (*config.ClusterConfig, Check) {
	c := Check{Section: SectionConfig, Name: "Profile"}
	cc, err := h.LoadConfig(profile)
	switch {
	case err == nil:
		c.Status, c.Message = Pass, fmt.Sprintf("%q loaded", profile)
		return cc, c
	case config.IsNotExist(err):
		c.Status = Fail
		c.Message = fmt.Sprintf("profile %q does not exist", profile)
		c.Fix = fmt.Sprintf("Create it with: minikube start -p %s", profile)
	default:
		c.Status = Fail
		c.Message = fmt.Sprintf("profile %q cannot be loaded", profile)
		c.Details = err.Error()
		c.Fix = fmt.Sprintf("Recreate it with: minikube delete -p %s && minikube start -p %s", profile, profile)
	}
	return nil, c
}

func kubectl(h Host) Check {
	c := Check{Section: SectionHost, Name: "kubectl"}
	if _, err := h.LookPath("kubectl"); err != nil {
		c.Status = Warn
		c.Message = "kubectl is not on PATH"
		c.Fix = "Use `minikube kubectl -- <args>`, or install kubectl: https://kubernetes.io/docs/tasks/tools/"
		return c
	}
	c.Status, c.Message = Pass, "kubectl is on PATH"
	return c
}

// driver uses the driver's own status checker, the one `minikube start`
// uses, so each driver's binaries and daemons are checked correctly.
func driver(name string, h Host) Check {
	c := Check{Section: SectionHost, Name: "Driver"}
	st, known := h.DriverStatus(name)
	switch {
	case !known:
		c.Status = Fail
		c.Message = fmt.Sprintf("driver %q is not supported by this minikube binary", name)
		c.Fix = "See `minikube start --help` for supported drivers"
	case !st.Installed:
		c.Status = Fail
		c.Message = fmt.Sprintf("%s is not installed", name)
		c.Fix = st.Fix
		if st.Doc != "" {
			c.Details = "Documentation: " + st.Doc
		}
	case !st.Healthy:
		c.Status = Fail
		c.Message = fmt.Sprintf("%s is installed but not healthy", name)
		if st.Error != nil {
			c.Details = st.Error.Error()
		}
		c.Fix = st.Fix
		if c.Fix == "" {
			c.Fix = fmt.Sprintf("Start %s, then run `minikube doctor` again", name)
		}
	case st.NeedsImprovement:
		c.Status = Warn
		c.Message = fmt.Sprintf("%s works but could be improved", name)
		c.Fix = st.Fix
	default:
		c.Status, c.Message = Pass, name+" is healthy"
		if st.Version != "" {
			c.Message = fmt.Sprintf("%s %s is healthy", name, st.Version)
		}
	}
	return c
}

// clusterChecks queries node status once and derives every cluster check
// from it.
func clusterChecks(cc *config.ClusterConfig, h Host) []Check {
	statuses, err := h.ClusterStatus(cc)
	host := Check{Section: SectionCluster, Name: "Cluster status"}
	if err != nil || len(statuses) == 0 {
		host.Status, host.Message = Fail, "cluster status is unavailable"
		host.Fix = fmt.Sprintf("Start it with: minikube start -p %s", cc.Name)
		if err != nil {
			host.Details = firstLine(err.Error())
			if strings.Contains(err.Error(), "No such container") || strings.Contains(err.Error(), "no such container") {
				host.Message = "the cluster's container no longer exists (it was removed outside minikube)"
				host.Fix = fmt.Sprintf("Recreate it with: minikube delete -p %s && minikube start -p %s", cc.Name, cc.Name)
			}
		}
		return []Check{host,
			skipped(SectionCluster, "API server", "cluster status unavailable"),
			skipped(SectionCluster, "Nodes", "cluster status unavailable"),
			skipped(SectionCluster, "Kubeconfig", "cluster status unavailable")}
	}

	cp := statuses[0]
	if cp.Host != "Running" {
		host.Status, host.Message = Fail, fmt.Sprintf("cluster is %s", strings.ToLower(cp.Host))
		host.Fix = fmt.Sprintf("Start it with: minikube start -p %s", cc.Name)
		return []Check{host,
			skipped(SectionCluster, "API server", "cluster is not running"),
			skipped(SectionCluster, "Nodes", "cluster is not running"),
			skipped(SectionCluster, "Kubeconfig", "cluster is not running")}
	}
	host.Status, host.Message = Pass, "running"
	checks := []Check{host}

	if cc.KubernetesConfig.KubernetesVersion == constants.NoKubernetesVersion {
		for _, name := range []string{"API server", "Nodes", "Kubeconfig"} {
			checks = append(checks, skipped(SectionCluster, name, "Kubernetes is disabled (--no-kubernetes)"))
		}
		return checks
	}

	api := Check{Section: SectionCluster, Name: "API server"}
	if cp.APIServer == "Running" {
		api.Status, api.Message = Pass, "running"
	} else {
		api.Status, api.Message = Fail, fmt.Sprintf("API server is %s", strings.ToLower(cp.APIServer))
		api.Fix = fmt.Sprintf("Restart the cluster with: minikube start -p %s", cc.Name)
	}

	nodes := Check{Section: SectionCluster, Name: "Nodes"}
	var notReady []string
	for _, st := range statuses {
		if st.Kubelet != "Running" && st.Kubelet != cluster.Irrelevant {
			notReady = append(notReady, fmt.Sprintf("%s (kubelet %s)", st.Name, strings.ToLower(st.Kubelet)))
		}
	}
	if len(notReady) == 0 {
		nodes.Status, nodes.Message = Pass, fmt.Sprintf("%d node(s), kubelet running on all", len(statuses))
	} else {
		nodes.Status, nodes.Message = Fail, "kubelet not running on "+strings.Join(notReady, ", ")
		nodes.Fix = fmt.Sprintf("Check the logs with: minikube logs -p %s", cc.Name)
	}

	kc := Check{Section: SectionCluster, Name: "Kubeconfig"}
	switch cp.Kubeconfig {
	case cluster.Configured:
		// Configured means the profile's kubeconfig entry has the right
		// endpoint; kubectl may still be using another context.
		kc.Status, kc.Message = Pass, "kubeconfig entry points at the cluster"
		if h.CurrentContext != nil {
			if cur, err := h.CurrentContext(); err == nil && cur != cc.Name {
				kc.Status = Warn
				kc.Message = fmt.Sprintf("kubeconfig entry is correct, but kubectl is using context %q", cur)
				kc.Fix = fmt.Sprintf("Switch with: kubectl config use-context %s", cc.Name)
			}
		}
	case cluster.Irrelevant:
		kc = skipped(SectionCluster, "Kubeconfig", "not applicable")
	default:
		kc.Status, kc.Message = Fail, fmt.Sprintf("kubeconfig is %s", strings.ToLower(cp.Kubeconfig))
		kc.Fix = fmt.Sprintf("Fix it with: minikube update-context -p %s", cc.Name)
	}
	return append(checks, api, nodes, kc)
}

// resources checks the values `minikube start` enforces. Zero means the
// profile was started with --cpus/--memory=no-limit.
func resources(cc *config.ClusterConfig) []Check {
	cpu := Check{Section: SectionResources, Name: "CPUs"}
	switch {
	case cc.CPUs == 0:
		cpu.Status, cpu.Message = Pass, "no limit"
	case cc.CPUs < minCPUs:
		cpu.Status = Fail
		cpu.Message = fmt.Sprintf("%d CPU allocated; Kubernetes needs at least %d", cc.CPUs, minCPUs)
		cpu.Fix = fmt.Sprintf("Recreate the cluster: minikube delete -p %s && minikube start -p %s --cpus=%d", cc.Name, cc.Name, minCPUs)
	default:
		cpu.Status, cpu.Message = Pass, fmt.Sprintf("%d CPUs", cc.CPUs)
	}

	mem := Check{Section: SectionResources, Name: "Memory"}
	fix := fmt.Sprintf("Recreate the cluster: minikube delete -p %s && minikube start -p %s --memory=%dmb", cc.Name, cc.Name, minRecommendedMemMB)
	switch {
	case cc.Memory == 0:
		mem.Status, mem.Message = Pass, "no limit"
	case cc.Memory < minUsableMemMB:
		mem.Status, mem.Fix = Fail, fix
		mem.Message = fmt.Sprintf("%d MB allocated; Kubernetes needs at least %d MB", cc.Memory, minUsableMemMB)
	case cc.Memory < minRecommendedMemMB:
		mem.Status, mem.Fix = Warn, fix
		mem.Message = fmt.Sprintf("%d MB allocated; %d MB or more is recommended", cc.Memory, minRecommendedMemMB)
	default:
		mem.Status, mem.Message = Pass, fmt.Sprintf("%d MB", cc.Memory)
	}
	return []Check{cpu, mem}
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return strings.TrimSpace(s[:i])
	}
	return s
}
