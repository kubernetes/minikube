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
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"k8s.io/minikube/pkg/minikube/localpath"
)

// UniqueProfileName returns a reasonably unique profile name
func UniqueProfileName(prefix string) string {
	if *forceProfile != "" {
		return *forceProfile
	}
	if NoneDriver() {
		return "minikube"
	}
	// example: prefix-20200413162239-3215
	return fmt.Sprintf("%s-%s-%d", prefix, time.Now().Format("20060102150405"), os.Getpid())
}

// auditContains checks if the provided string is contained within the logs.
func auditContains(substr string) (bool, error) {
	f, err := os.Open(localpath.AuditLog())
	if err != nil {
		return false, fmt.Errorf("Unable to open file %s: %v", localpath.AuditLog(), err)
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		if strings.Contains(s.Text(), substr) {
			return true, nil
		}
	}
	return false, nil
}
