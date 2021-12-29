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

package addons

import (
	"fmt"
	"net/url"
	"os"

	"github.com/docker/machine/libmachine/log"
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/assets"
)

type registryDeclaration struct {
	Addons []string
}

type addonAssetDeclaration struct {
	Source      string
	Target      string
	Permissions string
}

type addonImageDeclaration struct {
	Registry string
	Image    string
}

type addonDeclaration struct {
	Name       string
	Maintainer string
	Templates  []addonAssetDeclaration
	Assets     []addonAssetDeclaration
	Images     map[string]addonImageDeclaration
}

var defaultRegistries = []string{
	"embedfs://addon.DefaultAddonRegistryAssets/addon-registry.yaml",
}

var Addons = loadAddons()

func loadAddons() map[string]*AddonPackage {
	addons := make(map[string]*AddonPackage)

	for _, registry := range defaultRegistries {
		err := loadRegistry(registry, addons)

		if err != nil {
			panic(err)
		}
	}

	config, err := LoadAddonsConfig()

	if err != nil {
		panic(err)
	}

	if config != nil {
		for _, registry := range config.CustomRegistries {
			if !registry.Enabled {
				continue
			}

			err := loadRegistry(registry.Path, addons)

			if err != nil {
				log.Errorf("failed to load custom registry %s: %s", registry.Path, err)
			}
		}
	}

	return addons
}

func loadRegistry(path string, addons map[string]*AddonPackage) error {
	uri, err := url.Parse(path)
	if err != nil || uri.Scheme == "" {
		path = "file:///" + os.ExpandEnv(path)
		uri, err = url.Parse(path)
		if err != nil {
			return errors.Wrapf(err, "failed to create URL")
		}
	}

	var regConfig registryDeclaration
	err = assets.UnmarshalLoad(uri, &regConfig)
	if err != nil {
		return err
	}

	for _, addonPath := range regConfig.Addons {
		err = LoadAddon(uri, addonPath, path, addons)
		if err != nil {
			return err
		}
	}

	return nil
}

func LoadAddon(regURI *url.URL, addonPath string, path string, addons map[string]*AddonPackage) error {
	addonURI, err := url.Parse(addonPath)
	if err != nil {
		return errors.Wrapf(err, "parsing addon URL %s", path)
	}

	addonURI = regURI.ResolveReference(addonURI)

	var addonConfig addonDeclaration
	err = assets.UnmarshalLoad(addonURI, &addonConfig)
	if err != nil {
		return err
	}

	if _, exists := addons[addonConfig.Name]; exists {
		return fmt.Errorf("addon %s already exists", addonConfig.Name)
	}

	assets := make([]assets.Asset, 0, len(addonConfig.Templates)+len(addonConfig.Assets))
	assets, err = loadAssets(addonURI, addonConfig.Assets, false, assets)
	if err != nil {
		return err
	}
	assets, err = loadAssets(addonURI, addonConfig.Templates, true, assets)
	if err != nil {
		return err
	}

	images := make(map[string]AddonImage, len(addonConfig.Images))
	for name, image := range addonConfig.Images {
		images[name] = AddonImage{
			name:     name,
			image:    image.Image,
			registry: image.Registry,
		}
	}

	addons[addonConfig.Name] = &AddonPackage{
		name:       addonConfig.Name,
		maintainer: addonConfig.Maintainer,
		assets:     assets,
		images:     images,
	}
	return nil
}

func loadAssets(addonURI *url.URL, configs []addonAssetDeclaration, isTemplate bool, results []assets.Asset) ([]assets.Asset, error) {
	for _, assetConfig := range configs {
		permissions := assetConfig.Permissions
		if permissions == "" {
			permissions = "0640"
		}

		assetURI, err := url.Parse(assetConfig.Source)
		if err != nil {
			return results, errors.Wrapf(err, "parsing asset URL %s", assetConfig.Source)
		}

		assetData, err := assets.LoadAsset(addonURI, assetURI, assetConfig.Target, permissions, isTemplate)
		if err != nil {
			return results, errors.Wrapf(err, "loading asset %s", assetConfig.Source)
		}

		results = append(results, assetData)
	}

	return results, nil
}
