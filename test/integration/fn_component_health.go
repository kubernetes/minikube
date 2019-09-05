// +build integration

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

package integration

import (
	"context"
	"bytes"
	"encoding/json"
	"testing"
	api "k8s.io/kubernetes/pkg/apis/core"
)

func validateComponentHealth(ctx context.Context,t *testing.T, profile string) {
	rr, err := RunCmd(ctx, t, "kubectl", "--context", profile, "get", "cs", "-o=json")
	if err != nil {
		t.Fatalf("%s failed: %v", rr.Cmd.Args, err)
	}
	cs := api.ComponentStatusList{}
	d := json.NewDecoder(bytes.NewReader(rr.Stdout.Bytes()))
	if err := d.Decode(cs); err != nil {
		t.Fatalf("decode: %v", err)
	}

	for _, i := range cs.Items {
		status := api.ConditionFalse
		for _, c := range i.Conditions {
			if c.Type != api.ComponentHealthy {
				continue
			}
			status = c.Status
		}
		if status != api.ConditionTrue {
			t.Errorf("component %s is not Healthy! Status: %s", i.GetName(), status)
		}
	}
}
