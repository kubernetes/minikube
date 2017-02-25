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

package assets

import (
	"os"
	"path/filepath"
	"strconv"

	"github.com/golang/glog"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/util"
)

type Addon struct {
	Assets    []*MemoryAsset
	enabled   bool
	addonName string
}

func NewAddon(assets []*MemoryAsset, enabled bool, addonName string) *Addon {
	a := &Addon{
		Assets:    assets,
		enabled:   enabled,
		addonName: addonName,
	}
	return a
}

func (a *Addon) IsEnabled() (bool, error) {
	addonStatusText, err := config.Get(a.addonName)
	if err == nil {
		addonStatus, err := strconv.ParseBool(addonStatusText)
		if err != nil {
			return false, err
		}
		return addonStatus, nil
	}
	return a.enabled, nil
}

var Addons = map[string]*Addon{
	"addon-manager": NewAddon([]*MemoryAsset{
		NewMemoryAsset(
			"deploy/addons/addon-manager.yaml",
			"/etc/kubernetes/manifests/",
			"addon-manager.yaml",
			"0640"),
	}, true, "addon-manager"),
	"dashboard": NewAddon([]*MemoryAsset{
		NewMemoryAsset(
			"deploy/addons/dashboard/dashboard-rc.yaml",
			constants.AddonsPath,
			"dashboard-rc.yaml",
			"0640"),
		NewMemoryAsset(
			"deploy/addons/dashboard/dashboard-svc.yaml",
			constants.AddonsPath,
			"dashboard-svc.yaml",
			"0640"),
	}, true, "dashboard"),
	"default-storageclass": NewAddon([]*MemoryAsset{
		NewMemoryAsset(
			"deploy/addons/storageclass/storageclass.yaml",
			constants.AddonsPath,
			"storageclass.yaml",
			"0640"),
	}, true, "default-storageclass"),
	"kube-dns": NewAddon([]*MemoryAsset{
		NewMemoryAsset(
			"deploy/addons/kube-dns/kube-dns-rc.yaml",
			constants.AddonsPath,
			"kube-dns-rc.yaml",
			"0640"),
		NewMemoryAsset(
			"deploy/addons/kube-dns/kube-dns-svc.yaml",
			constants.AddonsPath,
			"kube-dns-svc.yaml",
			"0640"),
	}, true, "kube-dns"),
	"heapster": NewAddon([]*MemoryAsset{
		NewMemoryAsset(
			"deploy/addons/heapster/influxGrafana-rc.yaml",
			constants.AddonsPath,
			"influxGrafana-rc.yaml",
			"0640"),
		NewMemoryAsset(
			"deploy/addons/heapster/grafana-svc.yaml",
			constants.AddonsPath,
			"grafana-svc.yaml",
			"0640"),
		NewMemoryAsset(
			"deploy/addons/heapster/influxdb-svc.yaml",
			constants.AddonsPath,
			"influxdb-svc.yaml",
			"0640"),
		NewMemoryAsset(
			"deploy/addons/heapster/heapster-rc.yaml",
			constants.AddonsPath,
			"heapster-rc.yaml",
			"0640"),
		NewMemoryAsset(
			"deploy/addons/heapster/heapster-svc.yaml",
			constants.AddonsPath,
			"heapster-svc.yaml",
			"0640"),
	}, false, "heapster"),
	"ingress": NewAddon([]*MemoryAsset{
		NewMemoryAsset(
			"deploy/addons/ingress/ingress-configmap.yaml",
			constants.AddonsPath,
			"ingress-configmap.yaml",
			"0640"),
		NewMemoryAsset(
			"deploy/addons/ingress/ingress-rc.yaml",
			constants.AddonsPath,
			"ingress-rc.yaml",
			"0640"),
		NewMemoryAsset(
			"deploy/addons/ingress/ingress-svc.yaml",
			constants.AddonsPath,
			"ingress-svc.yaml",
			"0640"),
	}, false, "ingress"),
	"registry-creds": NewAddon([]*MemoryAsset{
		NewMemoryAsset(
			"deploy/addons/registry-creds/registry-creds-rc.yaml",
			constants.AddonsPath,
			"registry-creds-rc.yaml",
			"0640"),
	}, false, "registry-creds"),
}

func AddMinikubeAddonsDirToAssets(assetList *[]CopyableFile) {
	// loop over .minikube/addons and add them to assets
	searchDir := constants.MakeMiniPath("addons")
	err := filepath.Walk(searchDir, func(addonFile string, f os.FileInfo, err error) error {
		isDir, err := util.IsDirectory(addonFile)
		if err == nil && !isDir {
			f, err := NewFileAsset(addonFile, constants.AddonsPath, filepath.Base(addonFile), "0640")
			if err == nil {
				*assetList = append(*assetList, f)
			}
		} else if err != nil {
			glog.Infoln("Error encountered while walking .minikube/addons: ", err)
		}
		return nil
	})
	if err != nil {
		glog.Infoln("Error encountered while walking .minikube/addons: ", err)
	}
}
