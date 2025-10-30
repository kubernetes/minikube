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

	"k8s.io/minikube/pkg/drivers/common"
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
	*common.CommonDriver
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
		IPFamily:      strings.ToLower(d.NodeConfig.IPFamily),
		IPv6:          d.NodeConfig.StaticIPv6,
	}
	if params.Memory != "0" {
		params.Memory += "mb"
	}

	networkName := d.NodeConfig.Network
	if networkName == "" {
		networkName = d.NodeConfig.ClusterName
	}
	staticIP := d.NodeConfig.StaticIP
	// NEW: create network with IPv6/dual awareness
	gateway, err := oci.CreateNetworkWithIPFamily(
		d.OCIBinary,
		networkName,
		d.NodeConfig.Subnet,
		d.NodeConfig.Subnetv6, // NEW
		staticIP,
		d.NodeConfig.StaticIPv6, // NEW
		params.IPFamily,         // NEW
	)
	if err != nil {
		msg := "Unable to create dedicated network, this might result in cluster IP change after restart: {{.error}}"
		args := out.V{"error": err}
		if staticIP != "" {
			exit.Message(reason.IfDedicatedNetwork, msg, args)
		}
		out.WarningT(msg, args)
	}
	// Always attach to the created user network (even if gateway is nil in IPv6-only)
	params.Network = networkName

	// Now decide static IPs per family
	switch params.IPFamily {
	case "ipv6":
		if d.NodeConfig.StaticIPv6 != "" {
			params.IPv6 = d.NodeConfig.StaticIPv6
		}
	case "dual":
		// IPv4 part (only if Docker reported a v4 gateway)
		if g4 := gateway.To4(); g4 != nil {
			if staticIP != "" {
				params.IP = staticIP
			} else {
				ip := make(net.IP, len(g4))
				copy(ip, g4)
				index := driver.IndexFromMachineName(d.NodeConfig.MachineName)
				if int(ip[3])+index > 253 {
					return fmt.Errorf("too many machines to calculate an IPv4")
				}
				ip[3] += byte(index)
				klog.Infof("calculated static IPv4 %q for the %q container", ip.String(), d.NodeConfig.MachineName)
				params.IP = ip.String()
			}
		}
		if d.NodeConfig.StaticIPv6 != "" {
			params.IPv6 = d.NodeConfig.StaticIPv6
		}
	default: // ipv4
		if staticIP != "" {
			params.IP = staticIP
		} else if gateway != nil {
			if g4 := gateway.To4(); g4 != nil {
				ip := make(net.IP, len(g4))
				copy(ip, g4)
				index := driver.IndexFromMachineName(d.NodeConfig.MachineName)
				if int(ip[3])+index > 253 {
					return fmt.Errorf("too many machines to calculate an IP")
				}
				ip[3] += byte(index)
				klog.Infof("calculated static IP %q for the %q container", ip.String(), d.NodeConfig.MachineName)
				params.IP = ip.String()
			}
		}
	}
	drv := d.DriverName()
	// Default listen address: v4 localhost for ipv4, v6 localhost for ipv6-only
	listAddr := oci.DefaultBindIPV4
	// IPv6-only clusters must publish on IPv6 loopback so the host can reach them
	if params.IPFamily == "ipv6" {
		listAddr = "::1"
	}

	if d.NodeConfig.ListenAddress != "" && d.NodeConfig.ListenAddress != listAddr {
		out.Step(style.Tip, "minikube is not meant for production use. You are opening non-local traffic")
		out.WarningT("Listening to {{.listenAddr}}. This is not recommended and can cause a security vulnerability. Use at your own risk",
			out.V{"listenAddr": d.NodeConfig.ListenAddress})
		listAddr = d.NodeConfig.ListenAddress
	} else if oci.IsExternalDaemonHost(drv) {
		if params.IPFamily == "ipv6" {
			out.WarningT("Listening to :: on external docker host {{.host}}. Please be advised",
				out.V{"host": oci.DaemonHost(drv)})
			listAddr = "::"
		} else {
			out.WarningT("Listening to 0.0.0.0 on external docker host {{.host}}. Please be advised",
				out.V{"host": oci.DaemonHost(drv)})
			listAddr = "0.0.0.0"
		}
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
		// If preload doesn't exist, don't bother extracting tarball to volume
		if !download.PreloadExists(d.NodeConfig.KubernetesVersion, d.NodeConfig.ContainerRuntime, d.DriverName()) {
			return
		}
		t := time.Now()
		klog.Infof("Starting extracting preloaded images to volume ...")
		// Extract preloaded images to container
		if err := oci.ExtractTarballToVolume(d.NodeConfig.OCIBinary, download.TarballPath(d.NodeConfig.KubernetesVersion, d.NodeConfig.ContainerRuntime), params.Name, d.NodeConfig.ImageDigest); err != nil {
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

	if err := d.prepareSSH(); err != nil {
		return errors.Wrap(err, "prepare kic ssh")
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
	ip4, ip6, err := oci.ContainerIPs(d.OCIBinary, d.MachineName)
	if err != nil {
		return "", err
	}
	switch strings.ToLower(d.NodeConfig.IPFamily) {
	case "ipv6":
		if ip6 != "" {
			return ip6, nil
		}
	}
	// default / dual prefers IPv4 for backward compat
	return ip4, nil
}

// GetExternalIP returns an IP which is accessible from outside
func (d *Driver) GetExternalIP() (string, error) {
	host := oci.DaemonHost(d.DriverName())
	// For local daemons and IPv6-only clusters, ports are published on ::1
	if strings.ToLower(d.NodeConfig.IPFamily) == "ipv6" && !oci.IsExternalDaemonHost(d.DriverName()) {
		return "::1", nil
	}
	return host, nil
}

// GetSSHHostname returns hostname for use with ssh
func (d *Driver) GetSSHHostname() (string, error) {
	host := oci.DaemonHost(d.DriverName())
	// For local daemons and IPv6-only clusters, ports are published on ::1
	if strings.ToLower(d.NodeConfig.IPFamily) == "ipv6" && !oci.IsExternalDaemonHost(d.DriverName()) {
		return "::1", nil
	}
	return host, nil
}

// GetSSHPort returns port for use with ssh
func (d *Driver) GetSSHPort() (int, error) {
	p, err := oci.ForwardedPort(d.OCIBinary, d.MachineName, constants.SSHPort)
	if err != nil {
		return p, errors.Wrap(err, "get ssh host-port")
	}
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
		// even though we can't stop the containers inside, we still wanna stop the minikube container itself
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
