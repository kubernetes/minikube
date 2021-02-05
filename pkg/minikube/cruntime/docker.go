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

package cruntime

import (
	"fmt"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/pkg/errors"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/bootstrapper/images"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/docker"
	"k8s.io/minikube/pkg/minikube/download"
	"k8s.io/minikube/pkg/minikube/style"
	"k8s.io/minikube/pkg/minikube/sysinit"
)

// KubernetesContainerPrefix is the prefix of each Kubernetes container
const KubernetesContainerPrefix = "k8s_"

// ErrISOFeature is the error returned when disk image is missing features
type ErrISOFeature struct {
	missing string
}

// NewErrISOFeature creates a new ErrISOFeature
func NewErrISOFeature(missing string) *ErrISOFeature {
	return &ErrISOFeature{
		missing: missing,
	}
}

func (e *ErrISOFeature) Error() string {
	return e.missing
}

// Docker contains Docker runtime state
type Docker struct {
	Socket string
	Runner CommandRunner
	Init   sysinit.Manager
}

// Name is a human readable name for Docker
func (r *Docker) Name() string {
	return "Docker"
}

// Style is the console style for Docker
func (r *Docker) Style() style.Enum {
	return style.Docker
}

// Version retrieves the current version of this runtime
func (r *Docker) Version() (string, error) {
	// Note: the server daemon has to be running, for this call to return successfully
	c := exec.Command("docker", "version", "--format", "{{.Server.Version}}")
	rr, err := r.Runner.RunCmd(c)
	if err != nil {
		return "", err
	}
	return strings.Split(rr.Stdout.String(), "\n")[0], nil
}

// SocketPath returns the path to the socket file for Docker
func (r *Docker) SocketPath() string {
	if r.Socket != "" {
		return r.Socket
	}
	return "/var/run/dockershim.sock"
}

// Available returns an error if it is not possible to use this runtime on a host
func (r *Docker) Available() error {
	_, err := exec.LookPath("docker")
	return err
}

// Active returns if docker is active on the host
func (r *Docker) Active() bool {
	return r.Init.Active("docker")
}

// Enable idempotently enables Docker on a host
func (r *Docker) Enable(disOthers, forceSystemd bool) error {
	containerdWasActive := r.Init.Active("containerd")

	if disOthers {
		if err := disableOthers(r, r.Runner); err != nil {
			klog.Warningf("disableOthers: %v", err)
		}
	}

	if err := populateCRIConfig(r.Runner, r.SocketPath()); err != nil {
		return err
	}

	if forceSystemd {
		if err := r.forceSystemd(); err != nil {
			return err
		}
		return r.Init.Restart("docker")
	}

	if containerdWasActive && !dockerBoundToContainerd(r.Runner) {
		// Make sure to use the internal containerd
		return r.Init.Restart("docker")
	}

	return r.Init.Start("docker")
}

// Restart restarts Docker on a host
func (r *Docker) Restart() error {
	return r.Init.Restart("docker")
}

// Disable idempotently disables Docker on a host
func (r *Docker) Disable() error {
	// because #10373
	if err := r.Init.ForceStop("docker.socket"); err != nil {
		klog.ErrorS(err, "Failed to stop", "service", "docker.socket")
	}
	return r.Init.ForceStop("docker")
}

// ImageExists checks if an image exists
func (r *Docker) ImageExists(name string, sha string) bool {
	// expected output looks like [SHA_ALGO:SHA]
	c := exec.Command("docker", "image", "inspect", "--format", "{{.Id}}", name)
	rr, err := r.Runner.RunCmd(c)
	if err != nil {
		return false
	}
	if !strings.Contains(rr.Output(), sha) {
		return false
	}
	return true
}

// LoadImage loads an image into this runtime
func (r *Docker) LoadImage(path string) error {
	klog.Infof("Loading image: %s", path)
	c := exec.Command("docker", "load", "-i", path)
	if _, err := r.Runner.RunCmd(c); err != nil {
		return errors.Wrap(err, "loadimage docker.")
	}
	return nil
}

// CGroupDriver returns cgroup driver ("cgroupfs" or "systemd")
func (r *Docker) CGroupDriver() (string, error) {
	// Note: the server daemon has to be running, for this call to return successfully
	c := exec.Command("docker", "info", "--format", "{{.CgroupDriver}}")
	rr, err := r.Runner.RunCmd(c)
	if err != nil {
		return "", err
	}
	return strings.Split(rr.Stdout.String(), "\n")[0], nil
}

// KubeletOptions returns kubelet options for a runtime.
func (r *Docker) KubeletOptions() map[string]string {
	return map[string]string{
		"container-runtime": "docker",
	}
}

// ListContainers returns a list of containers
func (r *Docker) ListContainers(o ListOptions) ([]string, error) {
	args := []string{"ps"}
	switch o.State {
	case All:
		args = append(args, "-a")
	case Running:
		args = append(args, "--filter", "status=running")
	case Paused:
		args = append(args, "--filter", "status=paused")
	}

	nameFilter := KubernetesContainerPrefix + o.Name
	if len(o.Namespaces) > 0 {
		// Example result: k8s.*(kube-system|kubernetes-dashboard)
		nameFilter = fmt.Sprintf("%s.*_(%s)_", nameFilter, strings.Join(o.Namespaces, "|"))
	}

	args = append(args, fmt.Sprintf("--filter=name=%s", nameFilter), "--format={{.ID}}")
	rr, err := r.Runner.RunCmd(exec.Command("docker", args...))
	if err != nil {
		return nil, errors.Wrapf(err, "docker")
	}
	var ids []string
	for _, line := range strings.Split(rr.Stdout.String(), "\n") {
		if line != "" {
			ids = append(ids, line)
		}
	}
	return ids, nil
}

// KillContainers forcibly removes a running container based on ID
func (r *Docker) KillContainers(ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	klog.Infof("Killing containers: %s", ids)
	args := append([]string{"rm", "-f"}, ids...)
	c := exec.Command("docker", args...)
	if _, err := r.Runner.RunCmd(c); err != nil {
		return errors.Wrap(err, "Killing containers docker.")
	}
	return nil
}

// StopContainers stops a running container based on ID
func (r *Docker) StopContainers(ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	klog.Infof("Stopping containers: %s", ids)
	args := append([]string{"stop"}, ids...)
	c := exec.Command("docker", args...)
	if _, err := r.Runner.RunCmd(c); err != nil {
		return errors.Wrap(err, "docker")
	}
	return nil
}

// PauseContainers pauses a running container based on ID
func (r *Docker) PauseContainers(ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	klog.Infof("Pausing containers: %s", ids)
	args := append([]string{"pause"}, ids...)
	c := exec.Command("docker", args...)
	if _, err := r.Runner.RunCmd(c); err != nil {
		return errors.Wrap(err, "docker")
	}
	return nil
}

// UnpauseContainers unpauses a container based on ID
func (r *Docker) UnpauseContainers(ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	klog.Infof("Unpausing containers: %s", ids)
	args := append([]string{"unpause"}, ids...)
	c := exec.Command("docker", args...)
	if _, err := r.Runner.RunCmd(c); err != nil {
		return errors.Wrap(err, "docker")
	}
	return nil
}

// ContainerLogCmd returns the command to retrieve the log for a container based on ID
func (r *Docker) ContainerLogCmd(id string, len int, follow bool) string {
	var cmd strings.Builder
	cmd.WriteString("docker logs ")
	if len > 0 {
		cmd.WriteString(fmt.Sprintf("--tail %d ", len))
	}
	if follow {
		cmd.WriteString("--follow ")
	}

	cmd.WriteString(id)
	return cmd.String()
}

// SystemLogCmd returns the command to retrieve system logs
func (r *Docker) SystemLogCmd(len int) string {
	return fmt.Sprintf("sudo journalctl -u docker -n %d", len)
}

// ForceSystemd forces the docker daemon to use systemd as cgroup manager
func (r *Docker) forceSystemd() error {
	klog.Infof("Forcing docker to use systemd as cgroup manager...")
	daemonConfig := `{
"exec-opts": ["native.cgroupdriver=systemd"],
"log-driver": "json-file",
"log-opts": {
	"max-size": "100m"
},
"storage-driver": "overlay2"
}
`
	ma := assets.NewMemoryAsset([]byte(daemonConfig), "/etc/docker", "daemon.json", "0644")
	return r.Runner.Copy(ma)
}

// Preload preloads docker with k8s images:
// 1. Copy over the preloaded tarball into the VM
// 2. Extract the preloaded tarball to the correct directory
// 3. Remove the tarball within the VM
func (r *Docker) Preload(cfg config.KubernetesConfig) error {
	if !download.PreloadExists(cfg.KubernetesVersion, cfg.ContainerRuntime) {
		return nil
	}
	k8sVersion := cfg.KubernetesVersion
	cRuntime := cfg.ContainerRuntime

	// If images already exist, return
	images, err := images.Kubeadm(cfg.ImageRepository, k8sVersion)
	if err != nil {
		return errors.Wrap(err, "getting images")
	}
	if dockerImagesPreloaded(r.Runner, images) {
		klog.Info("Images already preloaded, skipping extraction")
		return nil
	}

	refStore := docker.NewStorage(r.Runner)
	if err := refStore.Save(); err != nil {
		klog.Infof("error saving reference store: %v", err)
	}

	tarballPath := download.TarballPath(k8sVersion, cRuntime)
	targetDir := "/"
	targetName := "preloaded.tar.lz4"
	dest := path.Join(targetDir, targetName)

	c := exec.Command("which", "lz4")
	if _, err := r.Runner.RunCmd(c); err != nil {
		return NewErrISOFeature("lz4")
	}

	// Copy over tarball into host
	fa, err := assets.NewFileAsset(tarballPath, targetDir, targetName, "0644")
	if err != nil {
		return errors.Wrap(err, "getting file asset")
	}
	t := time.Now()
	if err := r.Runner.Copy(fa); err != nil {
		return errors.Wrap(err, "copying file")
	}
	klog.Infof("Took %f seconds to copy over tarball", time.Since(t).Seconds())

	// extract the tarball to /var in the VM
	if rr, err := r.Runner.RunCmd(exec.Command("sudo", "tar", "-I", "lz4", "-C", "/var", "-xf", dest)); err != nil {
		return errors.Wrapf(err, "extracting tarball: %s", rr.Output())
	}

	//  remove the tarball in the VM
	if err := r.Runner.Remove(fa); err != nil {
		klog.Infof("error removing tarball: %v", err)
	}

	// save new reference store again
	if err := refStore.Save(); err != nil {
		klog.Infof("error saving reference store: %v", err)
	}
	// update reference store
	if err := refStore.Update(); err != nil {
		klog.Infof("error updating reference store: %v", err)
	}
	return r.Restart()
}

// dockerImagesPreloaded returns true if all images have been preloaded
func dockerImagesPreloaded(runner command.Runner, images []string) bool {
	rr, err := runner.RunCmd(exec.Command("docker", "images", "--format", "{{.Repository}}:{{.Tag}}"))
	if err != nil {
		return false
	}
	preloadedImages := map[string]struct{}{}
	for _, i := range strings.Split(rr.Stdout.String(), "\n") {
		i = trimDockerIO(i)
		preloadedImages[i] = struct{}{}
	}

	klog.Infof("Got preloaded images: %s", rr.Output())

	// Make sure images == imgs
	for _, i := range images {
		i = trimDockerIO(i)
		if _, ok := preloadedImages[i]; !ok {
			klog.Infof("%s wasn't preloaded", i)
			return false
		}
	}
	return true
}

// Remove docker.io prefix since it won't be included in images names
// when we call 'docker images'
func trimDockerIO(name string) string {
	name = strings.TrimPrefix(name, "docker.io/")
	return name
}

func dockerBoundToContainerd(runner command.Runner) bool {
	// NOTE: assumes systemd
	rr, err := runner.RunCmd(exec.Command("sudo", "systemctl", "cat", "docker.service"))
	if err != nil {
		klog.Warningf("unable to check if docker is bound to containerd")
		return false
	}

	if strings.Contains(rr.Stdout.String(), "\nBindsTo=containerd") {
		return true
	}

	return false
}

// ImagesPreloaded returns true if all images have been preloaded
func (r *Docker) ImagesPreloaded(images []string) bool {
	return dockerImagesPreloaded(r.Runner, images)
}
