/*
Copyright 2021 The Kubernetes Authors All rights reserved.

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

package assets

import (
	"encoding/json"
	"fmt"
	"strings"

	"k8s.io/minikube/deploy/addons"
)

// AliyunMirror list of images from Aliyun mirror
var AliyunMirror = loadAliyunMirror()

func loadAliyunMirror() map[string]string {
	data, err := addons.AliyunMirror.ReadFile("aliyun_mirror.json")
	if err != nil {
		panic(fmt.Sprintf("Failed to load aliyun_mirror.json: %v", err))
	}
	var mirror map[string]string
	err = json.Unmarshal(data, &mirror)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse aliyun_mirror.json: %v", err))
	}
	return mirror
}

// FixAddonImagesAndRegistries fixes images & registries in addon
func FixAddonImagesAndRegistries(addon *Addon, images map[string]string, registries map[string]string) (customImages, customRegistries map[string]string) {
	customImages = make(map[string]string)
	customRegistries = make(map[string]string)
	if images == nil {
		images = addon.Images
	}
	if addon.Registries == nil {
		addon.Registries = make(map[string]string)
	}
	if registries == nil {
		registries = make(map[string]string)
	}
	for name, image := range images {
		registry, found := registries[name]
		if !found {
			registry = addon.Registries[name]
		}
		img := image

		if registry != "" && registry != "docker.io" {
			img = registry + "/" + image
		}
		parts := strings.SplitN(img, ":", 2)
		imageName := parts[0]
		tag := parts[1]
		mirror, found := AliyunMirror[imageName]
		if found {
			parts := strings.SplitN(mirror, "/", 2)
			mirrorRegistry := parts[0]
			mirrorImage := parts[1] + ":" + tag
			customImages[name] = mirrorImage
			customRegistries[name] = mirrorRegistry
		} else {
			customImages[name] = image
			customRegistries[name] = registry
		}
	}
	return customImages, customRegistries
}
