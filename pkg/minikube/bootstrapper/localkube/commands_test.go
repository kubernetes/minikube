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
	gflag "flag"
	"fmt"
	"strings"
	"testing"

	"k8s.io/minikube/pkg/minikube/bootstrapper"
	"k8s.io/minikube/pkg/util"
)

func TestGetStartCommandCustomValues(t *testing.T) {
	flagMap := map[string]string{
		"v":       "10",
		"vmodule": "cluster*=5",
	}
	flagMapToSetFlags(flagMap)
	startCommand, err := GetStartCommand(bootstrapper.KubernetesConfig{})
	if err != nil {
		t.Fatalf("Error generating start command: %s", err)
	}

	for flag, val := range flagMap {
		if val != "" {
			if expectedFlag := getSingleFlagValue(flag, val); !strings.Contains(startCommand, getSingleFlagValue(flag, val)) {
				t.Fatalf("Expected GetStartCommand to contain: %s.", expectedFlag)
			}
		}
	}
}

func TestGetStartCommandExtraOptions(t *testing.T) {
	k := bootstrapper.KubernetesConfig{
		ExtraOptions: util.ExtraOptionSlice{
			util.ExtraOption{Component: "a", Key: "b", Value: "c"},
			util.ExtraOption{Component: "d", Key: "e.f", Value: "g"},
		},
	}
	startCommand, err := GetStartCommand(k)
	if err != nil {
		t.Fatalf("Error generating start command: %s", err)
	}
	for _, arg := range []string{"--extra-config=a.b=c", "--extra-config=d.e.f=g"} {
		if !strings.Contains(startCommand, arg) {
			t.Fatalf("Error, expected to find argument: %s. Got: %s", arg, startCommand)
		}
	}
}

func flagMapToSetFlags(flagMap map[string]string) {
	for flag, val := range flagMap {
		gflag.Set(flag, val)
	}
}
func getSingleFlagValue(flag, val string) string {
	return fmt.Sprintf("--%s %s", flag, val)
}
