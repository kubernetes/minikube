/*
Copyright 2025 The Kubernetes Authors All rights reserved.

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

	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/minikube/service"
)

const configDefaultValue = "changeme"

// Top level configs for RegistryCreds addons
type registryCredsAddonConfig struct {
	EnableAWSEcr string                         `json:"enableAWSEcr"`
	EcrConfigs   registryCredsAddonConfigAWSEcr `json:"awsEcrConfigs"`

	EnableGCR  string                      `json:"enableGCR"`
	GcrConfigs registryCredsAddonConfigGCR `json:"gcrConfigs"`

	EnableDockerRegistry string                         `json:"enableDockerRegistry"`
	DockerConfigs        registryCredsAddonConfigDocker `json:"dockerConfigs"`

	EnableACR  string                      `json:"enableACR"`
	AcrConfigs registryCredsAddonConfigACR `json:"acrConfigs"`
}

// Registry Creds addon config for AWS ECR
type registryCredsAddonConfigAWSEcr struct {
	AccessID     string `json:"awsAccessID"`
	AccessKey    string `json:"awsAccessKey"`
	SessionToken string `json:"awsSessionToken"`
	Region       string `json:"awsRegion"`
	Account      string `json:"awsAccount"`
	Role         string `json:"awsRole"`
}

// Registry Creds addon config for GCR
type registryCredsAddonConfigGCR struct {
	GcrPath string `json:"gcrPath"`
	GcrURL  string `json:"gcrURL"`
}

// Registry Creds addon config for Docker Registry
type registryCredsAddonConfigDocker struct {
	DockerServer string `json:"dockerServer"`
	DockerUser   string `json:"dockerUser"`
	DockerPass   string `json:"dockerPass"`
}

// Registry Creds addon config for Docker Azure container registry
type registryCredsAddonConfigACR struct {
	AcrURL      string `json:"acrURL"`
	AcrClientID string `json:"acrClientID"`
	AcrPassword string `json:"acrPassword"`
}

// Processes registry-creds addon config from configFile if it exists otherwise resorts to default behavior
func processRegistryCredsConfig(profile string, ac *addonConfig) {
	// Default values
	awsAccessID := configDefaultValue
	awsAccessKey := configDefaultValue
	awsSessionToken := ""
	awsRegion := configDefaultValue
	awsAccount := configDefaultValue
	awsRole := configDefaultValue
	gcrApplicationDefaultCredentials := configDefaultValue
	dockerServer := configDefaultValue
	dockerUser := configDefaultValue
	dockerPass := configDefaultValue
	gcrURL := "https://gcr.io"
	acrURL := configDefaultValue
	acrClientID := configDefaultValue
	acrPassword := configDefaultValue

	regCredsConf := &ac.RegistryCreds
	awsEcrAction := regCredsConf.EnableAWSEcr // regCredsConf. "enableAWSEcr")

	switch awsEcrAction {
	case "prompt", "":
		enableAWSECR := AskForYesNoConfirmation("\nDo you want to enable AWS Elastic Container Registry?", posResponses, negResponses)
		if enableAWSECR {
			awsAccessID = AskForStaticValue("-- Enter AWS Access Key ID: ")
			awsAccessKey = AskForStaticValue("-- Enter AWS Secret Access Key: ")
			awsSessionToken = AskForStaticValueOptional("-- (Optional) Enter AWS Session Token: ")
			awsRegion = AskForStaticValue("-- Enter AWS Region: ")
			awsAccount = AskForStaticValue("-- Enter 12 digit AWS Account ID (Comma separated list): ")
			awsRole = AskForStaticValueOptional("-- (Optional) Enter ARN of AWS role to assume: ")
		}
	case "enable":
		out.Ln("Loading AWS ECR configs from: %s", addonConfigFile)
		// Then read the configs
		awsAccessID = regCredsConf.EcrConfigs.AccessID
		awsAccessKey = regCredsConf.EcrConfigs.AccessKey
		awsSessionToken = regCredsConf.EcrConfigs.SessionToken
		awsRegion = regCredsConf.EcrConfigs.Region
		awsAccount = regCredsConf.EcrConfigs.Account
		awsRole = regCredsConf.EcrConfigs.Role
	case "disable":
		out.Ln("Ignoring AWS ECR configs")
	default:
		out.Ln("Disabling AWS ECR.  Invalid value for enableAWSEcr (%s).  Must be one of 'disable', 'enable' or 'prompt'", awsEcrAction)
	}

	gcrPath := ""
	gcrAction := regCredsConf.EnableGCR

	switch gcrAction {
	case "prompt", "":
		enableGCR := AskForYesNoConfirmation("\nDo you want to enable Google Container Registry?", posResponses, negResponses)
		if enableGCR {
			gcrPath = AskForStaticValue("-- Enter path to credentials (e.g. /home/user/.config/gcloud/application_default_credentials.json):")
			gcrchangeURL := AskForYesNoConfirmation("-- Do you want to change the GCR URL (Default https://gcr.io)?", posResponses, negResponses)

			if gcrchangeURL {
				gcrURL = AskForStaticValue("-- Enter GCR URL (e.g. https://asia.gcr.io):")
			}
		}
	case "enable":
		out.Ln("Loading GCR configs from: %s", addonConfigFile)
		// Then read the configs
		gcrPath = regCredsConf.GcrConfigs.GcrPath
		gcrURL = regCredsConf.GcrConfigs.GcrURL
	case "disable":
		out.Ln("Ignoring GCR configs")
	default:
		out.Ln("Disabling GCR.  Invalid value for enableGCR (%s).  Must be one of 'disable', 'enable' or 'prompt'", gcrAction)
	}

	if gcrPath != "" {
		// Read file from disk
		dat, err := os.ReadFile(gcrPath)

		if err != nil {
			exit.Message(reason.Usage, "Error reading {{.path}}: {{.error}}", out.V{"path": gcrPath, "error": err})
		} else {
			gcrApplicationDefaultCredentials = string(dat)
		}
	}

	dockerRegistryAction := regCredsConf.EnableDockerRegistry

	switch dockerRegistryAction {
	case "prompt", "":
		enableDR := AskForYesNoConfirmation("\nDo you want to enable Docker Registry?", posResponses, negResponses)
		if enableDR {
			dockerServer = AskForStaticValue("-- Enter docker registry server url: ")
			dockerUser = AskForStaticValue("-- Enter docker registry username: ")
			dockerPass = AskForPasswordValue("-- Enter docker registry password: ")
		}
	case "enable":
		out.Ln("Loading Docker Registry configs from: %s", addonConfigFile)
		dockerServer = regCredsConf.DockerConfigs.DockerServer
		dockerUser = regCredsConf.DockerConfigs.DockerUser
		dockerPass = regCredsConf.DockerConfigs.DockerPass
	case "disable":
		out.Ln("Ignoring Docker Registry configs")
	default:
		out.Ln("Disabling Docker Registry.  Invalid value for enableDockerRegistry (%s).  Must be one of 'disable', 'enable' or 'prompt'", dockerRegistryAction)
	}

	acrAction := regCredsConf.EnableACR

	switch acrAction {
	case "prompt", "":
		enableACR := AskForYesNoConfirmation("\nDo you want to enable Azure Container Registry?", posResponses, negResponses)
		if enableACR {
			acrURL = AskForStaticValue("-- Enter Azure Container Registry (ACR) URL: ")
			acrClientID = AskForStaticValue("-- Enter client ID (service principal ID) to access ACR: ")
			acrPassword = AskForPasswordValue("-- Enter service principal password to access Azure Container Registry: ")
		}
	case "enable":
		out.Ln("Loading ACR configs from: ", addonConfigFile)
		acrURL = regCredsConf.AcrConfigs.AcrURL
		acrClientID = regCredsConf.AcrConfigs.AcrClientID
		acrPassword = regCredsConf.AcrConfigs.AcrPassword
	case "disable":
		out.Ln("Ignoring ACR configs")
	default:
		out.Stringf("Disabling ACR.  Invalid value for enableACR (%s).  Must be one of 'disable', 'enable' or 'prompt'", acrAction)
	}

	namespace := "kube-system"

	// Create ECR Secret
	err := service.CreateSecret(
		profile,
		namespace,
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
		exit.Message(reason.InternalCommandRunner, "ERROR creating `registry-creds-ecr` secret: {{.error}}", out.V{"error": err})
	}

	// Create GCR Secret
	err = service.CreateSecret(
		profile,
		namespace,
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
		exit.Message(reason.InternalCommandRunner, "ERROR creating `registry-creds-gcr` secret: {{.error}}", out.V{"error": err})
	}

	// Create Docker Secret
	err = service.CreateSecret(
		profile,
		namespace,
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
		profile,
		namespace,
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
}
