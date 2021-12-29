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
	"fmt"
	"strings"

	"github.com/blang/semver/v4"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/util"
	"k8s.io/minikube/pkg/version"
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

func supportStorageProvisionerVersion(addonName string, image AddonImage) AddonImage {
	if addonName != "storage-provisioner" || image.name != "StorageProvisioner" {
		return image
	}

	var labelIndex = strings.LastIndex(image.image, ":")
	var slashIndex = strings.LastIndex(image.image, "/")

	// these are equal if neither '/' nor ':' is found
	if labelIndex <= slashIndex {
		image.image = fmt.Sprintf("%s:%s", image.Image(), version.GetStorageProvisionerVersion())
	}
	return image
}

// maintain backwards compatibility for ingress and ingress-dns addons with k8s < v1.19 by replacing default addons' images with compatible versions
func supportLegacyIngress(addon *AddonPackage, cc config.ClusterConfig) error {
	v, err := util.ParseKubernetesVersion(cc.KubernetesConfig.KubernetesVersion)
	if err != nil {
		return errors.Wrap(err, "parsing Kubernetes version")
	}
	if semver.MustParseRange("<1.19.0")(v) {
		if addon.Name() == "ingress" {
			addon.images = map[string]AddonImage{
				// https://github.com/kubernetes/ingress-nginx/blob/0a2ec01eb4ec0e1b29c4b96eb838a2e7bfe0e9f6/deploy/static/provider/kind/deploy.yaml#L328
				"IngressController": {
					name:     "IngressController",
					image:    "ingress-nginx/controller:v0.49.3@sha256:35fe394c82164efa8f47f3ed0be981b3f23da77175bbb8268a9ae438851c8324",
					registry: "",
				},
				// issues: https://github.com/kubernetes/ingress-nginx/issues/7418 and https://github.com/jet/kube-webhook-certgen/issues/30
				"KubeWebhookCertgenCreate": {
					name:     "KubeWebhookCertgenCreate",
					image:    "jettech/kube-webhook-certgen:v1.5.1@sha256:950833e19ade18cd389d647efb88992a7cc077abedef343fa59e012d376d79b7",
					registry: "docker.io",
				},
				"KubeWebhookCertgenPatch": {
					name:     "KubeWebhookCertgenPatch",
					image:    "jettech/kube-webhook-certgen:v1.5.1@sha256:950833e19ade18cd389d647efb88992a7cc077abedef343fa59e012d376d79b7",
					registry: "",
				},
			}
			return nil
		}
		if addon.Name() == "ingress-dns" {
			addon.images = map[string]AddonImage{
				"IngressDNS": {
					name:     "IngressDNS",
					image:    "cryptexlabs/minikube-ingress-dns:0.3.0@sha256:e252d2a4c704027342b303cc563e95d2e71d2a0f1404f55d676390e28d5093ab",
					registry: "",
				},
			}
			return nil
		}
		return fmt.Errorf("supportLegacyIngress called for unexpected addon %q - nothing to do here", addon.Name())
	}

	return nil
}
