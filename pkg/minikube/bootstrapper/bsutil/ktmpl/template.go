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

package ktmpl

import (
	"fmt"
	"sort"
)

// printMapInOrder sorts the keys and prints the map in order, combining key
// value pairs with the separator character
//
// Note: this is not necessary, but makes testing easy
func printMapInOrder(m map[string]string, sep string) []string {
	if m == nil {
		return nil
	}
	keys := []string{}
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for i, k := range keys {
		keys[i] = fmt.Sprintf("%s%s\"%s\"", k, sep, m[k])
	}
	return keys
}
