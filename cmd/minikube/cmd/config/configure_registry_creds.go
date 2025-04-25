package config

import (
	"os"

	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/minikube/service"
)

// Top level configs for RegistryCreds addons
type RegistryCredsAddonConfig struct {
	EnableAWSEcr string                         `json:"enableAWSEcr"`
	EcrConfigs   RegistryCredsAddonConfigAWSEcr `json:"awsEcrConfigs"`

	EnableGCR  string                      `json:"enableGCR"`
	GcrConfigs RegistryCredsAddonConfigGCR `json:"gcrConfigs"`

	EnableDockerRegistry string                         `json:"enableDockerRegistry"`
	DockerConfigs        RegistryCredsAddonConfigDocker `json:"dockerConfigs"`

	EnableACR  string                      `json:"enableACR"`
	AcrConfigs RegistryCredsAddonConfigACR `json:"acrConfigs"`
}

type RegistryCredsAddonConfigAWSEcr struct {
	AccessID     string `json:"awsAccessID"`
	AccessKey    string `json:"awsAccessKey"`
	SessionToken string `json:"awsSessionToken"`
	Region       string `json:"awsRegion"`
	Account      string `json:"awsAccount"`
	Role         string `json:"awsRole"`
}

type RegistryCredsAddonConfigGCR struct {
	GcrPath string `json:"gcrPath"`
	GcrURL  string `json:"gcrURL"`
}

type RegistryCredsAddonConfigDocker struct {
	DockerServer string `json:"dockerServer"`
	DockerUser   string `json:"dockerUser"`
	DockerPass   string `json:"dockerPass"`
}

type RegistryCredsAddonConfigACR struct {
	AcrURL      string `json:"acrURL"`
	AcrClientID string `json:"acrClientID"`
	AcrPassword string `json:"acrPassword"`
}

// Processes registry-creds addon config from configFile if it exists otherwise resorts to default behavior
func processRegistryCredsConfig(profile string, addonConfig *AddonConfig) {
	// Default values
	awsAccessID := "MINIKUBE_DEFAULT_VALUE"
	awsAccessKey := "MINIKUBE_DEFAULT_VALUE"
	awsSessionToken := ""
	awsRegion := "MINIKUBE_DEFAULT_VALUE"
	awsAccount := "MINIKUBE_DEFAULT_VALUE"
	awsRole := "MINIKUBE_DEFAULT_VALUE"
	gcrApplicationDefaultCredentials := "MINIKUBE_DEFAULT_VALUE"
	dockerServer := "MINIKUBE_DEFAULT_VALUE"
	dockerUser := "MINIKUBE_DEFAULT_VALUE"
	dockerPass := "MINIKUBE_DEFAULT_VALUE"
	gcrURL := "https://gcr.io"
	acrURL := "MINIKUBE_DEFAULT_VALUE"
	acrClientID := "MINIKUBE_DEFAULT_VALUE"
	acrPassword := "MINIKUBE_DEFAULT_VALUE"

	regCredsConf := &addonConfig.RegistryCreds
	awsEcrAction := regCredsConf.EnableAWSEcr // regCredsConf. "enableAWSEcr")
	if awsEcrAction == "prompt" || awsEcrAction == "" {
		enableAWSECR := AskForYesNoConfirmation("\nDo you want to enable AWS Elastic Container Registry?", posResponses, negResponses)
		if enableAWSECR {
			awsAccessID = AskForStaticValue("-- Enter AWS Access Key ID: ")
			awsAccessKey = AskForStaticValue("-- Enter AWS Secret Access Key: ")
			awsSessionToken = AskForStaticValueOptional("-- (Optional) Enter AWS Session Token: ")
			awsRegion = AskForStaticValue("-- Enter AWS Region: ")
			awsAccount = AskForStaticValue("-- Enter 12 digit AWS Account ID (Comma separated list): ")
			awsRole = AskForStaticValueOptional("-- (Optional) Enter ARN of AWS role to assume: ")
		}
	} else if awsEcrAction == "enable" {
		out.Ln("Loading AWS ECR configs from: %s", addonConfigFile)
		// Then read the configs
		awsAccessID = regCredsConf.EcrConfigs.AccessID
		awsAccessKey = regCredsConf.EcrConfigs.AccessKey
		awsSessionToken = regCredsConf.EcrConfigs.SessionToken
		awsRegion = regCredsConf.EcrConfigs.Region
		awsAccount = regCredsConf.EcrConfigs.Account
		awsRole = regCredsConf.EcrConfigs.Role
	} else if awsEcrAction == "disable" {
		out.Ln("Ignoring AWS ECR configs")
	} else {
		out.Ln("Disabling AWS ECR.  Invalid value for enableAWSEcr (%s).  Must be one of 'disable', 'enable' or 'prompt'", awsEcrAction)
	}

	gcrPath := ""
	gcrAction := regCredsConf.EnableGCR
	if gcrAction == "prompt" || gcrAction == "" {
		enableGCR := AskForYesNoConfirmation("\nDo you want to enable Google Container Registry?", posResponses, negResponses)
		if enableGCR {
			gcrPath = AskForStaticValue("-- Enter path to credentials (e.g. /home/user/.config/gcloud/application_default_credentials.json):")
			gcrchangeURL := AskForYesNoConfirmation("-- Do you want to change the GCR URL (Default https://gcr.io)?", posResponses, negResponses)

			if gcrchangeURL {
				gcrURL = AskForStaticValue("-- Enter GCR URL (e.g. https://asia.gcr.io):")
			}
		}
	} else if gcrAction == "enable" {
		out.Ln("Loading GCR configs from: %s", addonConfigFile)
		// Then read the configs
		gcrPath = regCredsConf.GcrConfigs.GcrPath
		gcrURL = regCredsConf.GcrConfigs.GcrURL
	} else if gcrAction == "disable" {
		out.Ln("Ignoring GCR configs")
	} else {
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
	if dockerRegistryAction == "prompt" || dockerRegistryAction == "" {
		enableDR := AskForYesNoConfirmation("\nDo you want to enable Docker Registry?", posResponses, negResponses)
		if enableDR {
			dockerServer = AskForStaticValue("-- Enter docker registry server url: ")
			dockerUser = AskForStaticValue("-- Enter docker registry username: ")
			dockerPass = AskForPasswordValue("-- Enter docker registry password: ")
		}
	} else if dockerRegistryAction == "enable" {
		out.Ln("Loading Docker Registry configs from: %s", addonConfigFile)
		dockerServer = regCredsConf.DockerConfigs.DockerServer
		dockerUser = regCredsConf.DockerConfigs.DockerUser
		dockerPass = regCredsConf.DockerConfigs.DockerPass
	} else if dockerRegistryAction == "disable" {
		out.Ln("Ignoring Docker Registry configs")
	} else {
		out.Ln("Disabling Docker Registry.  Invalid value for enableDockerRegistry (%s).  Must be one of 'disable', 'enable' or 'prompt'", dockerRegistryAction)
	}

	acrAction := regCredsConf.EnableACR
	if acrAction == "prompt" || acrAction == "" {
		enableACR := AskForYesNoConfirmation("\nDo you want to enable Azure Container Registry?", posResponses, negResponses)
		if enableACR {
			acrURL = AskForStaticValue("-- Enter Azure Container Registry (ACR) URL: ")
			acrClientID = AskForStaticValue("-- Enter client ID (service principal ID) to access ACR: ")
			acrPassword = AskForPasswordValue("-- Enter service principal password to access Azure Container Registry: ")
		}
	} else if acrAction == "enable" {
		out.Ln("Loading ACR configs from: ", addonConfigFile)
		acrURL = regCredsConf.AcrConfigs.AcrURL
		acrClientID = regCredsConf.AcrConfigs.AcrClientID
		acrPassword = regCredsConf.AcrConfigs.AcrPassword
	} else if acrAction == "disable" {
		out.Ln("Ignoring ACR configs")
	} else {
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
