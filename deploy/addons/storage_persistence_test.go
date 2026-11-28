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

package addons

import (
	"encoding/json"
	"path"
	"strings"
	"testing"

	"sigs.k8s.io/yaml"
)

// TestRancherVolumesUsePersistentGuestDisk checks the packaged provisioning
// configuration against Minikube's persistent VM mount, independently of Pods.
func TestRancherVolumesUsePersistentGuestDisk(t *testing.T) {
	content, err := StorageProvisionerRancherAssets.ReadFile("storage-provisioner-rancher/storage-provisioner-rancher.yaml.tmpl")
	if err != nil {
		t.Fatal(err)
	}
	// The final document contains the ConfigMap; image templates inside its
	// block scalar remain opaque to YAML parsing.
	documents := strings.Split(string(content), "\n---\n")
	var configMap struct {
		Data map[string]string `json:"data"`
	}
	if err := yaml.Unmarshal([]byte(documents[len(documents)-1]), &configMap); err != nil {
		t.Fatal(err)
	}
	var configuration struct {
		Nodes []struct {
			Node  string   `json:"node"`
			Paths []string `json:"paths"`
		} `json:"nodePathMap"`
	}
	if err := json.Unmarshal([]byte(configMap.Data["config.json"]), &configuration); err != nil {
		t.Fatal(err)
	}
	if len(configuration.Nodes) != 1 || configuration.Nodes[0].Node != "DEFAULT_PATH_FOR_NON_LISTED_NODES" || len(configuration.Nodes[0].Paths) == 0 {
		t.Fatal("expected a default provisioning path for every local node")
	}
	for _, directory := range configuration.Nodes[0].Paths {
		if !path.IsAbs(directory) || !strings.HasPrefix(path.Clean(directory), "/var/lib/minikube/") {
			t.Errorf("volume directory %q falls outside the persistent guest data mount", directory)
		}
	}
}
