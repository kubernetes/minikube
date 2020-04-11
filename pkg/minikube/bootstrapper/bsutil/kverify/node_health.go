/*
Copyright 2020 The Kubernetes Authors All rights reserved.

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

// Package kverify verifies a running kubernetes cluster is healthy
package kverify

import (
	"fmt"
	"runtime"
	"time"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/out"
)

// NodePressure verfies that node is not under disk, memory, pid or network pressure.
func NodePressure(cs *kubernetes.Clientset, drver string) error {
	glog.Info("waiting to verify node health ...")
	start := time.Now()
	defer func() {
		glog.Infof("duration metric: took %s to wait for k8s-apps to be running ...", time.Since(start))
	}()

	ns, err := cs.CoreV1().Nodes().List(meta.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "list nodes")
	}

	for _, n := range ns.Items {
		glog.Infof("node storage ephemeral capacity is %s", n.Status.Capacity.StorageEphemeral())
		glog.Infof("node cpu capacity is %s", n.Status.Capacity.Cpu().AsDec())
		for _, c := range n.Status.Conditions {
			if c.Type == v1.NodeDiskPressure && c.Status != v1.ConditionTrue {
				out.Ln("")
				out.ErrT(out.FailureType, "node {{.name}} has unwanted condition {{.condition_type}} : Reason {{.reason}} Message: {{.message}}", out.V{"name": n.Name, "condition_type": c.Type, "reason": c.Reason, "message": c.Message})
				out.WarningT("The node on {{.name}} has ran out of disk space. please consider allocating more disk using or pruning un-used images", out.V{"name": n.Name})
				if driver.IsKIC(drver) && runtime.GOOS != "linux" {
					out.T(out.Stopped, "Please increase Docker Desktop's disk image size.")
					if runtime.GOOS == "darwin" {
						out.T(out.Documentation, "Documentation: {{.url}}", out.V{"url": "https://docs.docker.com/docker-for-mac/space/"})
					}
					if runtime.GOOS == "windows" {
						out.T(out.Documentation, "Documentation: {{.url}}", out.V{"url": "https://docs.docker.com/docker-for-windows/"})
					}
				} else { // VM-drivers
					out.T(out.Stopped, "You can specify a larger disk for your cluster using `minikube start --disk` ")
				}
				out.Ln("") // if there is error message, lets make an empty space for better visilibtly
				return fmt.Errorf("node %q has unwanted condition %q : Reason %q Message: %q ", n.Name, c.Type, c.Reason, c.Message)
			}

			if c.Type == v1.NodeMemoryPressure && c.Status == v1.ConditionTrue {
				out.Ln("")
				out.ErrT(out.FailureType, "node {{.name}} has unwanted condition {{.condition_type}} : Reason {{.reason}} Message: {{.message}}", out.V{"name": n.Name, "condition_type": c.Type, "reason": c.Reason, "message": c.Message})
				out.WarningT("The node on {{.name}} has ran of memory.", out.V{"name": n.Name})
				if driver.IsKIC(drver) && runtime.GOOS != "linux" {
					out.T(out.Stopped, "Please increase Docker Desktop's memory.")
					if runtime.GOOS == "darwin" {
						out.T(out.Documentation, "Documentation: {{.url}}", out.V{"url": "https://docs.docker.com/docker-for-mac/space/"})
					}
					if runtime.GOOS == "windows" {
						out.T(out.Documentation, "Documentation: {{.url}}", out.V{"url": "https://docs.docker.com/docker-for-windows/"})
					}
				} else {
					out.T(out.Stopped, "You can specify a larger memory size for your cluster using `minikube start --memory` ")
				}
				out.Ln("") // if there is error message, lets make an empty space for better visilibtly
				return fmt.Errorf("node %q has unwanted condition %q : Reason %q Message: %q ", n.Name, c.Type, c.Reason, c.Message)
			}

			if c.Type == v1.NodePIDPressure && c.Status == v1.ConditionTrue {
				out.Ln("")
				out.ErrT(out.FailureType, "node {{.name}} has unwanted condition {{.condition_type}} : Reason {{.reason}} Message: {{.message}}", out.V{"name": n.Name, "condition_type": c.Type, "reason": c.Reason, "message": c.Message})
				out.WarningT("The node has ran out of available PIDs.", out.V{"name": n.Name})
				out.Ln("")
				return fmt.Errorf("node %q has unwanted condition %q : Reason %q Message: %q ", n.Name, c.Type, c.Reason, c.Message)
			}

			if c.Type == v1.NodeNetworkUnavailable && c.Status == v1.ConditionTrue {
				out.Ln("")
				out.ErrT(out.FailureType, "node {{.name}} has unwanted condition {{.condition_type}} : Reason {{.reason}} Message: {{.message}}", out.V{"name": n.Name, "condition_type": c.Type, "reason": c.Reason, "message": c.Message})
				out.WarningT("The node netowrking is not configured correctly.", out.V{"name": n.Name})
				out.Ln("")
				return fmt.Errorf("node %q has unwanted condition %q : Reason %q Message: %q ", n.Name, c.Type, c.Reason, c.Message)
			}

			if c.Type == v1.NodeReady && c.Status == v1.ConditionFalse {
				out.Ln("")
				out.ErrT(out.FailureType, "node {{.name}} has unwanted condition {{.condition_type}} : Reason {{.reason}} Message: {{.message}}", out.V{"name": n.Name, "condition_type": c.Type, "reason": c.Reason, "message": c.Message})
				out.WarningT("The node is not ready.", out.V{"name": n.Name})
				out.T(out.Tip, "get more information by running `kubectl describe nodes -A`")
				out.Ln("") // if there is error message, lets make an empty space for better visilibtly
				return fmt.Errorf("node %q has unwanted condition %q : Reason %q Message: %q ", n.Name, c.Type, c.Reason, c.Message)
			}
		}
	}
	glog.Infof("duration metric: took %s to wait for node-health ...", time.Since(start))
	return nil
}
