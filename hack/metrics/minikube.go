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
	"runtime"

	"log"

	"cloud.google.com/go/storage"
	"github.com/pkg/errors"
)

const (
	bucketName = "priya-test-bucket/latest"
)

// download minikube latest to a tmp file
func downloadMinikube() (string, error) {
	b := binary()
	tmp, err := ioutil.TempFile("", b)
	if err != nil {
		return "", errors.Wrap(err, "creating tmp file")
	}
	if err := tmp.Close(); err != nil {
		return "", errors.Wrap(err, "closing tmp file")
	}
	client, err := storage.NewClient(context.Background())
	if err != nil {
		return "", errors.Wrap(err, "creating client")
	}
	ctx := context.Background()
	rc, err := client.Bucket(bucketName).Object(b).NewReader(ctx)
	if err != nil {
		return "", errors.Wrap(err, "gcs new reader")
	}
	defer rc.Close()

	data, err := ioutil.ReadAll(rc)
	if err != nil {
		return "", errors.Wrap(err, "ioutil read all")
	}
	log.Printf("downloading gs://%s/%s to %v", bucketName, b, tmp.Name())
	if err := ioutil.WriteFile(tmp.Name(), data, 0777); err != nil {
		return "", errors.Wrap(err, "writing file")
	}
	if err := os.Chmod(tmp.Name(), 0700); err != nil {
		return "", errors.Wrap(err, "chmod")
	}
	return tmp.Name(), nil
}

func binary() string {
	b := fmt.Sprintf("minikube-%s-amd64", runtime.GOOS)
	if runtime.GOOS == "windows" {
		b += ".exe"
	}
	return b
}
