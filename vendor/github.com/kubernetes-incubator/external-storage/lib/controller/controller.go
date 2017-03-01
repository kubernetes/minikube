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
	"fmt"
	"os/exec"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/kubernetes-incubator/external-storage/lib/leaderelection"
	rl "github.com/kubernetes-incubator/external-storage/lib/leaderelection/resourcelock"
	"k8s.io/client-go/kubernetes"
	core_v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/storage/v1beta1"
	"k8s.io/client-go/pkg/fields"
	"k8s.io/client-go/pkg/runtime"
	"k8s.io/client-go/pkg/types"
	"k8s.io/client-go/pkg/util/uuid"
	"k8s.io/client-go/pkg/version"
	"k8s.io/client-go/pkg/watch"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/kubernetes/pkg/util/goroutinemap"
)

// annClass annotation represents the storage class associated with a resource:
// - in PersistentVolumeClaim it represents required class to match.
//   Only PersistentVolumes with the same class (i.e. annotation with the same
//   value) can be bound to the claim. In case no such volume exists, the
//   controller will provision a new one using StorageClass instance with
//   the same name as the annotation value.
// - in PersistentVolume it represents storage class to which the persistent
//   volume belongs.
const annClass = "volume.beta.kubernetes.io/storage-class"

// This annotation is added to a PV that has been dynamically provisioned by
// Kubernetes. Its value is name of volume plugin that created the volume.
// It serves both user (to show where a PV comes from) and Kubernetes (to
// recognize dynamically provisioned PVs in its decisions).
const annDynamicallyProvisioned = "pv.kubernetes.io/provisioned-by"

const annStorageProvisioner = "volume.beta.kubernetes.io/storage-provisioner"

// Number of retries when we create a PV object for a provisioned volume.
const createProvisionedPVRetryCount = 5

// Interval between retries when we create a PV object for a provisioned volume.
const createProvisionedPVInterval = 10 * time.Second

// ProvisionController is a controller that provisions PersistentVolumes for
// PersistentVolumeClaims.
type ProvisionController struct {
	client kubernetes.Interface

	// How often the controller relists PVCs, PVs, & storage classes. OnUpdate
	// will be called even if nothing has changed, meaning failed operations may
	// be retried on a PVC/PV every resyncPeriod regardless of whether it changed
	resyncPeriod time.Duration

	// The name of the provisioner for which this controller dynamically
	// provisions volumes. The value of annDynamicallyProvisioned and
	// annStorageProvisioner to set & watch for, respectively
	provisionerName string

	// The provisioner the controller will use to provision and delete volumes.
	// Presumably this implementer of Provisioner carries its own
	// volume-specific options and such that it needs in order to provision
	// volumes.
	provisioner Provisioner

	// Whether we are running in a 1.4 cluster before out-of-tree dynamic
	// provisioning is officially supported
	is1dot4 bool

	claimSource      cache.ListerWatcher
	claimController  *cache.Controller
	volumeSource     cache.ListerWatcher
	volumeController *cache.Controller
	classSource      cache.ListerWatcher
	classReflector   *cache.Reflector

	volumes cache.Store
	claims  cache.Store
	classes cache.Store

	eventRecorder record.EventRecorder

	// Map of scheduled/running operations.
	runningOperations goroutinemap.GoRoutineMap

	// Number of retries when we create a PV object for a provisioned volume.
	createProvisionedPVRetryCount int

	// Interval between retries when we create a PV object for a provisioned volume.
	createProvisionedPVInterval time.Duration

	// Identity of this controller, generated at creation time.
	identity types.UID

	// Parameters of LeaderElectionConfig: set to defaults except in tests
	leaseDuration, renewDeadline, retryPeriod, termLimit time.Duration

	// Map of claim UID to LeaderElector: for checking if this controller
	// is the leader of a given claim
	leaderElectors map[types.UID]*leaderelection.LeaderElector

	mapMutex *sync.Mutex

	// threshold for max number of retries on failure of provisioner
	failedRetryThreshold int

	// map of failed claims
	failedClaimsStats map[types.UID]int

	failedClaimsStatsMutex *sync.Mutex
}

// NewProvisionController creates a new provision controller
func NewProvisionController(
	client kubernetes.Interface,
	resyncPeriod time.Duration,
	provisionerName string,
	provisioner Provisioner,
	serverGitVersion string,
	exponentialBackOffOnError bool,
	failedRetryThreshold int,
	leaseDuration time.Duration,
	renewDeadline time.Duration,
	retryPeriod time.Duration,
	termLimit time.Duration,
) *ProvisionController {
	identity := uuid.NewUUID()

	broadcaster := record.NewBroadcaster()
	broadcaster.StartRecordingToSink(&core_v1.EventSinkImpl{Interface: client.Core().Events(v1.NamespaceAll)})
	var eventRecorder record.EventRecorder
	out, err := exec.Command("hostname").Output()
	if err != nil {
		eventRecorder = broadcaster.NewRecorder(v1.EventSource{Component: fmt.Sprintf("%s %s", provisionerName, string(identity))})
	} else {
		eventRecorder = broadcaster.NewRecorder(v1.EventSource{Component: fmt.Sprintf("%s %s %s", provisionerName, strings.TrimSpace(string(out)), string(identity))})
	}

	gitVersion := version.MustParse(serverGitVersion)
	gitVersion1dot5 := version.MustParse("1.5.0")
	is1dot4 := gitVersion.LT(gitVersion1dot5)

	controller := &ProvisionController{
		client:                        client,
		resyncPeriod:                  resyncPeriod,
		provisionerName:               provisionerName,
		provisioner:                   provisioner,
		is1dot4:                       is1dot4,
		eventRecorder:                 eventRecorder,
		runningOperations:             goroutinemap.NewGoRoutineMap(exponentialBackOffOnError),
		createProvisionedPVRetryCount: createProvisionedPVRetryCount,
		createProvisionedPVInterval:   createProvisionedPVInterval,
		identity:                      identity,
		leaseDuration:                 leaseDuration,
		renewDeadline:                 renewDeadline,
		retryPeriod:                   retryPeriod,
		termLimit:                     termLimit,
		leaderElectors:                make(map[types.UID]*leaderelection.LeaderElector),
		mapMutex:                      &sync.Mutex{},
		failedClaimsStats:             make(map[types.UID]int),
		failedRetryThreshold:          failedRetryThreshold,
		failedClaimsStatsMutex:        &sync.Mutex{},
	}

	controller.claimSource = &cache.ListWatch{
		ListFunc: func(options api.ListOptions) (runtime.Object, error) {
			var out v1.ListOptions
			v1.Convert_api_ListOptions_To_v1_ListOptions(&options, &out, nil)
			return client.Core().PersistentVolumeClaims(v1.NamespaceAll).List(out)
		},
		WatchFunc: func(options api.ListOptions) (watch.Interface, error) {
			var out v1.ListOptions
			v1.Convert_api_ListOptions_To_v1_ListOptions(&options, &out, nil)
			return client.Core().PersistentVolumeClaims(v1.NamespaceAll).Watch(out)
		},
	}
	controller.claims, controller.claimController = cache.NewInformer(
		controller.claimSource,
		&v1.PersistentVolumeClaim{},
		resyncPeriod,
		cache.ResourceEventHandlerFuncs{
			AddFunc:    controller.addClaim,
			UpdateFunc: controller.updateClaim,
			DeleteFunc: nil,
		},
	)

	controller.volumeSource = &cache.ListWatch{
		ListFunc: func(options api.ListOptions) (runtime.Object, error) {
			var out v1.ListOptions
			v1.Convert_api_ListOptions_To_v1_ListOptions(&options, &out, nil)
			return client.Core().PersistentVolumes().List(out)
		},
		WatchFunc: func(options api.ListOptions) (watch.Interface, error) {
			var out v1.ListOptions
			v1.Convert_api_ListOptions_To_v1_ListOptions(&options, &out, nil)

			return client.Core().PersistentVolumes().Watch(out)
		},
	}
	controller.volumes, controller.volumeController = cache.NewInformer(
		controller.volumeSource,
		&v1.PersistentVolume{},
		resyncPeriod,
		cache.ResourceEventHandlerFuncs{
			AddFunc:    nil,
			UpdateFunc: controller.updateVolume,
			DeleteFunc: nil,
		},
	)

	controller.classSource = &cache.ListWatch{
		ListFunc: func(options api.ListOptions) (runtime.Object, error) {
			var out v1.ListOptions
			v1.Convert_api_ListOptions_To_v1_ListOptions(&options, &out, nil)
			return client.Storage().StorageClasses().List(out)
		},
		WatchFunc: func(options api.ListOptions) (watch.Interface, error) {
			var out v1.ListOptions
			v1.Convert_api_ListOptions_To_v1_ListOptions(&options, &out, nil)
			return client.Storage().StorageClasses().Watch(out)
		},
	}
	controller.classes = cache.NewStore(cache.DeletionHandlingMetaNamespaceKeyFunc)
	controller.classReflector = cache.NewReflector(
		controller.classSource,
		&v1beta1.StorageClass{},
		controller.classes,
		resyncPeriod,
	)

	return controller
}

// Run starts all of this controller's control loops
func (ctrl *ProvisionController) Run(stopCh <-chan struct{}) {
	glog.Infof("Starting provisioner controller %s!", string(ctrl.identity))
	go ctrl.claimController.Run(stopCh)
	go ctrl.volumeController.Run(stopCh)
	go ctrl.classReflector.RunUntil(stopCh)
	<-stopCh
}

// On add claim, check if the added claim should have a volume provisioned for
// it and provision one if so.
func (ctrl *ProvisionController) addClaim(obj interface{}) {
	claim, ok := obj.(*v1.PersistentVolumeClaim)
	if !ok {
		glog.Errorf("Expected PersistentVolumeClaim but addClaim received %+v", obj)
		return
	}

	if ctrl.shouldProvision(claim) {
		ctrl.mapMutex.Lock()
		le, ok := ctrl.leaderElectors[claim.UID]
		ctrl.mapMutex.Unlock()
		if ok && le.IsLeader() {
			opName := fmt.Sprintf("provision-%s[%s]", claimToClaimKey(claim), string(claim.UID))
			ctrl.scheduleOperation(opName, func() error {
				err := ctrl.provisionClaimOperation(claim)
				ctrl.updateStats(claim, err)
				return err
			})
		} else {
			opName := fmt.Sprintf("lock-provision-%s[%s]", claimToClaimKey(claim), string(claim.UID))
			ctrl.scheduleOperation(opName, func() error {
				ctrl.lockProvisionClaimOperation(claim)
				return nil
			})
		}
	}
}

// On update claim, pass the new claim to addClaim. Updates occur at least every
// resyncPeriod.
func (ctrl *ProvisionController) updateClaim(oldObj, newObj interface{}) {
	// If they are exactly the same it must be a forced resync (every
	// resyncPeriod).
	if reflect.DeepEqual(oldObj, newObj) {
		ctrl.addClaim(newObj)
		return
	}

	// If not a forced resync, we filter out the frequent leader election record
	// annotation changes by checking if the only update is in the annotation
	oldClaim, ok := oldObj.(*v1.PersistentVolumeClaim)
	if !ok {
		glog.Errorf("Expected PersistentVolumeClaim but handler received %#v", oldObj)
		return
	}
	newClaim, ok := newObj.(*v1.PersistentVolumeClaim)
	if !ok {
		glog.Errorf("Expected PersistentVolumeClaim but handler received %#v", newObj)
		return
	}

	skipAddClaim, err := ctrl.isOnlyRecordUpdate(oldClaim, newClaim)
	if err != nil {
		glog.Errorf("Error checking if only record was updated in claim: %v", oldClaim)
		return
	}

	if !skipAddClaim {
		ctrl.addClaim(newObj)
	}
}

// On update volume, check if the updated volume should be deleted and delete if
// so. Updates occur at least every resyncPeriod.
func (ctrl *ProvisionController) updateVolume(oldObj, newObj interface{}) {
	volume, ok := newObj.(*v1.PersistentVolume)
	if !ok {
		glog.Errorf("Expected PersistentVolume but handler received %#v", newObj)
		return
	}

	if ctrl.shouldDelete(volume) {
		opName := fmt.Sprintf("delete-%s[%s]", volume.Name, string(volume.UID))
		ctrl.scheduleOperation(opName, func() error {
			return ctrl.deleteVolumeOperation(volume)
		})
	}
}

// isOnlyRecordUpdate checks if the only update between the old & new claim is
// the leader election record annotation.
func (ctrl *ProvisionController) isOnlyRecordUpdate(oldClaim, newClaim *v1.PersistentVolumeClaim) (bool, error) {
	old, err := ctrl.removeRecord(oldClaim)
	if err != nil {
		return false, err
	}
	new, err := ctrl.removeRecord(newClaim)
	if err != nil {
		return false, err
	}
	return reflect.DeepEqual(old, new), nil
}

// removeRecord returns a claim with its leader election record annotation and
// ResourceVersion set blank
func (ctrl *ProvisionController) removeRecord(claim *v1.PersistentVolumeClaim) (*v1.PersistentVolumeClaim, error) {
	clone, err := api.Scheme.DeepCopy(claim)
	if err != nil {
		return nil, fmt.Errorf("Error cloning claim: %v", err)
	}
	claimClone, ok := clone.(*v1.PersistentVolumeClaim)
	if !ok {
		return nil, fmt.Errorf("Unexpected claim cast error: %v", claimClone)
	}

	if claimClone.Annotations == nil {
		claimClone.Annotations = make(map[string]string)
	}
	claimClone.Annotations[rl.LeaderElectionRecordAnnotationKey] = ""

	claimClone.ResourceVersion = ""

	return claimClone, nil
}

func (ctrl *ProvisionController) shouldProvision(claim *v1.PersistentVolumeClaim) bool {
	if claim.Spec.VolumeName != "" {
		return false
	}

	ctrl.failedClaimsStatsMutex.Lock()
	if failureCount, exists := ctrl.failedClaimsStats[claim.UID]; exists == true {

		if failureCount >= ctrl.failedRetryThreshold {
			glog.Errorf("Exceeded failedRetryThreshold threshold: %d, for claim %q, provisioner will not attempt retries for this claim", ctrl.failedRetryThreshold, claimToClaimKey(claim))
			ctrl.failedClaimsStatsMutex.Unlock()
			return false
		}
	}
	ctrl.failedClaimsStatsMutex.Unlock()

	// Kubernetes 1.5 provisioning with annDynamicallyProvisioned
	if provisioner, found := claim.Annotations[annDynamicallyProvisioned]; found {
		if provisioner == ctrl.provisionerName {
			return true
		}
		return false
	}

	// Kubernetes 1.4 provisioning, evaluating class.Provisioner
	claimClass := getClaimClass(claim)
	_, err := ctrl.getStorageClass(claimClass)
	if err != nil {
		glog.Errorf("Claim %q: %v", claimToClaimKey(claim), err)
		return false
	}
	return true
}

func (ctrl *ProvisionController) shouldDelete(volume *v1.PersistentVolume) bool {
	// In 1.5+ we delete only if the volume is in state Released. In 1.4 we must
	// delete if the volume is in state Failed too.
	if !ctrl.is1dot4 {
		if volume.Status.Phase != v1.VolumeReleased {
			return false
		}
	} else {
		if volume.Status.Phase != v1.VolumeReleased && volume.Status.Phase != v1.VolumeFailed {
			return false
		}
	}

	if volume.Spec.PersistentVolumeReclaimPolicy != v1.PersistentVolumeReclaimDelete {
		return false
	}

	if !hasAnnotation(volume.ObjectMeta, annDynamicallyProvisioned) {
		return false
	}

	if ann := volume.Annotations[annDynamicallyProvisioned]; ann != ctrl.provisionerName {
		return false
	}

	return true
}

// lockProvisionClaimOperation wraps provisionClaimOperation. In case other
// controllers are serving the same claims, to prevent them all from creating
// volumes for a claim & racing to submit their PV, each controller creates a
// LeaderElector to instead race for the leadership (lock), where only the
// leader is tasked with provisioning & may try to do so
func (ctrl *ProvisionController) lockProvisionClaimOperation(claim *v1.PersistentVolumeClaim) {
	stoppedLeading := false
	rl := rl.ProvisionPVCLock{
		PVCMeta: claim.ObjectMeta,
		Client:  ctrl.client,
		LockConfig: rl.ResourceLockConfig{
			Identity:      string(ctrl.identity),
			EventRecorder: ctrl.eventRecorder,
		},
	}
	le, err := leaderelection.NewLeaderElector(leaderelection.LeaderElectionConfig{
		Lock:          &rl,
		LeaseDuration: ctrl.leaseDuration,
		RenewDeadline: ctrl.renewDeadline,
		RetryPeriod:   ctrl.retryPeriod,
		TermLimit:     ctrl.termLimit,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(_ <-chan struct{}) {
				opName := fmt.Sprintf("provision-%s[%s]", claimToClaimKey(claim), string(claim.UID))
				ctrl.scheduleOperation(opName, func() error {
					err := ctrl.provisionClaimOperation(claim)
					ctrl.updateStats(claim, err)
					return err

				})
			},
			OnStoppedLeading: func() {
				stoppedLeading = true
			},
		},
	})
	if err != nil {
		glog.Errorf("Error creating LeaderElector, can't provision for claim %q: %v", claimToClaimKey(claim), err)
		return
	}

	ctrl.mapMutex.Lock()
	ctrl.leaderElectors[claim.UID] = le
	ctrl.mapMutex.Unlock()

	// To determine when to stop trying to acquire/renew the lock, watch for
	// provisioning success/failure. (The leader could get the result of its
	// operation but it has to watch anyway)
	stopCh := make(chan struct{})
	successCh, err := ctrl.watchProvisioning(claim, stopCh)
	if err != nil {
		glog.Errorf("Error watching for provisioning success, can't provision for claim %q: %v", claimToClaimKey(claim), err)
	}

	le.Run(successCh)

	close(stopCh)

	// If we were the leader and stopped, give others a chance to acquire
	// (whether they exist & want to or not). Else, there must have been a
	// success so just proceed.
	if stoppedLeading {
		time.Sleep(ctrl.leaseDuration + ctrl.retryPeriod)
	}

	ctrl.mapMutex.Lock()
	delete(ctrl.leaderElectors, claim.UID)
	ctrl.mapMutex.Unlock()
}

func (ctrl *ProvisionController) updateStats(claim *v1.PersistentVolumeClaim, err error) {
	ctrl.failedClaimsStatsMutex.Lock()
	defer ctrl.failedClaimsStatsMutex.Unlock()
	if err != nil {
		if failureRetryCount, exists := ctrl.failedClaimsStats[claim.UID]; exists == true {
			failureRetryCount = failureRetryCount + 1
			ctrl.failedClaimsStats[claim.UID] = failureRetryCount
		} else {
			ctrl.failedClaimsStats[claim.UID] = 1
		}

	} else {
		delete(ctrl.failedClaimsStats, claim.UID)
	}
}

// provisionClaimOperation attempts to provision a volume for the given claim.
// Returns an error for use by goroutinemap when expbackoff is enabled: if nil,
// the operation is deleted, else the operation may be retried with expbackoff.
func (ctrl *ProvisionController) provisionClaimOperation(claim *v1.PersistentVolumeClaim) error {
	// Most code here is identical to that found in controller.go of kube's PV controller...
	claimClass := getClaimClass(claim)
	glog.V(4).Infof("provisionClaimOperation [%s] started, class: %q", claimToClaimKey(claim), claimClass)

	//  A previous doProvisionClaim may just have finished while we were waiting for
	//  the locks. Check that PV (with deterministic name) hasn't been provisioned
	//  yet.
	pvName := ctrl.getProvisionedVolumeNameForClaim(claim)
	volume, err := ctrl.client.Core().PersistentVolumes().Get(pvName)
	if err == nil && volume != nil {
		// Volume has been already provisioned, nothing to do.
		glog.V(4).Infof("provisionClaimOperation [%s]: volume already exists, skipping", claimToClaimKey(claim))
		return nil
	}

	// Prepare a claimRef to the claim early (to fail before a volume is
	// provisioned)
	claimRef, err := v1.GetReference(claim)
	if err != nil {
		glog.Errorf("Unexpected error getting claim reference to claim %q: %v", claimToClaimKey(claim), err)
		return nil
	}

	storageClass, err := ctrl.getStorageClass(claimClass)
	if err != nil {
		glog.Errorf("claim %v: %v", claimToClaimKey(claim), err)
		return nil
	}

	options := VolumeOptions{
		// TODO SHOULD be set to `Delete` unless user manually congiures other reclaim policy.
		PersistentVolumeReclaimPolicy: v1.PersistentVolumeReclaimDelete,
		PVName:     pvName,
		PVC:        claim,
		Parameters: storageClass.Parameters,
	}

	volume, err = ctrl.provisioner.Provision(options)
	if err != nil {
		strerr := fmt.Sprintf("Failed to provision volume with StorageClass %q: %v", storageClass.Name, err)
		glog.Errorf("Failed to provision volume for claim %q with StorageClass %q: %v", claimToClaimKey(claim), storageClass.Name, err)
		ctrl.eventRecorder.Event(claim, v1.EventTypeWarning, "ProvisioningFailed", strerr)
		return err
	}

	glog.Infof("volume %q for claim %q created", volume.Name, claimToClaimKey(claim))

	// Set ClaimRef and the PV controller will bind and set annBoundByController for us
	volume.Spec.ClaimRef = claimRef

	setAnnotation(&volume.ObjectMeta, annDynamicallyProvisioned, ctrl.provisionerName)
	setAnnotation(&volume.ObjectMeta, annClass, claimClass)

	// Try to create the PV object several times
	for i := 0; i < ctrl.createProvisionedPVRetryCount; i++ {
		glog.V(4).Infof("provisionClaimOperation [%s]: trying to save volume %s", claimToClaimKey(claim), volume.Name)
		if _, err = ctrl.client.Core().PersistentVolumes().Create(volume); err == nil {
			// Save succeeded.
			glog.Infof("volume %q for claim %q saved", volume.Name, claimToClaimKey(claim))
			break
		}
		// Save failed, try again after a while.
		glog.Infof("failed to save volume %q for claim %q: %v", volume.Name, claimToClaimKey(claim), err)
		time.Sleep(ctrl.createProvisionedPVInterval)
	}

	if err != nil {
		// Save failed. Now we have a storage asset outside of Kubernetes,
		// but we don't have appropriate PV object for it.
		// Emit some event here and try to delete the storage asset several
		// times.
		strerr := fmt.Sprintf("Error creating provisioned PV object for claim %s: %v. Deleting the volume.", claimToClaimKey(claim), err)
		glog.Error(strerr)
		ctrl.eventRecorder.Event(claim, v1.EventTypeWarning, "ProvisioningFailed", strerr)

		for i := 0; i < ctrl.createProvisionedPVRetryCount; i++ {
			if err = ctrl.provisioner.Delete(volume); err == nil {
				// Delete succeeded
				glog.V(4).Infof("provisionClaimOperation [%s]: cleaning volume %s succeeded", claimToClaimKey(claim), volume.Name)
				break
			}
			// Delete failed, try again after a while.
			glog.Infof("failed to delete volume %q: %v", volume.Name, err)
			time.Sleep(ctrl.createProvisionedPVInterval)
		}

		if err != nil {
			// Delete failed several times. There is an orphaned volume and there
			// is nothing we can do about it.
			strerr := fmt.Sprintf("Error cleaning provisioned volume for claim %s: %v. Please delete manually.", claimToClaimKey(claim), err)
			glog.Error(strerr)
			ctrl.eventRecorder.Event(claim, v1.EventTypeWarning, "ProvisioningCleanupFailed", strerr)
		}
	} else {
		glog.Infof("volume %q provisioned for claim %q", volume.Name, claimToClaimKey(claim))
		msg := fmt.Sprintf("Successfully provisioned volume %s", volume.Name)
		ctrl.eventRecorder.Event(claim, v1.EventTypeNormal, "ProvisioningSucceeded", msg)
	}

	return nil
}

// watchProvisioning returns a channel to which it sends the results of all
// provisioning attempts for the given claim. The PVC being modified to no
// longer need provisioning is considered a success.
func (ctrl *ProvisionController) watchProvisioning(claim *v1.PersistentVolumeClaim, stopChannel chan struct{}) (<-chan bool, error) {
	stopWatchPVC := make(chan struct{})
	pvcCh, err := ctrl.watchPVC(claim, stopWatchPVC)
	if err != nil {
		glog.Infof("cannot start watcher for PVC %s/%s: %v", claim.Namespace, claim.Name, err)
		return nil, err
	}

	successCh := make(chan bool, 0)

	go func() {
		defer close(stopWatchPVC)
		defer close(successCh)

		for {
			select {
			case _ = <-stopChannel:
				return

			case event := <-pvcCh:
				switch event.Object.(type) {
				case *v1.PersistentVolumeClaim:
					// PVC changed
					claim := event.Object.(*v1.PersistentVolumeClaim)
					glog.V(4).Infof("claim update received: %s %s/%s %s", event.Type, claim.Namespace, claim.Name, claim.Status.Phase)
					switch event.Type {
					case watch.Added, watch.Modified:
						if claim.Spec.VolumeName != "" {
							successCh <- true
						} else if !ctrl.shouldProvision(claim) {
							glog.Infof("claim %s/%s was modified to not ask for this provisioner", claim.Namespace, claim.Name)
							successCh <- true
						}

					case watch.Deleted:
						glog.Infof("claim %s/%s was deleted", claim.Namespace, claim.Name)
						successCh <- true

					case watch.Error:
						glog.Infof("claim %s/%s watcher failed", claim.Namespace, claim.Name)
						successCh <- true
					default:
					}
				case *v1.Event:
					// Event received
					claimEvent := event.Object.(*v1.Event)
					glog.V(4).Infof("claim event received: %s %s/%s %s/%s %s", event.Type, claimEvent.Namespace, claimEvent.Name, claimEvent.InvolvedObject.Namespace, claimEvent.InvolvedObject.Name, claimEvent.Reason)
					if claimEvent.Reason == "ProvisioningSucceeded" {
						successCh <- true
					} else if claimEvent.Reason == "ProvisioningFailed" {
						successCh <- false
					}
				}
			}
		}
	}()

	return successCh, nil
}

// watchPVC returns a watch on the given PVC and ProvisioningFailed &
// ProvisioningSucceeded events involving it
func (ctrl *ProvisionController) watchPVC(claim *v1.PersistentVolumeClaim, stopChannel chan struct{}) (<-chan watch.Event, error) {
	pvcSelector, _ := fields.ParseSelector("metadata.name=" + claim.Name)
	options := api.ListOptions{
		FieldSelector:   pvcSelector,
		Watch:           true,
		ResourceVersion: claim.ResourceVersion,
	}

	pvcWatch, err := ctrl.claimSource.Watch(options)
	if err != nil {
		return nil, err
	}

	failWatch, err := ctrl.getPVCEventWatch(claim, v1.EventTypeWarning, "ProvisioningFailed")
	if err != nil {
		pvcWatch.Stop()
		return nil, err
	}

	successWatch, err := ctrl.getPVCEventWatch(claim, v1.EventTypeNormal, "ProvisioningSucceeded")
	if err != nil {
		failWatch.Stop()
		pvcWatch.Stop()
		return nil, err
	}

	eventCh := make(chan watch.Event, 0)

	go func() {
		defer successWatch.Stop()
		defer failWatch.Stop()
		defer pvcWatch.Stop()
		defer close(eventCh)

		for {
			select {
			case _ = <-stopChannel:
				return

			case pvcEvent, ok := <-pvcWatch.ResultChan():
				if !ok {
					return
				}
				eventCh <- pvcEvent

			case failEvent, ok := <-failWatch.ResultChan():
				if !ok {
					return
				}
				eventCh <- failEvent

			case successEvent, ok := <-successWatch.ResultChan():
				if !ok {
					return
				}
				eventCh <- successEvent
			}
		}
	}()

	return eventCh, nil
}

// getPVCEventWatch returns a watch on the given PVC for the given event from
// this point forward.
func (ctrl *ProvisionController) getPVCEventWatch(claim *v1.PersistentVolumeClaim, eventType, reason string) (watch.Interface, error) {
	claimKind := "PersistentVolumeClaim"
	claimUID := string(claim.UID)
	fieldSelector := ctrl.client.Core().Events(claim.Namespace).GetFieldSelector(&claim.Name, &claim.Namespace, &claimKind, &claimUID).String() + ",type=" + eventType + ",reason=" + reason

	list, err := ctrl.client.Core().Events(claim.Namespace).List(v1.ListOptions{
		FieldSelector: fieldSelector,
	})
	if err != nil {
		return nil, err
	}

	resourceVersion := ""
	if len(list.Items) >= 1 {
		resourceVersion = list.Items[len(list.Items)-1].ResourceVersion
	}

	return ctrl.client.Core().Events(claim.Namespace).Watch(v1.ListOptions{
		FieldSelector:   fieldSelector,
		Watch:           true,
		ResourceVersion: resourceVersion,
	})
}

func (ctrl *ProvisionController) deleteVolumeOperation(volume *v1.PersistentVolume) error {
	glog.V(4).Infof("deleteVolumeOperation [%s] started", volume.Name)

	// This method may have been waiting for a volume lock for some time.
	// Our check does not have to be as sophisticated as PV controller's, we can
	// trust that the PV controller has set the PV to Released/Failed and it's
	// ours to delete
	newVolume, err := ctrl.client.Core().PersistentVolumes().Get(volume.Name)
	if err != nil {
		return nil
	}
	if !ctrl.shouldDelete(newVolume) {
		glog.Infof("volume %q no longer needs deletion, skipping", volume.Name)
		return nil
	}

	if err := ctrl.provisioner.Delete(volume); err != nil {
		if ierr, ok := err.(*IgnoredError); ok {
			// Delete ignored, do nothing and hope another provisioner will delete it.
			glog.Infof("deletion of volume %q ignored: %v", volume.Name, ierr)
			return nil
		}
		// Delete failed, emit an event.
		glog.Errorf("Deletion of volume %q failed: %v", volume.Name, err)
		ctrl.eventRecorder.Event(volume, v1.EventTypeWarning, "VolumeFailedDelete", err.Error())
		return err
	}

	glog.Infof("volume %q deleted", volume.Name)

	glog.V(4).Infof("deleteVolumeOperation [%s]: success", volume.Name)
	// Delete the volume
	if err = ctrl.client.Core().PersistentVolumes().Delete(volume.Name, nil); err != nil {
		// Oops, could not delete the volume and therefore the controller will
		// try to delete the volume again on next update.
		glog.Infof("failed to delete volume %q from database: %v", volume.Name, err)
		return nil
	}

	glog.Infof("volume %q deleted from database", volume.Name)
	return nil
}

// getProvisionedVolumeNameForClaim returns PV.Name for the provisioned volume.
// The name must be unique.
func (ctrl *ProvisionController) getProvisionedVolumeNameForClaim(claim *v1.PersistentVolumeClaim) string {
	return "pvc-" + string(claim.UID)
}

// scheduleOperation starts given asynchronous operation on given volume. It
// makes sure the operation is already not running.
func (ctrl *ProvisionController) scheduleOperation(operationName string, operation func() error) {
	glog.Infof("scheduleOperation[%s]", operationName)

	err := ctrl.runningOperations.Run(operationName, operation)
	if err != nil {
		if goroutinemap.IsAlreadyExists(err) {
			glog.V(4).Infof("operation %q is already running, skipping", operationName)
		} else {
			glog.Errorf("Error scheduling operaion %q: %v", operationName, err)
		}
	}
}

func (ctrl *ProvisionController) getStorageClass(name string) (*v1beta1.StorageClass, error) {
	classObj, found, err := ctrl.classes.GetByKey(name)
	if err != nil {
		return nil, fmt.Errorf("Error getting StorageClass %q: %v", name, err)
	}
	if !found {
		return nil, fmt.Errorf("StorageClass %q not found", name)
		// 3. It tries to find a StorageClass instance referenced by annotation
		//    `claim.Annotations["volume.beta.kubernetes.io/storage-class"]`. If not
		//    found, it SHOULD report an error (by sending an event to the claim) and it
		//    SHOULD retry periodically with step i.
	}
	storageClass, ok := classObj.(*v1beta1.StorageClass)
	if !ok {
		return nil, fmt.Errorf("Cannot convert object to StorageClass: %+v", classObj)
	}
	if storageClass.Provisioner != ctrl.provisionerName {
		// class.Provisioner has either changed since shouldProvision() or
		// annDynamicallyProvisioned contains different provisioner than
		// class.Provisioner.
		return nil, fmt.Errorf("Unknown provisioner %q requested in storage class %q", storageClass.Provisioner, name)
	}
	return storageClass, nil
}

func hasAnnotation(obj v1.ObjectMeta, ann string) bool {
	_, found := obj.Annotations[ann]
	return found
}

func setAnnotation(obj *v1.ObjectMeta, ann string, value string) {
	if obj.Annotations == nil {
		obj.Annotations = make(map[string]string)
	}
	obj.Annotations[ann] = value
}

// getClaimClass returns name of class that is requested by given claim.
// Request for `nil` class is interpreted as request for class "",
// i.e. for a classless PV.
func getClaimClass(claim *v1.PersistentVolumeClaim) string {
	// TODO: change to PersistentVolumeClaim.Spec.Class value when this
	// attribute is introduced.
	if class, found := claim.Annotations[annClass]; found {
		return class
	}

	return ""
}

func claimToClaimKey(claim *v1.PersistentVolumeClaim) string {
	return fmt.Sprintf("%s/%s", claim.Namespace, claim.Name)
}
