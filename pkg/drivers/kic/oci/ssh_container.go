/*
Copyright 2024 The Kubernetes Authors All rights reserved.

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

package oci

import (
	"fmt"
	"strings"
	"sync"

	"k8s.io/klog/v2"
)

var (
	// containerSSHPorts tracks SSH port mappings for containers
	containerSSHPorts = make(map[string]int)
	containerPortsMux sync.RWMutex
)

// EnsureContainerSSHAccess ensures SSH access to a container for remote Docker contexts
func EnsureContainerSSHAccess(containerName string) error {
	if IsRemoteDockerContext() && IsSSHDockerContext() {
		klog.Infof("Setting up SSH access for container %s on remote Docker context", containerName)
		
		// Get the forwarded SSH port on the remote host
		remotePort, err := ForwardedPort(Docker, containerName, 22)
		if err != nil {
			return fmt.Errorf("failed to get SSH port for container %s: %w", containerName, err)
		}
		
		// For SSH-based remote contexts, we need to create a local tunnel
		// from a local port to the container's SSH port on the remote host
		ctx, err := GetCurrentContext()
		if err != nil {
			return fmt.Errorf("failed to get Docker context: %w", err)
		}
		
		// Create SSH tunnel: local port -> remote host -> container SSH port
		tm := GetTunnelManager()
		tunnelKey := fmt.Sprintf("container-ssh-%s", containerName)
		
		// Check if tunnel already exists
		if existingTunnels := tm.GetActiveTunnels(); existingTunnels != nil {
			for key, tunnel := range existingTunnels {
				if key == tunnelKey && tunnel.Status == "active" {
					klog.Infof("SSH tunnel already exists for container %s on port %d", containerName, tunnel.LocalPort)
					containerPortsMux.Lock()
					containerSSHPorts[containerName] = tunnel.LocalPort
					containerPortsMux.Unlock()
					return nil
				}
			}
		}
		
		// Create new tunnel
		tunnel, err := tm.CreateContainerSSHTunnel(ctx, containerName, remotePort)
		if err != nil {
			return fmt.Errorf("failed to create SSH tunnel for container %s: %w", containerName, err)
		}
		
		containerPortsMux.Lock()
		containerSSHPorts[containerName] = tunnel.LocalPort
		containerPortsMux.Unlock()
		
		klog.Infof("Created SSH tunnel for container %s: localhost:%d -> remote:%d", 
			containerName, tunnel.LocalPort, remotePort)
	}
	return nil
}

// GetContainerSSHPort returns the SSH port for a container
func GetContainerSSHPort(containerName string) (int, error) {
	containerPortsMux.RLock()
	port, exists := containerSSHPorts[containerName]
	containerPortsMux.RUnlock()
	
	if exists {
		klog.V(3).Infof("Using cached SSH port %d for container %s", port, containerName)
		return port, nil
	}
	
	// For remote SSH contexts, ensure the tunnel is created
	if IsRemoteDockerContext() && IsSSHDockerContext() {
		klog.Infof("Container %s SSH port not cached, ensuring SSH access", containerName)
		if err := EnsureContainerSSHAccess(containerName); err != nil {
			return 0, fmt.Errorf("failed to ensure SSH access for container %s: %w", containerName, err)
		}
		
		// Try again from cache
		containerPortsMux.RLock()
		port, exists = containerSSHPorts[containerName]
		containerPortsMux.RUnlock()
		
		if exists {
			return port, nil
		}
	}
	
	// For local contexts, return the direct port
	port, err := ForwardedPort(Docker, containerName, 22)
	if err != nil {
		return 0, fmt.Errorf("failed to get SSH port for container %s: %w", containerName, err)
	}
	
	return port, nil
}

// CleanupContainerTunnels cleans up any SSH tunnels for containers
func CleanupContainerTunnels() {
	// Clean up all container SSH tunnels
	tm := GetTunnelManager()
	activeTunnels := tm.GetActiveTunnels()
	
	for key := range activeTunnels {
		if strings.HasPrefix(key, "container-ssh-") {
			if err := tm.StopTunnel(key); err != nil {
				klog.Warningf("Failed to stop tunnel %s: %v", key, err)
			}
		}
	}
	
	// Clear the port cache
	containerPortsMux.Lock()
	containerSSHPorts = make(map[string]int)
	containerPortsMux.Unlock()
	
	klog.Info("Cleaned up container SSH tunnels and port mappings")
}