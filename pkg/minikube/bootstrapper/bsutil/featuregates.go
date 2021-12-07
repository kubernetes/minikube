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
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/minikube/third_party/kubeadm/app/features"
)

// supportedFG indicates whether a feature name is supported by the bootstrapper
func supportedFG(featureName string) bool {
	for k := range features.InitFeatureGates {
		if featureName == k {
			return true
		}
	}
	return false
}

// parseFeatureArgs parses feature args into extra args
func parseFeatureArgs(featureGates string) (map[string]bool, string, error) {
	kubeadmFeatureArgs := map[string]bool{}
	componentFeatureArgs := ""
	for _, s := range strings.Split(featureGates, ",") {
		if len(s) == 0 {
			continue
		}

		fg := strings.SplitN(s, "=", 2)
		if len(fg) != 2 {
			return nil, "", fmt.Errorf("missing value for key \"%v\"", s)
		}

		k := strings.TrimSpace(fg[0])
		v := strings.TrimSpace(fg[1])

		if !supportedFG(k) {
			componentFeatureArgs = fmt.Sprintf("%s%s,", componentFeatureArgs, s)
			continue
		}

		boolValue, err := strconv.ParseBool(v)
		if err != nil {
			return nil, "", errors.Wrapf(err, "failed to convert bool value \"%v\"", v)
		}
		kubeadmFeatureArgs[k] = boolValue
	}
	componentFeatureArgs = strings.TrimRight(componentFeatureArgs, ",")
	return kubeadmFeatureArgs, componentFeatureArgs, nil
}
