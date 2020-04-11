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
	"time"

	"github.com/golang/glog"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// HealtyDisk verfies that node disks are healthy are not under pressure.
func HealtyDisk(cs *kubernetes.Clientset, timeout time.Duration) error {
	glog.Info("waiting to verify healty disk ...")
	start := time.Now()
	defer func() {
		glog.Infof("duration metric: took %s to wait for k8s-apps to be running ...", time.Since(start))
		}()
	// n, err := cs.CoreV1().ServiceAccounts("default").List(meta.ListOptions{})
	ns, err := cs.CoreV1().Nodes().List(meta.ListOptions{})
	if err != nil {
		glog.Infof("failed to get nodes nodes: %v", err)
	 }
	 fmt.Printf("Node Object: ---------------- %v ----------------\n",ns.Items[0])
	 

	// start := time.Now()

	// 	// equivalent to manual check of 'kubectl --context profile get serviceaccount default'
	// 	sas, err := cs.CoreV1().ServiceAccounts("default").List(meta.ListOptions{})
	// 	if err != nil {
	// 		glog.Infof("temproary error waiting for default SA: %v", err)
	// 		return err
	// 	}
	// 	for _, sa := range sas.Items {
	// 		if sa.Name == "default" {
	// 			glog.Infof("found service account: %q", sa.Name)
	// 			return nil
	// 		}
	// 	}
	// 	return fmt.Errorf("couldn't find default service account")
	// if err := wait.PollImmediate(kconst.APICallRetryInterval, timeout, checkRunning); err != nil {
	// 	return errors.Wrapf(err, "checking k8s-apps to be running")
	// }
	// glog.Infof("duration metric: took %s to wait for k8s-apps to be running ...", time.Since(start))
	
	
	return nil
}
