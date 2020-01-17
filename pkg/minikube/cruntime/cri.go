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

// getCrictlPath returns the absolute path of crictl
func getCrictlPath(cr CommandRunner) string {
	cmd := "crictl"
	rr, err := cr.RunCmd(exec.Command("which", cmd))
	if err != nil {
		return cmd
	}
	return strings.Split(rr.Stdout.String(), "\n")[0]
}

// listCRIContainers returns a list of containers using crictl
func listCRIContainers(cr CommandRunner, filter string) ([]string, error) {
	var err error
	var rr *command.RunResult
	state := "Running"
	crictl := getCrictlPath(cr)
	if filter != "" {
		c := exec.Command("sudo", crictl, "ps", "-a", fmt.Sprintf("--name=%s", filter), fmt.Sprintf("--state=%s", state), "--quiet")
		rr, err = cr.RunCmd(c)
	} else {
		rr, err = cr.RunCmd(exec.Command("sudo", crictl, "ps", "-a", fmt.Sprintf("--state=%s", state), "--quiet"))
	}
	if err != nil {
		return nil, err
	}
	var ids []string
	for _, line := range strings.Split(rr.Stderr.String(), "\n") {
		if line != "" {
			ids = append(ids, line)
		}
	}
	return ids, nil
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
		return errors.Wrap(err, "kill cri containers.")
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
	args := append([]string{crictl, "rm"}, ids...)
	c := exec.Command("sudo", args...)
	if _, err := cr.RunCmd(c); err != nil {
		return errors.Wrap(err, "stop cri containers")
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
