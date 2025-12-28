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
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/blang/semver/v4"

	"github.com/pkg/errors"

	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/util/retry"
)

// container maps to 'runc list -f json'
type container struct {
	ID          string            `json:"id"`
	Status      string            `json:"status"`
	Annotations map[string]string `json:"annotations"`
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

// timeoutOverride flag overrides the default 2s timeout for crictl commands
const timeoutOverrideFlag = "--timeout=10s"

// runcBinaryName is the default binary name for runc
const runcBinaryName = "runc"

// listCRIContainers returns a list of containers
func listCRIContainers(cr CommandRunner, runtime string, root string, o ListContainersOptions) ([]string, error) {
	if runtime == "" {
		runtime = runcBinaryName
	}
	args := []string{runtime}
	if root != "" {
		args = append(args, "--root", root)
	}

	args = append(args, "list", "-f", "json")
	rr, err := cr.RunCmd(exec.Command("sudo", args...))
	if err != nil {
		return nil, errors.Wrap(err, runtime)
	}

	var cs []container
	if err := json.Unmarshal(rr.Stdout.Bytes(), &cs); err != nil {
		return nil, errors.Wrapf(err, "unmarshal %s list", runtime)
	}

	klog.Infof("list returned %d containers", len(cs))
	var fids []string
	for _, c := range cs {
		// Filter by State
		if o.State != All && o.State.String() != c.Status {
			continue
		}

		// Filter by Name
		// crictl matches partial name? minikube usage implies we look for specific components.
		// We check io.kubernetes.container.name and io.kubernetes.pod.name
		if o.Name != "" {
			name := c.Annotations["io.kubernetes.container.name"]
			podName := c.Annotations["io.kubernetes.pod.name"]
			if !strings.Contains(name, o.Name) && !strings.Contains(podName, o.Name) {
				continue
			}
		}

		// Filter by Namespaces
		if len(o.Namespaces) > 0 {
			ns := c.Annotations["io.kubernetes.pod.namespace"]
			found := false
			for _, want := range o.Namespaces {
				if ns == want {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		fids = append(fids, c.ID)
	}
	return fids, nil
}

// pauseCRIContainers pauses a list of containers
func pauseCRIContainers(cr CommandRunner, runtime string, root string, ids []string) error {
	if runtime == "" {
		runtime = runcBinaryName
	}
	baseArgs := []string{runtime}
	if root != "" {
		baseArgs = append(baseArgs, "--root", root)
	}
	baseArgs = append(baseArgs, "pause")
	for _, id := range ids {
		args := baseArgs
		args = append(args, id)
		if _, err := cr.RunCmd(exec.Command("sudo", args...)); err != nil {
			return errors.Wrap(err, runtime)
		}
	}
	return nil
}

// getCrictlPath returns the absolute path of crictl
func getCrictlPath(cr CommandRunner) string {
	cmd := "crictl"
	rr, err := cr.RunCmd(exec.Command("which", cmd))
	if err != nil {
		return cmd
	}
	return strings.Split(rr.Stdout.String(), "\n")[0]
}

// unpauseCRIContainers pauses a list of containers
func unpauseCRIContainers(cr CommandRunner, runtime string, root string, ids []string) error {
	if runtime == "" {
		runtime = runcBinaryName
	}
	args := []string{runtime}
	if root != "" {
		args = append(args, "--root", root)
	}
	args = append(args, "resume")
	cargs := args
	for _, id := range ids {
		cargs := append(cargs, id)
		if _, err := cr.RunCmd(exec.Command("sudo", cargs...)); err != nil {
			return errors.Wrap(err, runtime)
		}
	}
	return nil
}

// killCRIContainers kills a list of containers using crictl
func killCRIContainers(cr CommandRunner, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	klog.Infof("Killing containers: %s", ids)

	crictl := getCrictlPath(cr)
	args := append([]string{crictl, "rm", "--force"}, ids...)
	c := exec.Command("sudo", args...)
	if _, err := cr.RunCmd(c); err != nil {
		return errors.Wrap(err, "crictl")
	}
	return nil
}

// pullCRIImage pulls image using crictl
func pullCRIImage(cr CommandRunner, name string) error {
	klog.Infof("Pulling image: %s", name)

	crictl := getCrictlPath(cr)
	args := append([]string{crictl, "pull"}, name)
	c := exec.Command("sudo", args...)
	if _, err := cr.RunCmd(c); err != nil {
		return errors.Wrap(err, "crictl")
	}
	return nil
}

// removeCRIImage remove image using crictl
// verifyRemoval only used for CRIO due to #22242
func removeCRIImage(cr CommandRunner, name string, verifyRemoval bool) error {
	klog.Infof("Removing image: %s", name)

	crictl := getCrictlPath(cr)
	args := append([]string{crictl, "rmi"}, name)
	c := exec.Command("sudo", args...)
	var err error
	success := false
	if _, err = cr.RunCmd(c); err == nil {
		success = true
	} else if _, err := cr.RunCmd(exec.Command("sudo", crictl, "rmi", AddDockerIO(name))); err == nil {
		// see https://github.com/containers/podman/issues/15974
		success = true
	} else if _, err := cr.RunCmd(exec.Command("sudo", crictl, "rmi", AddLocalhostPrefix(name))); err == nil {
		success = true
	}

	if !success {
		return errors.Wrap(err, "crictl")
	}

	if !verifyRemoval {
		return nil
	}

	// Verify that the image is removed
	checkFunc := func() error {
		c := exec.Command("sudo", crictl, "images", "--quiet", name)
		rr, err := cr.RunCmd(c)
		if err != nil {
			return err
		}
		if len(strings.TrimSpace(rr.Stdout.String())) > 0 {
			return fmt.Errorf("image %s still exists", name)
		}
		return nil
	}

	if err := retry.Expo(checkFunc, 250*time.Millisecond, 10*time.Second); err != nil {
		return errors.Wrapf(err, "image %s still exists after removal", name)
	}
	return nil
}

// stopCRIContainers stops containers using crictl
func stopCRIContainers(cr CommandRunner, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	klog.Infof("Stopping containers: %s", ids)

	crictl := getCrictlPath(cr)
	// bring crictl stop timeout on par with docker:
	// - docker stop --help => -t, --time int   Seconds to wait for stop before killing it (default 10)
	// - crictl stop --help => --timeout value, -t value  Seconds to wait to kill the container after a graceful stop is requested (default: 0)
	// to prevent "stuck" containers blocking ports (eg, "[ERROR Port-2379|2380]: Port 2379|2380 is in use" for etcd during "hot" k8s upgrade)
	args := append([]string{crictl, "stop", "--timeout=10"}, ids...)
	c := exec.Command("sudo", args...)
	if _, err := cr.RunCmd(c); err != nil {
		return errors.Wrap(err, "crictl")
	}
	return nil
}

// populateCRIConfig sets up /etc/crictl.yaml
func populateCRIConfig(cr CommandRunner, socket string) error {
	cPath := "/etc/crictl.yaml"
	tmpl := "runtime-endpoint: unix://{{.Socket}}\n"
	t, err := template.New("crictl").Parse(tmpl)
	if err != nil {
		return err
	}
	opts := struct{ Socket string }{Socket: socket}
	var b bytes.Buffer
	if err := t.Execute(&b, opts); err != nil {
		return err
	}
	c := exec.Command("/bin/bash", "-c", fmt.Sprintf("sudo mkdir -p %s && printf %%s \"%s\" | sudo tee %s", path.Dir(cPath), b.String(), cPath))
	if rr, err := cr.RunCmd(c); err != nil {
		return errors.Wrapf(err, "Run: %q", rr.Command())
	}
	return nil
}

// getCRIInfo returns current information
func getCRIInfo(cr CommandRunner) (map[string]interface{}, error) {
	args := []string{"crictl", timeoutOverrideFlag, "info"}
	c := exec.Command("sudo", args...)
	rr, err := cr.RunCmd(c)
	if err != nil {
		return nil, errors.Wrap(err, "get cri info")
	}
	info := rr.Stdout.String()
	jsonMap := make(map[string]interface{})
	err = json.Unmarshal([]byte(info), &jsonMap)
	if err != nil {
		return nil, err
	}
	return jsonMap, nil
}

// listCRIImages lists images using crictl
func listCRIImages(cr CommandRunner) ([]ListImage, error) {
	c := exec.Command("sudo", "crictl", timeoutOverrideFlag, "images", "--output", "json")
	rr, err := cr.RunCmd(c)
	if err != nil {
		return nil, errors.Wrapf(err, "crictl images")
	}

	var jsonImages crictlImages
	err = json.Unmarshal(rr.Stdout.Bytes(), &jsonImages)
	if err != nil {
		klog.Errorf("failed to unmarshal images, will assume images are not preloaded")
		return nil, err
	}

	images := []ListImage{}
	for _, img := range jsonImages.Images {
		images = append(images, ListImage{
			ID:          img.ID,
			RepoDigests: img.RepoDigests,
			RepoTags:    img.RepoTags,
			Size:        img.Size,
		})
	}
	return images, nil
}

// criContainerLogCmd returns the command to retrieve the log for a container based on ID
func criContainerLogCmd(cr CommandRunner, id string, length int, follow bool) string {
	crictl := getCrictlPath(cr)
	var cmd strings.Builder
	cmd.WriteString("sudo ")
	cmd.WriteString(crictl)
	cmd.WriteString(" logs ")
	if length > 0 {
		cmd.WriteString(fmt.Sprintf("--tail %d ", length))
	}
	if follow {
		cmd.WriteString("--follow ")
	}

	cmd.WriteString(id)
	return cmd.String()
}

// addRepoTagToImageName makes sure the image name has a repo tag in it.
// in crictl images list have the repo tag prepended to them
// for example "kubernetesui/dashboard:v2.0.0 will show up as "docker.io/kubernetesui/dashboard:v2.0.0"
// warning this is only meant for kubernetes images where we know the GCR addresses have .io in them
// not mean to be used for public images
func addRepoTagToImageName(imgName string) string {
	if !strings.Contains(imgName, ".io/") {
		return "docker.io/" + imgName
	} // else it already has repo name dont add anything
	return imgName
}

// kubeletCRIOptions returns the container runtime options for the kubelet
func kubeletCRIOptions(cr Manager, kubernetesVersion semver.Version) map[string]string {
	opts := map[string]string{
		"container-runtime-endpoint": fmt.Sprintf("unix://%s", cr.SocketPath()),
	}
	if kubernetesVersion.LT(semver.MustParse("1.24.0-alpha.0")) {
		opts["container-runtime"] = "remote"
	}
	return opts
}

func checkCNIPlugins(kubernetesVersion semver.Version) error {
	if kubernetesVersion.LT(semver.Version{Major: 1, Minor: 24}) {
		return nil
	}
	_, err := os.Stat("/opt/cni/bin")
	return err
}

// Add localhost prefix if the registry part is missing
func AddLocalhostPrefix(name string) string {
	return addRegistryPreix(name, "localhost")
}
