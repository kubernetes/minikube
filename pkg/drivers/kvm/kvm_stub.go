//go:build linux && !amd64

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

package kvm

import (
	"fmt"
	"runtime"

	"k8s.io/minikube/pkg/libmachine/drivers"
	"k8s.io/minikube/pkg/libmachine/mcnflag"
	"k8s.io/minikube/pkg/libmachine/state"
)

// This is a stub driver for unsupported architectures. All function fail with
// errNotSupported error or return an zero value.

var errNotSupported = fmt.Errorf("the kvm driver is not supported on %q", runtime.GOARCH)

func (d *Driver) Create() error                                       { return errNotSupported }
func (d *Driver) GetCreateFlags() []mcnflag.Flag                      { return nil }
func (d *Driver) GetIP() (string, error)                              { return "", errNotSupported }
func (d *Driver) GetMachineName() string                              { return "" }
func (d *Driver) GetSSHHostname() (string, error)                     { return "", errNotSupported }
func (d *Driver) GetSSHKeyPath() string                               { return "" }
func (d *Driver) GetSSHPort() (int, error)                            { return 0, errNotSupported }
func (d *Driver) GetSSHUsername() string                              { return "" }
func (d *Driver) GetURL() (string, error)                             { return "", errNotSupported }
func (d *Driver) GetState() (state.State, error)                      { return state.None, errNotSupported }
func (d *Driver) Kill() error                                         { return errNotSupported }
func (d *Driver) PreCreateCheck() error                               { return errNotSupported }
func (d *Driver) Remove() error                                       { return errNotSupported }
func (d *Driver) Restart() error                                      { return errNotSupported }
func (d *Driver) SetConfigFromFlags(opts drivers.DriverOptions) error { return errNotSupported }
func (d *Driver) Start() error                                        { return errNotSupported }
func (d *Driver) Stop() error                                         { return errNotSupported }
