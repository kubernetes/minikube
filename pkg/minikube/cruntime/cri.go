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

	"github.com/golang/glog"
)

// listCRIContainers returns a list of containers using crictl
func listCRIContainers(_ CommandRunner, _ string) ([]string, error) {
	// Should use crictl ps -a, but needs some massaging and testing.
	return []string{}, fmt.Errorf("unimplemented")
}

// pullImageCRI uses ctr to pull images into a CRI runtime
func pullImageCRI(cr CommandRunner, path string) error {
	glog.Infof("Loading image: %s", path)
	return cr.Run(fmt.Sprintf("sudo ctr cri load %s", path))
}

// criCRIContainers kills a list of containers using crictl
func killCRIContainers(CommandRunner, []string) error {
	return fmt.Errorf("unimplemented")
}

// stopCRIContainers stops containers using crictl
func stopCRIContainers(CommandRunner, []string) error {
	return fmt.Errorf("unimplemented")
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
