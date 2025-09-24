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
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/blang/semver/v4"
	"github.com/pkg/errors"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/bootstrapper/images"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/download"
	"k8s.io/minikube/pkg/minikube/style"
	"k8s.io/minikube/pkg/minikube/sysinit"
)

const (
	// crioConfigFile is the path to the CRI-O configuration
	crioConfigFile = "/etc/crio/crio.conf.d/02-crio.conf"
)

// CRIO contains CRIO runtime state
type CRIO struct {
	Socket            string
	Runner            CommandRunner
	ImageRepository   string
	KubernetesVersion semver.Version
	Init              sysinit.Manager
}

// generateCRIOConfig sets up pause image and cgroup manager for cri-o in crioConfigFile
func generateCRIOConfig(cr CommandRunner, imageRepository string, kv semver.Version, cgroupDriver string) error {
	pauseImage := images.Pause(kv, imageRepository)
	klog.Infof("configure cri-o to use %q pause image...", pauseImage)
	c := exec.Command("sh", "-c", fmt.Sprintf(`sudo sed -i 's|^.*pause_image = .*$|pause_image = %q|' %s`, pauseImage, crioConfigFile))
	if _, err := cr.RunCmd(c); err != nil {
		return errors.Wrap(err, "update pause_image")
	}

	// configure cgroup driver
	if cgroupDriver == constants.UnknownCgroupDriver {
		klog.Warningf("unable to configure cri-o to use unknown cgroup driver, will use default %q instead", constants.DefaultCgroupDriver)
		cgroupDriver = constants.DefaultCgroupDriver
	}
	klog.Infof("configuring cri-o to use %q as cgroup driver...", cgroupDriver)
	if _, err := cr.RunCmd(exec.Command("sh", "-c", fmt.Sprintf(`sudo sed -i 's|^.*cgroup_manager = .*$|cgroup_manager = %q|' %s`, cgroupDriver, crioConfigFile))); err != nil {
		return errors.Wrap(err, "configuring cgroup_manager")
	}
	// explicitly set conmon_cgroup to avoid errors like:
	// - level=fatal msg="Validating runtime config: conmon cgroup should be 'pod' or a systemd slice"
	// - level=fatal msg="Validating runtime config: cgroupfs manager conmon cgroup should be 'pod' or empty"
	// ref: https://github.com/cri-o/cri-o/pull/3940
	// ref: https://github.com/cri-o/cri-o/issues/6047
	// ref: https://kubernetes.io/docs/setup/production-environment/container-runtimes/#cgroup-driver
	if _, err := cr.RunCmd(exec.Command("sh", "-c", fmt.Sprintf(`sudo sed -i '/conmon_cgroup = .*/d' %s`, crioConfigFile))); err != nil {
		return errors.Wrap(err, "removing conmon_cgroup")
	}
	if _, err := cr.RunCmd(exec.Command("sh", "-c", fmt.Sprintf(`sudo sed -i '/cgroup_manager = .*/a conmon_cgroup = %q' %s`, "pod", crioConfigFile))); err != nil {
		return errors.Wrap(err, "configuring conmon_cgroup")
	}

	// we might still want to try removing '/etc/cni/net.mk' in case of upgrade from previous minikube version that had/used it
	if _, err := cr.RunCmd(exec.Command("sh", "-c", `sudo rm -rf /etc/cni/net.mk`)); err != nil {
		klog.Warningf("unable to remove /etc/cni/net.mk directory: %v", err)
	}

	// add 'net.ipv4.ip_unprivileged_port_start=0' sysctl so that containers that run with non-root user can bind to otherwise privilege ports (like coredns v1.11.0+)
	// note: 'net.ipv4.ip_unprivileged_port_start' sysctl was marked as safe since Kubernetes v1.22 (Aug 4, 2021) (ref: https://github.com/kubernetes/kubernetes/blob/master/CHANGELOG/CHANGELOG-1.22.md#feature-9)
	// note: cri-o supports 'default_sysctls' option since v1.12.0 (Oct 19, 2018) (ref: https://github.com/cri-o/cri-o/releases/tag/v1.12.0; https://github.com/cri-o/cri-o/pull/1721)
	if kv.GTE(semver.Version{Major: 1, Minor: 22}) {
		// remove any existing 'net.ipv4.ip_unprivileged_port_start' settings
		if _, err := cr.RunCmd(exec.Command("sh", "-c", fmt.Sprintf(`sudo sed -i '/^ *"net.ipv4.ip_unprivileged_port_start=.*"/d' %s`, crioConfigFile))); err != nil {
			return errors.Wrap(err, "removing net.ipv4.ip_unprivileged_port_start")
		}
		// insert 'default_sysctls' list, if not already present
		if _, err := cr.RunCmd(exec.Command("sh", "-c", fmt.Sprintf(`sudo grep -q "^ *default_sysctls" %s || sudo sed -i '/conmon_cgroup = .*/a default_sysctls = \[\n\]' %s`, crioConfigFile, crioConfigFile))); err != nil {
			return errors.Wrap(err, "inserting default_sysctls")
		}
		// add 'net.ipv4.ip_unprivileged_port_start' to 'default_sysctls' list
		if _, err := cr.RunCmd(exec.Command("sh", "-c", fmt.Sprintf(`sudo sed -i -r 's|^default_sysctls *= *\[|&\n  "net.ipv4.ip_unprivileged_port_start=0",|' %s`, crioConfigFile))); err != nil {
			return errors.Wrap(err, "configuring net.ipv4.ip_unprivileged_port_start")
		}
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
		return "", errors.Wrap(err, "crio version")
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
		return errors.Wrapf(err, "check crio available")
	}
	return checkCNIPlugins(r.KubernetesVersion)
}

// Active returns if CRIO is active on the host
func (r *CRIO) Active() bool {
	return r.Init.Active("crio")
}

// enableIPForwarding configures IP forwarding, which is handled normally by Docker
// Context: https://github.com/kubernetes/kubeadm/issues/1062
func enableIPForwarding(cr CommandRunner) error {
	// The bridge-netfilter module enables (ip|ip6)tables rules to apply on Linux bridges.
        // NOTE: br_netfilter isn't available everywhere (e.g., some WSL2 kernels) – treat as best-effort.
        if _, err := cr.RunCmd(exec.Command("sudo", "modprobe", "br_netfilter")); err != nil {
                klog.Warningf("modprobe br_netfilter failed (may be OK on this kernel): %v", err)
        }

        // Enable bridge netfilter hooks for both IPv4 and IPv6, and enable forwarding.
        // Best-effort: warn but don't fail hard if a sysctl isn't present.
        sysctls := []string{
                "sysctl -w net.bridge.bridge-nf-call-iptables=1",
                "sysctl -w net.bridge.bridge-nf-call-ip6tables=1",
                "sysctl -w net.ipv4.ip_forward=1",
                "sysctl -w net.ipv6.conf.all.forwarding=1",
        }
        for _, s := range sysctls {
                if _, err := cr.RunCmd(exec.Command("sudo", "sh", "-c", s)); err != nil {
                        klog.Warningf("failed to run %q (continuing): %v", s, err)
                }
        }
        return nil
}

// enableRootless enables configurations for running CRI-O in Rootless Docker.
//
// 1. Create /etc/systemd/system/crio.service.d/10-rootless.conf to set _CRIO_ROOTLESS=1
// 2. Reload systemd
//
// See https://kubernetes.io/docs/tasks/administer-cluster/kubelet-in-userns/#configuring-cri
func (r *CRIO) enableRootless() error {
	files := map[string]string{
		"/etc/systemd/system/crio.service.d/10-rootless.conf": `[Service]
Environment="_CRIO_ROOTLESS=1"
`,
	}
	for target, content := range files {
		targetDir := filepath.Dir(target)
		c := exec.Command("sudo", "mkdir", "-p", targetDir)
		if _, err := r.Runner.RunCmd(c); err != nil {
			return errors.Wrapf(err, "failed to create directory %q", targetDir)
		}
		asset := assets.NewMemoryAssetTarget([]byte(content), target, "0644")
		err := r.Runner.Copy(asset)
		asset.Close()
		if err != nil {
			return errors.Wrapf(err, "failed to create %q", target)
		}
	}
	// reload systemd to apply our changes on /etc/systemd
	if err := r.Init.Reload("crio"); err != nil {
		return err
	}
	if r.Init.Active("crio") {
		if err := r.Init.Restart("crio"); err != nil {
			return err
		}
	}
	return nil
}

// Enable idempotently enables CRIO on a host
func (r *CRIO) Enable(disOthers bool, cgroupDriver string, inUserNamespace bool) error {
	if disOthers {
		if err := disableOthers(r, r.Runner); err != nil {
			klog.Warningf("disableOthers: %v", err)
		}
	}
	if err := populateCRIConfig(r.Runner, r.SocketPath()); err != nil {
		return err
	}
	if err := generateCRIOConfig(r.Runner, r.ImageRepository, r.KubernetesVersion, cgroupDriver); err != nil {
		return err
	}
	if err := enableIPForwarding(r.Runner); err != nil {
		return err
	}
	if inUserNamespace {
		if err := CheckKernelCompatibility(r.Runner, 5, 11); err != nil {
			// For using overlayfs
			return fmt.Errorf("kernel >= 5.11 is required for rootless mode: %w", err)
		}
		if err := CheckKernelCompatibility(r.Runner, 5, 13); err != nil {
			// For avoiding SELinux error with overlayfs
			klog.Warningf("kernel >= 5.13 is recommended for rootless mode %v", err)
		}
		if err := r.enableRootless(); err != nil {
			return err
		}
	}
	// NOTE: before we start crio explicitly here, crio might be already started automatically
	return r.Init.Restart("crio")
}

// Disable idempotently disables CRIO on a host
func (r *CRIO) Disable() error {
	return r.Init.ForceStop("crio")
}

// ImageExists checks if image exists based on image name and optionally image sha
func (r *CRIO) ImageExists(name string, sha string) bool {
	// expected output looks like [NAME@sha256:SHA]
	c := exec.Command("sudo", "podman", "image", "inspect", "--format", "{{.Id}}", name)
	rr, err := r.Runner.RunCmd(c)
	if err != nil {
		return false
	}
	if sha != "" && !strings.Contains(rr.Output(), sha) {
		return false
	}
	return true
}

// ListImages returns a list of images managed by this container runtime
func (r *CRIO) ListImages(ListImagesOptions) ([]ListImage, error) {
	return listCRIImages(r.Runner)
}

// LoadImage loads an image into this runtime
func (r *CRIO) LoadImage(imgPath string) error {
	klog.Infof("Loading image: %s", imgPath)
	c := exec.Command("sudo", "podman", "load", "-i", imgPath)
	if _, err := r.Runner.RunCmd(c); err != nil {
		return errors.Wrap(err, "crio load image")
	}
	return nil
}

// PullImage pulls an image
func (r *CRIO) PullImage(name string) error {
	return pullCRIImage(r.Runner, name)
}

// SaveImage saves an image from this runtime
func (r *CRIO) SaveImage(name string, destPath string) error {
	klog.Infof("Saving image %s: %s", name, destPath)
	c := exec.Command("sudo", "podman", "save", name, "-o", destPath)
	if _, err := r.Runner.RunCmd(c); err != nil {
		return errors.Wrap(err, "crio save image")
	}
	return nil
}

// RemoveImage removes a image
func (r *CRIO) RemoveImage(name string) error {
	return removeCRIImage(r.Runner, name)
}

// TagImage tags an image in this runtime
func (r *CRIO) TagImage(source string, target string) error {
	klog.Infof("Tagging image %s: %s", source, target)
	c := exec.Command("sudo", "podman", "tag", source, target)
	if _, err := r.Runner.RunCmd(c); err != nil {
		return errors.Wrap(err, "crio tag image")
	}
	return nil
}

// BuildImage builds an image into this runtime
func (r *CRIO) BuildImage(src string, file string, tag string, push bool, env []string, opts []string) error {
	klog.Infof("Building image: %s", src)
	args := []string{"podman", "build"}
	if file != "" {
		args = append(args, "-f", file)
	}
	if tag != "" {
		args = append(args, "-t", tag)
	}
	args = append(args, src)
	for _, opt := range opts {
		args = append(args, "--"+opt)
	}
	args = append(args, "--cgroup-manager=cgroupfs")
	c := exec.Command("sudo", args...)
	e := os.Environ()
	e = append(e, env...)
	c.Env = e
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if _, err := r.Runner.RunCmd(c); err != nil {
		return errors.Wrap(err, "crio build image")
	}
	if tag != "" && push {
		c := exec.Command("sudo", "podman", "push", tag)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if _, err := r.Runner.RunCmd(c); err != nil {
			return errors.Wrap(err, "crio push image")
		}
	}
	return nil
}

// PushImage pushes an image
func (r *CRIO) PushImage(name string) error {
	klog.Infof("Pushing image %s", name)
	c := exec.Command("sudo", "podman", "push", name)
	if _, err := r.Runner.RunCmd(c); err != nil {
		return errors.Wrap(err, "crio push image")
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
	cgroupManager := "systemd" // default
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
	return kubeletCRIOptions(r, r.KubernetesVersion)
}

// ListContainers returns a list of managed by this container runtime
func (r *CRIO) ListContainers(o ListContainersOptions) ([]string, error) {
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
func (r *CRIO) ContainerLogCmd(id string, length int, follow bool) string {
	return criContainerLogCmd(r.Runner, id, length, follow)
}

// SystemLogCmd returns the command to retrieve system logs
func (r *CRIO) SystemLogCmd(length int) string {
	return fmt.Sprintf("sudo journalctl -u crio -n %d", length)
}

// Preload preloads the container runtime with k8s images
func (r *CRIO) Preload(cc config.ClusterConfig) error {
	if !download.PreloadExists(cc.KubernetesConfig.KubernetesVersion, cc.KubernetesConfig.ContainerRuntime, cc.Driver) {
		return nil
	}

	k8sVersion := cc.KubernetesConfig.KubernetesVersion
	cRuntime := cc.KubernetesConfig.ContainerRuntime

	// If images already exist, return
	imgs, err := images.Kubeadm(cc.KubernetesConfig.ImageRepository, k8sVersion)
	if err != nil {
		return errors.Wrap(err, "getting images")
	}
	if crioImagesPreloaded(r.Runner, imgs) {
		klog.Info("Images already preloaded, skipping extraction")
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
	defer func() {
		if err := fa.Close(); err != nil {
			klog.Warningf("error closing the file %s: %v", fa.GetSourcePath(), err)
		}
	}()

	t := time.Now()
	if err := r.Runner.Copy(fa); err != nil {
		return errors.Wrap(err, "copying file")
	}
	klog.Infof("duration metric: took %s to copy over tarball", time.Since(t))

	t = time.Now()
	// extract the tarball to /var in the VM
	if rr, err := r.Runner.RunCmd(exec.Command("sudo", "tar", "--xattrs", "--xattrs-include", "security.capability", "-I", "lz4", "-C", "/var", "-xf", dest)); err != nil {
		return errors.Wrapf(err, "extracting tarball: %s", rr.Output())
	}
	klog.Infof("duration metric: took %s to extract the tarball", time.Since(t))

	//  remove the tarball in the VM
	if err := r.Runner.Remove(fa); err != nil {
		klog.Infof("error removing tarball: %v", err)
	}

	return nil
}

// crioImagesPreloaded returns true if all images have been preloaded
func crioImagesPreloaded(runner command.Runner, imgs []string) bool {
	rr, err := runner.RunCmd(exec.Command("sudo", "crictl", "images", "--output", "json"))
	if err != nil {
		return false
	}

	var jsonImages crictlImages
	err = json.Unmarshal(rr.Stdout.Bytes(), &jsonImages)
	if err != nil {
		klog.Errorf("failed to unmarshal images, will assume images are not preloaded")
		return false
	}

	// Make sure images == imgs
	for _, i := range imgs {
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
			klog.Infof("couldn't find preloaded image for %q. assuming images are not preloaded.", i)
			return false
		}
	}
	klog.Infof("all images are preloaded for cri-o runtime.")
	return true
}

// ImagesPreloaded returns true if all images have been preloaded
func (r *CRIO) ImagesPreloaded(imgs []string) bool {
	return crioImagesPreloaded(r.Runner, imgs)
}
