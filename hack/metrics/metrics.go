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
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	_ "cloud.google.com/go/storage"
	mexporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/metric"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	pkgtrace "k8s.io/minikube/pkg/trace"
)

const (
	customMetricName = "custom.googleapis.com/minikube/start_time"
	profile          = "cloud-monitoring"
)

var (
	labels  string
	tmpFile string
)

func main() {
	if err := execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	flag.StringVar(&labels, "label", "", "Comma separated list of labels to add to metrics in key:value format")
	flag.StringVar(&tmpFile, "file", "/tmp/minikube", "Download minikube to this file for testing")
	flag.Parse()
}

func execute() error {
	projectID := os.Getenv(pkgtrace.ProjectEnvVar)
	if projectID == "" {
		return fmt.Errorf("metrics collector requires a valid GCP project id set via the %s env variable", pkgtrace.ProjectEnvVar)
	}
	ctx := context.Background()
	if err := downloadMinikube(ctx, tmpFile); err != nil {
		return errors.Wrap(err, "downloading minikube")
	}

	for _, cr := range []string{"docker", "containerd", "crio"} {
		if err := exportMinikubeStart(ctx, projectID, cr); err != nil {
			log.Printf("error exporting minikube start data for runtime %v: %v", cr, err)
		}
	}
	return nil
}

func exportMinikubeStart(ctx context.Context, projectID, containerRuntime string) error {
	mp, attrs, err := getMeterProvider(projectID, containerRuntime)
	if err != nil {
		return errors.Wrap(err, "creating meter provider")
	}
	defer func() { _ = mp.Shutdown(ctx) }()

	meter := mp.Meter("minikube")
	latency, err := meter.Float64Histogram(customMetricName)
	if err != nil {
		return errors.Wrap(err, "creating histogram")
	}

	st, err := minikubeStartTime(ctx, projectID, tmpFile, containerRuntime)
	if err != nil {
		return errors.Wrap(err, "collecting start time")
	}
	fmt.Printf("Latency: %f\n", st)
	latency.Record(ctx, st, metric.WithAttributes(attrs...))
	time.Sleep(30 * time.Second)
	return nil
}

func getMeterProvider(projectID, containerRuntime string) (*sdkmetric.MeterProvider, []attribute.KeyValue, error) {
	exporter, err := mexporter.New(mexporter.WithProjectID(projectID))
	if err != nil {
		return nil, nil, err
	}
	reader := sdkmetric.NewPeriodicReader(exporter, sdkmetric.WithInterval(11*time.Second))
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	return mp, getAttributes(containerRuntime), nil
}

func getAttributes(containerRuntime string) []attribute.KeyValue {
	attrs := []attribute.KeyValue{attribute.String("container-runtime", containerRuntime)}
	if labels == "" {
		return attrs
	}
	separated := strings.Split(labels, ",")
	for _, s := range separated {
		sep := strings.Split(s, ":")
		if len(sep) != 2 {
			continue
		}
		log.Printf("Adding label %s=%s to metrics...", sep[0], sep[1])
		attrs = append(attrs, attribute.String(sep[0], sep[1]))
	}
	return attrs
}

func minikubeStartTime(ctx context.Context, projectID, minikubePath, containerRuntime string) (float64, error) {
	defer deleteMinikube(ctx, minikubePath)

	cmd := exec.CommandContext(ctx, minikubePath, "start", "--driver=docker", "-p", profile, "--memory=3072", "--trace=gcp", fmt.Sprintf("--container-runtime=%s", containerRuntime))
	cmd.Env = append(os.Environ(), fmt.Sprintf("%s=%s", pkgtrace.ProjectEnvVar, projectID))
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr

	t := time.Now()
	log.Printf("Running [%v]....", cmd.Args)
	if err := cmd.Run(); err != nil {
		return 0, errors.Wrapf(err, "running %v", cmd.Args)
	}
	totalTime := time.Since(t).Seconds()
	return totalTime, nil
}

func deleteMinikube(ctx context.Context, minikubePath string) {
	cmd := exec.CommandContext(ctx, minikubePath, "delete", "-p", profile)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Printf("error deleting: %v", err)
	}
}
