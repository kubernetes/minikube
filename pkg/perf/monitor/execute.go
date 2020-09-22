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

package monitor

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pkg/errors"
)

// RunMkcmp runs minikube built at the given pr against minikube at master
func RunMkcmp(ctx context.Context, pr int) (string, error) {
	// run 'git pull' so that minikube dir is up to date
	if _, err := runCmdInMinikube(ctx, []string{"git", "pull", "origin", "master"}); err != nil {
		return "", errors.Wrap(err, "running git pull")
	}
	mkcmpPath := "out/mkcmp"
	minikubePath := "out/minikube"
	if _, err := runCmdInMinikube(ctx, []string{"make", mkcmpPath, minikubePath}); err != nil {
		return "", errors.Wrap(err, "building minikube and mkcmp at head")
	}
	return runCmdInMinikube(ctx, []string{mkcmpPath, minikubePath, fmt.Sprintf("pr://%d", pr)})
}

// runCmdInMinikube runs the cmd and return stdout
func runCmdInMinikube(ctx context.Context, command []string) (string, error) {
	cmd := exec.CommandContext(ctx, command[0], command[1:]...)
	cmd.Dir = minikubeDir()
	cmd.Env = os.Environ()

	buf := bytes.NewBuffer([]byte{})
	cmd.Stdout = buf

	log.Printf("Running: %v", cmd.Args)
	if err := cmd.Run(); err != nil {
		return "", errors.Wrapf(err, "running %v: %v", cmd.Args, buf.String())
	}
	return buf.String(), nil
}

func minikubeDir() string {
	return filepath.Join(os.Getenv("HOME"), "minikube")
}
