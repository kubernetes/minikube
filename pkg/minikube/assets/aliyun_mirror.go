package assets

import (
	"encoding/json"
	"fmt"
	"strings"

	"k8s.io/minikube/deploy/addons"
)

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
