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
	"os"
        "io/ioutil"
        "strings"
	
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	pkgConfig "k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/console"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/exit"
	pkgutil "k8s.io/minikube/pkg/util"
)

var (
       profileList     bool
       profileSet              string
)



// ProfileCmd represents the profile command
var ProfileCmd = &cobra.Command{
	Use:   "profile [MINIKUBE_PROFILE_NAME].  You can return to the default minikube profile by running `minikube profile default`",
	Short: "Profile gets or sets the current minikube profile",
	Long:  "profile sets the current minikube profile, or gets the current profile if no arguments are provided.  This is used to run and manage multiple minikube instance.  You can return to the default minikube profile by running `minikube profile default`",
	Run: func(cmd *cobra.Command, args []string) {
		if profileList {
			var profilePathStr strings.Builder
			profilePathStr.WriteString(constants.GetMinipath())
			profilePathStr.WriteString("/profiles")
			files, err := ioutil.ReadDir(profilePathStr.String())
			if err != nil {
				exit.WithError("Getting profile failed", err)
			} 
			console.OutLn("Available profiles are:")
			for _,file := range files {
				console.OutLn("%s", file.Name())
			}
			os.Exit(0)
		} 

		if (profileSet == "") && (len(args) == 0) {
			profile := viper.GetString(pkgConfig.MachineProfile)
			console.OutLn("Current Profile: %s", profile)
			os.Exit(0)
		}

		if ((profileSet != "") && (len(args) > 1)) {
			exit.Usage("usage: minikube profile -l | -s [MINIKUBE_PROFILE_NAME] ")
		}

		profile := profileSet
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
