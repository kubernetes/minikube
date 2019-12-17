/*
Copyright 2019 The Kubernetes Authors All rights reserved.

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

import (
	"sync"
)

// this is a separate struct so we can more easily ensure that this portion is
// thread safe
type nodeCache struct {
	mu                sync.RWMutex
	kubernetesVersion string
	ipv4              string
	ipv6              string
	ports             map[int32]int32
	role              string
}

func (cache *nodeCache) set(setter func(*nodeCache)) {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	setter(cache)
}

func (cache *nodeCache) KubeVersion() string {
	cache.mu.RLock()
	defer cache.mu.RUnlock()
	return cache.kubernetesVersion
}

func (cache *nodeCache) IP() (string, string) {
	cache.mu.RLock()
	defer cache.mu.RUnlock()
	return cache.ipv4, cache.ipv6
}

func (cache *nodeCache) HostPort(p int32) (int32, bool) {
	cache.mu.RLock()
	defer cache.mu.RUnlock()
	if cache.ports == nil {
		return 0, false
	}
	v, ok := cache.ports[p]
	return v, ok
}

func (cache *nodeCache) Role() string {
	cache.mu.RLock()
	defer cache.mu.RUnlock()
	return cache.role
}

func (n *Node) String() string {
	return n.name
}

// Name returns the node's name
func (n *Node) Name() string {
	return n.name
}
