/*
Copyright 2022 The Kubernetes Authors All rights reserved.

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

package virtualbox

import (
	"time"

	"k8s.io/minikube/pkg/libmachine/drivers"
	"k8s.io/minikube/pkg/libmachine/mcnutils"
)

// IPWaiter waits for an IP to be configured.
type IPWaiter interface {
	Wait(d *Driver) error
}

func NewIPWaiter() IPWaiter {
	return &sshIPWaiter{}
}

type sshIPWaiter struct{}

func (w *sshIPWaiter) Wait(d *Driver) error {
	// Wait for SSH over NAT to be available before returning to user
	if err := drivers.WaitForSSH(d); err != nil {
		return err
	}

	// Bail if we don't get an IP from DHCP after a given number of seconds.
	if err := mcnutils.WaitForSpecific(d.hostOnlyIPAvailable, 5, 4*time.Second); err != nil {
		return err
	}

	var err error
	d.IPAddress, err = d.GetIP()

	return err
}
