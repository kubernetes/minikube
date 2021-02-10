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

package trace

import (
	"context"
	"fmt"
	"os"

	"go.opentelemetry.io/otel/api/trace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"k8s.io/klog/v2"

	texporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel/api/global"
)

const (
	// ProjectEnvVar is the name of the env variable that the user must pass in their GCP project ID through
	ProjectEnvVar = "MINIKUBE_GCP_PROJECT_ID"
	// this is the name of the parent span to help identify it
	// in the Cloud Trace UI.
	parentSpanName = "minikube start"
)

type gcpTracer struct {
	projectID string
	parentCtx context.Context
	trace.Tracer
	spans   map[string]trace.Span
	cleanup func()
}

// StartSpan starts a span for the next step of
// `minikube start` via the GCP tracer
func (t *gcpTracer) StartSpan(name string) {
	_, span := t.Tracer.Start(t.parentCtx, name)
	t.spans[name] = span
}

// EndSpan ends the most recent span, indicating
// that one step of `minikube start` has completed
func (t *gcpTracer) EndSpan(name string) {
	span, ok := t.spans[name]
	if !ok {
		klog.Warningf("cannot end span %s as it was never started", name)
		return
	}
	span.End()
}

func (t *gcpTracer) Cleanup() {
	span, ok := t.spans[parentSpanName]
	if ok {
		span.End()
	}
	t.cleanup()
}

func initGCPTracer() (*gcpTracer, error) {
	projectID := os.Getenv(ProjectEnvVar)
	if projectID == "" {
		return nil, fmt.Errorf("GCP tracer requires a valid GCP project id set via the %s env variable", ProjectEnvVar)
	}

	_, flush, err := texporter.InstallNewPipeline(
		[]texporter.Option{
			texporter.WithProjectID(projectID),
		},
		sdktrace.WithConfig(sdktrace.Config{
			DefaultSampler: sdktrace.AlwaysSample(),
		}),
	)
	if err != nil {
		return nil, errors.Wrap(err, "installing pipeline")
	}

	t := global.Tracer(parentSpanName)

	ctx, span := t.Start(context.Background(), parentSpanName)
	return &gcpTracer{
		projectID: projectID,
		parentCtx: ctx,
		cleanup:   flush,
		Tracer:    t,
		spans: map[string]trace.Span{
			parentSpanName: span,
		},
	}, nil
}
