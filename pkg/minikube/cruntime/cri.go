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
	"os/exec"
	"path"
	"strings"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/command"
)

// container maps to 'runc list -f json'
type container struct {
	ID     string
	Status string
}

// crictlList returns the output of 'crictl ps' in an efficient manner
func crictlList(cr CommandRunner, root string, o ListOptions) (*command.RunResult, error) {
	glog.Infof("listing CRI containers in root %s: %+v", root, o)

	// Use -a because otherwise paused containers are missed
	baseCmd := []string{"crictl", "ps", "-a", "--quiet"}

	if o.Name != "" {
		baseCmd = append(baseCmd, fmt.Sprintf("--name=%s", o.Name))
	}

	// shortcut for all namespaces
	if len(o.Namespaces) == 0 {
		return cr.RunCmd(exec.Command("sudo", baseCmd...))
	}

	// Gather containers for all namespaces without causing extraneous shells to be launched
	cmds := []string{}
	for _, ns := range o.Namespaces {
		cmd := fmt.Sprintf("%s --label io.kubernetes.pod.namespace=%s", strings.Join(baseCmd, " "), ns)
		cmds = append(cmds, cmd)
	}

	return cr.RunCmd(exec.Command("sudo", "-s", "eval", strings.Join(cmds, "; ")))
}

// listCRIContainers returns a list of containers
func listCRIContainers(cr CommandRunner, root string, o ListOptions) ([]string, error) {
	rr, err := crictlList(cr, root, o)
	if err != nil {
		return nil, errors.Wrap(err, "crictl list")
	}

	// Avoid an id named ""
	var ids []string
	seen := map[string]bool{}
	for _, id := range strings.Split(rr.Stdout.String(), "\n") {
		glog.Infof("found id: %q", id)
		if id != "" && !seen[id] {
			ids = append(ids, id)
			seen[id] = true
		}
	}

	if len(ids) == 0 {
		return nil, nil
	}
	if o.State == All {
		return ids, nil
	}

	// crictl does not understand paused pods
	cs := []container{}
	args := []string{"runc"}
	if root != "" {
		args = append(args, "--root", root)
	}

	args = append(args, "list", "-f", "json")
	rr, err = cr.RunCmd(exec.Command("sudo", args...))
	if err != nil {
		return nil, errors.Wrap(err, "runc")
	}
	content := rr.Stdout.Bytes()
	glog.Infof("JSON = %s", content)
	d := json.NewDecoder(bytes.NewReader(content))
	if err := d.Decode(&cs); err != nil {
		return nil, err
	}

	if len(cs) == 0 {
		return nil, fmt.Errorf("list returned 0 containers, but ps returned %d", len(ids))
	}

	glog.Infof("list returned %d containers", len(cs))
	var fids []string
	for _, c := range cs {
		glog.Infof("container: %+v", c)
		if !seen[c.ID] {
			glog.Infof("skipping %s - not in ps", c.ID)
			continue
		}
		if o.State != All && o.State.String() != c.Status {
			glog.Infof("skipping %s: state = %q, want %q", c, c.Status, o.State)
			continue
		}
		fids = append(fids, c.ID)
	}
	return fids, nil
}

// pauseContainers pauses a list of containers
func pauseCRIContainers(cr CommandRunner, root string, ids []string) error {
	args := []string{"runc"}
	if root != "" {
		args = append(args, "--root", root)
	}
	args = append(args, "pause")

	for _, id := range ids {
		cargs := append(args, id)
		if _, err := cr.RunCmd(exec.Command("sudo", cargs...)); err != nil {
			return errors.Wrap(err, "runc")
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
func unpauseCRIContainers(cr CommandRunner, root string, ids []string) error {
	args := []string{"runc"}
	if root != "" {
		args = append(args, "--root", root)
	}
	args = append(args, "resume")

	for _, id := range ids {
		cargs := append(args, id)
		if _, err := cr.RunCmd(exec.Command("sudo", cargs...)); err != nil {
			return errors.Wrap(err, "runc")
		}
	}
	return nil
}

// criCRIContainers kills a list of containers using crictl
func killCRIContainers(cr CommandRunner, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	glog.Infof("Killing containers: %s", ids)

	crictl := getCrictlPath(cr)
	args := append([]string{crictl, "rm"}, ids...)
	c := exec.Command("sudo", args...)
	if _, err := cr.RunCmd(c); err != nil {
		return errors.Wrap(err, "crictl")
	}
	return nil
}

// stopCRIContainers stops containers using crictl
func stopCRIContainers(cr CommandRunner, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	glog.Infof("Stopping containers: %s", ids)

	crictl := getCrictlPath(cr)
	args := append([]string{crictl, "stop"}, ids...)
	c := exec.Command("sudo", args...)
	if _, err := cr.RunCmd(c); err != nil {
		return errors.Wrap(err, "crictl")
	}
	return nil
}

// populateCRIConfig sets up /etc/crictl.yaml
func populateCRIConfig(cr CommandRunner, socket string) error {
	cPath := "/etc/crictl.yaml"
	tmpl := `runtime-endpoint: unix://{{.Socket}}
image-endpoint: unix://{{.Socket}}
`
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
	args := []string{"crictl", "info"}
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

// criContainerLogCmd returns the command to retrieve the log for a container based on ID
func criContainerLogCmd(cr CommandRunner, id string, len int, follow bool) string {
	crictl := getCrictlPath(cr)
	var cmd strings.Builder
	cmd.WriteString("sudo ")
	cmd.WriteString(crictl)
	cmd.WriteString(" logs ")
	if len > 0 {
		cmd.WriteString(fmt.Sprintf("--tail %d ", len))
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
// warning this is only meant for kuberentes images where we know the GCR addreses have .io in them
// not mean to be used for public images
func addRepoTagToImageName(imgName string) string {
	if !strings.Contains(imgName, ".io/") {
		return "docker.io/" + imgName
	} // else it already has repo name dont add anything
	return imgName
}
