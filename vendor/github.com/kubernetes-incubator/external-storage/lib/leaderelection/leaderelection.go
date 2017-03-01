/*
Copyright 2015 The Kubernetes Authors.

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

// This is a modified version of kube's leaderelection package which uses a
// dummy endpoints object & its annotation as a lock. Here a pvc is used and
// the lock is to help ensure only one provisioner (the leader) is trying to
// provision a volume for the pvc at a time. So the election lasts only until
// the task is completed. Adds also a 'TermLimit.'
// https://github.com/kubernetes/kubernetes/tree/release-1.5/pkg/client/leaderelection

package leaderelection

import (
	"fmt"
	"reflect"
	"time"

	rl "github.com/kubernetes-incubator/external-storage/lib/leaderelection/resourcelock"
	"k8s.io/client-go/pkg/api/errors"
	"k8s.io/client-go/pkg/api/unversioned"
	"k8s.io/client-go/pkg/util/runtime"
	"k8s.io/client-go/pkg/util/wait"

	"github.com/golang/glog"
)

const (
	JitterFactor         = 1.2
	DefaultLeaseDuration = 15 * time.Second
	DefaultRenewDeadline = 10 * time.Second
	DefaultRetryPeriod   = 2 * time.Second
	DefaultTermLimit     = 30 * time.Second
)

// NewLeaderElector creates a LeaderElector from a LeaderElecitionConfig
func NewLeaderElector(lec LeaderElectionConfig) (*LeaderElector, error) {
	if lec.LeaseDuration <= lec.RenewDeadline {
		return nil, fmt.Errorf("leaseDuration must be greater than renewDeadline")
	}
	if lec.RenewDeadline <= time.Duration(JitterFactor*float64(lec.RetryPeriod)) {
		return nil, fmt.Errorf("renewDeadline must be greater than retryPeriod*JitterFactor")
	}
	if lec.Lock == nil {
		return nil, fmt.Errorf("Lock must not be nil.")
	}
	return &LeaderElector{
		config: lec,
	}, nil
}

type LeaderElectionConfig struct {
	// Lock is the resource that will be used for locking
	Lock rl.Interface

	// LeaseDuration is the duration that non-leader candidates will
	// wait to force acquire leadership. This is measured against time of
	// last observed ack.
	LeaseDuration time.Duration
	// RenewDeadline is the duration that the acting master will retry
	// refreshing leadership before giving up.
	RenewDeadline time.Duration
	// RetryPeriod is the duration the LeaderElector clients should wait
	// between tries of actions.
	RetryPeriod time.Duration
	// TermLimit is the maximum duration that a leader may remain the leader
	// to complete the task before it must give up its leadership. 0 for forever
	// or indefinite.
	TermLimit time.Duration

	// Callbacks are callbacks that are triggered during certain lifecycle
	// events of the LeaderElector
	Callbacks LeaderCallbacks
}

// LeaderCallbacks are callbacks that are triggered during certain
// lifecycle events of the LeaderElector. These are invoked asynchronously.
//
// possible future callbacks:
//  * OnChallenge()
type LeaderCallbacks struct {
	// OnStartedLeading is called when a LeaderElector client starts leading
	OnStartedLeading func(stop <-chan struct{})
	// OnStoppedLeading is called when a LeaderElector client stops leading
	OnStoppedLeading func()
	// OnNewLeader is called when the client observes a leader that is
	// not the previously observed leader. This includes the first observed
	// leader when the client starts.
	OnNewLeader func(identity string)
}

// LeaderElector is a leader election client.
//
// possible future methods:
//  * (le *LeaderElector) IsLeader()
//  * (le *LeaderElector) GetLeader()
type LeaderElector struct {
	config LeaderElectionConfig
	// internal bookkeeping
	observedRecord rl.LeaderElectionRecord
	observedTime   time.Time
	// used to implement OnNewLeader(), may lag slightly from the
	// value observedRecord.HolderIdentity if the transition has
	// not yet been reported.
	reportedLeader string
}

// Run starts the leader election loop
func (le *LeaderElector) Run(task <-chan bool) {
	defer func() {
		runtime.HandleCrash()
	}()
	over := le.acquire(task)
	if over {
		return
	}
	stop := make(chan struct{})
	go le.config.Callbacks.OnStartedLeading(stop)
	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(le.config.TermLimit)
		timeout <- true
	}()
	le.renew(task, timeout)
	close(stop)
	le.config.Callbacks.OnStoppedLeading()
}

// GetLeader returns the identity of the last observed leader or returns the empty string if
// no leader has yet been observed.
func (le *LeaderElector) GetLeader() string {
	return le.observedRecord.HolderIdentity
}

// IsLeader returns true if the last observed leader was this client else returns false.
func (le *LeaderElector) IsLeader() bool {
	return le.observedRecord.HolderIdentity == le.config.Lock.Identity()
}

// acquire loops calling tryAcquireOrRenew and returns immediately when tryAcquireOrRenew succeeds
// or the task has successfully finished in which case there is no longer a need to acquire
func (le *LeaderElector) acquire(task <-chan bool) bool {
	over := false
	stop := make(chan struct{})
	glog.Infof("attempting to acquire leader lease...")
	wait.JitterUntil(func() {
		select {
		case taskSucceeded := <-task:
			if taskSucceeded {
				// if the leader succeeded at the task, stop trying to acquire
				desc := le.config.Lock.Describe()
				glog.Infof("stopped trying to acquire lease %v, task succeeded", desc)
				over = true
				close(stop)
				return
			}
		default:
		}
		succeeded := le.tryAcquireOrRenew()
		le.maybeReportTransition()
		desc := le.config.Lock.Describe()
		if !succeeded {
			glog.V(4).Infof("failed to acquire lease %v", desc)
			return
		}
		le.config.Lock.RecordEvent("became leader")
		glog.Infof("sucessfully acquired lease %v", desc)
		close(stop)
	}, le.config.RetryPeriod, JitterFactor, true, stop)

	return over
}

// renew loops calling tryAcquireOrRenew and returns immediately when tryAcquireOrRenew fails
// or the task has either succeeded or failed in which case leadership must be given up
func (le *LeaderElector) renew(task <-chan bool, timeout <-chan bool) {
	stop := make(chan struct{})
	wait.Until(func() {
		select {
		case taskSucceeded := <-task:
			// if the leader (us) either succeeded or failed at the task, stop trying to renew
			desc := le.config.Lock.Describe()
			taskSucceededStr := "succeeded"
			if !taskSucceeded {
				taskSucceededStr = "failed"
			}
			glog.Infof("stopped trying to renew lease %v, task %s", desc, taskSucceededStr)
			close(stop)
			return
		case <-timeout:
			// our term limit has ended, let somebody else have a try
			desc := le.config.Lock.Describe()
			glog.Infof("stopped trying to renew lease %v, timeout reached", desc)
			close(stop)
			return
		default:
		}
		err := wait.Poll(le.config.RetryPeriod, le.config.RenewDeadline, func() (bool, error) {
			return le.tryAcquireOrRenew(), nil
		})
		le.maybeReportTransition()
		desc := le.config.Lock.Describe()
		if err == nil {
			glog.V(4).Infof("succesfully renewed lease %v", desc)
			return
		}
		le.config.Lock.RecordEvent("stopped leading")
		glog.Infof("failed to renew lease %v", desc)
		close(stop)
	}, 0, stop)
}

// tryAcquireOrRenew tries to acquire a leader lease if it is not already acquired,
// else it tries to renew the lease if it has already been acquired. Returns true
// on success else returns false.
func (le *LeaderElector) tryAcquireOrRenew() bool {
	now := unversioned.Now()
	leaderElectionRecord := rl.LeaderElectionRecord{
		HolderIdentity:       le.config.Lock.Identity(),
		LeaseDurationSeconds: int(le.config.LeaseDuration / time.Second),
		RenewTime:            now,
		AcquireTime:          now,
	}

	// 1. obtain or create the ElectionRecord
	oldLeaderElectionRecord, err := le.config.Lock.Get()
	if err != nil {
		if !errors.IsNotFound(err) {
			glog.Errorf("error retrieving resource lock %v: %v", le.config.Lock.Describe(), err)
			return false
		}
		if err = le.config.Lock.Create(leaderElectionRecord); err != nil {
			glog.Errorf("error initially creating leader election record: %v", err)
			return false
		}
		le.observedRecord = leaderElectionRecord
		le.observedTime = time.Now()
		return true
	}

	// 2. Record obtained, check the Identity & Time
	if !reflect.DeepEqual(le.observedRecord, *oldLeaderElectionRecord) {
		le.observedRecord = *oldLeaderElectionRecord
		le.observedTime = time.Now()
	}
	if le.observedTime.Add(le.config.LeaseDuration).After(now.Time) &&
		oldLeaderElectionRecord.HolderIdentity != le.config.Lock.Identity() {
		glog.V(4).Infof("lock is held by %v and has not yet expired", oldLeaderElectionRecord.HolderIdentity)
		return false
	}

	// 3. We're going to try to update. The leaderElectionRecord is set to it's default
	// here. Let's correct it before updating.
	if oldLeaderElectionRecord.HolderIdentity == le.config.Lock.Identity() {
		leaderElectionRecord.AcquireTime = oldLeaderElectionRecord.AcquireTime
	} else {
		leaderElectionRecord.LeaderTransitions = oldLeaderElectionRecord.LeaderTransitions + 1
	}

	// update the lock itself
	if err = le.config.Lock.Update(leaderElectionRecord); err != nil {
		glog.Errorf("Failed to update lock: %v", err)
		return false
	}
	le.observedRecord = leaderElectionRecord
	le.observedTime = time.Now()
	return true
}

func (l *LeaderElector) maybeReportTransition() {
	if l.observedRecord.HolderIdentity == l.reportedLeader {
		return
	}
	l.reportedLeader = l.observedRecord.HolderIdentity
	if l.config.Callbacks.OnNewLeader != nil {
		go l.config.Callbacks.OnNewLeader(l.reportedLeader)
	}
}
