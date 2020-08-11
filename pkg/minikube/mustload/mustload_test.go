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

package mustload

import (
	"path/filepath"
	"testing"
)

func TestPartial(t *testing.T) {
	path := filepath.Join("..", "config", "testdata", "profile", ".minikube")
	name := "p1"
	api, cc := Partial(name, path)

	if cc.Name != name {
		t.Fatalf("cc.Name expected to be same as name(%s), but got %s", name, cc.Name)
	}
	if api == nil {
		t.Fatalf("expected to get not empty api struct")
	}
}
