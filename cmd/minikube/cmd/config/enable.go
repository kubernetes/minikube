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
	"errors"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/addons"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/mustload"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/minikube/style"
)

var addonsEnableCmd = &cobra.Command{
	Use:     "enable ADDON_NAME",
	Short:   "Enables the addon w/ADDON_NAME within minikube. For a list of available addons use: minikube addons list ",
	Long:    "Enables the addon w/ADDON_NAME within minikube. For a list of available addons use: minikube addons list ",
	Example: "minikube addons enable dashboard",
	Run: func(_ *cobra.Command, args []string) {
		if len(args) != 1 {
			exit.Message(reason.Usage, "usage: minikube addons enable ADDON_NAME")
		}
		_, cc := mustload.Partial(ClusterFlagValue())
		if cc.KubernetesConfig.KubernetesVersion == constants.NoKubernetesVersion {
			exit.Message(reason.Usage, "You cannot enable addons on a cluster without Kubernetes, to enable Kubernetes on your cluster, run: minikube start --kubernetes-version=stable")
		}

		err := addons.VerifyNotPaused(ClusterFlagValue(), true)
		if err != nil {
			exit.Error(reason.InternalAddonEnablePaused, "enabled failed", err)
		}
		addon := args[0]
		isDeprecated, replacement, msg := addons.Deprecations(addon)
		if isDeprecated && replacement == "" {
			exit.Message(reason.InternalAddonEnable, msg)
		} else if isDeprecated {
			out.Styled(style.Waiting, msg)
			addon = replacement
		}
		addonBundle, ok := assets.Addons[addon]
		if ok {
			maintainer := addonBundle.Maintainer
			if isOfficialMaintainer(maintainer) {
				out.Styled(style.Tip, `{{.addon}} is an addon maintained by {{.maintainer}}. For any concerns contact minikube on GitHub.
You can view the list of minikube maintainers at: https://github.com/kubernetes/minikube/blob/master/OWNERS`,
					out.V{"addon": addon, "maintainer": maintainer})
			} else {
				out.Styled(style.Warning, `{{.addon}} is a 3rd party addon and is not maintained or verified by minikube maintainers, enable at your own risk.`,
					out.V{"addon": addon})
				if addonBundle.VerifiedMaintainer != "" {
					out.Styled(style.Tip, `{{.addon}} is maintained by {{.maintainer}} for any concerns contact {{.verifiedMaintainer}} on GitHub.`,
						out.V{"addon": addon, "maintainer": maintainer, "verifiedMaintainer": addonBundle.VerifiedMaintainer})
				} else {
					out.Styled(style.Warning, `{{.addon}} does not currently have an associated maintainer.`,
						out.V{"addon": addon})
				}
			}
		}
		if images != "" {
			viper.Set(config.AddonImages, images)
		}
		if registries != "" {
			viper.Set(config.AddonRegistries, registries)
		}
		err = addons.SetAndSave(ClusterFlagValue(), addon, "true")
		if err != nil && !errors.Is(err, addons.ErrSkipThisAddon) {
			exit.Error(reason.InternalAddonEnable, "enable failed", err)
		}
		if err == nil {
			out.Step(style.AddonEnable, "The '{{.addonName}}' addon is enabled", out.V{"addonName": addon})
		}
	},
}

func isOfficialMaintainer(maintainer string) bool {
	// using map[string]struct{} as an empty struct occupies 0 bytes in memory
	officialMaintainers := map[string]struct{}{"Google": {}, "Kubernetes": {}, "minikube": {}}
	_, ok := officialMaintainers[maintainer]
	return ok
}

var (
	images     string
	registries string
)

func init() {
	addonsEnableCmd.Flags().StringVar(&images, "images", "", "Images used by this addon. Separated by commas.")
	addonsEnableCmd.Flags().StringVar(&registries, "registries", "", "Registries used by this addon. Separated by commas.")
	addonsEnableCmd.Flags().BoolVar(&addons.Force, "force", false, "If true, will perform potentially dangerous operations. Use with discretion.")
	addonsEnableCmd.Flags().BoolVar(&addons.Refresh, "refresh", false, "If true, pods might get deleted and restarted on addon enable")
	AddonsCmd.AddCommand(addonsEnableCmd)
}
