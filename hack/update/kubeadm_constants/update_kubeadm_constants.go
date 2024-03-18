/*
Copyright 2021 The Kubernetes Authors All rights reserved.

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

package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"text/template"
	"time"

	"github.com/google/go-github/v60/github"
	"golang.org/x/mod/semver"
	"k8s.io/klog/v2"
	"k8s.io/minikube/hack/update"
	"k8s.io/minikube/pkg/minikube/constants"
)

const (
	// default context timeout
	cxTimeout                 = 5 * time.Minute
	kubeadmReleaseURL         = "https://dl.k8s.io/release/%s/bin/linux/%s/kubeadm"
	kubeadmBinaryName         = "kubeadm-linux-%s-%s"
	minikubeConstantsFilePath = "pkg/minikube/constants/constants_kubeadm_images.go"
	kubeadmImagesTemplate     = `
		{{- range $version, $element := .}}
		"{{$version}}": {
			{{- range $image, $tag := $element}}
			"{{$image}}": "{{$tag}}",
			{{- end}}
		},{{- end}}`
)

// Data contains kubeadm Images map
type Data struct {
	ImageMap string
}

func main() {
	minver := constants.OldestKubernetesVersion

	releases := []string{}

	ghc := github.NewClient(nil)

	opts := &github.ListOptions{PerPage: 100}
	for {
		rls, resp, err := ghc.Repositories.ListReleases(context.Background(), "kubernetes", "kubernetes", opts)
		if err != nil {
			klog.Fatal(err)
		}
		for _, rl := range rls {
			ver := rl.GetTagName()
			if !semver.IsValid(ver) {
				continue
			}
			// skip out-of-range versions
			if minver != "" && semver.Compare(minver, ver) == 1 {
				continue
			}
			releases = append([]string{ver}, releases...)
		}
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	for _, imageVersion := range releases {
		if _, ok := constants.KubeadmImages[imageVersion]; ok {
			continue
		}
		imageMapString, err := getKubeadmImagesMapString(imageVersion)
		if err != nil {
			klog.Fatalln(err)
		}

		var data Data
		schema := map[string]update.Item{
			minikubeConstantsFilePath: {
				Replace: map[string]string{},
			},
		}

		data = Data{ImageMap: imageMapString}
		schema[minikubeConstantsFilePath].Replace[`KubeadmImages = .*`] =
			`KubeadmImages = map[string]map[string]string{ {{.ImageMap}}`
		update.Apply(schema, data)
	}
}

func getKubeadmImagesMapString(version string) (string, error) {
	arch := runtime.GOARCH
	url := fmt.Sprintf(kubeadmReleaseURL, version, arch)
	fileName := fmt.Sprintf(kubeadmBinaryName, arch, version)
	if err := downloadFile(url, fileName); err != nil {
		klog.Errorf("failed to download kubeadm binary: %v", err)
		return "", err
	}

	kubeadmCommand := fmt.Sprintf("./%s", fileName)
	args := []string{"config", "images", "list"}
	imageListString, err := executeCommand(kubeadmCommand, args...)
	if err != nil {
		klog.Errorf("failed to execute kubeadm command %s", kubeadmCommand)
		return "", err
	}

	if err := os.Remove(fileName); err != nil {
		klog.Errorf("failed to remove binary %s", fileName)
	}

	return formatKubeadmImageList(version, imageListString)
}

func formatKubeadmImageList(version, data string) (string, error) {
	templateData := make(map[string]map[string]string)
	templateData[version] = make(map[string]string)
	lines := strings.Split(data, "\n")
	for _, line := range lines {
		imageTag := strings.Split(line, ":")
		if len(imageTag) != 2 {
			continue
		}
		// removing the repo from image name
		imageName := strings.Split(imageTag[0], "/")
		imageTag[0] = strings.Join(imageName[1:], "/")
		if !isKubeImage(imageTag[0]) {
			templateData[version][imageTag[0]] = imageTag[1]
		}
	}

	imageTemplate := template.New("kubeadmImage")
	t, err := imageTemplate.Parse(kubeadmImagesTemplate)
	if err != nil {
		klog.Errorf("failed to create kubeadm image map template: %v", err)
		return "", err
	}

	var bytesBuffer bytes.Buffer
	if err := t.Execute(&bytesBuffer, &templateData); err != nil {
		return "", err
	}

	return bytesBuffer.String(), nil
}

func isKubeImage(name string) bool {
	kubeImages := map[string]bool{
		"kube-apiserver":          true,
		"kube-controller-manager": true,
		"kube-proxy":              true,
		"kube-scheduler":          true,
	}
	return kubeImages[name]
}

func downloadFile(url, fileName string) error {
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("non success status code, while downloading file: %s from: %s", fileName, url)
	}

	if _, err := io.Copy(file, response.Body); err != nil {
		return err
	}

	return os.Chmod(fileName, os.ModePerm)
}

func executeCommand(command string, args ...string) (string, error) {
	output, err := exec.Command(command, args...).Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}
