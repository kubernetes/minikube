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

package audit

import (
	"time"

	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
)

// entry represents the execution of a command.
type entry struct {
	data map[string]string
}

// Type returns the cloud events compatible type of this struct.
func (e *entry) Type() string {
	return "io.k8s.sigs.minikube.audit"
}

// newEntry returns a new audit type.
func newEntry(command string, args string, user string, startTime time.Time, endTime time.Time) *entry {
	return &entry{
		map[string]string{
			"args":      args,
			"command":   command,
			"endTime":   endTime.Format(constants.TimeFormat),
			"profile":   viper.GetString(config.ProfileName),
			"startTime": startTime.Format(constants.TimeFormat),
			"user":      user,
		},
	}
}
