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
	Join() error

	// LabelAndUntaint applies labels and removes taints after join is complete.
	// Preconditions: node registered with apiserver
	// Side effects: updates node labels and taints in etcd
	LabelAndUntaint() error
}
