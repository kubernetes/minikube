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

package controller

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"sync"
	"testing"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	fakev1core "k8s.io/client-go/kubernetes/typed/core/v1/fake"
	"k8s.io/client-go/pkg/api/resource"
	"k8s.io/client-go/pkg/api/testapi"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/storage/v1beta1"
	"k8s.io/client-go/pkg/conversion"
	"k8s.io/client-go/pkg/runtime"
	"k8s.io/client-go/pkg/types"
	"k8s.io/client-go/pkg/watch"
	testclient "k8s.io/client-go/testing"
	fcache "k8s.io/client-go/tools/cache/testing"

	"k8s.io/minikube/pkg/localkube/leaderelection"
	rl "k8s.io/minikube/pkg/localkube/leaderelection/resourcelock"
)

const (
	resyncPeriod         = 100 * time.Millisecond
	failedRetryThreshold = 5
)

// TODO clean this up, e.g. remove redundant params (provisionerName: "foo.bar/baz")
func TestController(t *testing.T) {
	tests := []struct {
		name            string
		objs            []runtime.Object
		provisionerName string
		provisioner     Provisioner
		verbs           []string
		reaction        testclient.ReactionFunc
		expectedVolumes []v1.PersistentVolume
	}{
		{
			name: "provision for claim-1 but not claim-2",
			objs: []runtime.Object{
				newStorageClass("class-1", "foo.bar/baz"),
				newStorageClass("class-2", "abc.def/ghi"),
				newClaim("claim-1", "uid-1-1", "class-1", "", nil),
				newClaim("claim-2", "uid-1-2", "class-2", "", nil),
			},
			provisionerName: "foo.bar/baz",
			provisioner:     newTestProvisioner(),
			expectedVolumes: []v1.PersistentVolume{
				*newProvisionedVolume(newStorageClass("class-1", "foo.bar/baz"), newClaim("claim-1", "uid-1-1", "class-1", "", nil)),
			},
		},
		{
			name: "delete volume-1 but not volume-2",
			objs: []runtime.Object{
				newVolume("volume-1", v1.VolumeReleased, v1.PersistentVolumeReclaimDelete, map[string]string{annDynamicallyProvisioned: "foo.bar/baz"}),
				newVolume("volume-2", v1.VolumeReleased, v1.PersistentVolumeReclaimDelete, map[string]string{annDynamicallyProvisioned: "abc.def/ghi"}),
			},
			provisionerName: "foo.bar/baz",
			provisioner:     newTestProvisioner(),
			expectedVolumes: []v1.PersistentVolume{
				*newVolume("volume-2", v1.VolumeReleased, v1.PersistentVolumeReclaimDelete, map[string]string{annDynamicallyProvisioned: "abc.def/ghi"}),
			},
		},
		{
			name: "don't provision for claim-1 because it's already bound",
			objs: []runtime.Object{
				newClaim("claim-1", "uid-1-1", "class-1", "volume-1", nil),
			},
			provisionerName: "foo.bar/baz",
			provisioner:     newTestProvisioner(),
			expectedVolumes: []v1.PersistentVolume(nil),
		},
		{
			name: "don't provision for claim-1 because its class doesn't exist",
			objs: []runtime.Object{
				newClaim("claim-1", "uid-1-1", "class-1", "", nil),
			},
			provisionerName: "foo.bar/baz",
			provisioner:     newTestProvisioner(),
			expectedVolumes: []v1.PersistentVolume(nil),
		},
		{
			name: "don't delete volume-1 because it's still bound",
			objs: []runtime.Object{
				newVolume("volume-1", v1.VolumeBound, v1.PersistentVolumeReclaimDelete, map[string]string{annDynamicallyProvisioned: "foo.bar/baz"}),
			},
			provisionerName: "foo.bar/baz",
			provisioner:     newTestProvisioner(),
			expectedVolumes: []v1.PersistentVolume{
				*newVolume("volume-1", v1.VolumeBound, v1.PersistentVolumeReclaimDelete, map[string]string{annDynamicallyProvisioned: "foo.bar/baz"}),
			},
		},
		{
			name: "don't delete volume-1 because its reclaim policy is not delete",
			objs: []runtime.Object{
				newVolume("volume-1", v1.VolumeReleased, v1.PersistentVolumeReclaimRetain, map[string]string{annDynamicallyProvisioned: "foo.bar/baz"}),
			},
			provisionerName: "foo.bar/baz",
			provisioner:     newTestProvisioner(),
			expectedVolumes: []v1.PersistentVolume{
				*newVolume("volume-1", v1.VolumeReleased, v1.PersistentVolumeReclaimRetain, map[string]string{annDynamicallyProvisioned: "foo.bar/baz"}),
			},
		},
		{
			name: "provisioner fails to provision for claim-1: no pv is created",
			objs: []runtime.Object{
				newStorageClass("class-1", "foo.bar/baz"),
				newClaim("claim-1", "uid-1-1", "class-1", "", nil),
			},
			provisionerName: "foo.bar/baz",
			provisioner:     newBadTestProvisioner(),
			expectedVolumes: []v1.PersistentVolume(nil),
		},
		{
			name: "provisioner fails to delete volume-1: pv is not deleted",
			objs: []runtime.Object{
				newVolume("volume-1", v1.VolumeReleased, v1.PersistentVolumeReclaimDelete, map[string]string{annDynamicallyProvisioned: "foo.bar/baz"}),
			},
			provisionerName: "foo.bar/baz",
			provisioner:     newBadTestProvisioner(),
			expectedVolumes: []v1.PersistentVolume{
				*newVolume("volume-1", v1.VolumeReleased, v1.PersistentVolumeReclaimDelete, map[string]string{annDynamicallyProvisioned: "foo.bar/baz"}),
			},
		},
		{
			name: "try to provision for claim-1 but fail to save the pv object",
			objs: []runtime.Object{
				newStorageClass("class-1", "foo.bar/baz"),
				newClaim("claim-1", "uid-1-1", "class-1", "", nil),
			},
			provisionerName: "foo.bar/baz",
			provisioner:     newTestProvisioner(),
			verbs:           []string{"create"},
			reaction: func(action testclient.Action) (handled bool, ret runtime.Object, err error) {
				return true, nil, errors.New("fake error")
			},
			expectedVolumes: []v1.PersistentVolume(nil),
		},
		{
			name: "try to delete volume-1 but fail to delete the pv object",
			objs: []runtime.Object{
				newVolume("volume-1", v1.VolumeReleased, v1.PersistentVolumeReclaimDelete, map[string]string{annDynamicallyProvisioned: "foo.bar/baz"}),
			},
			provisionerName: "foo.bar/baz",
			provisioner:     newTestProvisioner(),
			verbs:           []string{"delete"},
			reaction: func(action testclient.Action) (handled bool, ret runtime.Object, err error) {
				return true, nil, errors.New("fake error")
			},
			expectedVolumes: []v1.PersistentVolume{
				*newVolume("volume-1", v1.VolumeReleased, v1.PersistentVolumeReclaimDelete, map[string]string{annDynamicallyProvisioned: "foo.bar/baz"}),
			},
		},
	}
	for _, test := range tests {
		client := fake.NewSimpleClientset(test.objs...)
		if len(test.verbs) != 0 {
			for _, v := range test.verbs {
				client.Fake.PrependReactor(v, "persistentvolumes", test.reaction)
			}
		}
		ctrl := newTestProvisionController(client, resyncPeriod, test.provisionerName, test.provisioner, "v1.5.0", false, failedRetryThreshold)
		stopCh := make(chan struct{})
		go ctrl.Run(stopCh)

		time.Sleep(2 * resyncPeriod)
		ctrl.runningOperations.Wait()

		pvList, _ := client.Core().PersistentVolumes().List(v1.ListOptions{})
		if !reflect.DeepEqual(test.expectedVolumes, pvList.Items) {
			t.Logf("test case: %s", test.name)
			t.Errorf("expected PVs:\n %v\n but got:\n %v\n", test.expectedVolumes, pvList.Items)
		}
		close(stopCh)
	}
}

func TestMultipleControllers(t *testing.T) {
	tests := []struct {
		name            string
		provisionerName string
		numControllers  int
		numClaims       int
		expectedCalls   int
	}{
		{
			name:            "call provision exactly once",
			provisionerName: "foo.bar/baz",
			numControllers:  5,
			numClaims:       1,
			expectedCalls:   1,
		},
	}
	for _, test := range tests {
		client := fake.NewSimpleClientset()

		// Create a reactor to reject Updates if object has already been modified,
		// like etcd.
		claimSource := fcache.NewFakePVCControllerSource()
		reactor := claimReactor{
			fake:        &fakev1core.FakeCoreV1{Fake: &client.Fake},
			claims:      make(map[string]*v1.PersistentVolumeClaim),
			lock:        sync.Mutex{},
			claimSource: claimSource,
		}
		reactor.claims["claim-1"] = newClaim("claim-1", "uid-1-1", "class-1", "", nil)
		client.PrependReactor("update", "persistentvolumeclaims", reactor.React)
		client.PrependReactor("get", "persistentvolumeclaims", reactor.React)

		// Create a fake watch so each controller can get ProvisioningSucceeded
		fakeWatch := watch.NewFakeWithChanSize(test.numControllers, false)
		client.PrependWatchReactor("events", testclient.DefaultWatchReactor(fakeWatch, nil))
		client.PrependReactor("create", "events", func(action testclient.Action) (bool, runtime.Object, error) {
			obj := action.(testclient.CreateAction).GetObject()
			for i := 0; i < test.numControllers; i++ {
				fakeWatch.Add(obj)
			}
			return true, obj, nil
		})

		provisioner := newTestProvisioner()
		ctrls := make([]*ProvisionController, test.numControllers)
		stopChs := make([]chan struct{}, test.numControllers)
		for i := 0; i < test.numControllers; i++ {
			ctrls[i] = NewProvisionController(client, 15*time.Second, test.provisionerName, provisioner, "v1.5.0", false, failedRetryThreshold, leaderelection.DefaultLeaseDuration, leaderelection.DefaultRenewDeadline, leaderelection.DefaultRetryPeriod, leaderelection.DefaultTermLimit)
			ctrls[i].createProvisionedPVInterval = 10 * time.Millisecond
			ctrls[i].claimSource = claimSource
			ctrls[i].claims.Add(newClaim("claim-1", "uid-1-1", "class-1", "", nil))
			ctrls[i].classes.Add(newStorageClass("class-1", "foo.bar/baz"))
			stopChs[i] = make(chan struct{})
		}

		for i := 0; i < test.numControllers; i++ {
			go ctrls[i].addClaim(newClaim("claim-1", "uid-1-1", "class-1", "", nil))
		}

		// Sleep for 3 election retry periods
		time.Sleep(3 * ctrls[0].retryPeriod)

		if test.expectedCalls != len(provisioner.provisionCalls) {
			t.Logf("test case: %s", test.name)
			t.Errorf("expected provision calls:\n %v\n but got:\n %v\n", test.expectedCalls, len(provisioner.provisionCalls))
		}

		for _, stopCh := range stopChs {
			close(stopCh)
		}
	}
}

func TestShouldProvision(t *testing.T) {
	tests := []struct {
		name            string
		provisionerName string
		class           *v1beta1.StorageClass
		claim           *v1.PersistentVolumeClaim
		expectedShould  bool
	}{
		{
			name:            "should provision",
			provisionerName: "foo.bar/baz",
			class:           newStorageClass("class-1", "foo.bar/baz"),
			claim:           newClaim("claim-1", "1-1", "class-1", "", nil),
			expectedShould:  true,
		},
		{
			name:            "claim already bound",
			provisionerName: "foo.bar/baz",
			class:           newStorageClass("class-1", "foo.bar/baz"),
			claim:           newClaim("claim-1", "1-1", "class-1", "foo", nil),
			expectedShould:  false,
		},
		{
			name:            "no such class",
			provisionerName: "foo.bar/baz",
			class:           newStorageClass("class-1", "foo.bar/baz"),
			claim:           newClaim("claim-1", "1-1", "class-2", "", nil),
			expectedShould:  false,
		},
		{
			name:            "not this provisioner's job",
			provisionerName: "foo.bar/baz",
			class:           newStorageClass("class-1", "abc.def/ghi"),
			claim:           newClaim("claim-1", "1-1", "class-1", "", nil),
			expectedShould:  false,
		},
		// Kubernetes 1.5 provisioning - annDynamicallyProvisioned is set
		// and only this annotation is evaluated
		{
			name:            "should provision 1.5",
			provisionerName: "foo.bar/baz",
			class:           newStorageClass("class-2", "abc.def/ghi"),
			claim: newClaim("claim-1", "1-1", "class-1", "",
				map[string]string{annDynamicallyProvisioned: "foo.bar/baz"}),
			expectedShould: true,
		},
		{
			name:            "unknown provisioner 1.5",
			provisionerName: "foo.bar/baz",
			class:           newStorageClass("class-1", "foo.bar/baz"),
			claim: newClaim("claim-1", "1-1", "class-1", "",
				map[string]string{annDynamicallyProvisioned: "abc.def/ghi"}),
			expectedShould: false,
		},
	}
	for _, test := range tests {
		client := fake.NewSimpleClientset(test.claim)
		provisioner := newTestProvisioner()
		ctrl := newTestProvisionController(client, resyncPeriod, test.provisionerName, provisioner, "v1.5.0", false, failedRetryThreshold)

		err := ctrl.classes.Add(test.class)
		if err != nil {
			t.Logf("test case: %s", test.name)
			t.Errorf("error adding class %v to cache: %v", test.class, err)
		}

		should := ctrl.shouldProvision(test.claim)
		if test.expectedShould != should {
			t.Logf("test case: %s", test.name)
			t.Errorf("expected should provision %v but got %v\n", test.expectedShould, should)
		}
	}
}

func TestShouldDelete(t *testing.T) {
	tests := []struct {
		name             string
		provisionerName  string
		volume           *v1.PersistentVolume
		serverGitVersion string
		expectedShould   bool
	}{
		{
			name:             "should delete",
			provisionerName:  "foo.bar/baz",
			volume:           newVolume("volume-1", v1.VolumeReleased, v1.PersistentVolumeReclaimDelete, map[string]string{annDynamicallyProvisioned: "foo.bar/baz"}),
			serverGitVersion: "v1.5.0",
			expectedShould:   true,
		},
		{
			name:             "1.4 and failed: should delete",
			provisionerName:  "foo.bar/baz",
			volume:           newVolume("volume-1", v1.VolumeFailed, v1.PersistentVolumeReclaimDelete, map[string]string{annDynamicallyProvisioned: "foo.bar/baz"}),
			serverGitVersion: "v1.4.0",
			expectedShould:   true,
		},
		{
			name:             "1.5 and failed: shouldn't delete",
			provisionerName:  "foo.bar/baz",
			volume:           newVolume("volume-1", v1.VolumeFailed, v1.PersistentVolumeReclaimDelete, map[string]string{annDynamicallyProvisioned: "foo.bar/baz"}),
			serverGitVersion: "v1.5.0",
			expectedShould:   false,
		},
		{
			name:             "volume still bound",
			provisionerName:  "foo.bar/baz",
			volume:           newVolume("volume-1", v1.VolumeBound, v1.PersistentVolumeReclaimDelete, map[string]string{annDynamicallyProvisioned: "foo.bar/baz"}),
			serverGitVersion: "v1.5.0",
			expectedShould:   false,
		},
		{
			name:             "non-delete reclaim policy",
			provisionerName:  "foo.bar/baz",
			volume:           newVolume("volume-1", v1.VolumeReleased, v1.PersistentVolumeReclaimRetain, map[string]string{annDynamicallyProvisioned: "foo.bar/baz"}),
			serverGitVersion: "v1.5.0",
			expectedShould:   false,
		},
		{
			name:             "not this provisioner's job",
			provisionerName:  "foo.bar/baz",
			volume:           newVolume("volume-1", v1.VolumeReleased, v1.PersistentVolumeReclaimDelete, map[string]string{annDynamicallyProvisioned: "abc.def/ghi"}),
			serverGitVersion: "v1.5.0",
			expectedShould:   false,
		},
	}
	for _, test := range tests {
		client := fake.NewSimpleClientset()
		provisioner := newTestProvisioner()
		ctrl := newTestProvisionController(client, resyncPeriod, test.provisionerName, provisioner, test.serverGitVersion, false, failedRetryThreshold)

		should := ctrl.shouldDelete(test.volume)
		if test.expectedShould != should {
			t.Logf("test case: %s", test.name)
			t.Errorf("expected should delete %v but got %v\n", test.expectedShould, should)
		}
	}
}

func TestIsOnlyRecordUpdate(t *testing.T) {
	tests := []struct {
		name       string
		old        *v1.PersistentVolumeClaim
		new        *v1.PersistentVolumeClaim
		expectedIs bool
	}{
		{
			name:       "is only record update",
			old:        newClaim("claim-1", "1-1", "class-1", "", map[string]string{rl.LeaderElectionRecordAnnotationKey: "a"}),
			new:        newClaim("claim-1", "1-1", "class-1", "", map[string]string{rl.LeaderElectionRecordAnnotationKey: "b"}),
			expectedIs: true,
		},
		{
			name:       "is seen as only record update, stayed exactly the same",
			old:        newClaim("claim-1", "1-1", "class-1", "", map[string]string{rl.LeaderElectionRecordAnnotationKey: "a"}),
			new:        newClaim("claim-1", "1-1", "class-1", "", map[string]string{rl.LeaderElectionRecordAnnotationKey: "a"}),
			expectedIs: true,
		},
		{
			name:       "isn't only record update, class changed as well",
			old:        newClaim("claim-1", "1-1", "class-1", "", map[string]string{rl.LeaderElectionRecordAnnotationKey: "a"}),
			new:        newClaim("claim-1", "1-1", "class-2", "", map[string]string{rl.LeaderElectionRecordAnnotationKey: "b"}),
			expectedIs: false,
		},
		{
			name:       "isn't only record update, only class changed",
			old:        newClaim("claim-1", "1-1", "class-1", "", map[string]string{rl.LeaderElectionRecordAnnotationKey: "a"}),
			new:        newClaim("claim-1", "1-1", "class-2", "", map[string]string{rl.LeaderElectionRecordAnnotationKey: "a"}),
			expectedIs: false,
		},
	}
	for _, test := range tests {
		client := fake.NewSimpleClientset()
		provisioner := newTestProvisioner()
		ctrl := newTestProvisionController(client, resyncPeriod, "foo.bar/baz", provisioner, "v1.5.0", false, failedRetryThreshold)

		is, _ := ctrl.isOnlyRecordUpdate(test.old, test.new)
		if test.expectedIs != is {
			t.Logf("test case: %s", test.name)
			t.Errorf("expected is only record update %v but got %v\n", test.expectedIs, is)
		}
	}
}

func newTestProvisionController(
	client kubernetes.Interface,
	resyncPeriod time.Duration,
	provisionerName string,
	provisioner Provisioner,
	serverGitVersion string,
	exponentialBackOffOnError bool,
	failedRetryThreshold int,
) *ProvisionController {
	ctrl := NewProvisionController(client, resyncPeriod, provisionerName, provisioner, serverGitVersion, exponentialBackOffOnError, failedRetryThreshold, 2*resyncPeriod, resyncPeriod, resyncPeriod/2, 2*resyncPeriod)
	ctrl.createProvisionedPVInterval = 10 * time.Millisecond
	return ctrl
}

func newStorageClass(name, provisioner string) *v1beta1.StorageClass {
	return &v1beta1.StorageClass{
		ObjectMeta: v1.ObjectMeta{
			Name: name,
		},
		Provisioner: provisioner,
	}
}

func newClaim(name, claimUID, provisioner, volumeName string, annotations map[string]string) *v1.PersistentVolumeClaim {
	claim := &v1.PersistentVolumeClaim{
		ObjectMeta: v1.ObjectMeta{
			Name:            name,
			Namespace:       v1.NamespaceDefault,
			UID:             types.UID(claimUID),
			ResourceVersion: "0",
			Annotations:     map[string]string{annClass: provisioner},
			SelfLink:        testapi.Default.SelfLink("pvc", ""),
		},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce, v1.ReadOnlyMany},
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceName(v1.ResourceStorage): resource.MustParse("1Mi"),
				},
			},
			VolumeName: volumeName,
		},
		Status: v1.PersistentVolumeClaimStatus{
			Phase: v1.ClaimPending,
		},
	}
	for k, v := range annotations {
		claim.Annotations[k] = v
	}
	return claim
}

func newVolume(name string, phase v1.PersistentVolumePhase, policy v1.PersistentVolumeReclaimPolicy, annotations map[string]string) *v1.PersistentVolume {
	pv := &v1.PersistentVolume{
		ObjectMeta: v1.ObjectMeta{
			Name:        name,
			Annotations: annotations,
		},
		Spec: v1.PersistentVolumeSpec{
			PersistentVolumeReclaimPolicy: policy,
			AccessModes:                   []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce, v1.ReadOnlyMany},
			Capacity: v1.ResourceList{
				v1.ResourceName(v1.ResourceStorage): resource.MustParse("1Mi"),
			},
			PersistentVolumeSource: v1.PersistentVolumeSource{
				NFS: &v1.NFSVolumeSource{
					Server:   "foo",
					Path:     "bar",
					ReadOnly: false,
				},
			},
		},
		Status: v1.PersistentVolumeStatus{
			Phase: phase,
		},
	}

	return pv
}

// newProvisionedVolume returns the volume the test controller should provision for the
// given claim with the given class
func newProvisionedVolume(storageClass *v1beta1.StorageClass, claim *v1.PersistentVolumeClaim) *v1.PersistentVolume {
	// pv.Spec MUST be set to match requirements in claim.Spec, especially access mode and PV size. The provisioned volume size MUST NOT be smaller than size requested in the claim, however it MAY be larger.
	options := VolumeOptions{
		PersistentVolumeReclaimPolicy: v1.PersistentVolumeReclaimDelete,
		PVName:     "pvc-" + string(claim.ObjectMeta.UID),
		PVC:        claim,
		Parameters: storageClass.Parameters,
	}
	volume, _ := newTestProvisioner().Provision(options)

	// pv.Spec.ClaimRef MUST point to the claim that led to its creation (including the claim UID).
	volume.Spec.ClaimRef, _ = v1.GetReference(claim)

	// pv.Annotations["pv.kubernetes.io/provisioned-by"] MUST be set to name of the external provisioner. This provisioner will be used to delete the volume.
	// pv.Annotations["volume.beta.kubernetes.io/storage-class"] MUST be set to name of the storage class requested by the claim.
	volume.Annotations = map[string]string{annDynamicallyProvisioned: storageClass.Provisioner, annClass: storageClass.Name}

	// TODO implement options.ProvisionerSelector parsing
	// pv.Labels MUST be set to match claim.spec.selector. The provisioner MAY add additional labels.

	return volume
}

func newTestProvisioner() *testProvisioner {
	return &testProvisioner{make(chan bool, 16)}
}

type testProvisioner struct {
	provisionCalls chan bool
}

var _ Provisioner = &testProvisioner{}

func (p *testProvisioner) Provision(options VolumeOptions) (*v1.PersistentVolume, error) {
	p.provisionCalls <- true

	// Sleep to simulate work done by Provision...for long enough that
	// TestMultipleControllers will consistently fail with lock disabled. If
	// Provision happens too fast, the first controller creates the PV too soon
	// and the next controllers won't call Provision even though they're clearly
	// racing when there's no lock
	time.Sleep(50 * time.Millisecond)

	pv := &v1.PersistentVolume{
		ObjectMeta: v1.ObjectMeta{
			Name: options.PVName,
		},
		Spec: v1.PersistentVolumeSpec{
			PersistentVolumeReclaimPolicy: options.PersistentVolumeReclaimPolicy,
			AccessModes:                   options.PVC.Spec.AccessModes,
			Capacity: v1.ResourceList{
				v1.ResourceName(v1.ResourceStorage): options.PVC.Spec.Resources.Requests[v1.ResourceName(v1.ResourceStorage)],
			},
			PersistentVolumeSource: v1.PersistentVolumeSource{
				NFS: &v1.NFSVolumeSource{
					Server:   "foo",
					Path:     "bar",
					ReadOnly: false,
				},
			},
		},
	}

	return pv, nil
}

func (p *testProvisioner) Delete(volume *v1.PersistentVolume) error {
	return nil
}

func newBadTestProvisioner() Provisioner {
	return &badTestProvisioner{}
}

type badTestProvisioner struct {
}

var _ Provisioner = &badTestProvisioner{}

func (p *badTestProvisioner) Provision(options VolumeOptions) (*v1.PersistentVolume, error) {
	return nil, errors.New("fake error")
}

func (p *badTestProvisioner) Delete(volume *v1.PersistentVolume) error {
	return errors.New("fake error")
}

type claimReactor struct {
	fake        *fakev1core.FakeCoreV1
	claims      map[string]*v1.PersistentVolumeClaim
	lock        sync.Mutex
	claimSource *fcache.FakePVCControllerSource
}

func (r *claimReactor) React(action testclient.Action) (handled bool, ret runtime.Object, err error) {
	r.lock.Lock()
	defer r.lock.Unlock()
	switch {
	case action.Matches("update", "persistentvolumeclaims"):
		obj := action.(testclient.UpdateAction).GetObject()

		claim := obj.(*v1.PersistentVolumeClaim)

		// Check and bump object version
		storedClaim, found := r.claims[claim.Name]
		if found {
			storedVer, _ := strconv.Atoi(storedClaim.ResourceVersion)
			requestedVer, _ := strconv.Atoi(claim.ResourceVersion)
			if storedVer != requestedVer {
				return true, obj, errors.New("VersionError")
			}
			claim.ResourceVersion = strconv.Itoa(storedVer + 1)
		} else {
			return true, nil, fmt.Errorf("Cannot update claim %s: claim not found", claim.Name)
		}

		r.claims[claim.Name] = claim
		r.claimSource.Modify(claim)
		return true, claim, nil
	case action.Matches("get", "persistentvolumeclaims"):
		name := action.(testclient.GetAction).GetName()
		claim, found := r.claims[name]
		if found {
			clone, err := conversion.NewCloner().DeepCopy(claim)
			if err != nil {
				return true, nil, fmt.Errorf("Error cloning claim %s: %v", name, err)
			}
			claimClone, ok := clone.(*v1.PersistentVolumeClaim)
			if !ok {
				return true, nil, fmt.Errorf("Error casting clone of claim %s: %v", name, claimClone)
			}
			return true, claimClone, nil
		}
		return true, nil, fmt.Errorf("Cannot find claim %s", name)
	}

	return false, nil, nil
}
