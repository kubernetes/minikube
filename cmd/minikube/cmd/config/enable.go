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
	"fmt"

	"github.com/blang/semver/v4"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/addons"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/minikube/style"
	"k8s.io/minikube/pkg/util"
)

var addonsEnableCmd = &cobra.Command{
	Use:     "enable ADDON_NAME",
	Short:   "Enables the addon w/ADDON_NAME within minikube. For a list of available addons use: minikube addons list ",
	Long:    "Enables the addon w/ADDON_NAME within minikube. For a list of available addons use: minikube addons list ",
	Example: "minikube addons enable dashboard",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			exit.Message(reason.Usage, "usage: minikube addons enable ADDON_NAME")
		}
		addon := args[0]
		// replace heapster as metrics-server because heapster is deprecated
		if addon == "heapster" {
			out.Styled(style.Waiting, "using metrics-server addon, heapster is deprecated")
			addon = "metrics-server"
		}
		if addon == "ambassador" {
			out.Styled(style.Warning, "The ambassador addon has stopped working as of v1.23.0, for more details visit: https://github.com/datawire/ambassador-operator/issues/73")
		}
		if addon == "olm" {
			out.Styled(style.Warning, "The OLM addon has stopped working, for more details visit: https://github.com/operator-framework/operator-lifecycle-manager/issues/2534")
		}
		addonBundle, ok := assets.Addons[addon]
		if ok {
			maintainer := addonBundle.Maintainer
			if maintainer == "Google" || maintainer == "Kubernetes" {
				out.Styled(style.Tip, `{{.addon}} is an addon maintained by {{.maintainer}}. For any concerns contact minikube on GitHub.
You can view the list of minikube maintainers at: https://github.com/kubernetes/minikube/blob/master/OWNERS`,
					out.V{"addon": addon, "maintainer": maintainer})
			} else {
				out.Styled(style.Warning, `{{.addon}} is a 3rd party addon and not maintained or verified by minikube maintainers, enable at your own risk.`,
					out.V{"addon": addon})
				if addonBundle.VerifiedMaintainer != "" {
					out.Styled(style.Tip, `{{.addon}} is maintained by {{.maintainer}} for any concerns contact {{.verifiedMaintainer}} on GitHub.`,
						out.V{"addon": addon, "maintainer": maintainer, "verifiedMaintainer": addonBundle.VerifiedMaintainer})
				}
			}
		}
		viper.Set(config.AddonImages, images)
		viper.Set(config.AddonRegistries, registries)
		err := addons.SetAndSave(ClusterFlagValue(), addon, "true")
		if err != nil && !errors.Is(err, addons.ErrSkipThisAddon) {
			exit.Error(reason.InternalAddonEnable, "enable failed", err)
		}
		if addon == "dashboard" {
			tipProfileArg := ""
			if ClusterFlagValue() != constants.DefaultClusterName {
				tipProfileArg = fmt.Sprintf(" -p %s", ClusterFlagValue())
			}
			out.Styled(style.Tip, `Some dashboard features require the metrics-server addon. To enable all features please run:

	minikube{{.profileArg}} addons enable metrics-server	

`, out.V{"profileArg": tipProfileArg})

		}
		if addon == "headlamp" {
			out.Styled(style.Tip, `To access Headlamp, use the following command:
minikube service headlamp -n headlamp

`)
			tokenGenerationTip := "To authenticate in Headlamp, fetch the Authentication Token using the following command:"
			createSvcAccountToken := "kubectl create token headlamp --duration 24h -n headlamp"
			getSvcAccountToken := `export SECRET=$(kubectl get secrets --namespace headlamp -o custom-columns=":metadata.name" | grep "headlamp-token")
kubectl get secret $SECRET --namespace headlamp --template=\{\{.data.token\}\} | base64 --decode`

			clusterName := ClusterFlagValue()
			clusterVersion := ClusterKubernetesVersion(clusterName)
			parsedClusterVersion, err := util.ParseKubernetesVersion(clusterVersion)
			if err != nil {
				tokenGenerationTip = fmt.Sprintf("%s\nIf Kubernetes Version is <1.24:\n%s\n\nIf Kubernetes Version is >=1.24:\n%s\n", tokenGenerationTip, createSvcAccountToken, getSvcAccountToken)
			} else {
				if parsedClusterVersion.GTE(semver.Version{Major: 1, Minor: 24}) {
					tokenGenerationTip = fmt.Sprintf("%s\n%s", tokenGenerationTip, createSvcAccountToken)
				} else {
					tokenGenerationTip = fmt.Sprintf("%s\n%s", tokenGenerationTip, getSvcAccountToken)
				}
			}
			out.Styled(style.Tip, fmt.Sprintf("%s\n", tokenGenerationTip))

			tipProfileArg := ""
			if clusterName != constants.DefaultClusterName {
				tipProfileArg = fmt.Sprintf(" -p %s", clusterName)
			}
			out.Styled(style.Tip, `Headlamp can display more detailed information when metrics-server is installed. To install it, run:

minikube{{.profileArg}} addons enable metrics-server	

`, out.V{"profileArg": tipProfileArg})

		}
		if err == nil {
			out.Step(style.AddonEnable, "The '{{.addonName}}' addon is enabled", out.V{"addonName": addon})
		}
	},
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
