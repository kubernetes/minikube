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
	"errors"
	"fmt"
	"net"
	"os"
	"regexp"
	"time"

	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
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
	"k8s.io/minikube/pkg/minikube/style"
	"k8s.io/minikube/pkg/minikube/sysinit"
)

var addonConfigFile = ""
var posResponses = []string{"yes", "y"}
var negResponses = []string{"no", "n"}

// Typed addon configs
type addonConfig struct {
	RegistryCreds registryCredsAddonConfig `json:"registry-creds"`
}

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
		addonConfig := loadAddonConfigFile(addon, addonConfigFile)

		// allows for additional prompting of information when enabling addons
		switch addon {
		case "registry-creds":
			processRegistryCredsConfig(profile, addonConfig)

		case "metallb":
			processMetalLBConfig(profile, addonConfig)

		case "ingress":
			processIngressConfig(profile, addonConfig)

		case "registry-aliases":
			processRegistryAliasesConfig(profile, addonConfig)

		case "auto-pause":
			processAutoPauseConfig(profile, addonConfig)

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
	addonsConfigureCmd.Flags().StringVarP(&addonConfigFile, "config-file", "f", "", "An optional configuration file to read addon specific configs from instead of being prompted each time.")
	AddonsCmd.AddCommand(addonsConfigureCmd)
}

// Helper method to load a config file for addons
func loadAddonConfigFile(addon, configFilePath string) (ac *addonConfig) {
	type configFile struct {
		Addons addonConfig `json:"addons"`
	}
	var cf configFile

	if configFilePath != "" {
		out.Ln("Reading %s configs from %s", addon, configFilePath)
		confData, err := os.ReadFile(configFilePath)
		if err != nil && errors.Is(err, os.ErrNotExist) { // file does not exist
			klog.Warningf("config file (%s) does not exist: %v", configFilePath, err)
			exit.Message(reason.Usage, "config file does not exist")
		}

		if err != nil { // file cannot be opened
			klog.Errorf("error opening config file (%s): %v", configFilePath, err)
			// err = errors2.Wrapf(err, "config file (%s) does not exist", configFilePath)
			exit.Message(reason.Kind{ExitCode: reason.ExProgramConfig, Advice: "provide a valid config file"},
				fmt.Sprintf("error opening config file: %s", configFilePath))
		}

		if err = json.Unmarshal(confData, &cf); err != nil {
			// err = errors2.Wrapf(err, "error reading config file (%s)", configFilePath)
			klog.Errorf("error reading config file (%s): %v", configFilePath, err)
			exit.Message(reason.Kind{ExitCode: reason.ExProgramConfig, Advice: "provide a valid config file"},
				fmt.Sprintf("error reading config file: %v", err))
		}

		return &cf.Addons
	}
	return nil
}

// Processes metallb addon config from configFile if it exists otherwise resorts to default behavior
func processMetalLBConfig(profile string, _ *addonConfig) {
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
}

// Processes ingress addon config from configFile if it exists otherwise resorts to default behavior
func processIngressConfig(profile string, _ *addonConfig) {
	_, cfg := mustload.Partial(profile)

	validator := func(s string) bool {
		format := regexp.MustCompile("^.+/.+$")
		return format.MatchString(s)
	}

	customCert := AskForStaticValidatedValue("-- Enter custom cert (format is \"namespace/secret\"): ", validator)
	if cfg.KubernetesConfig.CustomIngressCert != "" {
		overwrite := AskForYesNoConfirmation("A custom cert for ingress has already been set. Do you want overwrite it?", posResponses, negResponses)
		if !overwrite {
			return
		}
	}

	cfg.KubernetesConfig.CustomIngressCert = customCert

	if err := config.SaveProfile(profile, cfg); err != nil {
		out.ErrT(style.Fatal, "Failed to save config {{.profile}}", out.V{"profile": profile})
	}
}

// Processes auto-pause addon config from configFile if it exists otherwise resorts to default behavior
func processAutoPauseConfig(profile string, _ *addonConfig) {
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
	if err = config.SaveProfile(profile, cfg); err != nil {
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
}

// Processes registry-aliases addon config from configFile if it exists otherwise resorts to default behavior
func processRegistryAliasesConfig(profile string, _ *addonConfig) {
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
}
