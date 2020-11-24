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
	"log"
	"os"
	"os/exec"
	"runtime"
	"time"

	_ "cloud.google.com/go/storage"
	"contrib.go.opencensus.io/exporter/stackdriver"
	"github.com/pkg/errors"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
	"go.opencensus.io/trace"
)

const (
	projectEnvVar    = "MINIKUBE_GCP_PROJECT_ID"
	customMetricName = "custom.googleapis.com/minikube/start_time"
	profile          = "cloud-monitoring"
)

var (
	// The task latency in seconds
	latencyS = stats.Float64("repl/start_time", "start time in seconds", "s")
)

func main() {
	if err := execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func execute() error {
	projectID := os.Getenv(projectEnvVar)
	if projectID == "" {
		return fmt.Errorf("metrics collector requires a valid GCP project id set via the %s env variable", projectEnvVar)
	}

	osMethod, err := tag.NewKey("os")
	if err != nil {
		return errors.Wrap(err, "new tag key")
	}

	ctx, err := tag.New(context.Background(), tag.Insert(osMethod, runtime.GOOS))
	if err != nil {
		return errors.Wrap(err, "new tag")
	}
	// Register the view. It is imperative that this step exists,
	// otherwise recorded metrics will be dropped and never exported.
	v := &view.View{
		Name:        customMetricName,
		Measure:     latencyS,
		Description: "minikube start time",
		Aggregation: view.LastValue(),
	}
	if err := view.Register(v); err != nil {
		return errors.Wrap(err, "registering view")
	}

	sd, err := stackdriver.NewExporter(stackdriver.Options{
		ProjectID: projectID,
		// MetricPrefix helps uniquely identify your metrics.
		MetricPrefix: "minikube_start",
		// ReportingInterval sets the frequency of reporting metrics
		// to stackdriver backend.
		ReportingInterval: 1 * time.Second,
	})
	if err != nil {
		return errors.Wrap(err, "getting stackdriver exporter")
	}
	// It is imperative to invoke flush before your main function exits
	defer sd.Flush()

	// Register it as a trace exporter
	trace.RegisterExporter(sd)

	if err := sd.StartMetricsExporter(); err != nil {
		return errors.Wrap(err, "starting metric exporter")
	}
	defer sd.StopMetricsExporter()

	for {
		st, err := minikubeStartTime(ctx)
		if err != nil {
			log.Printf("error collecting start time: %v", err)
			continue
		}
		fmt.Printf("Latency: %f\n", st)
		stats.Record(ctx, latencyS.M(st))
		time.Sleep(30 * time.Second)
	}
}

func minikubeStartTime(ctx context.Context) (float64, error) {
	minikubePath, err := downloadMinikube()
	if err != nil {
		return 0, errors.Wrap(err, "downloading minikube")
	}
	defer os.Remove(minikubePath)
	defer deleteMinikube(ctx, minikubePath)

	cmd := exec.CommandContext(ctx, minikubePath, "start", "--driver=docker", "-p", profile, "--memory=2000")
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr

	t := time.Now()
	log.Print("Running minikube start....")
	if err := cmd.Run(); err != nil {
		return 0, errors.Wrapf(err, "running %v", cmd.Args)
	}
	return time.Since(t).Seconds(), nil
}

func deleteMinikube(ctx context.Context, minikubePath string) {
	cmd := exec.CommandContext(ctx, minikubePath, "delete", "-p", profile)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Printf("error deleting: %v", err)
	}
}
