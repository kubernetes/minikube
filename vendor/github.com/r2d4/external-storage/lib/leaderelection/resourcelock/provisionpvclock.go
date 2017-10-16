/*
Copyright 2016 The Kubernetes Authors.

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

package resourcelock

import (
	"encoding/json"
	"errors"
	"fmt"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
)

// ProvisionPVCLock is a lock on an existing PVC to provision a PV for
type ProvisionPVCLock struct {
	// PVCMeta should contain a Name and a Namespace of a PVC
	// object that the LeaderElector will attempt to lead.
	PVCMeta    metav1.ObjectMeta
	Client     clientset.Interface
	LockConfig Config
	p          *v1.PersistentVolumeClaim
}

// Get returns the LeaderElectionRecord
func (pl *ProvisionPVCLock) Get() (*LeaderElectionRecord, error) {
	var record LeaderElectionRecord
	var err error
	pl.p, err = pl.Client.Core().PersistentVolumeClaims(pl.PVCMeta.Namespace).Get(pl.PVCMeta.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	// TODO there should be a way to give up if the pvc is already bound...we are doing a Get regardless
	if pl.p.Annotations == nil {
		pl.p.Annotations = make(map[string]string)
	}
	if recordBytes, found := pl.p.Annotations[LeaderElectionRecordAnnotationKey]; found {
		if err := json.Unmarshal([]byte(recordBytes), &record); err != nil {
			return nil, err
		}
	}
	return &record, nil
}

// Create is not allowed, the PVC should already exist
func (pl *ProvisionPVCLock) Create(ler LeaderElectionRecord) error {
	return errors.New("create not allowed, PVC should already exist")
}

// Update will update and existing annotation on a given resource.
func (pl *ProvisionPVCLock) Update(ler LeaderElectionRecord) error {
	if pl.p == nil {
		return errors.New("PVC not initialized, call get first")
	}
	recordBytes, err := json.Marshal(ler)
	if err != nil {
		return err
	}
	pl.p.Annotations[LeaderElectionRecordAnnotationKey] = string(recordBytes)
	pl.p, err = pl.Client.Core().PersistentVolumeClaims(pl.PVCMeta.Namespace).Update(pl.p)
	return err
}

// RecordEvent in leader election while adding meta-data
func (pl *ProvisionPVCLock) RecordEvent(s string) {
	// events := fmt.Sprintf("%v %v", pl.LockConfig.Identity, s)
	// pl.LockConfig.EventRecorder.Eventf(&v1.PersistentVolumeClaim{ObjectMeta: pl.p.ObjectMeta}, v1.EventTypeNormal, "LeaderElection", events)
}

// Describe is used to convert details on current resource lock
// into a string
func (pl *ProvisionPVCLock) Describe() string {
	return fmt.Sprintf("to provision for pvc %v/%v", pl.PVCMeta.Namespace, pl.PVCMeta.Name)
}

// Identity returns the Identity of the lock
func (pl *ProvisionPVCLock) Identity() string {
	return pl.LockConfig.Identity
}
