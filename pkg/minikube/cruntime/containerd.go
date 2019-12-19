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
	"fmt"
	"os/exec"
	"path"
	"strings"
	"text/template"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/bootstrapper/images"
	"k8s.io/minikube/pkg/minikube/out"
)

const (
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
	KubernetesVersion string
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

// DefaultCNI returns whether to use CNI networking by default
func (r *Containerd) DefaultCNI() bool {
	return true
}

// Active returns if containerd is active on the host
func (r *Containerd) Active() bool {
	c := exec.Command("systemctl", "is-active", "--quiet", "service", "containerd")
	_, err := r.Runner.RunCmd(c)
	return err == nil
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
func generateContainerdConfig(cr CommandRunner, imageRepository string) error {
	cPath := containerdConfigFile
	t, err := template.New("containerd.config.toml").Parse(containerdConfigTemplate)
	if err != nil {
		return err
	}
	pauseImage := images.Pause(imageRepository)
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
func (r *Containerd) Enable(disOthers bool) error {
	if disOthers {
		if err := disableOthers(r, r.Runner); err != nil {
			glog.Warningf("disableOthers: %v", err)
		}
	}
	if err := populateCRIConfig(r.Runner, r.SocketPath()); err != nil {
		return err
	}
	if err := generateContainerdConfig(r.Runner, r.ImageRepository); err != nil {
		return err
	}
	if err := enableIPForwarding(r.Runner); err != nil {
		return err
	}
	// Otherwise, containerd will fail API requests with 'Unimplemented'
	c := exec.Command("sudo", "systemctl", "restart", "containerd")
	if _, err := r.Runner.RunCmd(c); err != nil {
		return errors.Wrap(err, "restart containerd")
	}
	return nil
}

// Disable idempotently disables containerd on a host
func (r *Containerd) Disable() error {
	c := exec.Command("sudo", "systemctl", "stop", "containerd")
	if _, err := r.Runner.RunCmd(c); err != nil {
		return errors.Wrapf(err, "stop containerd")
	}
	return nil
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
func (r *Containerd) ListContainers(filter string) ([]string, error) {
	return listCRIContainers(r.Runner, filter)
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
	return criContainerLogCmd(id, len, follow)
}

// SystemLogCmd returns the command to retrieve system logs
func (r *Containerd) SystemLogCmd(len int) string {
	return fmt.Sprintf("sudo journalctl -u containerd -n %d", len)
}
