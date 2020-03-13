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

package machine

import (
	"sync"

	"github.com/docker/machine/libmachine/host"
	"github.com/golang/glog"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/drvcmd"
)

var cachedRunner sync.Map

// Runner returns best available command runner for this host
func Runner(h *host.Host) (command.Runner, error) {
	result, ok := cachedRunner.Load(h.Name)
	if ok {
		glog.Errorf("NOTE: returning cached runner for %q", h.Name)
		return result.(command.Runner), nil
	}
	cr, err := drvcmd.Runner(h.Driver)
	if err != nil {
		return nil, err
	}
	cachedRunner.Store(h.Name, cr)
	return cr, nil
}
