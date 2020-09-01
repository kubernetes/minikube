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
	"io/ioutil"
	"net"

	"github.com/spf13/cobra"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/mustload"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/minikube/service"
	"k8s.io/minikube/pkg/minikube/style"
)

var addonsConfigureCmd = &cobra.Command{
	Use:   "configure ADDON_NAME",
	Short: "Configures the addon w/ADDON_NAME within minikube (example: minikube addons configure registry-creds). For a list of available addons use: minikube addons list ",
	Long:  "Configures the addon w/ADDON_NAME within minikube (example: minikube addons configure registry-creds). For a list of available addons use: minikube addons list ",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			exit.Message(reason.Usage, "usage: minikube addons configure ADDON_NAME")
		}

		addon := args[0]
		// allows for additional prompting of information when enabling addons
		switch addon {
		case "registry-creds":
			posResponses := []string{"yes", "y"}
			negResponses := []string{"no", "n"}

			// Default values
			awsAccessID := "changeme"
			awsAccessKey := "changeme"
			awsSessionToken := ""
			awsRegion := "changeme"
			awsAccount := "changeme"
			awsRole := "changeme"
			gcrApplicationDefaultCredentials := "changeme"
			dockerServer := "changeme"
			dockerUser := "changeme"
			dockerPass := "changeme"
			gcrURL := "https://gcr.io"
			acrURL := "changeme"
			acrClientID := "changeme"
			acrPassword := "changeme"

			enableAWSECR := AskForYesNoConfirmation("\nDo you want to enable AWS Elastic Container Registry?", posResponses, negResponses)
			if enableAWSECR {
				awsAccessID = AskForStaticValue("-- Enter AWS Access Key ID: ")
				awsAccessKey = AskForStaticValue("-- Enter AWS Secret Access Key: ")
				awsSessionToken = AskForStaticValueOptional("-- (Optional) Enter AWS Session Token: ")
				awsRegion = AskForStaticValue("-- Enter AWS Region: ")
				awsAccount = AskForStaticValue("-- Enter 12 digit AWS Account ID (Comma separated list): ")
				awsRole = AskForStaticValueOptional("-- (Optional) Enter ARN of AWS role to assume: ")
			}

			enableGCR := AskForYesNoConfirmation("\nDo you want to enable Google Container Registry?", posResponses, negResponses)
			if enableGCR {
				gcrPath := AskForStaticValue("-- Enter path to credentials (e.g. /home/user/.config/gcloud/application_default_credentials.json):")
				gcrchangeURL := AskForYesNoConfirmation("-- Do you want to change the GCR URL (Default https://gcr.io)?", posResponses, negResponses)

				if gcrchangeURL {
					gcrURL = AskForStaticValue("-- Enter GCR URL (e.g. https://asia.gcr.io):")
				}

				// Read file from disk
				dat, err := ioutil.ReadFile(gcrPath)

				if err != nil {
					out.FailureT("Error reading {{.path}}: {{.error}}", out.V{"path": gcrPath, "error": err})
				} else {
					gcrApplicationDefaultCredentials = string(dat)
				}
			}

			enableDR := AskForYesNoConfirmation("\nDo you want to enable Docker Registry?", posResponses, negResponses)
			if enableDR {
				dockerServer = AskForStaticValue("-- Enter docker registry server url: ")
				dockerUser = AskForStaticValue("-- Enter docker registry username: ")
				dockerPass = AskForPasswordValue("-- Enter docker registry password: ")
			}

			enableACR := AskForYesNoConfirmation("\nDo you want to enable Azure Container Registry?", posResponses, negResponses)
			if enableACR {
				acrURL = AskForStaticValue("-- Enter Azure Container Registry (ACR) URL: ")
				acrClientID = AskForStaticValue("-- Enter client ID (service principal ID) to access ACR: ")
				acrPassword = AskForPasswordValue("-- Enter service principal password to access Azure Container Registry: ")
			}

			cname := ClusterFlagValue()

			// Create ECR Secret
			err := service.CreateSecret(
				cname,
				"kube-system",
				"registry-creds-ecr",
				map[string]string{
					"AWS_ACCESS_KEY_ID":     awsAccessID,
					"AWS_SECRET_ACCESS_KEY": awsAccessKey,
					"AWS_SESSION_TOKEN":     awsSessionToken,
					"aws-account":           awsAccount,
					"aws-region":            awsRegion,
					"aws-assume-role":       awsRole,
				},
				map[string]string{
					"app":                           "registry-creds",
					"cloud":                         "ecr",
					"kubernetes.io/minikube-addons": "registry-creds",
				})
			if err != nil {
				out.FailureT("ERROR creating `registry-creds-ecr` secret: {{.error}}", out.V{"error": err})
			}

			// Create GCR Secret
			err = service.CreateSecret(
				cname,
				"kube-system",
				"registry-creds-gcr",
				map[string]string{
					"application_default_credentials.json": gcrApplicationDefaultCredentials,
					"gcrurl":                               gcrURL,
				},
				map[string]string{
					"app":                           "registry-creds",
					"cloud":                         "gcr",
					"kubernetes.io/minikube-addons": "registry-creds",
				})

			if err != nil {
				out.FailureT("ERROR creating `registry-creds-gcr` secret: {{.error}}", out.V{"error": err})
			}

			// Create Docker Secret
			err = service.CreateSecret(
				cname,
				"kube-system",
				"registry-creds-dpr",
				map[string]string{
					"DOCKER_PRIVATE_REGISTRY_SERVER":   dockerServer,
					"DOCKER_PRIVATE_REGISTRY_USER":     dockerUser,
					"DOCKER_PRIVATE_REGISTRY_PASSWORD": dockerPass,
				},
				map[string]string{
					"app":                           "registry-creds",
					"cloud":                         "dpr",
					"kubernetes.io/minikube-addons": "registry-creds",
				})

			if err != nil {
				out.WarningT("ERROR creating `registry-creds-dpr` secret")
			}

			// Create Azure Container Registry Secret
			err = service.CreateSecret(
				cname,
				"kube-system",
				"registry-creds-acr",
				map[string]string{
					"ACR_URL":       acrURL,
					"ACR_CLIENT_ID": acrClientID,
					"ACR_PASSWORD":  acrPassword,
				},
				map[string]string{
					"app":                           "registry-creds",
					"cloud":                         "acr",
					"kubernetes.io/minikube-addons": "registry-creds",
				})

			if err != nil {
				out.WarningT("ERROR creating `registry-creds-acr` secret")
			}

		case "metallb":
			profile := ClusterFlagValue()
			_, cfg := mustload.Partial(profile)

			validator := func(s string) bool {
				return net.ParseIP(s) != nil
			}

			if cfg.KubernetesConfig.LoadBalancerStartIP == "" {
				cfg.KubernetesConfig.LoadBalancerStartIP = AskForStaticValidatedValue("-- Enter Load Balancer Start IP: ", validator)
			}

			if cfg.KubernetesConfig.LoadBalancerEndIP == "" {
				cfg.KubernetesConfig.LoadBalancerEndIP = AskForStaticValidatedValue("-- Enter Load Balancer End IP: ", validator)
			}

			if err := config.SaveProfile(profile, cfg); err != nil {
				out.ErrT(style.Fatal, "Failed to save config {{.profile}}", out.V{"profile": profile})
			}

		default:
			out.FailureT("{{.name}} has no available configuration options", out.V{"name": addon})
			return
		}

		out.SuccessT("{{.name}} was successfully configured", out.V{"name": addon})
	},
}

func init() {
	AddonsCmd.AddCommand(addonsConfigureCmd)
}
