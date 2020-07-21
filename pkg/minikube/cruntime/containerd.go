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
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os/exec"
	"path"
	"strings"
	"text/template"
	"time"

	"github.com/blang/semver"
	"github.com/golang/glog"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/bootstrapper/images"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/download"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/sysinit"
)

const (
	containerdNamespaceRoot = "/run/containerd/runc/k8s.io"
	// ContainerdConfFile is the path to the containerd configuration
	containerdConfigFile     = "/etc/containerd/config.toml"
	containerdConfigTemplate = `root = "/var/lib/containerd"
state = "/run/containerd"
oom_score = 0

[grpc]
  address = "/run/containerd/containerd.sock"
  uid = 0
  gid = 0
  max_recv_message_size = 16777216
  max_send_message_size = 16777216

[debug]
  address = ""
  uid = 0
  gid = 0
  level = ""

[metrics]
  address = ""
  grpc_histogram = false

[cgroup]
  path = ""

[plugins]
  [plugins.cgroups]
    no_prometheus = false
  [plugins.cri]
    stream_server_address = ""
    stream_server_port = "10010"
    enable_selinux = false
    sandbox_image = "{{ .PodInfraContainerImage }}"
    stats_collect_period = 10
    systemd_cgroup = false
    enable_tls_streaming = false
    max_container_log_line_size = 16384
    [plugins.cri.containerd]
      snapshotter = "overlayfs"
      no_pivot = true
      [plugins.cri.containerd.default_runtime]
        runtime_type = "io.containerd.runtime.v1.linux"
        runtime_engine = ""
        runtime_root = ""
      [plugins.cri.containerd.untrusted_workload_runtime]
        runtime_type = ""
        runtime_engine = ""
        runtime_root = ""
    [plugins.cri.cni]
      bin_dir = "/opt/cni/bin"
      conf_dir = "/etc/cni/net.d"
      conf_template = ""
    [plugins.cri.registry]
      [plugins.cri.registry.mirrors]
        [plugins.cri.registry.mirrors."docker.io"]
          endpoint = ["https://registry-1.docker.io"]
  [plugins.diff-service]
    default = ["walking"]
  [plugins.linux]
    shim = "containerd-shim"
    runtime = "runc"
    runtime_root = ""
    no_shim = false
    shim_debug = false
  [plugins.scheduler]
    pause_threshold = 0.02
    deletion_threshold = 0
    mutation_threshold = 100
    schedule_delay = "0s"
    startup_delay = "100ms"
`
)

// Containerd contains containerd runtime state
type Containerd struct {
	Socket            string
	Runner            CommandRunner
	ImageRepository   string
	KubernetesVersion semver.Version
	Init              sysinit.Manager
}

// Name is a human readable name for containerd
func (r *Containerd) Name() string {
	return "containerd"
}

// Style is the console style for containerd
func (r *Containerd) Style() out.StyleEnum {
	return out.Containerd
}

// Version retrieves the current version of this runtime
func (r *Containerd) Version() (string, error) {
	c := exec.Command("containerd", "--version")
	rr, err := r.Runner.RunCmd(c)
	if err != nil {
		return "", errors.Wrapf(err, "containerd check version.")
	}
	// containerd github.com/containerd/containerd v1.2.0 c4446665cb9c30056f4998ed953e6d4ff22c7c39
	words := strings.Split(rr.Stdout.String(), " ")
	if len(words) >= 4 && words[0] == "containerd" {
		return strings.Replace(words[2], "v", "", 1), nil
	}
	return "", fmt.Errorf("unknown version: %q", rr.Stdout.String())
}

// SocketPath returns the path to the socket file for containerd
func (r *Containerd) SocketPath() string {
	if r.Socket != "" {
		return r.Socket
	}
	return "/run/containerd/containerd.sock"
}

// Active returns if containerd is active on the host
func (r *Containerd) Active() bool {
	return r.Init.Active("containerd")
}

// Available returns an error if it is not possible to use this runtime on a host
func (r *Containerd) Available() error {
	c := exec.Command("which", "containerd")
	if _, err := r.Runner.RunCmd(c); err != nil {
		return errors.Wrap(err, "check containerd availability.")
	}
	return nil
}

// generateContainerdConfig sets up /etc/containerd/config.toml
func generateContainerdConfig(cr CommandRunner, imageRepository string, kv semver.Version) error {
	cPath := containerdConfigFile
	t, err := template.New("containerd.config.toml").Parse(containerdConfigTemplate)
	if err != nil {
		return err
	}
	pauseImage := images.Pause(kv, imageRepository)
	opts := struct{ PodInfraContainerImage string }{PodInfraContainerImage: pauseImage}
	var b bytes.Buffer
	if err := t.Execute(&b, opts); err != nil {
		return err
	}
	c := exec.Command("/bin/bash", "-c", fmt.Sprintf("sudo mkdir -p %s && printf %%s \"%s\" | base64 -d | sudo tee %s", path.Dir(cPath), base64.StdEncoding.EncodeToString(b.Bytes()), cPath))
	if _, err := cr.RunCmd(c); err != nil {
		return errors.Wrap(err, "generate containerd cfg.")
	}
	return nil
}

// Enable idempotently enables containerd on a host
func (r *Containerd) Enable(disOthers, _ bool) error {
	if disOthers {
		if err := disableOthers(r, r.Runner); err != nil {
			glog.Warningf("disableOthers: %v", err)
		}
	}
	if err := populateCRIConfig(r.Runner, r.SocketPath()); err != nil {
		return err
	}
	if err := generateContainerdConfig(r.Runner, r.ImageRepository, r.KubernetesVersion); err != nil {
		return err
	}
	if err := enableIPForwarding(r.Runner); err != nil {
		return err
	}

	// Otherwise, containerd will fail API requests with 'Unimplemented'
	return r.Init.Restart("containerd")
}

// Disable idempotently disables containerd on a host
func (r *Containerd) Disable() error {
	return r.Init.ForceStop("containerd")
}

// ImageExists checks if an image exists, expected input format
func (r *Containerd) ImageExists(name string, sha string) bool {
	c := exec.Command("/bin/bash", "-c", fmt.Sprintf("sudo ctr -n=k8s.io images check | grep %s | grep %s", name, sha))
	if _, err := r.Runner.RunCmd(c); err != nil {
		return false
	}
	return true
}

// LoadImage loads an image into this runtime
func (r *Containerd) LoadImage(path string) error {
	glog.Infof("Loading image: %s", path)
	c := exec.Command("sudo", "ctr", "-n=k8s.io", "images", "import", path)
	if _, err := r.Runner.RunCmd(c); err != nil {
		return errors.Wrapf(err, "ctr images import")
	}
	return nil
}

// CGroupDriver returns cgroup driver ("cgroupfs" or "systemd")
func (r *Containerd) CGroupDriver() (string, error) {
	info, err := getCRIInfo(r.Runner)
	if err != nil {
		return "", err
	}
	if info["config"] == nil {
		return "", errors.Wrapf(err, "missing config")
	}
	config, ok := info["config"].(map[string]interface{})
	if !ok {
		return "", errors.Wrapf(err, "config not map")
	}
	cgroupManager := "cgroupfs" // default
	switch config["systemdCgroup"] {
	case false:
		cgroupManager = "cgroupfs"
	case true:
		cgroupManager = "systemd"
	}
	return cgroupManager, nil
}

// KubeletOptions returns kubelet options for a containerd
func (r *Containerd) KubeletOptions() map[string]string {
	return map[string]string{
		"container-runtime":          "remote",
		"container-runtime-endpoint": fmt.Sprintf("unix://%s", r.SocketPath()),
		"image-service-endpoint":     fmt.Sprintf("unix://%s", r.SocketPath()),
		"runtime-request-timeout":    "15m",
	}
}

// ListContainers returns a list of managed by this container runtime
func (r *Containerd) ListContainers(o ListOptions) ([]string, error) {
	return listCRIContainers(r.Runner, containerdNamespaceRoot, o)
}

// PauseContainers pauses a running container based on ID
func (r *Containerd) PauseContainers(ids []string) error {
	return pauseCRIContainers(r.Runner, containerdNamespaceRoot, ids)
}

// UnpauseContainers unpauses a running container based on ID
func (r *Containerd) UnpauseContainers(ids []string) error {
	return unpauseCRIContainers(r.Runner, containerdNamespaceRoot, ids)
}

// KillContainers removes containers based on ID
func (r *Containerd) KillContainers(ids []string) error {
	return killCRIContainers(r.Runner, ids)
}

// StopContainers stops containers based on ID
func (r *Containerd) StopContainers(ids []string) error {
	return stopCRIContainers(r.Runner, ids)
}

// ContainerLogCmd returns the command to retrieve the log for a container based on ID
func (r *Containerd) ContainerLogCmd(id string, len int, follow bool) string {
	return criContainerLogCmd(r.Runner, id, len, follow)
}

// SystemLogCmd returns the command to retrieve system logs
func (r *Containerd) SystemLogCmd(len int) string {
	return fmt.Sprintf("sudo journalctl -u containerd -n %d", len)
}

// Preload preloads the container runtime with k8s images
func (r *Containerd) Preload(cfg config.KubernetesConfig) error {
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
	if containerdImagesPreloaded(r.Runner, images) {
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

	return r.Restart()
}

// Restart restarts Docker on a host
func (r *Containerd) Restart() error {
	return r.Init.Restart("containerd")
}

// containerdImagesPreloaded returns true if all images have been preloaded
func containerdImagesPreloaded(runner command.Runner, images []string) bool {
	rr, err := runner.RunCmd(exec.Command("sudo", "crictl", "images", "--output", "json"))
	if err != nil {
		return false
	}
	type criImages struct {
		Images []struct {
			ID          string      `json:"id"`
			RepoTags    []string    `json:"repoTags"`
			RepoDigests []string    `json:"repoDigests"`
			Size        string      `json:"size"`
			UID         interface{} `json:"uid"`
			Username    string      `json:"username"`
		} `json:"images"`
	}

	var jsonImages criImages
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
	glog.Infof("all images are preloaded for containerd runtime.")
	return true
}
