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

package version

var (
	// APIVersion dictates which version of the libmachine API this is.
	APIVersion = 1

	// ConfigVersion dictates which version of the config.json format is
	// used. It needs to be bumped if there is a breaking change, and
	// therefore migration, introduced to the config file format.
	ConfigVersion = 3
)
