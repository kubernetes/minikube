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

package storage

import (
	"fmt"
	"os"
	"path"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	"github.com/r2d4/external-storage/lib/controller"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
)

const provisionerName = "k8s.io/minikube-hostpath"

type hostPathProvisioner struct {
	// The directory to create PV-backing directories in
	pvDir string

	// Identity of this hostPathProvisioner, generated. Used to identify "this"
	// provisioner's PVs.
	identity types.UID
}

// NewHostPathProvisioner creates a new Provisioner using host paths
func NewHostPathProvisioner() controller.Provisioner {
	return &hostPathProvisioner{
		pvDir:    "/tmp/hostpath-provisioner",
		identity: uuid.NewUUID(),
	}
}

var _ controller.Provisioner = &hostPathProvisioner{}

// Provision creates a storage asset and returns a PV object representing it.
func (p *hostPathProvisioner) Provision(options controller.VolumeOptions) (*v1.PersistentVolume, error) {
	glog.Infof("Provisioning volume %v", options)
	path := path.Join(p.pvDir, options.PVName)
	if err := os.MkdirAll(path, 0777); err != nil {
		return nil, err
	}

	// Explicitly chmod created dir, so we know mode is set to 0777 regardless of umask
	if err := os.Chmod(path, 0777); err != nil {
		return nil, err
	}

	pv := &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
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
	glog.Infof("Deleting volume %v", volume)
	ann, ok := volume.Annotations["hostPathProvisionerIdentity"]
	if !ok {
		return errors.New("identity annotation not found on PV")
	}
	if ann != string(p.identity) {
		return &controller.IgnoredError{Reason: "identity annotation on PV does not match ours"}
	}

	path := path.Join(p.pvDir, volume.Name)
	if err := os.RemoveAll(path); err != nil {
		return errors.Wrap(err, "removing hostpath PV")
	}

	return nil
}

// StartStorageProvisioner will start storage provisioner server
func StartStorageProvisioner() error {
	glog.Infof("Initializing the Minikube storage provisioner...")
	config, err := restclient.InClusterConfig()
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
	pc := controller.NewProvisionController(clientset, provisionerName, hostPathProvisioner, serverVersion.GitVersion)

	glog.Info("Storage provisioner initialized, now starting service!")
	pc.Run(wait.NeverStop)
	return nil
}
