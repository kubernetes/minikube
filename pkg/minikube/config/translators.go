/*
Copyright 2020 The Kubernetes Authors All rights reserved.

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

package config

import (
	"encoding/json"
	"errors"

	v152 "k8s.io/minikube/pkg/minikube/config/v152"
	v162 "k8s.io/minikube/pkg/minikube/config/v162"
)

func tryTranslate(vcontran []versionConfigTranslator, name string, miniHome ...string) (interface{}, error) {

	// Get the previous version translator
	previousVersion := vcontran[0]
	previousVersionConfig, err := previousVersion.TryLoadFromFile(name, miniHome...)
	// if the translator couldn't load the config from the file, then it's probably even an older version, so let's recurse again deeper
	if (err != nil ||
		(previousVersionConfig != nil && !previousVersion.IsValid(previousVersionConfig))) && len(vcontran) > 1 {
		var err error
		previousVersionConfig, err = tryTranslate(vcontran[1:], name, miniHome...)
		if err != nil {
			// Ah too bad, even the older versions couldn't translate it, this would bubble up to the end.
			return nil, err
		}
	}
	// Yes! The previous recurse iteration returned a successful and valid previousVersionConfig. Now let's translate it to the next version
	translatedConfig, err := previousVersion.TranslateToNextVersion(previousVersionConfig)
	return translatedConfig, err
}

var versionConfigTranslators = []versionConfigTranslator{
	{
		TryLoadFromFile: func(name string, miniHome ...string) (interface{}, error) {
			return v162.DefaultLoader.LoadConfigFromFile(name, miniHome...)
		},
		TranslateToNextVersion: func(config interface{}) (interface{}, error) {
			return translateFrom162ToNextVersion(config.(*v162.MachineConfig))
		},
		IsValid: func(config interface{}) bool {

			return v162.IsValid(config.(*v162.MachineConfig))
		},
	},
	{
		TryLoadFromFile: func(name string, miniHome ...string) (interface{}, error) {
			return v152.DefaultLoader.LoadConfigFromFile(name, miniHome...)
		},
		TranslateToNextVersion: func(config interface{}) (interface{}, error) {
			return translateFrom152ToNextVersion(config.(*v152.Config))
		},
		IsValid: func(config interface{}) bool {

			return v152.IsValid(config.(*v152.Config))
		},
	},
}

type versionConfigTranslator struct {
	TryLoadFromFile        tryLoadFromFile
	TranslateToNextVersion translateToNextVersion
	IsValid                isValid
}
type translateToNextVersion func(interface{}) (interface{}, error)
type tryLoadFromFile func(name string, miniHome ...string) (interface{}, error)
type isValid func(config interface{}) bool

func translateFrom162ToNextVersion(oldConfig *v162.MachineConfig) (*ClusterConfig, error) {

	hypervUseExternalSwitch := false

	if oldConfig.HypervVirtualSwitch != "" && oldConfig.HypervVirtualSwitch == "true" {
		hypervUseExternalSwitch = true
	}
	oldConfigBytes, err := json.Marshal(oldConfig)
	if err != nil {
		return nil, err
	}

	var newConfig ClusterConfig

	errorUnmarshalling := json.Unmarshal(oldConfigBytes, &newConfig)

	if errorUnmarshalling != nil {
		return nil, errorUnmarshalling
	}
	newConfig.HypervUseExternalSwitch = hypervUseExternalSwitch

	newConfig.Driver = oldConfig.VMDriver
	newConfig.Nodes = []Node{
		{
			Name:              oldConfig.KubernetesConfig.NodeName,
			IP:                oldConfig.KubernetesConfig.NodeIP,
			Port:              oldConfig.KubernetesConfig.NodePort,
			KubernetesVersion: oldConfig.KubernetesConfig.KubernetesVersion,
		},
	}
	return &newConfig, nil

}

func translateFrom152ToNextVersion(oldConfig *v152.Config) (*v162.MachineConfig, error) {

	// The structure of the config from version 1.5.2 was different. It was split from the root to two properties: MachineConfig and KubernetesConfig
	// so we have to accommodate to the next version which flattened the MachineConfig to the root, and made KubernetesConfig a property

	if oldConfig == nil {
		return nil, errors.New("oldConfig is nil")
	}

	oldMachineConfigBytes, err := json.Marshal(oldConfig.MachineConfig)
	if err != nil {
		return nil, err
	}

	var newConfig v162.MachineConfig

	errorUnmarshalling := json.Unmarshal(oldMachineConfigBytes, &newConfig)
	if errorUnmarshalling != nil {
		return nil, errorUnmarshalling
	}

	var newKubernetesConfig v162.KubernetesConfig
	oldKubernetesConfigBytes, err := json.Marshal(oldConfig.KubernetesConfig)

	if err != nil {
		return nil, err
	}

	errorUnmarshallingKubernetesConfig := json.Unmarshal(oldKubernetesConfigBytes, &newKubernetesConfig)
	if errorUnmarshallingKubernetesConfig != nil {
		return nil, errorUnmarshallingKubernetesConfig
	}

	newConfig.KubernetesConfig = newKubernetesConfig

	//TODO: do real translation here, find the difference between the old and the new configs and re-assign the properties
	return &newConfig, nil

}
