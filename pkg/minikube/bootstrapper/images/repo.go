/*
Copyright 2019 The Kubernetes Authors All rights reserved.

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

package images

// DefaultKubernetesRepo is the default Kubernetes repository
const DefaultKubernetesRepo = "registry.k8s.io"

// kubernetesRepo returns the official Kubernetes repository, or an alternate
func kubernetesRepo(mirror string) string {
	if mirror != "" {
		return mirror
	}
	return DefaultKubernetesRepo
}
