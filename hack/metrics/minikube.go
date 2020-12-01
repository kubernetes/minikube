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

package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"log"

	"cloud.google.com/go/storage"
	"github.com/pkg/errors"
)

const (
	bucketName = "minikube/latest"
)

// download minikube latest to file
func downloadMinikube(ctx context.Context, minikubePath string) error {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return errors.Wrap(err, "creating client")
	}

	if localMinikubeIsLatest(ctx, minikubePath, client) {
		log.Print("local minikube is latest, skipping download...")
		return nil
	}

	os.Remove(minikubePath)
	// download minikube binary from GCS
	obj := client.Bucket("minikube").Object(fmt.Sprintf("latest/%s", binary()))
	rc, err := obj.NewReader(ctx)
	if err != nil {
		return errors.Wrap(err, "gcs new reader")
	}
	defer rc.Close()

	data, err := ioutil.ReadAll(rc)
	if err != nil {
		return errors.Wrap(err, "ioutil read all")
	}
	log.Printf("downloading gs://%s/%s to %v", bucketName, binary(), minikubePath)
	if err := ioutil.WriteFile(minikubePath, data, 0777); err != nil {
		return errors.Wrap(err, "writing minikubePath")
	}
	if err := os.Chmod(minikubePath, 0700); err != nil {
		return errors.Wrap(err, "chmod")
	}
	return nil
}

// localMinikubeIsLatest returns true if the local version of minikube
// matches the latest version in GCS
func localMinikubeIsLatest(ctx context.Context, minikubePath string, client *storage.Client) bool {
	log.Print("checking if local minikube is latest...")
	obj := client.Bucket("minikube").Object(fmt.Sprintf("latest/%s", binary()))
	attrs, err := obj.Attrs(ctx)
	if err != nil {
		log.Printf("error getting %s object attrs: %v", obj.ObjectName(), err)
		return false
	}
	gcsMinikubeVersion, ok := attrs.Metadata["commit"]
	if !ok {
		log.Printf("there is no commit: %v", attrs.Metadata)
		return false
	}
	currentMinikubeVersion, err := exec.Command(minikubePath, "version", "--output=json").Output()
	if err != nil {
		log.Printf("error running [%s version]: %v", minikubePath, err)
		return false
	}
	return strings.Contains(string(currentMinikubeVersion), gcsMinikubeVersion)
}

func binary() string {
	b := fmt.Sprintf("minikube-%s-amd64", runtime.GOOS)
	if runtime.GOOS == "windows" {
		b += ".exe"
	}
	return b
}
