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
	"fmt"

	"github.com/pkg/errors"
)

var (
	tracer minikubeTracer
)

type minikubeTracer interface {
	StartSpan(string)
	EndSpan(string)
	Cleanup()
}

// Initialize intializes the global tracer variable
func Initialize(t string) error {
	tr, err := getTracer(t)
	if err != nil {
		return errors.Wrap(err, "getting tracer")
	}
	tracer = tr
	return nil
}

func getTracer(t string) (minikubeTracer, error) {
	switch t {
	case "gcp":
		return initGCPTracer()
	case "":
		return nil, nil
	}
	return nil, fmt.Errorf("%s is not a valid tracer, valid tracers include: [gcp]", t)
}

// StartSpan starts a span with the given name
func StartSpan(name string) {
	if tracer == nil {
		return
	}
	tracer.StartSpan(name)
}

// EndSpan ends a span with the given name
func EndSpan(name string) {
	if tracer == nil {
		return
	}
	tracer.EndSpan(name)
}

// Cleanup is responsible for trace related cleanup,
// such as flushing all data
func Cleanup() {
	if tracer == nil {
		return
	}
	tracer.Cleanup()
}
