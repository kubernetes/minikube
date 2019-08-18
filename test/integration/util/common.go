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

package util

import (
	"fmt"
	"testing"
	"time"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/minikube/pkg/kube"
)

// WaitForBusyboxRunning waits until busybox pod to be running
func WaitForBusyboxRunning(t *testing.T, namespace string, miniProfile string) error {
	client, err := kube.Client(miniProfile)
	if err != nil {
		return errors.Wrap(err, "getting kubernetes client")
	}
	selector := labels.SelectorFromSet(labels.Set(map[string]string{"integration-test": "busybox"}))
	return kube.WaitForPodsWithLabelRunning(client, namespace, selector)
}

// Logf writes logs to stdout if -v is set.
func Logf(str string, args ...interface{}) {
	if !testing.Verbose() {
		return
	}
	fmt.Printf(" %s | ", time.Now().Format("15:04:05"))
	fmt.Println(fmt.Sprintf(str, args...))
}
