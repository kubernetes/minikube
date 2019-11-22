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
)

// container maps to 'runc list -f json'
type container struct {
	pid         int
	id          string
	status      string
	annotations map[string]string
}

// listCRIContainers returns a list of containers
func listCRIContainers(cr CommandRunner, root string, o ListOptions) ([]string, error) {
	// First use crictl, because it reliably matches names
	args := []string{"crictl", "ps", "--quiet"}
	if o.State == All {
		args = append(args, "-a")
	}
	if o.Name != "" {
		args = append(args, fmt.Sprintf("--name=%s", o.Name))
	}
	rr, err := cr.RunCmd(exec.Command("sudo", args...))
	if err != nil {
		return nil, err
	}

	// Avoid an id named ""
	var ids []string
	seen := map[string]bool{}
	for _, id := range strings.Split(rr.Stderr.String(), "\n") {
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
	args = []string{"runc", "list", "-f", "json"}
	if root != "" {
		args = append(args, "--root", root)
	}

	if _, err := cr.RunCmd(exec.Command("sudo", args...)); err != nil {
		return nil, errors.Wrap(err, "runc")
	}

	d := json.NewDecoder(bytes.NewReader(rr.Stdout.Bytes()))
	if err := d.Decode(&cs); err != nil {
		return nil, err
	}

	var fids []string
	for _, c := range cs {
		if !seen[c.id] {
			continue
		}
		if o.State.String() != c.status {
			continue
		}
		fids = append(fids, c.id)
	}
	return fids, nil
}

// pauseCRIContainers pauses a list of containers
func pauseCRIContainers(cr CommandRunner, root string, ids []string) error {
	args := []string{"runc", "pause"}
	if root != "" {
		args = append(args, "--root", root)
	}

	for _, id := range ids {
		cargs := append(args, id)
		if _, err := cr.RunCmd(exec.Command("sudo", cargs...)); err != nil {
			return errors.Wrap(err, "runc")
		}
	}
	return nil
}

// unpauseCRIContainers pauses a list of containers
func unpauseCRIContainers(cr CommandRunner, root string, ids []string) error {
	args := []string{"runc", "unpause"}
	if root != "" {
		args = append(args, "--root", root)
	}

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

	args := append([]string{"crictl", "rm"}, ids...)
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
	args := append([]string{"crictl", "rm"}, ids...)
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

// criContainerLogCmd returns the command to retrieve the log for a container based on ID
func criContainerLogCmd(id string, len int, follow bool) string {
	var cmd strings.Builder
	cmd.WriteString("sudo crictl logs ")
	if len > 0 {
		cmd.WriteString(fmt.Sprintf("--tail %d ", len))
	}
	if follow {
		cmd.WriteString("--follow ")
	}

	cmd.WriteString(id)
	return cmd.String()
}
