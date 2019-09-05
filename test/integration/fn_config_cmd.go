// +build integration

/*
Copyright 2019 The Kubernetes Authors All rights reserved.

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
	"strings"
	"testing"
	"time"
)

func TestConfigCmd(t *testing.T) {
	MaybeParallel(t)
	profile := Profile("config")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer Cleanup(t, profile, cancel)

	tests := []struct {
		args    []string
		wantOut string
		wantErr string
	}{
		{[]string{"unset", "cpus"}, "", ""},
		{[]string{"get", "cpus"}, "", "Error: specified key could not be found in config"},
		{[]string{"set", "cpus 2"}, "! These changes will take effect upon a minikube delete and then a minikube start", ""},
		{[]string{"get", "cpus"}, "2", ""},
		{[]string{"unset", "cpus"}, "", ""},
		{[]string{"get", "cpus"}, "", "Error: specified key could not be found in config"},
	}

	for _, tc := range tests {
		args := append([]string{"-p", profile, "config"}, tc.args...)
		rr, err := RunCmd(ctx, t, Target(), args...)
		if err != nil {
			t.Errorf("%s failed: %v", rr.Cmd.Args, err)
		}

		got := strings.TrimSpace(rr.Stdout.String())
		if got != tc.wantOut {
			t.Errorf("config %s stdout got: %q, want: %q", tc.args, got, tc.wantOut)
		}
		got = strings.TrimSpace(rr.Stderr.String())
		if got != tc.wantErr {
			t.Errorf("config %s stderr got: %q, want: %q", tc.args, got, tc.wantErr)
		}
	}
}
