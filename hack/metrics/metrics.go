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
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	_ "cloud.google.com/go/storage"
	"contrib.go.opencensus.io/exporter/stackdriver"
	"github.com/pkg/errors"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
	pkgtrace "k8s.io/minikube/pkg/trace"
)

const (
	customMetricName = "custom.googleapis.com/minikube/start_time"
	profile          = "cloud-monitoring"
)

var (
	// The task latency in seconds
	latencyS = stats.Float64("repl/start_time", "start time in seconds", "s")
	labels   string
)

func main() {
	if err := execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	flag.StringVar(&labels, "label", "", "Comma separated list of labels to add to metrics in key:value format")
	flag.Parse()
}

func execute() error {
	projectID := os.Getenv(pkgtrace.ProjectEnvVar)
	if projectID == "" {
		return fmt.Errorf("metrics collector requires a valid GCP project id set via the %s env variable", pkgtrace.ProjectEnvVar)
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
		ReportingInterval:       1 * time.Second,
		DefaultMonitoringLabels: getLabels(),
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
	ctx := context.Background()

	// create a tmp file to download minikube to
	tmp, err := ioutil.TempFile("", binary())
	if err != nil {
		return errors.Wrap(err, "creating tmp file")
	}
	if err := tmp.Close(); err != nil {
		return errors.Wrap(err, "closing file")
	}
	defer os.Remove(tmp.Name())

	// start loop where we will track minikube start time
	for {
		st, err := minikubeStartTime(ctx, projectID, tmp.Name())
		if err != nil {
			log.Printf("error collecting start time: %v", err)
			continue
		}
		fmt.Printf("Latency: %f\n", st)
		stats.Record(ctx, latencyS.M(st))
		time.Sleep(30 * time.Second)
	}
}

func getLabels() *stackdriver.Labels {
	l := &stackdriver.Labels{}
	if labels == "" {
		return l
	}
	separated := strings.Split(labels, ",")
	for _, s := range separated {
		sep := strings.Split(s, ":")
		if len(sep) != 2 {
			continue
		}
		log.Printf("Adding label %s=%s to metrics...", sep[0], sep[1])
		l.Set(sep[0], sep[1], "")
	}
	return l
}

func minikubeStartTime(ctx context.Context, projectID, minikubePath string) (float64, error) {
	if err := downloadMinikube(ctx, minikubePath); err != nil {
		return 0, errors.Wrap(err, "downloading minikube")
	}
	defer deleteMinikube(ctx, minikubePath)

	cmd := exec.CommandContext(ctx, minikubePath, "start", "--driver=docker", "-p", profile, "--memory=2000", "--trace=gcp")
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
