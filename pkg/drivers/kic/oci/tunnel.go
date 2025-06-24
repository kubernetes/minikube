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
	"context"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"k8s.io/klog/v2"
)

// TunnelManager manages SSH tunnels for remote Docker contexts
type TunnelManager struct {
	tunnels map[string]*SSHTunnel
	mutex   sync.RWMutex
}

// SSHTunnel represents an active SSH tunnel
type SSHTunnel struct {
	LocalPort  int
	RemoteHost string
	RemotePort int
	SSHHost    string
	SSHUser    string
	SSHPort    int
	Process    *exec.Cmd
	Cancel     context.CancelFunc
	Status     string
	Metrics    *TunnelMetrics
}

// TunnelMetrics tracks tunnel performance and reliability metrics
type TunnelMetrics struct {
	CreatedAt           time.Time
	LastHealthCheck     time.Time
	LastSuccessfulCheck time.Time
	TotalHealthChecks   int64
	FailedHealthChecks  int64
	RestartCount        int64
	AvgLatency          time.Duration
	IPv4Available       bool
	IPv6Available       bool
	LastError           string
	UptimeSeconds       int64
	mutex               sync.RWMutex
}

var (
	globalTunnelManager *TunnelManager
	tunnelManagerOnce   sync.Once
)

// GetTunnelManager returns the global tunnel manager instance
func GetTunnelManager() *TunnelManager {
	tunnelManagerOnce.Do(func() {
		globalTunnelManager = &TunnelManager{
			tunnels: make(map[string]*SSHTunnel),
		}
	})
	return globalTunnelManager
}

// CreateAPIServerTunnel creates an SSH tunnel for API server access
func (tm *TunnelManager) CreateAPIServerTunnel(ctx *ContextInfo, remotePort int) (*SSHTunnel, error) {
	if !ctx.IsSSH {
		return nil, errors.New("context is not SSH-based")
	}

	u, err := url.Parse(ctx.Host)
	if err != nil {
		return nil, errors.Wrapf(err, "parsing SSH host %q", ctx.Host)
	}

	sshUser := u.User.Username()
	sshHost := u.Hostname()
	sshPort := 22
	if u.Port() != "" {
		sshPort, _ = strconv.Atoi(u.Port())
	}

	// Find available local port
	localPort, err := findAvailablePort()
	if err != nil {
		return nil, errors.Wrap(err, "finding available local port")
	}

	tunnelKey := fmt.Sprintf("%s:%d->%s:%d", sshHost, remotePort, "localhost", localPort)

	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	// Check if tunnel already exists
	if existing, exists := tm.tunnels[tunnelKey]; exists {
		if existing.Status == "active" {
			klog.Infof("SSH tunnel already active: %s", tunnelKey)
			return existing, nil
		}
		// Clean up stale tunnel
		tm.cleanupTunnel(existing)
		delete(tm.tunnels, tunnelKey)
	}

	tunnel := &SSHTunnel{
		LocalPort:  localPort,
		RemoteHost: "localhost", // API server runs on localhost inside the remote container
		RemotePort: remotePort,
		SSHHost:    sshHost,
		SSHUser:    sshUser,
		SSHPort:    sshPort,
		Status:     "starting",
		Metrics: &TunnelMetrics{
			CreatedAt: time.Now(),
		},
	}

	// Pre-flight SSH connectivity check
	if err := tm.checkSSHConnectivity(tunnel); err != nil {
		return nil, errors.Wrapf(err, "SSH connectivity pre-check failed for %s", tunnelKey)
	}

	// Start tunnel with retry logic
	if err := tm.startTunnelWithRetry(tunnel, 3); err != nil {
		return nil, errors.Wrapf(err, "starting SSH tunnel %s", tunnelKey)
	}

	tm.tunnels[tunnelKey] = tunnel
	klog.Infof("SSH tunnel created: %s", tunnelKey)

	return tunnel, nil
}

// startTunnelWithRetry starts the SSH tunnel process with retry logic
func (tm *TunnelManager) startTunnelWithRetry(tunnel *SSHTunnel, maxRetries int) error {
	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		klog.V(3).Infof("Starting SSH tunnel (attempt %d/%d)", attempt, maxRetries)
		
		if err := tm.startTunnel(tunnel); err != nil {
			lastErr = err
			klog.Warningf("SSH tunnel start attempt %d failed: %v", attempt, err)
			
			if attempt < maxRetries {
				// Clean up failed attempt
				tm.cleanupTunnel(tunnel)
				
				// Wait before retry with exponential backoff
				backoffDuration := time.Duration(attempt) * 500 * time.Millisecond
				klog.V(3).Infof("Retrying SSH tunnel creation in %v", backoffDuration)
				time.Sleep(backoffDuration)
			}
			continue
		}
		
		klog.Infof("SSH tunnel started successfully on attempt %d", attempt)
		return nil
	}
	
	return errors.Wrapf(lastErr, "failed to start SSH tunnel after %d attempts", maxRetries)
}

// startTunnel starts the SSH tunnel process
func (tm *TunnelManager) startTunnel(tunnel *SSHTunnel) error {
	ctx, cancel := context.WithCancel(context.Background())
	tunnel.Cancel = cancel

	// Build SSH command with both IPv4 and IPv6 bindings
	sshArgs := []string{
		// Bind to both IPv4 (127.0.0.1) and IPv6 (::1) localhost
		"-L", fmt.Sprintf("127.0.0.1:%d:%s:%d", tunnel.LocalPort, tunnel.RemoteHost, tunnel.RemotePort),
		"-L", fmt.Sprintf("[::1]:%d:%s:%d", tunnel.LocalPort, tunnel.RemoteHost, tunnel.RemotePort),
		"-N", // Don't execute remote command
		"-f", // Run in background
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "LogLevel=ERROR",
		"-o", "ServerAliveInterval=30",
		"-o", "ServerAliveCountMax=3",
		"-o", "ExitOnForwardFailure=yes",
		"-o", "ControlMaster=auto",
		"-o", "ControlPath=/tmp/minikube-ssh-%r@%h:%p",
		"-o", "ControlPersist=600", // Keep connection alive for 10 minutes
		fmt.Sprintf("%s@%s", tunnel.SSHUser, tunnel.SSHHost),
	}

	if tunnel.SSHPort != 22 {
		sshArgs = append([]string{"-p", strconv.Itoa(tunnel.SSHPort)}, sshArgs...)
	}

	klog.V(3).Infof("Starting SSH tunnel: ssh %s", strings.Join(sshArgs, " "))

	cmd := exec.CommandContext(ctx, "ssh", sshArgs...)
	cmd.Stderr = os.Stderr

	tunnel.Process = cmd

	if err := cmd.Start(); err != nil {
		cancel()
		return errors.Wrap(err, "starting SSH process")
	}

	// Wait for tunnel to be ready with extended timeout
	if err := tm.waitForTunnel(tunnel, 10*time.Second); err != nil {
		cancel()
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		return errors.Wrap(err, "waiting for tunnel to be ready")
	}

	tunnel.Status = "active"

	// Monitor tunnel in background
	go tm.monitorTunnel(tunnel)

	return nil
}

// waitForTunnel waits for the SSH tunnel to be ready
func (tm *TunnelManager) waitForTunnel(tunnel *SSHTunnel, timeout time.Duration) error {
	klog.V(3).Infof("Waiting for SSH tunnel to be ready (timeout: %v)", timeout)
	deadline := time.Now().Add(timeout)
	attempts := 0

	for time.Now().Before(deadline) {
		attempts++
		
		// Check if the process is still running
		if tunnel.Process != nil && tunnel.Process.Process != nil {
			// Check if process has exited
			if tunnel.Process.ProcessState != nil && tunnel.Process.ProcessState.Exited() {
				return errors.Errorf("SSH process exited unexpectedly: %v", tunnel.Process.ProcessState.String())
			}
		}
		
		// Try to connect to the local tunnel port (IPv4 first, then IPv6)
		ipv4Addr := fmt.Sprintf("127.0.0.1:%d", tunnel.LocalPort)
		ipv6Addr := fmt.Sprintf("[::1]:%d", tunnel.LocalPort)
		
		// Test IPv4 connectivity
		conn4, err4 := net.DialTimeout("tcp", ipv4Addr, 2*time.Second)
		ipv4Ready := err4 == nil
		if ipv4Ready {
			conn4.Close()
		}
		
		// Test IPv6 connectivity
		conn6, err6 := net.DialTimeout("tcp", ipv6Addr, 2*time.Second)
		ipv6Ready := err6 == nil
		if ipv6Ready {
			conn6.Close()
		}
		
		// Tunnel is ready if at least one protocol is working
		if ipv4Ready || ipv6Ready {
			klog.V(3).Infof("SSH tunnel ready after %d attempts in %v (IPv4: %v, IPv6: %v)", 
				attempts, time.Since(deadline.Add(-timeout)), ipv4Ready, ipv6Ready)
			return nil
		}
		
		if attempts%20 == 0 { // Log every 1 second (50ms * 20)
			klog.V(3).Infof("SSH tunnel not ready yet (attempt %d): IPv4 err=%v, IPv6 err=%v", attempts, err4, err6)
		}
		
		time.Sleep(50 * time.Millisecond)
	}

	return errors.Errorf("tunnel did not become ready within %v (after %d attempts)", timeout, attempts)
}

// monitorTunnel monitors the tunnel process and restarts if needed
func (tm *TunnelManager) monitorTunnel(tunnel *SSHTunnel) {
	defer func() {
		tm.mutex.Lock()
		tunnel.Status = "stopped"
		tm.mutex.Unlock()
	}()

	// Start health checking in a separate goroutine
	go tm.healthCheckTunnel(tunnel)

	if err := tunnel.Process.Wait(); err != nil {
		klog.Warningf("SSH tunnel process exited with error: %v", err)
		
		// Attempt auto-recovery if the tunnel was active
		if tunnel.Status == "active" {
			klog.Infof("Attempting to restart SSH tunnel...")
			if err := tm.restartTunnel(tunnel); err != nil {
				klog.Errorf("Failed to restart SSH tunnel: %v", err)
			}
		}
	} else {
		klog.Infof("SSH tunnel process exited cleanly")
	}
}

// healthCheckTunnel periodically checks tunnel health
func (tm *TunnelManager) healthCheckTunnel(tunnel *SSHTunnel) {
	ticker := time.NewTicker(30 * time.Second)
	metricsTicker := time.NewTicker(5 * time.Minute) // Log metrics every 5 minutes
	defer ticker.Stop()
	defer metricsTicker.Stop()

	for {
		select {
		case <-ticker.C:
			if tunnel.Status != "active" {
				return // Tunnel is no longer active
			}

			// Perform comprehensive health check
			if err := tm.performHealthCheck(tunnel); err != nil {
				klog.Warningf("SSH tunnel health check failed: %v", err)
				tunnel.Status = "unhealthy"
			} else {
				if tunnel.Status == "unhealthy" {
					klog.Infof("SSH tunnel health restored")
				}
				tunnel.Status = "active"
			}

		case <-metricsTicker.C:
			// Log comprehensive tunnel metrics periodically
			tm.LogTunnelMetrics()

		case <-time.After(15 * time.Minute):
			// Stop health checking after 15 minutes (extended from 5 minutes)
			klog.V(3).Infof("Stopping health monitoring for tunnel after 15 minutes")
			return
		}
	}
}

// restartTunnel attempts to restart a failed tunnel
func (tm *TunnelManager) restartTunnel(tunnel *SSHTunnel) error {
	klog.Infof("Restarting SSH tunnel to %s:%d", tunnel.SSHHost, tunnel.RemotePort)
	
	// Update metrics
	if tunnel.Metrics != nil {
		tunnel.Metrics.mutex.Lock()
		tunnel.Metrics.RestartCount++
		tunnel.Metrics.mutex.Unlock()
	}
	
	// Clean up the old process
	tm.cleanupTunnel(tunnel)
	
	// Wait a moment before restarting
	time.Sleep(1 * time.Second)
	
	// Reset tunnel status
	tunnel.Status = "restarting"
	
	// Start the tunnel again with retry logic
	return tm.startTunnelWithRetry(tunnel, 2) // Use fewer retries for restart
}

// cleanupTunnel cleans up a tunnel
func (tm *TunnelManager) cleanupTunnel(tunnel *SSHTunnel) {
	if tunnel.Cancel != nil {
		tunnel.Cancel()
	}
	if tunnel.Process != nil && tunnel.Process.Process != nil {
		tunnel.Process.Process.Kill()
	}
}

// StopTunnel stops a specific tunnel
func (tm *TunnelManager) StopTunnel(tunnelKey string) error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	tunnel, exists := tm.tunnels[tunnelKey]
	if !exists {
		return errors.Errorf("tunnel %s not found", tunnelKey)
	}

	tm.cleanupTunnel(tunnel)
	delete(tm.tunnels, tunnelKey)

	klog.Infof("SSH tunnel stopped: %s", tunnelKey)
	return nil
}

// StopAllTunnels stops all active tunnels
func (tm *TunnelManager) StopAllTunnels() {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	for key, tunnel := range tm.tunnels {
		tm.cleanupTunnel(tunnel)
		delete(tm.tunnels, key)
	}

	klog.Infof("All SSH tunnels stopped")
}

// GetActiveTunnels returns information about active tunnels
func (tm *TunnelManager) GetActiveTunnels() map[string]*SSHTunnel {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	result := make(map[string]*SSHTunnel)
	for key, tunnel := range tm.tunnels {
		if tunnel.Status == "active" {
			result[key] = tunnel
		}
	}

	return result
}

// findAvailablePort finds an available local port
func findAvailablePort() (int, error) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()

	addr := listener.Addr().(*net.TCPAddr)
	return addr.Port, nil
}

// GetAPIServerTunnelEndpoint returns the local endpoint for API server access
func GetAPIServerTunnelEndpoint(ctx *ContextInfo, apiServerPort int) (string, error) {
	if !ctx.IsRemote {
		return "", nil // No tunnel needed for local contexts
	}

	if !ctx.IsSSH {
		return "", errors.New("automatic tunneling only supported for SSH contexts")
	}

	tm := GetTunnelManager()
	tunnel, err := tm.CreateAPIServerTunnel(ctx, apiServerPort)
	if err != nil {
		return "", errors.Wrap(err, "creating API server tunnel")
	}

	return fmt.Sprintf("https://localhost:%d", tunnel.LocalPort), nil
}

// CreateContainerSSHTunnel creates an SSH tunnel for container SSH access
func (tm *TunnelManager) CreateContainerSSHTunnel(ctx *ContextInfo, containerName string, remotePort int) (*SSHTunnel, error) {
	if !ctx.IsSSH {
		return nil, errors.New("context is not SSH-based")
	}

	u, err := url.Parse(ctx.Host)
	if err != nil {
		return nil, errors.Wrapf(err, "parsing SSH host %q", ctx.Host)
	}

	sshUser := u.User.Username()
	sshHost := u.Hostname()
	sshPort := 22
	if u.Port() != "" {
		sshPort, _ = strconv.Atoi(u.Port())
	}

	// Find available local port
	localPort, err := findAvailablePort()
	if err != nil {
		return nil, errors.Wrap(err, "finding available local port")
	}

	tunnelKey := fmt.Sprintf("container-ssh-%s", containerName)

	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	// Check if tunnel already exists
	if existing, exists := tm.tunnels[tunnelKey]; exists {
		if existing.Status == "active" {
			klog.Infof("Container SSH tunnel already active: %s", tunnelKey)
			return existing, nil
		}
		// Clean up stale tunnel
		tm.cleanupTunnel(existing)
		delete(tm.tunnels, tunnelKey)
	}

	tunnel := &SSHTunnel{
		LocalPort:  localPort,
		RemoteHost: "127.0.0.1", // Container SSH is on localhost of the remote host
		RemotePort: remotePort,
		SSHHost:    sshHost,
		SSHUser:    sshUser,
		SSHPort:    sshPort,
		Status:     "starting",
		Metrics: &TunnelMetrics{
			CreatedAt: time.Now(),
		},
	}

	// Pre-flight SSH connectivity check
	if err := tm.checkSSHConnectivity(tunnel); err != nil {
		return nil, errors.Wrapf(err, "SSH connectivity pre-check failed for %s", tunnelKey)
	}

	// Start tunnel with retry logic
	if err := tm.startTunnelWithRetry(tunnel, 3); err != nil {
		return nil, errors.Wrapf(err, "starting SSH tunnel %s", tunnelKey)
	}

	tm.tunnels[tunnelKey] = tunnel
	klog.Infof("Container SSH tunnel created: %s (localhost:%d -> %s:%d via %s@%s)", 
		tunnelKey, localPort, tunnel.RemoteHost, remotePort, sshUser, sshHost)

	return tunnel, nil
}

// performHealthCheck performs comprehensive health check for SSH tunnels
func (tm *TunnelManager) performHealthCheck(tunnel *SSHTunnel) error {
	startTime := time.Now()
	
	// Update metrics at the start
	if tunnel.Metrics != nil {
		tunnel.Metrics.mutex.Lock()
		tunnel.Metrics.LastHealthCheck = startTime
		tunnel.Metrics.TotalHealthChecks++
		tunnel.Metrics.UptimeSeconds = int64(time.Since(tunnel.Metrics.CreatedAt).Seconds())
		tunnel.Metrics.mutex.Unlock()
	}
	
	// 1. Check if local tunnel port is responsive (test both IPv4 and IPv6)
	ipv4Addr := fmt.Sprintf("127.0.0.1:%d", tunnel.LocalPort)
	ipv6Addr := fmt.Sprintf("[::1]:%d", tunnel.LocalPort)
	
	conn4, err4 := net.DialTimeout("tcp", ipv4Addr, 2*time.Second)
	ipv4Working := err4 == nil
	if ipv4Working {
		conn4.Close()
	}
	
	conn6, err6 := net.DialTimeout("tcp", ipv6Addr, 2*time.Second)
	ipv6Working := err6 == nil
	if ipv6Working {
		conn6.Close()
	}
	
	// Update protocol availability in metrics
	if tunnel.Metrics != nil {
		tunnel.Metrics.mutex.Lock()
		tunnel.Metrics.IPv4Available = ipv4Working
		tunnel.Metrics.IPv6Available = ipv6Working
		tunnel.Metrics.mutex.Unlock()
	}
	
	// At least one protocol should be working
	if !ipv4Working && !ipv6Working {
		errorMsg := fmt.Sprintf("local tunnel port not responsive (IPv4: %v, IPv6: %v)", err4, err6)
		tm.updateMetricsOnError(tunnel, errorMsg)
		return errors.New(errorMsg)
	}
	
	klog.V(4).Infof("Tunnel health check: IPv4=%v, IPv6=%v", ipv4Working, ipv6Working)

	// 2. Verify SSH connectivity to remote host
	if err := tm.checkSSHConnectivity(tunnel); err != nil {
		errorMsg := fmt.Sprintf("SSH connectivity check failed: %v", err)
		tm.updateMetricsOnError(tunnel, errorMsg)
		return errors.Wrap(err, "SSH connectivity check failed")
	}

	// 3. Verify remote service is accessible through tunnel
	if err := tm.checkRemoteServiceConnectivity(tunnel); err != nil {
		klog.V(3).Infof("Remote service connectivity check failed (may be normal): %v", err)
		// Don't fail health check for remote service issues as the tunnel itself may be fine
	}

	// Health check completed successfully
	checkDuration := time.Since(startTime)
	tm.updateMetricsOnSuccess(tunnel, checkDuration)

	return nil
}

// checkSSHConnectivity verifies SSH connection to remote host
func (tm *TunnelManager) checkSSHConnectivity(tunnel *SSHTunnel) error {
	// Use a quick SSH command to verify connectivity
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sshArgs := []string{
		"-o", "ConnectTimeout=3",
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "LogLevel=ERROR",
		"-o", "BatchMode=yes", // Don't prompt for passwords
		"-o", "ControlMaster=auto",
		"-o", "ControlPath=/tmp/minikube-ssh-%r@%h:%p",
		"-o", "ControlPersist=600", // Keep connection alive for 10 minutes
		fmt.Sprintf("%s@%s", tunnel.SSHUser, tunnel.SSHHost),
		"echo", "ssh-health-check",
	}

	if tunnel.SSHPort != 22 {
		sshArgs = append([]string{"-p", strconv.Itoa(tunnel.SSHPort)}, sshArgs...)
	}

	cmd := exec.CommandContext(ctx, "ssh", sshArgs...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return errors.Wrapf(err, "SSH command failed: %s", string(output))
	}

	// Verify we got expected response
	if !strings.Contains(string(output), "ssh-health-check") {
		return errors.Errorf("unexpected SSH response: %s", string(output))
	}

	return nil
}

// checkRemoteServiceConnectivity verifies the remote service is accessible
func (tm *TunnelManager) checkRemoteServiceConnectivity(tunnel *SSHTunnel) error {
	// Try to connect to the remote service through the tunnel
	// This is a best-effort check since the service might not be ready
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Use curl through SSH to check if the remote service responds
	sshArgs := []string{
		"-o", "ConnectTimeout=2",
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "LogLevel=ERROR",
		"-o", "BatchMode=yes",
		fmt.Sprintf("%s@%s", tunnel.SSHUser, tunnel.SSHHost),
		"curl", "-k", "--connect-timeout", "2", "--max-time", "3",
		"--silent", "--output", "/dev/null", "--write-out", "%{http_code}",
		fmt.Sprintf("https://%s:%d/", tunnel.RemoteHost, tunnel.RemotePort),
	}

	if tunnel.SSHPort != 22 {
		sshArgs = append([]string{"-p", strconv.Itoa(tunnel.SSHPort)}, sshArgs...)
	}

	cmd := exec.CommandContext(ctx, "ssh", sshArgs...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return errors.Wrapf(err, "remote service check failed: %s", string(output))
	}

	// Any HTTP response code indicates the service is reachable
	// Even 404 or 401 means the service is responding
	responseCode := strings.TrimSpace(string(output))
	if responseCode == "" || responseCode == "000" {
		return errors.New("remote service not responding")
	}

	klog.V(3).Infof("Remote service health check: HTTP %s", responseCode)
	return nil
}

// updateMetricsOnError updates tunnel metrics when a health check fails
func (tm *TunnelManager) updateMetricsOnError(tunnel *SSHTunnel, errorMsg string) {
	if tunnel.Metrics == nil {
		return
	}
	
	tunnel.Metrics.mutex.Lock()
	defer tunnel.Metrics.mutex.Unlock()
	
	tunnel.Metrics.FailedHealthChecks++
	tunnel.Metrics.LastError = errorMsg
}

// updateMetricsOnSuccess updates tunnel metrics when a health check succeeds
func (tm *TunnelManager) updateMetricsOnSuccess(tunnel *SSHTunnel, latency time.Duration) {
	if tunnel.Metrics == nil {
		return
	}
	
	tunnel.Metrics.mutex.Lock()
	defer tunnel.Metrics.mutex.Unlock()
	
	tunnel.Metrics.LastSuccessfulCheck = time.Now()
	tunnel.Metrics.LastError = "" // Clear last error
	
	// Update average latency using exponential moving average
	if tunnel.Metrics.AvgLatency == 0 {
		tunnel.Metrics.AvgLatency = latency
	} else {
		// Weight: 80% old average, 20% new measurement
		tunnel.Metrics.AvgLatency = time.Duration(
			float64(tunnel.Metrics.AvgLatency)*0.8 + float64(latency)*0.2,
		)
	}
}

// GetTunnelMetrics returns a copy of tunnel metrics (thread-safe)
func (tm *TunnelManager) GetTunnelMetrics() map[string]TunnelMetrics {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()
	
	result := make(map[string]TunnelMetrics)
	for key, tunnel := range tm.tunnels {
		if tunnel.Metrics != nil {
			tunnel.Metrics.mutex.RLock()
			result[key] = *tunnel.Metrics // Copy the metrics
			tunnel.Metrics.mutex.RUnlock()
		}
	}
	
	return result
}

// LogTunnelMetrics logs comprehensive tunnel metrics
func (tm *TunnelManager) LogTunnelMetrics() {
	metrics := tm.GetTunnelMetrics()
	
	if len(metrics) == 0 {
		klog.V(3).Infof("No active tunnels to report metrics for")
		return
	}
	
	klog.Infof("=== SSH Tunnel Health Metrics ===")
	for tunnelKey, m := range metrics {
		successRate := float64(0)
		if m.TotalHealthChecks > 0 {
			successRate = float64(m.TotalHealthChecks-m.FailedHealthChecks) / float64(m.TotalHealthChecks) * 100
		}
		
		klog.Infof("Tunnel: %s", tunnelKey)
		klog.Infof("  Uptime: %v (%d seconds)", time.Duration(m.UptimeSeconds)*time.Second, m.UptimeSeconds)
		klog.Infof("  Health Checks: %d total, %d failed (%.1f%% success rate)", 
			m.TotalHealthChecks, m.FailedHealthChecks, successRate)
		klog.Infof("  Restarts: %d", m.RestartCount)
		klog.Infof("  Avg Latency: %v", m.AvgLatency)
		klog.Infof("  Protocol Support: IPv4=%v, IPv6=%v", m.IPv4Available, m.IPv6Available)
		klog.Infof("  Last Successful Check: %v", m.LastSuccessfulCheck.Format("2006-01-02 15:04:05"))
		if m.LastError != "" {
			klog.Infof("  Last Error: %s", m.LastError)
		}
	}
	klog.Infof("=== End Tunnel Metrics ===")
}