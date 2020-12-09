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

	"github.com/spf13/cobra"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/kubeconfig"
	"k8s.io/minikube/pkg/minikube/mustload"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/minikube/style"
)

// ProfileCmd represents the profile command
var ProfileCmd = &cobra.Command{
	Use:   "profile [MINIKUBE_PROFILE_NAME].  You can return to the default minikube profile by running `minikube profile default`",
	Short: "Get or list the current profiles (clusters)",
	Long:  "profile sets the current minikube profile, or gets the current profile if no arguments are provided.  This is used to run and manage multiple minikube instance.  You can return to the default minikube profile by running `minikube profile default`",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			profile := ClusterFlagValue()
			out.Step(style.Empty, out.NoSpinner, profile)
			os.Exit(0)
		}

		if len(args) > 1 {
			exit.Message(reason.Usage, "usage: minikube profile [MINIKUBE_PROFILE_NAME]")
		}

		profile := args[0]
		// Check whether the profile name is container friendly
		if !config.ProfileNameValid(profile) {
			out.WarningT("Profile name '{{.profilename}}' is not valid", out.V{"profilename": profile})
			exit.Message(reason.Usage, "Only alphanumeric and dashes '-' are permitted. Minimum 1 character, starting with alphanumeric.")
		}
		/**
		we need to add code over here to check whether the profile
		name is in the list of reserved keywords
		*/
		if config.ProfileNameInReservedKeywords(profile) {
			exit.Message(reason.InternalReservedProfile, `Profile name "{{.profilename}}" is reserved keyword. To delete this profile, run: "{{.cmd}}"`, out.V{"profilename": profile, "cmd": mustload.ExampleCmd(profile, "delete")})
		}

		if profile == "default" {
			profile = "minikube"
		} else {
			// not validating when it is default profile
			errProfile, ok := ValidateProfile(profile)
			if !ok && errProfile != nil {
				out.FailureT(errProfile.Msg)
			}
		}

		if !config.ProfileExists(profile) {
			out.ErrT(style.Tip, `if you want to create a profile you can by this command: minikube start -p {{.profile_name}}`, out.V{"profile_name": profile})
			os.Exit(0)
		}

		err := Set(config.ProfileName, profile)
		if err != nil {
			exit.Error(reason.InternalConfigSet, "Setting profile failed", err)
		}
		cc, err := config.Load(profile)
		// might err when loading older version of cfg file that doesn't have KeepContext field
		if err != nil && !config.IsNotExist(err) {
			out.ErrT(style.Sad, `Error loading profile config: {{.error}}`, out.V{"error": err})
		}
		if err == nil {
			if cc.KeepContext {
				out.SuccessT("Skipped switching kubectl context for {{.profile_name}} because --keep-context was set.", out.V{"profile_name": profile})
				out.SuccessT("To connect to this cluster, use: kubectl --context={{.profile_name}}", out.V{"profile_name": profile})
			} else {
				err := kubeconfig.SetCurrentContext(profile, kubeconfig.PathFromEnv())
				if err != nil {
					out.ErrT(style.Sad, `Error while setting kubectl current context :  {{.error}}`, out.V{"error": err})
				}
			}
			out.SuccessT("minikube profile was successfully set to {{.profile_name}}", out.V{"profile_name": profile})
		}
	},
}
