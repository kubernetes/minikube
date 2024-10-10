/*
Copyright 2020 The Kubernetes Authors All rights reserved.

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

package cni

import (
	"bytes"
	"fmt"
	"os/exec"
	"path"

	// goembed needs this
	_ "embed"
	"text/template"

	"github.com/blang/semver/v4"
	"github.com/pkg/errors"

	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/kapi"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/bootstrapper/images"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/vmpath"
	"k8s.io/minikube/pkg/util"
)

// https://docs.projectcalico.org/manifests/calico.yaml
//
//go:embed calico.yaml
var calicoYaml string

//go:embed calico-apiserver.yaml
var calicoAPIServerYaml string

// calicoTmpl is from https://docs.projectcalico.org/manifests/calico.yaml
var calicoTmpl = template.Must(template.New("calico").Parse(calicoYaml))
var calicoAPIServerTmpl = template.Must(template.New("calico-apiserver").Parse(calicoAPIServerYaml))

// Calico is the Calico CNI manager
type Calico struct {
	cc config.ClusterConfig
}

type calicoTmplStruct struct {
	PodCIDR                   string
	DeploymentImageName       string
	DaemonSetImageName        string
	BinaryImageName           string
	LegacyPodDisruptionBudget bool
}

// String returns a string representation of this CNI
func (c Calico) String() string {
	return "Calico"
}

// manifest returns a Kubernetes manifest for a CNI
func (c Calico) manifest() (assets.CopyableFile, error) {
	k8sVersion, err := util.ParseKubernetesVersion(c.cc.KubernetesConfig.KubernetesVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Kubernetes version: %v", err)
	}

	input := &calicoTmplStruct{
		PodCIDR:                   DefaultPodCIDR,
		DeploymentImageName:       images.CalicoDeployment(c.cc.KubernetesConfig.ImageRepository),
		DaemonSetImageName:        images.CalicoDaemonSet(c.cc.KubernetesConfig.ImageRepository),
		BinaryImageName:           images.CalicoBin(c.cc.KubernetesConfig.ImageRepository),
		LegacyPodDisruptionBudget: k8sVersion.LT(semver.Version{Major: 1, Minor: 25}),
	}

	b := bytes.Buffer{}
	if err := calicoTmpl.Execute(&b, input); err != nil {
		return nil, err
	}
	return manifestAsset(b.Bytes()), nil
}

func (c Calico) apiServerManifest() (assets.CopyableFile, error) {
	b := bytes.Buffer{}
	if err := calicoAPIServerTmpl.Execute(&b, map[string]string{
		"APIServerImageName": images.CalicoAPIServer(c.cc.KubernetesConfig.ImageRepository),
	}); err != nil {
		return nil, err
	}
	return assets.NewMemoryAssetTarget(b.Bytes(), apiServerMainifestPath(), "0644"), nil
}

// Apply enables the CNI
func (c Calico) Apply(r Runner) error {
	manifest, err := c.manifest()
	if err != nil {
		return errors.Wrap(err, "manifest")
	}
	if err := applyManifest(c.cc, r, manifest); err != nil {
		return err
	}
	// then we apply the mainifest
	kubectl := kapi.KubectlBinaryPath(c.cc.KubernetesConfig.KubernetesVersion)

	apiServerManifest, err := c.apiServerManifest()
	if err != nil {
		return errors.Wrap(err, "apiserver manifest")
	}
	if err := r.Copy(apiServerManifest); err != nil {
		return errors.Wrapf(err, "copy")
	}
	cmd := exec.Command("sudo", kubectl, "apply", fmt.Sprintf("--kubeconfig=%s", path.Join(vmpath.GuestPersistentDir, "kubeconfig")), "-f", apiServerMainifestPath())
	if rr, err := r.RunCmd(cmd); err != nil {
		return errors.Wrapf(err, "cmd: %s output: %s", rr.Command(), rr.Output())
	}

	// according to https://docs.tigera.io/calico/3.26/operations/install-apiserver (Manifest Install)
	// there is something more we need to do
	// from now on, we won't return any error if any failure happens
	// because executing the previous manifest.yaml is sufficient for calico cni plugin to work
	// apiserver.yaml is just for managing the calico, it affect cni plugin's duty if the api server is down
	cmd = exec.Command("openssl", "req", "-x509", "-nodes", "-newkey", "rsa:4096", "-keyout", "apiserver.key", "-out", "apiserver.crt", "-days", "3650", "-subj", "/", "-addext", "subjectAltName = DNS:calico-api.calico-apiserver.svc")
	if out, err := r.RunCmd(cmd); err != nil {
		klog.Error("failed to generate certificate for calico api server, output:", out.Output(), "error: ", err)
		return nil
	}

	cmd = exec.Command("sudo", kubectl, "create", "secret", fmt.Sprintf("--kubeconfig=%s", path.Join(vmpath.GuestPersistentDir, "kubeconfig")), "-n", "calico-apiserver", "generic", "calico-apiserver-certs", "--from-file=apiserver.key", "--from-file=apiserver.crt")
	if out, err := r.RunCmd(cmd); err != nil {
		klog.Error("failed to generate secret for calico api server, output:", out.Output(), "error: ", err)
		return nil
	}
	cmd = exec.Command("sudo", "/bin/bash", "-c", kubectl, "patch", "apiservice", "v3.projectcalico.org", fmt.Sprintf("--kubeconfig=%s", path.Join(vmpath.GuestPersistentDir, "kubeconfig")), "-p",
		`kubectl patch apiservice v3.projectcalico.org -p \
    "{\"spec\": {\"caBundle\": \"$(kubectl get secret -n calico-apiserver calico-apiserver-certs -o go-template='{{ index .data "apiserver.crt" }}')\"}}"`,
	)
	if out, err := r.RunCmd(cmd); err != nil {
		klog.Error("failed to patch calico api server for calico api server, output:", out.Output(), "error: ", err)
		return nil
	}
	return nil

}

// CIDR returns the default CIDR used by this CNI
func (c Calico) CIDR() string {
	// Calico docs specify 192.168.0.0/16 - but we do this for compatibility with other CNI's.
	return DefaultPodCIDR
}
func apiServerMainifestPath() string {
	return path.Join(vmpath.GuestEphemeralDir, "calico-apiserver.yaml")
}
