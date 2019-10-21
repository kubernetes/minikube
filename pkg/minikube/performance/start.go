/*
Copyright 2017 The Kubernetes Authors All rights reserved.

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

package performance

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/pkg/errors"
)

var (
	runs = 1
	// For testing
	collectTimeMinikubeStart = timeMinikubeStart
)

// CompareMinikubeStart compares the time to run `minikube start` between two minikube binaries
func CompareMinikubeStart(ctx context.Context, out io.Writer, binaries []*Binary) error {
	durations, err := collectTimes(ctx, binaries)
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "Old binary: %v\nNew binary: %v\nAverage Old: %f\nAverage New: %f\n", durations[0], durations[1], average(durations[0]), average(durations[1]))
	return nil
}

func collectTimes(ctx context.Context, binaries []*Binary) ([][]float64, error) {
	durations := make([][]float64, len(binaries))
	for i := range durations {
		durations[i] = make([]float64, runs)
	}

	for r := 0; r < runs; r++ {
		log.Printf("Executing run %d...", r)
		for index, binary := range binaries {
			duration, err := collectTimeMinikubeStart(ctx, binary)
			if err != nil {
				return nil, errors.Wrapf(err, "timing run %d with %s", r, binary.path)
			}
			durations[index][r] = duration
		}
	}

	return durations, nil
}

func average(array []float64) float64 {
	total := float64(0)
	for _, a := range array {
		total += a
	}
	return total / float64(len(array))
}

// timeMinikubeStart returns the time it takes to execute `minikube start`
// It deletes the VM after `minikube start`.
func timeMinikubeStart(ctx context.Context, binary *Binary) (float64, error) {
	startCmd := exec.CommandContext(ctx, binary.path, "start")
	startCmd.Stdout = os.Stdout
	startCmd.Stderr = os.Stderr

	deleteCmd := exec.CommandContext(ctx, binary.path, "delete")
	defer func() {
		if err := deleteCmd.Run(); err != nil {
			log.Printf("error deleting minikube: %v", err)
		}
	}()

	log.Printf("Running: %v...", startCmd.Args)
	start := time.Now()
	if err := startCmd.Run(); err != nil {
		return 0, errors.Wrap(err, "starting minikube")
	}

	startDuration := time.Since(start).Seconds()
	return startDuration, nil
}
