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
	"encoding/json"
	"fmt"
	"net"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/blang/semver"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/bootstrapper/images"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/download"
	"k8s.io/minikube/pkg/minikube/style"
	"k8s.io/minikube/pkg/minikube/sysinit"
)

const (
	// CRIOConfFile is the path to the CRI-O configuration
	crioConfigFile = "/etc/crio/crio.conf"
)

// CRIO contains CRIO runtime state
type CRIO struct {
	Socket            string
	Runner            CommandRunner
	ImageRepository   string
	KubernetesVersion semver.Version
	Init              sysinit.Manager
}

// generateCRIOConfig sets up /etc/crio/crio.conf
func generateCRIOConfig(cr CommandRunner, imageRepository string, kv semver.Version) error {
	cPath := crioConfigFile
	pauseImage := images.Pause(kv, imageRepository)

	c := exec.Command("/bin/bash", "-c", fmt.Sprintf("sudo sed -e 's|^pause_image = .*$|pause_image = \"%s\"|' -i %s", pauseImage, cPath))
	if _, err := cr.RunCmd(c); err != nil {
		return errors.Wrap(err, "generateCRIOConfig.")
	}
	return nil
}

// Name is a human readable name for CRIO
func (r *CRIO) Name() string {
	return "CRI-O"
}

// Style is the console style for CRIO
func (r *CRIO) Style() style.Enum {
	return style.CRIO
}

// Version retrieves the current version of this runtime
func (r *CRIO) Version() (string, error) {
	c := exec.Command("crio", "--version")
	rr, err := r.Runner.RunCmd(c)
	if err != nil {
		return "", errors.Wrap(err, "crio version.")
	}

	// crio version 1.13.0
	// commit: ""
	line := strings.Split(rr.Stdout.String(), "\n")[0]
	return strings.Replace(line, "crio version ", "", 1), nil
}

// SocketPath returns the path to the socket file for CRIO
func (r *CRIO) SocketPath() string {
	if r.Socket != "" {
		return r.Socket
	}
	return "/var/run/crio/crio.sock"
}

// Available returns an error if it is not possible to use this runtime on a host
func (r *CRIO) Available() error {
	c := exec.Command("which", "crio")
	if _, err := r.Runner.RunCmd(c); err != nil {
		return errors.Wrapf(err, "check crio available.")
	}
	return nil
}

// Active returns if CRIO is active on the host
func (r *CRIO) Active() bool {
	return r.Init.Active("crio")
}

// Enable idempotently enables CRIO on a host
func (r *CRIO) Enable(disOthers, _ bool) error {
	if disOthers {
		if err := disableOthers(r, r.Runner); err != nil {
			glog.Warningf("disableOthers: %v", err)
		}
	}
	if err := populateCRIConfig(r.Runner, r.SocketPath()); err != nil {
		return err
	}
	if err := generateCRIOConfig(r.Runner, r.ImageRepository, r.KubernetesVersion); err != nil {
		return err
	}
	if err := enableIPForwarding(r.Runner); err != nil {
		return err
	}
	return r.Init.Start("crio")
}

// Disable idempotently disables CRIO on a host
func (r *CRIO) Disable() error {
	return r.Init.ForceStop("crio")
}

// ImageExists checks if an image exists
func (r *CRIO) ImageExists(name string, sha string) bool {
	// expected output looks like [NAME@sha256:SHA]
	c := exec.Command("sudo", "podman", "image", "inspect", "--format", "{{.Id}}", name)
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
func (r *CRIO) LoadImage(path string) error {
	glog.Infof("Loading image: %s", path)
	c := exec.Command("sudo", "podman", "load", "-i", path)
	if _, err := r.Runner.RunCmd(c); err != nil {
		return errors.Wrap(err, "crio load image")
	}
	return nil
}

// CGroupDriver returns cgroup driver ("cgroupfs" or "systemd")
func (r *CRIO) CGroupDriver() (string, error) {
	c := exec.Command("crio", "config")
	rr, err := r.Runner.RunCmd(c)
	if err != nil {
		return "", err
	}
	cgroupManager := "cgroupfs" // default
	for _, line := range strings.Split(rr.Stdout.String(), "\n") {
		if strings.HasPrefix(line, "cgroup_manager") {
			// cgroup_manager = "cgroupfs"
			f := strings.Split(strings.TrimSpace(line), " = ")
			if len(f) == 2 {
				cgroupManager = strings.Trim(f[1], "\"")
			}
		}
	}
	return cgroupManager, nil
}

// KubeletOptions returns kubelet options for a runtime.
func (r *CRIO) KubeletOptions() map[string]string {
	return map[string]string{
		"container-runtime":          "remote",
		"container-runtime-endpoint": r.SocketPath(),
		"image-service-endpoint":     r.SocketPath(),
		"runtime-request-timeout":    "15m",
	}
}

// ListContainers returns a list of managed by this container runtime
func (r *CRIO) ListContainers(o ListOptions) ([]string, error) {
	return listCRIContainers(r.Runner, "", o)
}

// PauseContainers pauses a running container based on ID
func (r *CRIO) PauseContainers(ids []string) error {
	return pauseCRIContainers(r.Runner, "", ids)
}

// UnpauseContainers unpauses a running container based on ID
func (r *CRIO) UnpauseContainers(ids []string) error {
	return unpauseCRIContainers(r.Runner, "", ids)
}

// KillContainers removes containers based on ID
func (r *CRIO) KillContainers(ids []string) error {
	return killCRIContainers(r.Runner, ids)
}

// StopContainers stops containers based on ID
func (r *CRIO) StopContainers(ids []string) error {
	return stopCRIContainers(r.Runner, ids)
}

// ContainerLogCmd returns the command to retrieve the log for a container based on ID
func (r *CRIO) ContainerLogCmd(id string, len int, follow bool) string {
	return criContainerLogCmd(r.Runner, id, len, follow)
}

// SystemLogCmd returns the command to retrieve system logs
func (r *CRIO) SystemLogCmd(len int) string {
	return fmt.Sprintf("sudo journalctl -u crio -n %d", len)
}

// Preload preloads the container runtime with k8s images
func (r *CRIO) Preload(cfg config.KubernetesConfig) error {
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
	if crioImagesPreloaded(r.Runner, images) {
		glog.Info("Images already preloaded, skipping extraction")
		return nil
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
	glog.Infof("Took %f seconds to copy over tarball", time.Since(t).Seconds())

	t = time.Now()
	// extract the tarball to /var in the VM
	if rr, err := r.Runner.RunCmd(exec.Command("sudo", "tar", "-I", "lz4", "-C", "/var", "-xvf", dest)); err != nil {
		return errors.Wrapf(err, "extracting tarball: %s", rr.Output())
	}
	glog.Infof("Took %f seconds t extract the tarball", time.Since(t).Seconds())

	//  remove the tarball in the VM
	if err := r.Runner.Remove(fa); err != nil {
		glog.Infof("error removing tarball: %v", err)
	}

	return nil
}

// crioImagesPreloaded returns true if all images have been preloaded
func crioImagesPreloaded(runner command.Runner, images []string) bool {
	rr, err := runner.RunCmd(exec.Command("sudo", "crictl", "images", "--output", "json"))
	if err != nil {
		return false
	}
	type crictlImages struct {
		Images []struct {
			ID          string      `json:"id"`
			RepoTags    []string    `json:"repoTags"`
			RepoDigests []string    `json:"repoDigests"`
			Size        string      `json:"size"`
			UID         interface{} `json:"uid"`
			Username    string      `json:"username"`
		} `json:"images"`
	}

	var jsonImages crictlImages
	err = json.Unmarshal(rr.Stdout.Bytes(), &jsonImages)
	if err != nil {
		glog.Errorf("failed to unmarshal images, will assume images are not preloaded")
		return false
	}

	// Make sure images == imgs
	for _, i := range images {
		found := false
		for _, ji := range jsonImages.Images {
			for _, rt := range ji.RepoTags {
				i = addRepoTagToImageName(i)
				if i == rt {
					found = true
					break
				}
			}
			if found {
				break
			}

		}
		if !found {
			glog.Infof("couldn't find preloaded image for %q. assuming images are not preloaded.", i)
			return false
		}
	}
	glog.Infof("all images are preloaded for cri-o runtime.")
	return true
}

func (r *CRIO) ImagesPreloaded(images []string) bool {
	return crioImagesPreloaded(r.Runner, images)
}

// UpdateCRIONet updates CRIO CNI network configuration and restarts it
func UpdateCRIONet(r CommandRunner, cidr string) error {
	glog.Infof("Updating CRIO to use CIDR: %q", cidr)
	ip, net, err := net.ParseCIDR(cidr)
	if err != nil {
		return errors.Wrap(err, "parse cidr")
	}

	oldNet := "10.88.0.0/16"
	oldGw := "10.88.0.1"

	newNet := cidr

	// Assume gateway is first IP in netmask (10.244.0.1, for instance)
	newGw := ip.Mask(net.Mask)
	newGw[3]++

	// Update subnets used by 100-crio-bridge.conf & 87-podman-bridge.conflist
	// avoids: "Error adding network: failed to set bridge addr: could not add IP address to \"cni0\": permission denied"
	sed := fmt.Sprintf("sed -i -e s#%s#%s# -e s#%s#%s# /etc/cni/net.d/*bridge*", oldNet, newNet, oldGw, newGw)
	if _, err := r.RunCmd(exec.Command("sudo", "/bin/bash", "-c", sed)); err != nil {
		glog.Errorf("netconf update failed: %v", err)
	}

	return sysinit.New(r).Restart("crio")
}
