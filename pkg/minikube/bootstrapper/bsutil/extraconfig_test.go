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

// Package bsutil will eventually be renamed to kubeadm package after getting rid of older one
package bsutil

import (
	"reflect"
	"testing"

	"k8s.io/minikube/pkg/minikube/config"
)

func TestFindInvalidExtraConfigFlags(t *testing.T) {
	defaultOpts := getExtraOpts()
	badOption1 := config.ExtraOption{Component: "bad_option_1"}
	badOption2 := config.ExtraOption{Component: "bad_option_2"}
	tests := []struct {
		name string
		opts config.ExtraOptionSlice
		want []string
	}{
		{
			name: "with valid options only",
			opts: defaultOpts,
			want: nil,
		},
		{
			name: "with invalid options",
			opts: append(defaultOpts, badOption1, badOption2),
			want: []string{"bad_option_1", "bad_option_2"},
		},
		{
			name: "with invalid options and duplicates",
			opts: append(defaultOpts, badOption2, badOption1, badOption1),
			want: []string{"bad_option_2", "bad_option_1"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FindInvalidExtraConfigFlags(tt.opts); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FindInvalidExtraConfigFlags() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestCreateFlagsFromExtraArgs checks that CreateFlagsFromExtraArgs correctly filters and formats
// extra args into kubeadm command line flags.
// This is important because kubeadm has strict rules about which flags can be combined with --config;
// passing config-file parameters as flags will cause kubeadm initialization to fail.
func TestCreateFlagsFromExtraArgs(t *testing.T) {
	tests := []struct {
		name string
		opts config.ExtraOptionSlice
		want string
	}{
		{
			name: "allowed command line args",
			opts: config.ExtraOptionSlice{
				{Component: Kubeadm, Key: "ignore-preflight-errors", Value: "all"},
				{Component: Kubeadm, Key: "dry-run", Value: "true"},
			},
			want: "--dry-run=true --ignore-preflight-errors=all",
		},
		{
			name: "filtered config file args",
			opts: config.ExtraOptionSlice{
				{Component: Kubeadm, Key: "pod-network-cidr", Value: "10.0.0.0/24"}, // Should be filtered out
				{Component: Kubeadm, Key: "ignore-preflight-errors", Value: "SystemVerification"},
			},
			want: "--ignore-preflight-errors=SystemVerification",
		},
		{
			name: "mixed components",
			opts: config.ExtraOptionSlice{
				{Component: Kubeadm, Key: "kubeconfig", Value: "/tmp/conf"},
				{Component: Kubelet, Key: "v", Value: "4"}, // Should be ignored by this function
			},
			want: "--kubeconfig=/tmp/conf",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CreateFlagsFromExtraArgs(tt.opts); got != tt.want {
				t.Errorf("CreateFlagsFromExtraArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}
