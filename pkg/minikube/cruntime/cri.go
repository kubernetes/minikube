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
	"fmt"
	"html/template"
	"path"
	"strings"

	"github.com/golang/glog"
)

// listCRIContainers returns a list of containers using crictl
func listCRIContainers(cr CommandRunner, filter string) ([]string, error) {
	var content string
	var err error
	state := "Running"
	if filter != "" {
		content, err = cr.CombinedOutput(fmt.Sprintf(`sudo crictl ps -a --name=%s --state=%s --quiet`, filter, state))
	} else {
		content, err = cr.CombinedOutput(fmt.Sprintf(`sudo crictl ps -a --state=%s --quiet`, state))
	}
	if err != nil {
		return nil, err
	}
	var ids []string
	for _, line := range strings.Split(content, "\n") {
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
	return cr.Run(fmt.Sprintf("sudo crictl rm %s", strings.Join(ids, " ")))
}

// stopCRIContainers stops containers using crictl
func stopCRIContainers(cr CommandRunner, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	glog.Infof("Stopping containers: %s", ids)
	return cr.Run(fmt.Sprintf("sudo crictl stop %s", strings.Join(ids, " ")))
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
	return cr.Run(fmt.Sprintf("sudo mkdir -p %s && printf %%s \"%s\" | sudo tee %s", path.Dir(cPath), b.String(), cPath))
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
