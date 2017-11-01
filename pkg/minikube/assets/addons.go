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
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/golang/glog"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/util"
)

type Addon struct {
	Assets    []*BinDataAsset
	enabled   bool
	addonName string
}

func NewAddon(assets []*BinDataAsset, enabled bool, addonName string) *Addon {
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
	"addon-manager": NewAddon([]*BinDataAsset{
		NewBinDataAsset(
			"deploy/addons/addon-manager.yaml",
			"/etc/kubernetes/manifests/",
			"addon-manager.yaml",
			"0640"),
	}, true, "addon-manager"),
	"dashboard": NewAddon([]*BinDataAsset{
		NewBinDataAsset(
			"deploy/addons/dashboard/dashboard-rc.yaml",
			constants.AddonsPath,
			"dashboard-rc.yaml",
			"0640"),
		NewBinDataAsset(
			"deploy/addons/dashboard/dashboard-svc.yaml",
			constants.AddonsPath,
			"dashboard-svc.yaml",
			"0640"),
	}, true, "dashboard"),
	"default-storageclass": NewAddon([]*BinDataAsset{
		NewBinDataAsset(
			"deploy/addons/storageclass/storageclass.yaml",
			constants.AddonsPath,
			"storageclass.yaml",
			"0640"),
	}, true, "default-storageclass"),
	"coredns": NewAddon([]*BinDataAsset{
		NewBinDataAsset(
			"deploy/addons/coredns/coreDNS-controller.yaml",
			constants.AddonsPath,
			"coreDNS-controller.yaml",
			"0640"),
		NewBinDataAsset(
			"deploy/addons/coredns/coreDNS-configmap.yaml",
			constants.AddonsPath,
			"coreDNS-configmap.yaml",
			"0640"),
		NewBinDataAsset(
			"deploy/addons/coredns/coreDNS-svc.yaml",
			constants.AddonsPath,
			"coreDNS-svc.yaml",
			"0640"),
		NewBinDataAsset(
			"deploy/addons/coredns/coreDNS-crbinding.yaml",
			constants.AddonsPath,
			"coreDNS-crbinding.yaml",
			"0640"),
		NewBinDataAsset(
			"deploy/addons/coredns/coreDNS-sa.yaml",
			constants.AddonsPath,
			"coreDNS-sa.yaml",
			"0640"),
		NewBinDataAsset(
			"deploy/addons/coredns/coreDNS-clusterrole.yaml",
			constants.AddonsPath,
			"coreDNS-clusterrole.yaml",
			"0640"),
	}, false, "coredns"),
	"kube-dns": NewAddon([]*BinDataAsset{
		NewBinDataAsset(
			"deploy/addons/kube-dns/kube-dns-controller.yaml",
			constants.AddonsPath,
			"kube-dns-controller.yaml",
			"0640"),
		NewBinDataAsset(
			"deploy/addons/kube-dns/kube-dns-cm.yaml",
			constants.AddonsPath,
			"kube-dns-cm.yaml",
			"0640"),
		NewBinDataAsset(
			"deploy/addons/kube-dns/kube-dns-svc.yaml",
			constants.AddonsPath,
			"kube-dns-svc.yaml",
			"0640"),
	}, true, "kube-dns"),
	"heapster": NewAddon([]*BinDataAsset{
		NewBinDataAsset(
			"deploy/addons/heapster/influxGrafana-rc.yaml",
			constants.AddonsPath,
			"influxGrafana-rc.yaml",
			"0640"),
		NewBinDataAsset(
			"deploy/addons/heapster/grafana-svc.yaml",
			constants.AddonsPath,
			"grafana-svc.yaml",
			"0640"),
		NewBinDataAsset(
			"deploy/addons/heapster/influxdb-svc.yaml",
			constants.AddonsPath,
			"influxdb-svc.yaml",
			"0640"),
		NewBinDataAsset(
			"deploy/addons/heapster/heapster-rc.yaml",
			constants.AddonsPath,
			"heapster-rc.yaml",
			"0640"),
		NewBinDataAsset(
			"deploy/addons/heapster/heapster-svc.yaml",
			constants.AddonsPath,
			"heapster-svc.yaml",
			"0640"),
	}, false, "heapster"),
	"efk": NewAddon([]*BinDataAsset{
		NewBinDataAsset(
			"deploy/addons/efk/elasticsearch-rc.yaml",
			constants.AddonsPath,
			"elasticsearch-rc.yaml",
			"0640"),
		NewBinDataAsset(
			"deploy/addons/efk/elasticsearch-svc.yaml",
			constants.AddonsPath,
			"elasticsearch-svc.yaml",
			"0640"),
		NewBinDataAsset(
			"deploy/addons/efk/fluentd-es-rc.yaml",
			constants.AddonsPath,
			"fluentd-es-rc.yaml",
			"0640"),
		NewBinDataAsset(
			"deploy/addons/efk/fluentd-es-configmap.yaml",
			constants.AddonsPath,
			"fluentd-es-configmap.yaml",
			"0640"),
		NewBinDataAsset(
			"deploy/addons/efk/kibana-rc.yaml",
			constants.AddonsPath,
			"kibana-rc.yaml",
			"0640"),
		NewBinDataAsset(
			"deploy/addons/efk/kibana-svc.yaml",
			constants.AddonsPath,
			"kibana-svc.yaml",
			"0640"),
	}, false, "efk"),
	"ingress": NewAddon([]*BinDataAsset{
		NewBinDataAsset(
			"deploy/addons/ingress/ingress-configmap.yaml",
			constants.AddonsPath,
			"ingress-configmap.yaml",
			"0640"),
		NewBinDataAsset(
			"deploy/addons/ingress/ingress-rc.yaml",
			constants.AddonsPath,
			"ingress-rc.yaml",
			"0640"),
		NewBinDataAsset(
			"deploy/addons/ingress/ingress-svc.yaml",
			constants.AddonsPath,
			"ingress-svc.yaml",
			"0640"),
	}, false, "ingress"),
	"registry": NewAddon([]*BinDataAsset{
		NewBinDataAsset(
			"deploy/addons/registry/registry-rc.yaml",
			constants.AddonsPath,
			"registry-rc.yaml",
			"0640"),
		NewBinDataAsset(
			"deploy/addons/registry/registry-svc.yaml",
			constants.AddonsPath,
			"registry-svc.yaml",
			"0640"),
	}, false, "registry"),
	"registry-creds": NewAddon([]*BinDataAsset{
		NewBinDataAsset(
			"deploy/addons/registry-creds/registry-creds-rc.yaml",
			constants.AddonsPath,
			"registry-creds-rc.yaml",
			"0640"),
	}, false, "registry-creds"),
}

func AddMinikubeDirToAssets(minipath string, vmpath string, assetList *[]CopyableFile) {
	// loop over $MINIKUBE_HOME/minipath and add them to assets
	searchDir := constants.MakeMiniPath(minipath)
	err := filepath.Walk(searchDir, func(miniFile string, f os.FileInfo, err error) error {
		isDir, err := util.IsDirectory(miniFile)
		if err == nil && !isDir {
			f, err := NewFileAsset(miniFile, vmpath, filepath.Base(miniFile), "0640")
			if err == nil {
				*assetList = append(*assetList, f)
			}
		} else if err != nil {
			glog.Infoln(fmt.Sprintf("Error encountered while walking %s: ", searchDir), err)
		}
		return nil
	})
	if err != nil {
		glog.Infoln(fmt.Sprintf("Error encountered while walking %s: ", searchDir), err)
	}
}
