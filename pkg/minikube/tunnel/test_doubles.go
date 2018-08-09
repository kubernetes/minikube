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
	"github.com/sirupsen/logrus"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/tunnel/types"
)

type recordingReporter struct {
	statesRecorded []*types.TunnelState
}

func (r *recordingReporter) Report(tunnelState *types.TunnelState) {
	logrus.Debugf("recordingReporter.Report: %v", tunnelState)
	r.statesRecorded = append(r.statesRecorded, tunnelState)
}

type fakeRouter struct {
	osRoutes []*types.Route
	osRouter
	errorResponse error
}

func (r *fakeRouter) EnsureRouteIsAdded(route *types.Route) error {
	if r.errorResponse == nil {
		r.osRoutes = append(r.osRoutes, route)
	}
	return r.errorResponse
}
func (r *fakeRouter) Cleanup(route *types.Route) error {
	if r.errorResponse == nil {
		r.osRoutes = []*types.Route{}
	}
	return r.errorResponse
}

type stubConfigLoader struct {
	c config.Config
	e error
}

func (l *stubConfigLoader) LoadConfigFromFile(profile string) (config.Config, error) {
	return l.c, l.e
}

////
////type fakeCoreClient struct {
////	fake.FakeCoreV1
////}
////
////func (c *fakeCoreClient) Services(namespace string) v1.ServiceInterface {
////	return &fake.FakeServices{&c.FakeCoreV1, namespace}
////}
//
//func (c *fakeCoreClient) RESTClient() rest.Interface {
//	var ret *rest.RESTClient
//	return ret
//}

type FakeRouter struct {
}
