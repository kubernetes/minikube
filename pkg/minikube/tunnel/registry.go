/*
Copyright 2018 The Kubernetes Authors All rights reserved.

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

package tunnel

//registry should be
// - configurable in terms of directory for testing
// - one per user, across multiple vms
// should have a list of tunnels:
// tunnel: route, pid, machinename
// when cleanup is called, all the non running tunnels should be checked for removal
// when a new tunnel is created it should register itself with the registry pid/machinename/route
