/*
Copyright 2018 The Kubernetes Authors All rights reserved.

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

package tunnel

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"k8s.io/minikube/pkg/minikube/config"
)

type recordingReporter struct {
	statesRecorded []*TunnelStatus
}

func (r *recordingReporter) Report(tunnelState *TunnelStatus) {
	logrus.Debugf("recordingReporter.Report: %v", tunnelState)
	r.statesRecorded = append(r.statesRecorded, tunnelState)
}

//simulating idempotent router behavior
//without checking for conflicting routes
type fakeRouter struct {
	rt            routingTable
	errorResponse error
}

func (r *fakeRouter) EnsureRouteIsAdded(route *Route) error {
	logrus.Debugf("fakerouter.EnsureRouteIsAdded %s", route)
	if r.errorResponse == nil {
		exists, e := isValidToAddOrDelete(r, route)
		if e != nil {
			return e
		}
		if !exists {
			r.rt = append(r.rt, routingTableLine{
				route: route,
				line:  fmt.Sprintf("fake router line: %s", route),
			})
		}

	}
	return r.errorResponse
}
func (r *fakeRouter) Cleanup(route *Route) error {
	logrus.Debugf("fake router cleanup: %v\n", route)
	if r.errorResponse == nil {
		exists, e := isValidToAddOrDelete(r, route)
		if e != nil {
			return e
		}
		if exists {
			for i := range r.rt {
				if r.rt[i].route.Equal(route) {
					r.rt = append(r.rt[:i], r.rt[i+1:]...)
					break
				}
			}
		}
	}
	return r.errorResponse
}

func (r *fakeRouter) Inspect(route *Route) (exists bool, conflict string, overlaps []string, err error) {
	err = r.errorResponse
	exists, conflict, overlaps = r.rt.Check(route)
	return
}

type stubConfigLoader struct {
	c config.Config
	e error
}

func (l *stubConfigLoader) LoadConfigFromFile(profile string) (config.Config, error) {
	return l.c, l.e
}
