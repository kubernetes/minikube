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

package perf

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/pkg/errors"
	"google.golang.org/api/option"
	"k8s.io/minikube/pkg/minikube/constants"
)

// Binary holds a minikube binary
type Binary struct {
	path string
	pr   int
}

const (
	prPrefix = "pr://"
	bucket   = "minikube-builds"
)

// NewBinary returns a new binary type
func NewBinary(b string) (*Binary, error) {
	// If it doesn't have the prefix, assume a path
	if !strings.HasPrefix(b, prPrefix) {
		return &Binary{
			path: b,
		}, nil
	}
	return newBinaryFromPR(b)
}

// Name returns the name of the binary
func (b *Binary) Name() string {
	if b.pr != 0 {
		return fmt.Sprintf("Minikube (PR %d)", b.pr)
	}
	return filepath.Base(b.path)
}

func (b *Binary) download() error {
	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithoutAuthentication())
	if err != nil {
		return errors.Wrap(err, "getting storage client")
	}
	defer client.Close()
	rc, err := client.Bucket(bucket).Object(fmt.Sprintf("%d/minikube-linux-amd64", b.pr)).NewReader(ctx)
	if err != nil {
		return errors.Wrap(err, "getting minikube object from gcs bucket")
	}
	defer rc.Close()

	if err := os.MkdirAll(filepath.Dir(b.path), 0777); err != nil {
		return err
	}

	f, err := os.OpenFile(b.path, os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, rc)
	return err
}

// newBinaryFromPR downloads the minikube binary built for the pr by Jenkins from GCS
func newBinaryFromPR(pr string) (*Binary, error) {
	pr = strings.TrimPrefix(pr, prPrefix)
	// try to convert to int
	i, err := strconv.Atoi(pr)
	if err != nil {
		return nil, errors.Wrapf(err, "converting %s to an integer", pr)
	}

	b := &Binary{
		path: localMinikubePath(i),
		pr:   i,
	}
	if err := b.download(); err != nil {
		return nil, errors.Wrapf(err, "downloading binary")
	}
	return b, nil
}

func localMinikubePath(pr int) string {
	return fmt.Sprintf("%s/minikube-binaries/%d/minikube", constants.DefaultMinipath, pr)
}
