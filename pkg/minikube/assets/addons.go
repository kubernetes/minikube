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
	"path"
	"path/filepath"
	"strconv"

	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/util"
)

// Addon is a named list of assets, that can be enabled
type Addon struct {
	Assets    []*BinDataAsset
	enabled   bool
	addonName string
}

// NewAddon creates a new Addon
func NewAddon(assets []*BinDataAsset, enabled bool, addonName string) *Addon {
	a := &Addon{
		Assets:    assets,
		enabled:   enabled,
		addonName: addonName,
	}
	return a
}

// IsEnabled checks if an Addon is enabled
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

// Addons is the list of addons
var Addons = map[string]*Addon{
	"addon-manager": NewAddon([]*BinDataAsset{
		NewBinDataAsset(
			"deploy/addons/addon-manager.yaml",
			"/etc/kubernetes/manifests/",
			"addon-manager.yaml",
			"0640",
			true),
	}, true, "addon-manager"),
	"dashboard": NewAddon([]*BinDataAsset{
		NewBinDataAsset(
			"deploy/addons/dashboard/dashboard-dp.yaml",
			constants.AddonsPath,
			"dashboard-dp.yaml",
			"0640",
			true),
		NewBinDataAsset(
			"deploy/addons/dashboard/dashboard-svc.yaml",
			constants.AddonsPath,
			"dashboard-svc.yaml",
			"0640",
			false),
	}, false, "dashboard"),
	"default-storageclass": NewAddon([]*BinDataAsset{
		NewBinDataAsset(
			"deploy/addons/storageclass/storageclass.yaml",
			constants.AddonsPath,
			"storageclass.yaml",
			"0640",
			false),
	}, true, "default-storageclass"),
	"storage-provisioner": NewAddon([]*BinDataAsset{
		NewBinDataAsset(
			"deploy/addons/storage-provisioner/storage-provisioner.yaml",
			constants.AddonsPath,
			"storage-provisioner.yaml",
			"0640",
			true),
	}, true, "storage-provisioner"),
	"storage-provisioner-gluster": NewAddon([]*BinDataAsset{
		NewBinDataAsset(
			"deploy/addons/storage-provisioner-gluster/storage-gluster-ns.yaml",
			constants.AddonsPath,
			"storage-gluster-ns.yaml",
			"0640",
			false),
		NewBinDataAsset(
			"deploy/addons/storage-provisioner-gluster/glusterfs-daemonset.yaml",
			constants.AddonsPath,
			"glusterfs-daemonset.yaml",
			"0640",
			false),
		NewBinDataAsset(
			"deploy/addons/storage-provisioner-gluster/heketi-deployment.yaml",
			constants.AddonsPath,
			"heketi-deployment.yaml",
			"0640",
			false),
		NewBinDataAsset(
			"deploy/addons/storage-provisioner-gluster/storage-provisioner-glusterfile.yaml",
			constants.AddonsPath,
			"storage-privisioner-glusterfile.yaml",
			"0640",
			false),
	}, false, "storage-provisioner-gluster"),
	"heapster": NewAddon([]*BinDataAsset{
		NewBinDataAsset(
			"deploy/addons/heapster/influx-grafana-rc.yaml",
			constants.AddonsPath,
			"influxGrafana-rc.yaml",
			"0640",
			true),
		NewBinDataAsset(
			"deploy/addons/heapster/grafana-svc.yaml",
			constants.AddonsPath,
			"grafana-svc.yaml",
			"0640",
			false),
		NewBinDataAsset(
			"deploy/addons/heapster/influxdb-svc.yaml",
			constants.AddonsPath,
			"influxdb-svc.yaml",
			"0640",
			false),
		NewBinDataAsset(
			"deploy/addons/heapster/heapster-rc.yaml",
			constants.AddonsPath,
			"heapster-rc.yaml",
			"0640",
			true),
		NewBinDataAsset(
			"deploy/addons/heapster/heapster-svc.yaml",
			constants.AddonsPath,
			"heapster-svc.yaml",
			"0640",
			false),
	}, false, "heapster"),
	"efk": NewAddon([]*BinDataAsset{
		NewBinDataAsset(
			"deploy/addons/efk/elasticsearch-rc.yaml",
			constants.AddonsPath,
			"elasticsearch-rc.yaml",
			"0640",
			true),
		NewBinDataAsset(
			"deploy/addons/efk/elasticsearch-svc.yaml",
			constants.AddonsPath,
			"elasticsearch-svc.yaml",
			"0640",
			false),
		NewBinDataAsset(
			"deploy/addons/efk/fluentd-es-rc.yaml",
			constants.AddonsPath,
			"fluentd-es-rc.yaml",
			"0640",
			true),
		NewBinDataAsset(
			"deploy/addons/efk/fluentd-es-configmap.yaml",
			constants.AddonsPath,
			"fluentd-es-configmap.yaml",
			"0640",
			false),
		NewBinDataAsset(
			"deploy/addons/efk/kibana-rc.yaml",
			constants.AddonsPath,
			"kibana-rc.yaml",
			"0640",
			false),
		NewBinDataAsset(
			"deploy/addons/efk/kibana-svc.yaml",
			constants.AddonsPath,
			"kibana-svc.yaml",
			"0640",
			false),
	}, false, "efk"),
	"ingress": NewAddon([]*BinDataAsset{
		NewBinDataAsset(
			"deploy/addons/ingress/ingress-configmap.yaml",
			constants.AddonsPath,
			"ingress-configmap.yaml",
			"0640",
			false),
		NewBinDataAsset(
			"deploy/addons/ingress/ingress-rbac.yaml",
			constants.AddonsPath,
			"ingress-rbac.yaml",
			"0640",
			false),
		NewBinDataAsset(
			"deploy/addons/ingress/ingress-dp.yaml",
			constants.AddonsPath,
			"ingress-dp.yaml",
			"0640",
			true),
		NewBinDataAsset(
			"deploy/addons/ingress/ingress-svc.yaml",
			constants.AddonsPath,
			"ingress-svc.yaml",
			"0640",
			false),
	}, false, "ingress"),
	"metrics-server": NewAddon([]*BinDataAsset{
		NewBinDataAsset(
			"deploy/addons/metrics-server/metrics-apiservice.yaml",
			constants.AddonsPath,
			"metrics-apiservice.yaml",
			"0640",
			false),
		NewBinDataAsset(
			"deploy/addons/metrics-server/metrics-server-deployment.yaml",
			constants.AddonsPath,
			"metrics-server-deployment.yaml",
			"0640",
			true),
		NewBinDataAsset(
			"deploy/addons/metrics-server/metrics-server-service.yaml",
			constants.AddonsPath,
			"metrics-server-service.yaml",
			"0640",
			false),
	}, false, "metrics-server"),
	"registry": NewAddon([]*BinDataAsset{
		NewBinDataAsset(
			"deploy/addons/registry/registry-rc.yaml",
			constants.AddonsPath,
			"registry-rc.yaml",
			"0640",
			false),
		NewBinDataAsset(
			"deploy/addons/registry/registry-svc.yaml",
			constants.AddonsPath,
			"registry-svc.yaml",
			"0640",
			false),
	}, false, "registry"),
	"registry-creds": NewAddon([]*BinDataAsset{
		NewBinDataAsset(
			"deploy/addons/registry-creds/registry-creds-rc.yaml",
			constants.AddonsPath,
			"registry-creds-rc.yaml",
			"0640",
			false),
	}, false, "registry-creds"),
	"freshpod": NewAddon([]*BinDataAsset{
		NewBinDataAsset(
			"deploy/addons/freshpod/freshpod-rc.yaml",
			constants.AddonsPath,
			"freshpod-rc.yaml",
			"0640",
			true),
	}, false, "freshpod"),
	"nvidia-driver-installer": NewAddon([]*BinDataAsset{
		NewBinDataAsset(
			"deploy/addons/gpu/nvidia-driver-installer.yaml",
			constants.AddonsPath,
			"nvidia-driver-installer.yaml",
			"0640",
			true),
	}, false, "nvidia-driver-installer"),
	"nvidia-gpu-device-plugin": NewAddon([]*BinDataAsset{
		NewBinDataAsset(
			"deploy/addons/gpu/nvidia-gpu-device-plugin.yaml",
			constants.AddonsPath,
			"nvidia-gpu-device-plugin.yaml",
			"0640",
			true),
	}, false, "nvidia-gpu-device-plugin"),
	"logviewer": NewAddon([]*BinDataAsset{
		NewBinDataAsset(
			"deploy/addons/logviewer/logviewer-dp-and-svc.yaml",
			constants.AddonsPath,
			"logviewer-dp-and-svc.yaml",
			"0640",
			false),
		NewBinDataAsset(
			"deploy/addons/logviewer/logviewer-rbac.yaml",
			constants.AddonsPath,
			"logviewer-rbac.yaml",
			"0640",
			false),
	}, false, "logviewer"),
	"gvisor": NewAddon([]*BinDataAsset{
		NewBinDataAsset(
			"deploy/addons/gvisor/gvisor-pod.yaml",
			constants.AddonsPath,
			"gvisor-pod.yaml",
			"0640",
			true),
		NewBinDataAsset(
			"deploy/addons/gvisor/gvisor-config.toml",
			constants.GvisorFilesPath,
			constants.GvisorConfigTomlTargetName,
			"0640",
			true),
		NewBinDataAsset(
			"deploy/addons/gvisor/gvisor-containerd-shim.toml",
			constants.GvisorFilesPath,
			constants.GvisorContainerdShimTargetName,
			"0640",
			false),
	}, false, "gvisor"),
}

// AddMinikubeDirAssets adds all addons and files to the list
// of files to be copied to the vm.
func AddMinikubeDirAssets(assets *[]CopyableFile) error {
	if err := addMinikubeDirToAssets(constants.MakeMiniPath("addons"), constants.AddonsPath, assets); err != nil {
		return errors.Wrap(err, "adding addons folder to assets")
	}
	if err := addMinikubeDirToAssets(constants.MakeMiniPath("files"), "", assets); err != nil {
		return errors.Wrap(err, "adding files rootfs to assets")
	}

	return nil
}

// AddMinikubeDirToAssets adds all the files in the basedir argument to the list
// of files to be copied to the vm.  If vmpath is left blank, the files will be
// transferred to the location according to their relative minikube folder path.
func addMinikubeDirToAssets(basedir, vmpath string, assets *[]CopyableFile) error {
	err := filepath.Walk(basedir, func(hostpath string, info os.FileInfo, err error) error {
		isDir, err := util.IsDirectory(hostpath)
		if err != nil {
			return errors.Wrapf(err, "checking if %s is directory", hostpath)
		}
		if !isDir {
			vmdir := vmpath
			if vmdir == "" {
				rPath, err := filepath.Rel(basedir, hostpath)
				if err != nil {
					return errors.Wrap(err, "generating relative path")
				}
				rPath = filepath.Dir(rPath)
				rPath = filepath.ToSlash(rPath)
				vmdir = path.Join("/", rPath)
			}
			permString := fmt.Sprintf("%o", info.Mode().Perm())
			// The conversion will strip the leading 0 if present, so add it back
			// if we need to.
			if len(permString) == 3 {
				permString = fmt.Sprintf("0%s", permString)
			}

			f, err := NewFileAsset(hostpath, vmdir, filepath.Base(hostpath), permString)
			if err != nil {
				return errors.Wrapf(err, "creating file asset for %s", hostpath)
			}
			*assets = append(*assets, f)
		}

		return nil
	})
	if err != nil {
		return errors.Wrap(err, "walking filepath")
	}
	return nil
}

// GenerateTemplateData generates template data for template assets
func GenerateTemplateData(cfg config.KubernetesConfig) interface{} {
	opts := struct {
		ImageRepository string
	}{
		ImageRepository: cfg.ImageRepository,
	}

	return opts
}
