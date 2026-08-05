/*
Copyright 2026 The Kubernetes Authors All rights reserved.

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

package https

import "fmt"

// TrustStore represents an OS-specific user certificate trust store.
type TrustStore interface {
	Install(caCertPath string, profile string) error
	Remove(profile string) error
}

// NewTrustStore returns the TrustStore implementation for the current OS.
func NewTrustStore() TrustStore {
	return newTrustStore()
}

// CAName returns the canonical name of the CA certificate for a profile.
func CAName(profile string) string {
	return fmt.Sprintf("minikube-%s", profile)
}
