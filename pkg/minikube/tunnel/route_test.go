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
	"k8s.io/minikube/pkg/minikube/tunnel/types"
	"net"
)

func parseRoute(ipString string, gwString string) *types.Route {
	ip := net.ParseIP(ipString)
	_, ipNet, _ := net.ParseCIDR(gwString)

	expectedRoute := &types.Route{
		Gateway:  ip,
		DestCIDR: ipNet,
	}
	return expectedRoute
}
