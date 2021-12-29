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

package addons

import (
	"strings"

	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/out"
)

// overrideDefaults creates a copy of `defaultMap` where `overrideMap` replaces any of its values that `overrideMap` contains.
func overrideDefaults(defaultMap, overrideMap map[string]string) map[string]string {
	return mergeMaps(defaultMap, filterKeySpace(defaultMap, overrideMap))
}

// parseMapString creates a map based on `str` which is encoded as <key1>=<value1>,<key2>=<value2>,...
func parseMapString(str string) map[string]string {
	mapResult := make(map[string]string)
	if str == "" {
		return mapResult
	}
	for _, pairText := range strings.Split(str, ",") {
		vals := strings.Split(pairText, "=")
		if len(vals) != 2 {
			out.WarningT("Ignoring invalid pair entry {{.pair}}", out.V{"pair": pairText})
			continue
		}
		mapResult[vals[0]] = vals[1]
	}
	return mapResult
}

// mergeMaps creates a map with the union of `sourceMap` and `overrideMap` where collisions take the value of `overrideMap`.
func mergeMaps(sourceMap, overrideMap map[string]string) map[string]string {
	result := make(map[string]string)
	for name, value := range sourceMap {
		result[name] = value
	}
	for name, value := range overrideMap {
		result[name] = value
	}
	return result
}

// filterKeySpace creates a map of the values in `targetMap` where the keys are also in `keySpace`.
func filterKeySpace(keySpace map[string]string, targetMap map[string]string) map[string]string {
	result := make(map[string]string)
	for name := range keySpace {
		if value, ok := targetMap[name]; ok {
			result[name] = value
		}
	}
	return result
}

// fixAddonImagesAndRegistries fixes images & registries in addon
func fixAddonImagesAndRegistries(addon *AddonPackage, images map[string]string, registries map[string]string) (customImages, customRegistries map[string]string) {
	customImages = make(map[string]string)
	customRegistries = make(map[string]string)
	addonImages := addon.GetImages()
	if images == nil {
		images = make(map[string]string, len(addonImages))
	}
	if registries == nil {
		registries = make(map[string]string, len(addonImages))
	}

	for name, image := range images {
		registry, found := registries[name]
		if !found {
			registry = addonImages[name].Registry()
		}
		img := image

		if registry != "" && registry != "docker.io" {
			img = registry + "/" + image
		}
		parts := strings.SplitN(img, ":", 2)
		imageName := parts[0]
		tag := parts[1]
		mirror, found := assets.AliyunMirror[imageName]
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
