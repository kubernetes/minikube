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

	"github.com/pkg/errors"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/command"
)

// runcContainer maps to 'runc list -f json'
type runcContainer struct {
	ID     string
	Status string
}

// criOutput maps to the output of `crictl ps -a --output=json`
type criOutput struct {
	Containers []criContainer `json:"containers"`
}

// criMetadata maps to the metadata object of containers in crictl output
type criMetadata struct {
	Name string `json:"name"`
}

// criContainer maps to containers in criOutput
type criContainer struct {
	ID       string      `json:"id"`
	Metadata criMetadata `json:"metadata"`
	CriState string      `json:"state"`
}

// State maps the cri-o states into State enum
// https://github.com/kubernetes/cri-api/blob/104a5b05531db3b886b68e0b91b6fdc3fe3c3738/pkg/apis/runtime/v1alpha2/api.proto#L889-L894
func (cc *criContainer) State() ContainerState {
	switch cc.CriState {
	case "CONTAINER_CREATED":
		return Created
	case "CONTAINER_RUNNING":
		return Running
	case "CONTAINER_EXITED":
		return Exited
	default:
		return Unknown
	}
}

// crictlList returns the output of 'crictl ps' in an efficient manner
func crictlList(cr CommandRunner, root string, o ListContainersOptions) (Containers, error) {
	klog.Infof("listing CRI containers in root %s: %+v", root, o)

	// Use -a because otherwise paused containers are missed
	baseCmd := []string{"crictl", "ps", "-a", "--output=json"}

	if o.Name != "" {
		baseCmd = append(baseCmd, fmt.Sprintf("--name=%s", o.Name))
	}

	var (
		rr  *command.RunResult
		err error
	)
	if len(o.Namespaces) == 0 {
		rr, err = cr.RunCmd(exec.Command("sudo", baseCmd...))
	} else {
		// Gather containers for all namespaces into one command without causing extraneous shells to be launched
		cmds := []string{}
		for _, ns := range o.Namespaces {
			cmd := fmt.Sprintf("%s --label io.kubernetes.pod.namespace=%s", strings.Join(baseCmd, " "), ns)
			cmds = append(cmds, cmd)
		}

		rr, err = cr.RunCmd(exec.Command("sudo", "-s", "eval", strings.Join(cmds, "; ")))
	}
	if err != nil {
		return nil, errors.Wrap(err, "crictl list")
	}

	var containers Containers
	// Since we may run multiple crictl within the same shell, multiple JSON objects need to be parsed from the buffer
	dec := json.NewDecoder(&rr.Stdout)
	for dec.More() {
		criOut := criOutput{}
		err := dec.Decode(&criOut)
		if err != nil {
			return nil, errors.Wrap(err, "crictl list JSON decoding")
		}
		for _, cc := range criOut.Containers {
			// avoid an ID ""
			if cc.ID == "" {
				continue
			}
			c := Container{
				ID:    cc.ID,
				Name:  cc.Metadata.Name,
				State: cc.State(),
			}
			containers = append(containers, c)
		}
	}
	return containers, nil
}

// listCRIPausedContainerIDs lists the paused container IDs with runc
// crictl don't understand paused containers so we need runc to tell use what containers are paused
func listCRIPausedContainerIDs(cr CommandRunner, root string) ([]string, error) {
	runcs := make([]runcContainer, 0)
	args := []string{"runc"}
	if root != "" {
		args = append(args, "--root", root)
	}

	args = append(args, "list", "-f", "json")
	rr, err := cr.RunCmd(exec.Command("sudo", args...))
	if err != nil {
		return nil, errors.Wrap(err, "runc")
	}
	content := rr.Stdout.Bytes()
	klog.Infof("JSON = %s", content)
	d := json.NewDecoder(bytes.NewReader(content))
	if err := d.Decode(&runcs); err != nil {
		return nil, err
	}

	if len(runcs) == 0 {
		return nil, nil
	}

	klog.Infof("runc list returned %d containers", len(runcs))

	pausedIDs := make([]string, 0)
	for _, rc := range runcs {
		if rc.Status == "paused" {
			pausedIDs = append(pausedIDs, rc.ID)
		}
	}
	return pausedIDs, nil
}

// listCRIContainers returns a list of containers
func listCRIContainers(cr CommandRunner, root string, o ListContainersOptions) (Containers, error) {
	containers, err := crictlList(cr, root, o)
	if err != nil {
		return nil, errors.Wrap(err, "crictl list")
	}

	if len(containers) == 0 {
		return nil, nil
	}

	// mark the paused containers for later filtering
	pausedIDs, err := listCRIPausedContainerIDs(cr, root)
	if err != nil {
		return nil, errors.Wrap(err, "runc list")
	}
	pausedIDMap := make(map[string]bool)
	for _, id := range pausedIDs {
		pausedIDMap[id] = true
	}
	for _, ctr := range containers {
		if _, ok := pausedIDMap[ctr.ID]; ok {
			ctr.State = Paused
		}
	}

	if o.State == All {
		return containers, nil
	}

	filtered := make([]Container, 0)
	for _, ctr := range containers {
		if o.State != ctr.State {
			klog.Infof("skipping %s: state = %q, want %q", ctr.ID, ctr.State, o.State)
			continue
		}
		filtered = append(filtered, ctr)
	}
	return filtered, nil
}

// pauseContainers pauses a list of containers
func pauseCRIContainers(cr CommandRunner, root string, ids []string) error {
	args := []string{"runc"}
	if root != "" {
		args = append(args, "--root", root)
	}
	args = append(args, "pause")
	for _, id := range ids {
		args := append(args, id)
		if _, err := cr.RunCmd(exec.Command("sudo", args...)); err != nil {
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
	cargs := args
	for _, id := range ids {
		cargs := append(cargs, id)
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
	klog.Infof("Killing containers: %s", ids)

	crictl := getCrictlPath(cr)
	args := append([]string{crictl, "rm"}, ids...)
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
func removeCRIImage(cr CommandRunner, name string) error {
	klog.Infof("Removing image: %s", name)

	crictl := getCrictlPath(cr)
	args := append([]string{crictl, "rmi"}, name)
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
	klog.Infof("Stopping containers: %s", ids)

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

// criContainerLogCmd returns the command to retrieve the log for a runcContainer based on ID
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
