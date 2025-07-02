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

package kic

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/ssh"
	"github.com/docker/machine/libmachine/state"
	"github.com/pkg/errors"
	"k8s.io/klog/v2"

	pkgdrivers "k8s.io/minikube/pkg/drivers"
	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/cruntime"
	"k8s.io/minikube/pkg/minikube/download"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/minikube/style"
	"k8s.io/minikube/pkg/minikube/sysinit"
	"k8s.io/minikube/pkg/util/retry"
)

// Driver represents a kic driver https://minikube.sigs.k8s.io/docs/reference/drivers/docker
type Driver struct {
	*drivers.BaseDriver
	*pkgdrivers.CommonDriver
	URL        string
	exec       command.Runner
	NodeConfig Config
	OCIBinary  string // docker,podman
}

// NewDriver returns a fully configured Kic driver
func NewDriver(c Config) *Driver {
	d := &Driver{
		BaseDriver: &drivers.BaseDriver{
			MachineName: c.MachineName,
			StorePath:   c.StorePath,
		},
		exec:       command.NewKICRunner(c.MachineName, c.OCIBinary),
		NodeConfig: c,
		OCIBinary:  c.OCIBinary,
	}
	return d
}

// Create a host using the driver's config
func (d *Driver) Create() error {
	ctx := context.Background()
	params := oci.CreateParams{
		Mounts:        d.NodeConfig.Mounts,
		Name:          d.NodeConfig.MachineName,
		Image:         d.NodeConfig.ImageDigest,
		ClusterLabel:  oci.ProfileLabelKey + "=" + d.MachineName,
		NodeLabel:     oci.NodeLabelKey + "=" + d.NodeConfig.MachineName,
		CPUs:          strconv.Itoa(d.NodeConfig.CPU),
		Memory:        strconv.Itoa(d.NodeConfig.Memory),
		Envs:          d.NodeConfig.Envs,
		ExtraArgs:     append([]string{"--expose", fmt.Sprintf("%d", d.NodeConfig.APIServerPort)}, d.NodeConfig.ExtraArgs...),
		OCIBinary:     d.NodeConfig.OCIBinary,
		APIServerPort: d.NodeConfig.APIServerPort,
		GPUs:          d.NodeConfig.GPUs,
	}
	if params.Memory != "0" {
		params.Memory += "mb"
	}

	networkName := d.NodeConfig.Network
	if networkName == "" {
		networkName = d.NodeConfig.ClusterName
	}
	staticIP := d.NodeConfig.StaticIP
	if gateway, err := oci.CreateNetwork(d.OCIBinary, networkName, d.NodeConfig.Subnet, staticIP); err != nil {
		msg := "Unable to create dedicated network, this might result in cluster IP change after restart: {{.error}}"
		args := out.V{"error": err}
		if staticIP != "" {
			exit.Message(reason.IfDedicatedNetwork, msg, args)
		}
		out.WarningT(msg, args)
	} else if gateway != nil && staticIP != "" {
		params.Network = networkName
		params.IP = staticIP
	} else if gateway != nil {
		params.Network = networkName
		ip := gateway.To4()
		// calculate the container IP based on guessing the machine index
		index := driver.IndexFromMachineName(d.NodeConfig.MachineName)
		if int(ip[3])+index > 253 { // reserve last client ip address for multi-control-plane loadbalancer vip address in ha cluster
			return fmt.Errorf("too many machines to calculate an IP")
		}
		ip[3] += byte(index)
		klog.Infof("calculated static IP %q for the %q container", ip.String(), d.NodeConfig.MachineName)
		params.IP = ip.String()
	}
	drv := d.DriverName()

	listAddr := oci.DefaultBindIPV4
	if d.NodeConfig.ListenAddress != "" && d.NodeConfig.ListenAddress != listAddr {
		out.Step(style.Tip, "minikube is not meant for production use. You are opening non-local traffic")
		out.WarningT("Listening to {{.listenAddr}}. This is not recommended and can cause a security vulnerability. Use at your own risk",
			out.V{"listenAddr": d.NodeConfig.ListenAddress})
		listAddr = d.NodeConfig.ListenAddress
	} else if oci.IsRemoteDockerContext() || oci.IsExternalDaemonHost(drv) {
		hostInfo := "remote Docker daemon"
		if ctx, err := oci.GetCurrentContext(); err == nil && ctx.Host != "" {
			hostInfo = fmt.Sprintf("remote Docker daemon (%s)", ctx.Host)
		} else if oci.IsExternalDaemonHost(drv) {
			hostInfo = fmt.Sprintf("external docker host %s", oci.DaemonHost(drv))
		}
		out.WarningT("Listening to 0.0.0.0 on {{.host}}. Please be advised",
			out.V{"host": hostInfo})
		listAddr = "0.0.0.0"
	}

	// control plane specific options
	params.PortMappings = append(params.PortMappings,
		oci.PortMapping{
			ListenAddress: listAddr,
			ContainerPort: int32(params.APIServerPort),
		},
		oci.PortMapping{
			ListenAddress: listAddr,
			ContainerPort: constants.SSHPort,
		},
		oci.PortMapping{
			ListenAddress: listAddr,
			ContainerPort: constants.DockerDaemonPort,
		},
		oci.PortMapping{
			ListenAddress: listAddr,
			ContainerPort: constants.RegistryAddonPort,
		},
		oci.PortMapping{
			ListenAddress: listAddr,
			ContainerPort: constants.AutoPauseProxyPort,
		},
	)

	exists, err := oci.ContainerExists(d.OCIBinary, params.Name, true)
	if err != nil {
		klog.Warningf("failed to check if container already exists: %v", err)
	}
	if exists {
		// if container was created by minikube it is safe to delete and recreate it.
		if oci.IsCreatedByMinikube(d.OCIBinary, params.Name) {
			klog.Info("Found already existing abandoned minikube container, will try to delete.")
			if err := oci.DeleteContainer(ctx, d.OCIBinary, params.Name); err != nil {
				klog.Errorf("Failed to delete a conflicting minikube container %s. You might need to restart your %s daemon and delete it manually and try again: %v", params.Name, params.OCIBinary, err)
			}
		} else {
			// The conflicting container name was not created by minikube
			// user has a container that conflicts with minikube profile name, will not delete users container.
			return errors.Wrapf(err, "user has a conflicting container name %q with minikube container. Needs to be deleted by user's consent", params.Name)
		}
	}

	if err := oci.PrepareContainerNode(params); err != nil {
		return errors.Wrap(err, "setting up container node")
	}

	var waitForPreload sync.WaitGroup
	waitForPreload.Add(1)
	var pErr error
	go func() {
		defer waitForPreload.Done()
		// Detect target architecture for preload
		arch := runtime.GOARCH
		if d.NodeConfig.OCIBinary == oci.Docker {
			if daemonArch, err := oci.DaemonArch(oci.Docker); err != nil {
				klog.Warningf("Failed to detect Docker daemon architecture, using local arch: %v", err)
			} else {
				arch = daemonArch
				klog.Infof("Detected Docker daemon architecture for preload: %s", arch)
			}
		}

		// If preload doesn't exist, don't bother extracting tarball to volume
		if !download.PreloadExistsWithArch(d.NodeConfig.KubernetesVersion, d.NodeConfig.ContainerRuntime, d.DriverName(), arch) {
			return
		}
		t := time.Now()
		klog.Infof("Starting extracting preloaded images to volume ...")
		// Extract preloaded images to container
		if err := oci.ExtractTarballToVolume(d.NodeConfig.OCIBinary, download.TarballPathWithArch(d.NodeConfig.KubernetesVersion, d.NodeConfig.ContainerRuntime, arch), params.Name, d.NodeConfig.ImageDigest); err != nil {
			if strings.Contains(err.Error(), "No space left on device") {
				pErr = oci.ErrInsufficientDockerStorage
				return
			}
			klog.Infof("Unable to extract preloaded tarball to volume: %v", err)
		} else {
			klog.Infof("duration metric: took %s to extract preloaded images to volume ...", time.Since(t))
		}
	}()
	waitForPreload.Wait()
	if pErr == oci.ErrInsufficientDockerStorage {
		return pErr
	}

	if err := oci.CreateContainerNode(params); err != nil {
		return errors.Wrap(err, "create kic node")
	}

	// For remote Docker contexts with SSH, ensure SSH tunnel is established
	if oci.IsRemoteDockerContext() && oci.IsSSHDockerContext() {
		klog.Warningf("TUNNEL SETUP: Ensuring SSH access for remote Docker container %s", d.MachineName)
		if err := oci.EnsureContainerSSHAccess(d.MachineName); err != nil {
			// SSH tunnel is critical for remote Docker contexts
			return errors.Wrapf(err, "failed to ensure SSH access for container %s", d.MachineName)
		}
		klog.Warningf("TUNNEL SETUP SUCCESS: SSH access ensured for container %s", d.MachineName)
	}

	if err := d.prepareSSH(); err != nil {
		return errors.Wrap(err, "prepare kic ssh")
	}

	// Setup SSH proxy configuration for remote Docker contexts
	if oci.IsRemoteDockerContext() {
		sshPort, err := oci.ForwardedPort(d.OCIBinary, d.MachineName, constants.SSHPort)
		if err != nil {
			return errors.Wrap(err, "get SSH port")
		}
		if err := oci.WriteSSHProxyConfig(d.MachineName, sshPort); err != nil {
			return errors.Wrap(err, "write SSH proxy config")
		}

		// Setup automatic API server tunnel for SSH contexts
		if oci.IsSSHDockerContext() {
			ctx, err := oci.GetCurrentContext()
			if err != nil {
				klog.Warningf("Failed to get Docker context for tunnel setup: %v", err)
			} else {
				apiServerPort, err := oci.ForwardedPort(d.OCIBinary, d.MachineName, constants.APIServerPort)
				if err != nil {
					klog.Warningf("Failed to get API server port for tunnel: %v", err)
				} else {
					tm := oci.GetTunnelManager()
					tunnel, err := tm.CreateAPIServerTunnel(ctx, apiServerPort)
					if err != nil {
						klog.Warningf("Failed to create API server tunnel: %v", err)
					} else {
						klog.Infof("API server tunnel created automatically: localhost:%d -> %s:%d", tunnel.LocalPort, tunnel.SSHHost, apiServerPort)
					}
				}
			}
		}
	}

	return nil
}

// prepareSSH will generate keys and copy to the container so minikube ssh works
func (d *Driver) prepareSSH() error {
	keyPath := d.GetSSHKeyPath()
	klog.Infof("Creating ssh key for kic: %s...", keyPath)
	if err := ssh.GenerateSSHKey(keyPath); err != nil {
		return errors.Wrap(err, "generate ssh key")
	}

	cmder := command.NewKICRunner(d.NodeConfig.MachineName, d.NodeConfig.OCIBinary)
	f, err := assets.NewFileAsset(d.GetSSHKeyPath()+".pub", "/home/docker/.ssh/", "authorized_keys", "0644")
	if err != nil {
		return errors.Wrap(err, "create pubkey assetfile ")
	}
	defer func() {
		if err := f.Close(); err != nil {
			klog.Warningf("error closing the file %s: %v", f.GetSourcePath(), err)
		}
	}()

	if err := cmder.Copy(f); err != nil {
		return errors.Wrap(err, "copying pub key")
	}

	// Double-check that the container has not crashed so that we may give a better error message
	s, err := oci.ContainerStatus(d.NodeConfig.OCIBinary, d.MachineName)
	if err != nil {
		return err
	}

	if s != state.Running {
		excerpt := oci.LogContainerDebug(d.OCIBinary, d.MachineName)
		return errors.Wrapf(oci.ErrExitedUnexpectedly, "container name %q state %s: log: %s", d.MachineName, s, excerpt)
	}

	if rr, err := cmder.RunCmd(exec.Command("chown", "docker:docker", "/home/docker/.ssh/authorized_keys")); err != nil {
		return errors.Wrapf(err, "apply authorized_keys file ownership, output %s", rr.Output())
	}

	if runtime.GOOS == "windows" {
		path, _ := exec.LookPath("powershell")
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()

		klog.Infof("ensuring only current user has permissions to key file located at : %s...", keyPath)

		// Get the SID of the current user
		currentUserSidCmd := exec.CommandContext(ctx, path, "-NoProfile", "-NonInteractive", "([System.Security.Principal.WindowsIdentity]::GetCurrent()).User.Value")
		currentUserSidOut, currentUserSidErr := currentUserSidCmd.CombinedOutput()
		if currentUserSidErr != nil {
			klog.Warningf("unable to determine current user's SID. minikube tunnel may not work.")
		} else {
			icaclsArguments := fmt.Sprintf(`"%s" /grant:r *%s:F /inheritancelevel:r`, keyPath, strings.TrimSpace(string(currentUserSidOut)))
			icaclsCmd := exec.CommandContext(ctx, path, "-NoProfile", "-NonInteractive", "icacls.exe", icaclsArguments)
			icaclsCmdOut, icaclsCmdErr := icaclsCmd.CombinedOutput()

			if icaclsCmdErr != nil {
				return errors.Wrap(icaclsCmdErr, fmt.Sprintf("unable to execute icacls to set permissions: %s", icaclsCmdOut))
			}
		}
	}

	return nil
}

// DriverName returns the name of the driver
func (d *Driver) DriverName() string {
	if d.NodeConfig.OCIBinary == oci.Podman {
		return oci.Podman
	}
	return oci.Docker
}

// GetIP returns an IP or hostname that this host is available at
func (d *Driver) GetIP() (string, error) {
	ip, _, err := oci.ContainerIPs(d.OCIBinary, d.MachineName)
	return ip, err
}

// GetExternalIP returns an IP which is accessible from outside
func (d *Driver) GetExternalIP() (string, error) {
	return oci.DaemonHost(d.DriverName()), nil
}

// GetSSHHostname returns hostname for use with ssh
func (d *Driver) GetSSHHostname() (string, error) {
	// For remote Docker contexts, handle SSH and TLS differently
	if oci.IsRemoteDockerContext() {
		if oci.IsSSHDockerContext() {
			// SSH contexts use tunneling, so use localhost
			klog.Infof("GetSSHHostname: Using localhost for SSH Docker context")
			return "127.0.0.1", nil
		}
		// For TLS contexts, SSH is not used - return localhost as dummy value
		klog.V(3).Infof("GetSSHHostname: Remote TLS context detected, SSH not needed")
		return "127.0.0.1", nil
	}
	return oci.DaemonHost(d.DriverName()), nil
}

// GetSSHPort returns port for use with ssh
func (d *Driver) GetSSHPort() (int, error) {
	// For remote SSH Docker contexts, return the tunnel port if available
	if oci.IsRemoteDockerContext() && oci.IsSSHDockerContext() {
		tunnelPort, err := oci.GetContainerSSHPort(d.MachineName)
		if err == nil && tunnelPort > 0 {
			klog.Infof("GetSSHPort: Returning SSH tunnel port %d for %s", tunnelPort, d.MachineName)
			return tunnelPort, nil
		}
		klog.Warningf("GetSSHPort: No SSH tunnel found for %s, will use direct port (may fail)", d.MachineName)
	}

	// For remote TLS Docker contexts, SSH is not used - return a dummy port to satisfy libmachine
	if oci.IsRemoteDockerContext() && !oci.IsSSHDockerContext() {
		klog.V(3).Infof("GetSSHPort: Remote TLS context detected, SSH not needed for %s", d.MachineName)
		return 22, nil // Return standard SSH port as a dummy value
	}

	// For local contexts or if tunnel not available, return the direct port
	p, err := oci.ForwardedPort(d.OCIBinary, d.MachineName, constants.SSHPort)
	if err != nil {
		return p, errors.Wrap(err, "get ssh host-port")
	}
	klog.Infof("GetSSHPort: Returning Docker port %d for %s", p, d.MachineName)
	return p, nil
}

// GetSSHUsername returns the ssh username
func (d *Driver) GetSSHUsername() string {
	return "docker"
}

// GetSSHKeyPath returns the ssh key path
func (d *Driver) GetSSHKeyPath() string {
	if d.SSHKeyPath == "" {
		d.SSHKeyPath = d.ResolveStorePath("id_rsa")
	}
	return d.SSHKeyPath
}

// GetURL returns a Docker URL inside this host
// e.g. tcp://1.2.3.4:2376
// more info https://github.com/docker/machine/blob/b170508bf44c3405e079e26d5fdffe35a64c6972/libmachine/provision/utils.go#L159_L175
func (d *Driver) GetURL() (string, error) {
	ip, err := d.GetIP()
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf("tcp://%s", net.JoinHostPort(ip, "2376"))
	return url, nil
}

// GetState returns the state that the host is in (running, stopped, etc)
func (d *Driver) GetState() (state.State, error) {
	return oci.ContainerStatus(d.OCIBinary, d.MachineName, true)
}

// Kill stops a host forcefully, including any containers that we are managing.
func (d *Driver) Kill() error {
	// on init this doesn't get filled when called from cmd
	d.exec = command.NewKICRunner(d.MachineName, d.OCIBinary)
	if err := sysinit.New(d.exec).ForceStop("kubelet"); err != nil {
		klog.Warningf("couldn't force stop kubelet. will continue with kill anyways: %v", err)
	}

	if err := oci.ShutDown(d.OCIBinary, d.MachineName); err != nil {
		klog.Warningf("couldn't shutdown the container, will continue with kill anyways: %v", err)
	}

	// Clean up any SSH tunnels for remote Docker contexts
	if oci.IsRemoteDockerContext() {
		oci.CleanupContainerTunnels()
	}

	cr := command.NewExecRunner(false) // using exec runner for interacting with daemon.
	if _, err := cr.RunCmd(oci.PrefixCmd(exec.Command(d.NodeConfig.OCIBinary, "kill", d.MachineName))); err != nil {
		return errors.Wrapf(err, "killing %q", d.MachineName)
	}
	return nil
}

// Remove will delete the Kic Node Container
func (d *Driver) Remove() error {
	if _, err := oci.ContainerID(d.OCIBinary, d.MachineName); err != nil {
		klog.Infof("could not find the container %s to remove it. will try anyways", d.MachineName)
	}

	// Clean up any SSH tunnels for remote Docker contexts
	if oci.IsRemoteDockerContext() {
		oci.CleanupContainerTunnels()
	}

	if err := oci.DeleteContainer(context.Background(), d.NodeConfig.OCIBinary, d.MachineName); err != nil {
		if strings.Contains(err.Error(), "is already in progress") {
			return errors.Wrap(err, "stuck delete")
		}
		if strings.Contains(err.Error(), "No such container:") {
			return nil // nothing was found to delete.
		}

	}

	// check there be no container left after delete
	if id, err := oci.ContainerID(d.OCIBinary, d.MachineName); err == nil && id != "" {
		return fmt.Errorf("expected no container ID be found for %q after delete. but got %q", d.MachineName, id)
	}

	if err := oci.RemoveNetwork(d.OCIBinary, d.NodeConfig.ClusterName); err != nil {
		klog.Warningf("failed to remove network (which might be okay) %s: %v", d.NodeConfig.ClusterName, err)
	}
	return nil
}

// Restart a host
func (d *Driver) Restart() error {
	s, err := d.GetState()
	if err != nil {
		klog.Warningf("get state during restart: %v", err)
	}
	if s == state.Stopped { // don't stop if already stopped
		return d.Start()
	}
	if err = d.Stop(); err != nil {
		return fmt.Errorf("stop during restart %v", err)
	}
	if err = d.Start(); err != nil {
		return fmt.Errorf("start during restart %v", err)
	}
	return nil
}

// Start an already created kic container
func (d *Driver) Start() error {
	if err := oci.StartContainer(d.NodeConfig.OCIBinary, d.MachineName); err != nil {
		oci.LogContainerDebug(d.OCIBinary, d.MachineName)
		if _, err := oci.DaemonInfo(d.OCIBinary); err != nil {
			return errors.Wrapf(oci.ErrDaemonInfo, "debug daemon info %q", d.MachineName)
		}
		return errors.Wrap(err, "start")
	}
	checkRunning := func() error {
		s, err := oci.ContainerStatus(d.NodeConfig.OCIBinary, d.MachineName)
		if err != nil {
			return err
		}
		if s != state.Running {
			return fmt.Errorf("expected container state be running but got %q", s)
		}
		klog.Infof("container %q state is running.", d.MachineName)
		return nil
	}

	if err := retry.Expo(checkRunning, 500*time.Microsecond, time.Second*30); err != nil {
		excerpt := oci.LogContainerDebug(d.OCIBinary, d.MachineName)
		_, err := oci.DaemonInfo(d.OCIBinary)
		if err != nil {
			return errors.Wrapf(oci.ErrDaemonInfo, "container name %q", d.MachineName)
		}

		return errors.Wrapf(oci.ErrExitedUnexpectedly, "container name %q: log: %s", d.MachineName, excerpt)
	}

	// Ensure SSH tunnel for remote Docker contexts after container is running
	if oci.IsRemoteDockerContext() && oci.IsSSHDockerContext() {
		klog.Warningf("Ensuring SSH tunnel after container start for %s", d.MachineName)
		if err := oci.EnsureContainerSSHAccess(d.MachineName); err != nil {
			klog.Errorf("Failed to ensure SSH access on Start: %v", err)
			// Don't fail start, tunnel might be established later
		} else {
			klog.Infof("SSH access ensured for container %s after restart", d.MachineName)
		}

		// Also re-establish API server tunnel if needed
		apiServerPort, err := oci.ForwardedPort(d.OCIBinary, d.MachineName, constants.APIServerPort)
		if err != nil {
			klog.Warningf("Failed to get API server port for tunnel: %v", err)
		} else {
			ctx, err := oci.GetCurrentContext()
			if err != nil {
				klog.Warningf("Failed to get Docker context for tunnel setup: %v", err)
			} else {
				tm := oci.GetTunnelManager()
				tunnel, err := tm.CreateAPIServerTunnel(ctx, apiServerPort)
				if err != nil {
					klog.Warningf("Failed to create API server tunnel on restart: %v", err)
				} else {
					klog.Infof("API server tunnel re-established on restart: localhost:%d -> %s:%d", tunnel.LocalPort, tunnel.SSHHost, apiServerPort)
				}
			}
		}
	}

	return nil
}

// Stop a host gracefully, including any containers that we are managing.
func (d *Driver) Stop() error {
	// on init this doesn't get filled when called from cmd
	d.exec = command.NewKICRunner(d.MachineName, d.OCIBinary)
	// docker does not send right SIG for systemd to know to stop the systemd.
	// to avoid bind address be taken on an upgrade. more info https://github.com/kubernetes/minikube/issues/7171
	if err := sysinit.New(d.exec).Stop("kubelet"); err != nil {
		klog.Warningf("couldn't stop kubelet. will continue with stop anyways: %v", err)
		if err := sysinit.New(d.exec).ForceStop("kubelet"); err != nil {
			klog.Warningf("couldn't force stop kubelet. will continue with stop anyways: %v", err)
		}
	}

	crMgr, err := cruntime.New(cruntime.Config{Type: d.NodeConfig.ContainerRuntime, Runner: d.exec})
	if err != nil { // won't return error because:
		// even though we can't stop the cotainers inside, we still wanna stop the minikube container itself
		klog.Errorf("unable to get container runtime: %v", err)
	} else {
		containers, err := crMgr.ListContainers(cruntime.ListContainersOptions{Namespaces: constants.DefaultNamespaces})
		if err != nil {
			klog.Infof("unable list containers : %v", err)
		}
		if len(containers) > 0 {
			if err := crMgr.StopContainers(containers); err != nil {
				klog.Infof("unable to stop containers : %v", err)
			}
			if err := crMgr.KillContainers(containers); err != nil {
				klog.Errorf("unable to kill containers : %v", err)
			}
		}
		klog.Infof("successfully stopped kubernetes!")

	}

	if err := killAPIServerProc(d.exec); err != nil {
		klog.Warningf("couldn't stop kube-apiserver proc: %v", err)
	}

	cmd := exec.Command(d.NodeConfig.OCIBinary, "stop", d.MachineName)
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "stopping %s", d.MachineName)
	}
	return nil
}

// RunSSHCommandFromDriver implements direct ssh control to the driver
func (d *Driver) RunSSHCommandFromDriver() error {
	return fmt.Errorf("driver does not support RunSSHCommandFromDriver commands")
}

// killAPIServerProc will kill an api server proc if it exists
// to ensure this never happens https://github.com/kubernetes/minikube/issues/7521
func killAPIServerProc(runner command.Runner) error {
	// first check if it exists
	rr, err := runner.RunCmd(exec.Command("pgrep", "kube-apiserver"))
	if err == nil { // this means we might have a running kube-apiserver
		pid, err := strconv.Atoi(rr.Stdout.String())
		if err == nil { // this means we have a valid pid
			klog.Warningf("Found a kube-apiserver running with pid %d, will try to kill the proc", pid)
			if _, err = runner.RunCmd(exec.Command("pkill", "-9", fmt.Sprint(pid))); err != nil {
				return errors.Wrap(err, "kill")
			}
		}
	}
	return nil
}
