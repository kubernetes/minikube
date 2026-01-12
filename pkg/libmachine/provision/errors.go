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

package provision

import (
	"errors"
	"fmt"
)

var (
	ErrDetectionFailed = errors.New("OS type not recognized")
)

type ErrDaemonAvailable struct {
	wrappedErr error
}

func (e ErrDaemonAvailable) Error() string {
	return fmt.Sprintf("Unable to verify the Docker daemon is listening: %s", e.wrappedErr)
}

func NewErrDaemonAvailable(err error) ErrDaemonAvailable {
	return ErrDaemonAvailable{
		wrappedErr: err,
	}
}
