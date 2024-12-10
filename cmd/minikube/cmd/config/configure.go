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
	"net"
	"os"
	"regexp"
	"time"

	"github.com/spf13/cobra"
	"k8s.io/minikube/pkg/addons"
	"k8s.io/minikube/pkg/minikube/assets"
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/cruntime"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/machine"
	"k8s.io/minikube/pkg/minikube/mustload"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/minikube/service"
	"k8s.io/minikube/pkg/minikube/style"
	"k8s.io/minikube/pkg/minikube/sysinit"
)

var posResponses = []string{"yes", "y"}
var negResponses = []string{"no", "n"}

var addonsConfigureCmd = &cobra.Command{
	Use:   "configure ADDON_NAME",
	Short: "Configures the addon w/ADDON_NAME within minikube (example: minikube addons configure registry-creds). For a list of available addons use: minikube addons list",
	Long:  "Configures the addon w/ADDON_NAME within minikube (example: minikube addons configure registry-creds). For a list of available addons use: minikube addons list",
	Run: func(_ *cobra.Command, args []string) {
		if len(args) != 1 {
			exit.Message(reason.Usage, "usage: minikube addons configure ADDON_NAME")
		}

		profile := ClusterFlagValue()

		addon := args[0]
		// allows for additional prompting of information when enabling addons
		switch addon {
		case "registry-creds":

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
				dat, err := os.ReadFile(gcrPath)

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
				out.FailureT("ERROR creating `registry-creds-ecr` secret: {{.error}}", out.V{"error": err})
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
				out.FailureT("ERROR creating `registry-creds-gcr` secret: {{.error}}", out.V{"error": err})
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

		case "metallb":
			_, cfg := mustload.Partial(profile)

			validator := func(s string) bool {
				return net.ParseIP(s) != nil
			}

			cfg.KubernetesConfig.LoadBalancerStartIP = AskForStaticValidatedValue("-- Enter Load Balancer Start IP: ", validator)

			cfg.KubernetesConfig.LoadBalancerEndIP = AskForStaticValidatedValue("-- Enter Load Balancer End IP: ", validator)

			if err := config.SaveProfile(profile, cfg); err != nil {
				out.ErrT(style.Fatal, "Failed to save config {{.profile}}", out.V{"profile": profile})
			}

			// Re-enable metallb addon in order to generate template manifest files with Load Balancer Start/End IP
			if err := addons.EnableOrDisableAddon(cfg, "metallb", "true"); err != nil {
				out.ErrT(style.Fatal, "Failed to configure metallb IP {{.profile}}", out.V{"profile": profile})
			}
		case "ingress":
			_, cfg := mustload.Partial(profile)

			validator := func(s string) bool {
				format := regexp.MustCompile("^.+/.+$")
				return format.MatchString(s)
			}

			customCert := AskForStaticValidatedValue("-- Enter custom cert (format is \"namespace/secret\"): ", validator)
			if cfg.KubernetesConfig.CustomIngressCert != "" {
				overwrite := AskForYesNoConfirmation("A custom cert for ingress has already been set. Do you want overwrite it?", posResponses, negResponses)
				if !overwrite {
					break
				}
			}

			cfg.KubernetesConfig.CustomIngressCert = customCert

			if err := config.SaveProfile(profile, cfg); err != nil {
				out.ErrT(style.Fatal, "Failed to save config {{.profile}}", out.V{"profile": profile})
			}
		case "registry-aliases":
			_, cfg := mustload.Partial(profile)
			validator := func(s string) bool {
				format := regexp.MustCompile(`^([a-zA-Z0-9-_]+\.[a-zA-Z0-9-_]+)+(\ [a-zA-Z0-9-_]+\.[a-zA-Z0-9-_]+)*$`)
				return format.MatchString(s)
			}
			registryAliases := AskForStaticValidatedValue("-- Enter registry aliases separated by space: ", validator)
			cfg.KubernetesConfig.RegistryAliases = registryAliases

			if err := config.SaveProfile(profile, cfg); err != nil {
				out.ErrT(style.Fatal, "Failed to save config {{.profile}}", out.V{"profile": profile})
			}
			addon := assets.Addons["registry-aliases"]
			if addon.IsEnabled(cfg) {
				// Re-enable registry-aliases addon in order to generate template manifest files with custom hosts
				if err := addons.EnableOrDisableAddon(cfg, "registry-aliases", "true"); err != nil {
					out.ErrT(style.Fatal, "Failed to configure registry-aliases {{.profile}}", out.V{"profile": profile})
				}
			}
		case "auto-pause":
			lapi, cfg := mustload.Partial(profile)
			intervalInput := AskForStaticValue("-- Enter interval time of auto-pause-interval (ex. 1m0s): ")
			intervalTime, err := time.ParseDuration(intervalInput)
			if err != nil {
				out.ErrT(style.Fatal, "Interval is an invalid duration: {{.error}}", out.V{"error": err})
			}
			if intervalTime != intervalTime.Abs() || intervalTime.String() == "0s" {
				out.ErrT(style.Fatal, "Interval must be greater than 0s")
			}
			cfg.AutoPauseInterval = intervalTime
			if err := config.SaveProfile(profile, cfg); err != nil {
				out.ErrT(style.Fatal, "Failed to save config {{.profile}}", out.V{"profile": profile})
			}
			addon := assets.Addons["auto-pause"]
			if addon.IsEnabled(cfg) {

				// see #17945: restart auto-pause service
				p, err := config.LoadProfile(profile)
				if err != nil {
					out.ErrT(style.Fatal, "failed to load profile: {{.error}}", out.V{"error": err})
				}
				if profileStatus(p, lapi).StatusCode/100 == 2 { // 2xx code
					co := mustload.Running(profile)
					// first unpause all nodes cluster immediately
					unpauseWholeCluster(co)
					// Re-enable auto-pause addon in order to update interval time
					if err := addons.EnableOrDisableAddon(cfg, "auto-pause", "true"); err != nil {
						out.ErrT(style.Fatal, "Failed to configure auto-pause {{.profile}}", out.V{"profile": profile})
					}
					// restart auto-pause service
					if err := sysinit.New(co.CP.Runner).Restart("auto-pause"); err != nil {
						out.ErrT(style.Fatal, "failed to restart auto-pause: {{.error}}", out.V{"error": err})
					}

				}
			}
		default:
			out.FailureT("{{.name}} has no available configuration options", out.V{"name": addon})
			return
		}

		out.SuccessT("{{.name}} was successfully configured", out.V{"name": addon})
	},
}

func unpauseWholeCluster(co mustload.ClusterController) {
	for _, n := range co.Config.Nodes {

		// Use node-name if available, falling back to cluster name
		name := n.Name
		if n.Name == "" {
			name = co.Config.Name
		}

		out.Step(style.Pause, "Unpausing node {{.name}} ... ", out.V{"name": name})

		machineName := config.MachineName(*co.Config, n)
		host, err := machine.LoadHost(co.API, machineName)
		if err != nil {
			exit.Error(reason.GuestLoadHost, "Error getting host", err)
		}

		r, err := machine.CommandRunner(host)
		if err != nil {
			exit.Error(reason.InternalCommandRunner, "Failed to get command runner", err)
		}

		cr, err := cruntime.New(cruntime.Config{Type: co.Config.KubernetesConfig.ContainerRuntime, Runner: r})
		if err != nil {
			exit.Error(reason.InternalNewRuntime, "Failed runtime", err)
		}

		_, err = cluster.Unpause(cr, r, nil) // nil means all namespaces
		if err != nil {
			exit.Error(reason.GuestUnpause, "Pause", err)
		}
	}
}

func init() {
	AddonsCmd.AddCommand(addonsConfigureCmd)
}
