package config

import (
	"encoding/json"

	v152 "k8s.io/minikube/pkg/minikube/config/v152"
	v162 "k8s.io/minikube/pkg/minikube/config/v162"
)

func tryTranslate(vcontran []VersionConfigTranslator, name string, miniHome ...string) (interface{}, error) {

	// Get the previous version translator
	previousVersion := vcontran[0]
	previousVersionConfig, err := previousVersion.TryLoadFromFile(name, miniHome...)
	// if the translator couldn't load the config from the file, then it's probably even an older version, so let's recurse again deeper
	if (err != nil || !previousVersion.IsValid(previousVersionConfig)) && len(vcontran) > 1 {
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

var versionConfigTranslators = []VersionConfigTranslator{
	{
		TryLoadFromFile: func(name string, miniHome ...string) (interface{}, error) {
			return v162.DefaultLoader.LoadConfigFromFile(name, miniHome...)
		},
		TranslateToNextVersion: func(config interface{}) (interface{}, error) {
			return translateFrom163ToNextVersion(config.(v162.MachineConfig))
		},
	},
	{
		TryLoadFromFile: func(name string, miniHome ...string) (interface{}, error) {
			return v152.DefaultLoader.LoadConfigFromFile(name, miniHome...)
		},
		TranslateToNextVersion: func(config interface{}) (interface{}, error) {
			return translateFrom152ToNextVersion(config.(v162.MachineConfig))
		},
	},
}

type VersionConfigTranslator struct {
	TryLoadFromFile        tryLoadFromFile
	TranslateToNextVersion translateToNextVersion
	IsValid                isValid
}
type translateToNextVersion func(interface{}) (interface{}, error)
type tryLoadFromFile func(name string, miniHome ...string) (interface{}, error)
type isValid func(interface{}) bool

func translateFrom163ToNextVersion(oldConfig v162.MachineConfig) (*ClusterConfig, error) {

	hypervUseExternalSwitch := false

	if oldConfig.HypervVirtualSwitch != "" && oldConfig.HypervVirtualSwitch == "true" {
		hypervUseExternalSwitch = true
	}
	oldConfigBytes, err := json.Marshal(oldConfig)
	if err != nil {
		return nil, err
	}

	var newConfig *ClusterConfig

	errorUnmarshalling := json.Unmarshal(oldConfigBytes, newConfig)

	if errorUnmarshalling != nil {
		return nil, errorUnmarshalling
	}
	newConfig.HypervUseExternalSwitch = hypervUseExternalSwitch
	return newConfig, nil

}

func translateFrom152ToNextVersion(oldConfig v162.MachineConfig) (*ClusterConfig, error) {

	oldConfigBytes, err := json.Marshal(oldConfig)
	if err != nil {
		return nil, err
	}

	var newConfig *ClusterConfig

	errorUnmarshalling := json.Unmarshal(oldConfigBytes, newConfig)
	if errorUnmarshalling != nil {
		return nil, errorUnmarshalling
	}

	//TODO: do real translation here, find the difference between the old and the new configs and re-assign the properties
	return newConfig, nil

}
