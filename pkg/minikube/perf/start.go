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

package perf

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"runtime"

	"github.com/pkg/errors"
)

const (
	// runs is the number of times each binary will be timed for 'minikube start'
	runs = 5
	// threshold is the time difference in seconds we start alerting on
	threshold = 5.0
)

// CompareMinikubeStart compares the time to run `minikube start` between two minikube binaries
func CompareMinikubeStart(ctx context.Context, out io.Writer, binaries []*Binary) error {
	drivers := []string{"kvm2", "docker"}
	if runtime.GOOS == "darwin" {
		drivers = []string{"hyperkit", "docker"}
	}
	runtimes := []string{"docker", "containerd"}
	for _, d := range drivers {
		for _, r := range runtimes {
			if !proceed(d, r) {
				continue
			}
			fmt.Printf("**%s driver with %s runtime**\n", d, r)
			if err := downloadArtifacts(ctx, binaries, d, r); err != nil {
				fmt.Printf("error downloading artifacts: %v", err)
				continue
			}
			rm, err := collectResults(ctx, binaries, d, r)
			if err != nil {
				fmt.Printf("error collecting results for %s driver: %v\n", d, err)
				continue
			}
			rm.summarizeResults(binaries)
			fmt.Println()
		}
	}
	return nil
}

func collectResults(ctx context.Context, binaries []*Binary, driver string, runtime string) (*resultManager, error) {
	rm := newResultManager()
	for run := 0; run < runs; run++ {
		log.Printf("Executing run %d/%d...", run+1, runs)
		for _, binary := range binaries {
			r, err := timeMinikubeStart(ctx, binary, driver, runtime)
			if err != nil {
				return nil, errors.Wrapf(err, "timing run %d with %s", run, binary.Name())
			}
			rm.addResult(binary, "start", *r)
			if !skipIngress(driver, runtime) {
				r, err = timeEnableIngress(ctx, binary)
				if err != nil {
					return nil, errors.Wrapf(err, "timing run %d with %s", run, binary.Name())
				}
				rm.addResult(binary, "ingress", *r)
			}
			deleteCmd := exec.CommandContext(ctx, binary.path, "delete")
			if err := deleteCmd.Run(); err != nil {
				log.Printf("error deleting minikube: %v", err)
			}
		}
	}
	return rm, nil
}

func average(nums []float64) float64 {
	total := float64(0)
	for _, a := range nums {
		total += a
	}
	return total / float64(len(nums))
}

func downloadArtifacts(ctx context.Context, binaries []*Binary, driver string, runtime string) error {
	for _, b := range binaries {
		c := exec.CommandContext(ctx, b.path, "start", fmt.Sprintf("--driver=%s", driver), "--download-only", fmt.Sprintf("--container-runtime=%s", runtime))
		c.Stderr = os.Stderr
		log.Printf("Running: %v...", c.Args)
		if err := c.Run(); err != nil {
			return errors.Wrap(err, "downloading artifacts")
		}
	}
	return nil
}

// timeMinikubeStart returns the time it takes to execute `minikube start`
func timeMinikubeStart(ctx context.Context, binary *Binary, driver string, runtime string) (*result, error) {
	startCmd := exec.CommandContext(ctx, binary.path, "start", fmt.Sprintf("--driver=%s", driver), fmt.Sprintf("--container-runtime=%s", runtime))
	startCmd.Stderr = os.Stderr

	r, err := timeCommandLogs(startCmd)
	if err != nil {
		return nil, errors.Wrapf(err, "timing cmd: %v", startCmd.Args)
	}
	return r, nil
}

// timeEnableIngress returns the time it takes to execute `minikube addons enable ingress`
// It deletes the VM after `minikube addons enable ingress`.
func timeEnableIngress(ctx context.Context, binary *Binary) (*result, error) {
	enableCmd := exec.CommandContext(ctx, binary.path, "addons", "enable", "ingress")
	enableCmd.Stderr = os.Stderr

	r, err := timeCommandLogs(enableCmd)
	if err != nil {
		return nil, errors.Wrapf(err, "timing cmd: %v", enableCmd.Args)
	}
	return r, nil
}

// Ingress doesn't currently work on MacOS with the docker driver
func skipIngress(driver string, cruntime string) bool {
	return (runtime.GOOS == "darwin" && driver == "docker") || cruntime == "containerd"
}

// We only want to run the tests if:
// 1. It's a VM driver and docker container runtime
// 2. It's docker driver with any container runtime
func proceed(driver string, runtime string) bool {
	return runtime == "docker" || driver == "docker"
}
