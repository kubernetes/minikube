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

package config

import (
	"encoding/json"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io/ioutil"
	minikubeConfig "k8s.io/minikube/pkg/minikube/config"
	pkgConfig "k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/console"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/exit"
	pkgutil "k8s.io/minikube/pkg/util"
	"os"
	"path/filepath"
	"reflect"
)

// ProfileCmd represents the profile command
var ProfileCmd = &cobra.Command{
	Use:   "profile [MINIKUBE_PROFILE_NAME].  You can return to the default minikube profile by running `minikube profile default`",
	Short: "Profile gets or sets the current minikube profile",
	Long:  "profile sets the current minikube profile, or gets the current profile if no arguments are provided.  This is used to run and manage multiple minikube instance.  You can return to the default minikube profile by running `minikube profile default`",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			profile := viper.GetString(pkgConfig.MachineProfile)
			console.OutLn("%s", profile)
			os.Exit(0)
		}

		if len(args) > 1 {
			exit.Usage("usage: minikube profile [MINIKUBE_PROFILE_NAME]")
		}

		profile := args[0]
		if profile == "default" {
			profile = "minikube"
		}
		err := Set(pkgConfig.MachineProfile, profile)
		if err != nil {
			exit.WithError("Setting profile failed", err)
		}
		cc, err := pkgConfig.Load()
		// might err when loading older version of cfg file that doesn't have KeepContext field
		if err != nil && !os.IsNotExist(err) {
			console.ErrLn("Error loading profile config: %v", err)
		}
		if err == nil {
			if cc.MachineConfig.KeepContext {
				console.Success("Skipped switching kubectl context for %s , because --keep-context", profile)
				console.Success("To connect to this cluster, use: kubectl --context=%s", profile)
			} else {
				err := pkgutil.SetCurrentContext(constants.KubeconfigPath, profile)
				if err != nil {
					console.ErrLn("Error while setting kubectl current context :  %v", err)
				}
			}
		}
		console.Success("minikube profile was successfully set to %s", profile)
	},
}

func GetAllProfiles() []string {
	miniPath := constants.GetMinipath()
	profilesPath := filepath.Join(miniPath, "profiles")
	fileInfos, err := ioutil.ReadDir(profilesPath)
	if err != nil {
		console.ErrLn("Unable to list in dir: %s \n Error: %v", profilesPath, err)
	}

	var profiles []string
	for _, fileInfo := range fileInfos {
		if fileInfo.IsDir() {
			profilePath := filepath.Join(profilesPath, fileInfo.Name())
			if isValidProfile(profilePath) {
				profiles = append(profiles, fileInfo.Name())
			}
		}
	}
	return profiles
}

func isValidProfile(profilePath string) bool {
	fileInfos, err := ioutil.ReadDir(profilePath)
	if err != nil {
		console.ErrLn("Unable to list in dir: %s \n Error: %v", profilePath, err)
	}

	hasConfigJson := false
	for _, fileInfo := range fileInfos {
		if fileInfo.Name() == "config.json" {
			hasConfigJson = true
		}
	}

	if !hasConfigJson {
		return false
	}

	// TODO: Use constants?
	profileConfigPath := filepath.Join(profilePath, "config.json")
	bytes, err := ioutil.ReadFile(profileConfigPath)
	if err != nil {
		console.ErrLn("Unable to read file: %s \n Error: %v", profileConfigPath, err)
	}

	var configObject minikubeConfig.Config
	errUnmarshal := json.Unmarshal(bytes, &configObject)

	if errUnmarshal != nil {
		console.ErrLn("Could not unmarshal config json to config object: %s \n Error: %v", profileConfigPath, err)
	}
	return IsProfileConfigValid(configObject)
}

func IsProfileConfigValid(configObject minikubeConfig.Config) bool {
	machineConfig := configObject.MachineConfig
	kubernetesConfig := configObject.KubernetesConfig
	if reflect.DeepEqual(machineConfig, minikubeConfig.MachineConfig{}) || reflect.DeepEqual(kubernetesConfig, minikubeConfig.KubernetesConfig{}) {
		return false
	}

	//TODO: Validate MachineConfig and KubernetesConfig?

	return true
}
