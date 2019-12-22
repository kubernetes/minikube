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

package kubeadm

import (
	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/assets"
)

func addAddons(files *[]assets.CopyableFile, data interface{}) error {
	// add addons to file list
	// custom addons
	if err := assets.AddMinikubeDirAssets(files); err != nil {
		return errors.Wrap(err, "adding minikube dir assets")
	}
	// bundled addons
	for _, addonBundle := range assets.Addons {
		if isEnabled, err := addonBundle.IsEnabled(); err == nil && isEnabled {
			for _, addon := range addonBundle.Assets {
				if addon.IsTemplate() {
					addonFile, err := addon.Evaluate(data)
					if err != nil {
						return errors.Wrapf(err, "evaluate bundled addon %s asset", addon.GetAssetName())
					}

					*files = append(*files, addonFile)
				} else {
					*files = append(*files, addon)
				}
			}
		} else if err != nil {
			return nil
		}
	}

	return nil
}
