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

package node

// Provisioner encapsulates OS-specific node lifecycle operations using a strategy pattern.
// Different OS implementations (Linux, Windows) use the same interface but with different
// join mechanisms and readiness verification strategies.
// This eliminates scattered OS conditionals from shared orchestration code.
//
// Note: This interface is minimal and grows incrementally as new provisioner implementations
// are added. The complete lifecycle will eventually include PreBootstrap and PostJoin methods,
// but they are added only when a second provisioner (Windows) needs them. This follows the
// principle of not building speculative abstractions.
type Provisioner interface {
	// Join generates the kubeadm join command and executes it to add the node to the cluster.
	// Handles retry logic with OS-specific failure recovery (e.g., kubeadm reset for Linux).
	// Preconditions: kubeadm binary present, network connectivity, cluster running
	// Side effects: joins node to cluster, writes kubeadm state to node
	//
	// Linux: Executes kubeadm join with exponential backoff retry. On failure, runs kubeadm reset
	// to clean up state before retrying. Recovery mechanism is synchronous and fast.
	//
	// Windows: Executes kubeadm join similarly but does not perform kubeadm reset on failure.
	// Windows nodes require additional API server registration time (handled by retry in LabelAndUntaint).
	Join() error

	// LabelAndUntaint applies labels and removes taints after join is complete.
	// Preconditions: node registered with apiserver
	// Side effects: updates node labels and taints in etcd
	//
	// Linux: Single attempt to apply labels/taints. Node should be registered immediately
	// after kubeadm join returns.
	//
	// Windows: Retries labeling for up to 3 minutes as Windows nodes take extra time to register
	// with the API server after kubeadm join completes.
	LabelAndUntaint() error

	// PostJoin performs OS-specific operations after node successfully joins the cluster.
	// Runs after LabelAndUntaint() completes. Used for CNI and network setup that must happen
	// after the node is labeled and integrated into the cluster.
	// Preconditions: node is labeled and registered with apiserver
	// Side effects: applies CNI manifests, configures network plugins
	//
	// Linux: No operation needed. CNI is configured externally or handled by control-plane addons.
	//
	// Windows: Applies Windows-specific CNI manifests (e.g., Flannel-Windows DaemonSet,
	// kube-proxy-windows DaemonSet) that are bundled in the minikube binary. These cannot
	// run on Linux nodes and must be applied only to Windows workers.
	// PostJoin() error
}
