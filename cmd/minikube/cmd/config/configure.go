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
	"fmt"
	"io/ioutil"
	"os"

	"k8s.io/minikube/pkg/minikube/service"

	"github.com/spf13/cobra"
)

var addonsConfigureCmd = &cobra.Command{
	Use:   "configure ADDON_NAME",
	Short: "Configures the addon w/ADDON_NAME within minikube (example: minikube addons configure registry-creds). For a list of available addons use: minikube addons list ",
	Long:  "Configures the addon w/ADDON_NAME within minikube (example: minikube addons configure registry-creds). For a list of available addons use: minikube addons list ",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Fprintln(os.Stderr, "usage: minikube addons configure ADDON_NAME")
			os.Exit(1)
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
					fmt.Println("Could not read file for application_default_credentials.json")
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

			// Create ECR Secret
			err := service.CreateSecret(
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
				fmt.Println("ERROR creating `registry-creds-ecr` secret")
			}

			// Create GCR Secret
			err = service.CreateSecret(
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
				fmt.Println("ERROR creating `registry-creds-gcr` secret")
			}

			// Create Docker Secret
			err = service.CreateSecret(
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
				fmt.Println("ERROR creating `registry-creds-dpr` secret")
			}

			break
		default:
			fmt.Fprintln(os.Stdout, fmt.Sprintf("%s has no available configuration options", addon))
			return
		}

		fmt.Fprintln(os.Stdout, fmt.Sprintf("%s was successfully configured", addon))
	},
}

func init() {
	AddonsCmd.AddCommand(addonsConfigureCmd)
}
