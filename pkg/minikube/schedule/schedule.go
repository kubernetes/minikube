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

package schedule

import (
	"log"
	"time"

	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/mustload"
)

// Daemonize daemonizes minikube so that scheduled stop happens as expected
func Daemonize(profiles []string, duration time.Duration) error {
	// save current time and expected duration in config
	scheduledStop := &config.ScheduledStopConfig{
		InitiationTime: time.Now().Unix(),
		Duration:       duration,
	}
	if err := killExistingScheduledStops(profiles); err != nil {
		log.Printf("error killing existing scheduled stops: %v", err)
	}
	for _, p := range profiles {
		_, cc := mustload.Partial(p)
		cc.ScheduledStop = scheduledStop
		if err := config.SaveProfile(p, cc); err != nil {
			return errors.Wrap(err, "saving profile")
		}
	}

	return daemonize(profiles, duration)
}
