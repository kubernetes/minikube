/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package localkube

import (
	"io/ioutil"
	"net/http"

	"github.com/golang/glog"
)

type HealthCheck func() bool

func healthCheck(addr string, lk LocalkubeServer) HealthCheck {
	return func() bool {
		glog.Infof("Performing healthcheck on %s\n", addr)

		transport, err := lk.GetTransport()
		if err != nil {
			return false
		}
		client := http.Client{Transport: transport}

		resp, err := client.Get(addr)
		if err != nil {
			glog.Errorf("Error performing healthcheck: %s", err)
			return false
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			glog.Errorf("Error reading healthcheck response: %s", err)
			return false
		}
		glog.Infof("Got healthcheck response: %s", body)
		return string(body) == "ok"
	}
}

func noop() bool {
	return true
}
