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
	"errors"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/golang/glog"
	"github.com/r2d4/external-storage/lib/controller"
	"github.com/r2d4/external-storage/lib/leaderelection"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/minikube/pkg/util"
)

const (
	resyncPeriod              = 15 * time.Second
	provisionerName           = "k8s.io/minikube-hostpath"
	exponentialBackOffOnError = false
	failedRetryThreshold      = 5
	leasePeriod               = leaderelection.DefaultLeaseDuration
	retryPeriod               = leaderelection.DefaultRetryPeriod
	renewDeadline             = leaderelection.DefaultRenewDeadline
	termLimit                 = leaderelection.DefaultTermLimit
)

type hostPathProvisioner struct {
	// The directory to create PV-backing directories in
	pvDir string

	// Identity of this hostPathProvisioner, generated. Used to identify "this"
	// provisioner's PVs.
	identity types.UID
}

func NewHostPathProvisioner() controller.Provisioner {
	return &hostPathProvisioner{
		pvDir:    "/tmp/hostpath-provisioner",
		identity: uuid.NewUUID(),
	}
}

var _ controller.Provisioner = &hostPathProvisioner{}

// Provision creates a storage asset and returns a PV object representing it.
func (p *hostPathProvisioner) Provision(options controller.VolumeOptions) (*v1.PersistentVolume, error) {
	path := path.Join(p.pvDir, options.PVName)

	if err := os.MkdirAll(path, 0777); err != nil {
		return nil, err
	}

	pv := &v1.PersistentVolume{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: options.PVName,
			Annotations: map[string]string{
				"hostPathProvisionerIdentity": string(p.identity),
			},
		},
		Spec: v1.PersistentVolumeSpec{
			PersistentVolumeReclaimPolicy: options.PersistentVolumeReclaimPolicy,
			AccessModes:                   options.PVC.Spec.AccessModes,
			Capacity: v1.ResourceList{
				v1.ResourceName(v1.ResourceStorage): options.PVC.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)],
			},
			PersistentVolumeSource: v1.PersistentVolumeSource{
				HostPath: &v1.HostPathVolumeSource{
					Path: path,
				},
			},
		},
	}

	return pv, nil
}

// Delete removes the storage asset that was created by Provision represented
// by the given PV.
func (p *hostPathProvisioner) Delete(volume *v1.PersistentVolume) error {
	ann, ok := volume.Annotations["hostPathProvisionerIdentity"]
	if !ok {
		return errors.New("identity annotation not found on PV")
	}
	if ann != string(p.identity) {
		return &controller.IgnoredError{"identity annotation on PV does not match ours"}
	}

	path := path.Join(p.pvDir, volume.Name)
	if err := os.RemoveAll(path); err != nil {
		return err
	}

	return nil
}

func (lk LocalkubeServer) NewStorageProvisionerServer() Server {
	return NewSimpleServer("storage-provisioner", serverInterval, StartStorageProvisioner(lk), noop)
}

func StartStorageProvisioner(lk LocalkubeServer) func() error {

	return func() error {
		config, err := clientcmd.BuildConfigFromFlags("", util.DefaultKubeConfigPath)
		if err != nil {
			return err
		}
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			glog.Fatalf("Failed to create client: %v", err)
		}

		// The controller needs to know what the server version is because out-of-tree
		// provisioners aren't officially supported until 1.5
		serverVersion, err := clientset.Discovery().ServerVersion()
		if err != nil {
			return fmt.Errorf("Error getting server version: %v", err)
		}

		// Create the provisioner: it implements the Provisioner interface expected by
		// the controller
		hostPathProvisioner := NewHostPathProvisioner()

		// Start the provision controller which will dynamically provision hostPath
		// PVs
		pc := controller.NewProvisionController(clientset, resyncPeriod, provisionerName, hostPathProvisioner, serverVersion.GitVersion, exponentialBackOffOnError, failedRetryThreshold, leasePeriod, renewDeadline, retryPeriod, termLimit)

		pc.Run(wait.NeverStop)
		return nil
	}
}
