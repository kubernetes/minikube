/*
Copyright 2022 The Kubernetes Authors All rights reserved.

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

package host

import (
	"bytes"
	"encoding/json"

	"k8s.io/minikube/pkg/libmachine/log"
)

type RawHost struct {
	Driver *json.RawMessage
}

func MigrateHostV2ToHostV3(hostV2 *V2, data []byte, storePath string) *Host {
	// Migrate to include RawDriver so that driver plugin will work
	// smoothly.
	rawHost := &RawHost{}
	if err := json.Unmarshal(data, &rawHost); err != nil {
		log.Warnf("Could not unmarshal raw host for RawDriver information: %s", err)
	}

	m := make(map[string]interface{})

	// Must migrate to include store path in driver since it was not
	// previously stored in drivers directly
	d := json.NewDecoder(bytes.NewReader(*rawHost.Driver))
	d.UseNumber()
	if err := d.Decode(&m); err != nil {
		log.Warnf("Could not unmarshal raw host into map[string]interface{}: %s", err)
	}

	m["StorePath"] = storePath

	// Now back to []byte
	rawDriver, err := json.Marshal(m)
	if err != nil {
		log.Warnf("Could not re-marshal raw driver: %s", err)
	}

	h := &Host{
		ConfigVersion: 2,
		DriverName:    hostV2.DriverName,
		Name:          hostV2.Name,
		HostOptions:   hostV2.HostOptions,
		RawDriver:     rawDriver,
	}

	return h
}
