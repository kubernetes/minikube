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

package oci

import "errors"

// FailFastError type is an error that could not be solved by trying again
type FailFastError error

// ErrWindowsContainers is thrown when docker been configured to run windows containers instead of Linux
var ErrWindowsContainers = FailFastError(errors.New("docker container type is windows"))

// ErrCPUCountLimit is thrown when docker daemon doesn't have enough CPUs for the requested container
var ErrCPUCountLimit = FailFastError(errors.New("not enough CPUs is available for container"))

// ErrNotRunningAfterCreate is thrown when container is created without error but it exists and it's status is not running
var ErrNotRunningAfterCreate = errors.New("container status is not running after creation")
