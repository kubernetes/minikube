/*
Copyright 2023 The Kubernetes Authors All rights reserved.

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

package cruntimeInstaller

import (
	"bytes"
	"fmt"
	"html/template"
	"os/exec"
	"path"

	"k8s.io/klog"
	"k8s.io/minikube/pkg/libmachine/libmachine/auth"
	"k8s.io/minikube/pkg/libmachine/libmachine/engine"
	"k8s.io/minikube/pkg/libmachine/libmachine/runner"
)

type crioInstaller struct {
	Options              *engine.Options
	ContainerRuntimeName string
	Commander            runner.Runner
	Provider             string
	AuthOptions          *auth.Options
}

// x7TODO: complete this
func (ci *crioInstaller) InstallCRuntime() error {
	if err := ci.SetCRuntimeOptions(); err != nil {
		klog.Infof("Error setting container-runtime (%s) options during provisioning %v",
			ci.ContainerRuntimeName, err)
	}

	return nil
}

func NewCRIOInstaller(opts *engine.Options, cmd runner.Runner, provider string, authOpts *auth.Options) *crioInstaller {
	return &crioInstaller{
		Options:              opts,
		ContainerRuntimeName: "CRI-O",
		Commander:            cmd,
		Provider:             provider,
		AuthOptions:          authOpts,
	}
}

func (ci *crioInstaller) SetCRuntimeOptions() error {
	// pass through --insecure-registry
	var (
		crioOptsTmpl = `
CRIO_MINIKUBE_OPTIONS='{{ range .EngineOptions.InsecureRegistry }}--insecure-registry {{.}} {{ end }}'
`
		crioOptsPath = "/etc/sysconfig/crio.minikube"
	)
	t, err := template.New("crioOpts").Parse(crioOptsTmpl)
	if err != nil {
		return err
	}
	var crioOptsBuf bytes.Buffer
	if err := t.Execute(&crioOptsBuf, ci.Commander); err != nil {
		return err
	}

	if _, err = ci.Commander.RunCmd(exec.Command("bash", "-c", fmt.Sprintf(`sudo mkdir -p %s && printf %%s \"%s\" | sudo tee %s && sudo systemctl restart crio`, path.Dir(crioOptsPath), crioOptsBuf.String(), crioOptsPath))); err != nil {
		return err
	}

	return nil
}
