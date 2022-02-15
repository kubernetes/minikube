/*
Copyright 2022 The Kubernetes Authors All rights reserved.

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
	"github.com/spf13/cobra"
	"k8s.io/minikube/pkg/minikube/download"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/reason"
)

var addonsToggle = &cobra.Command{
	Use:    "toggle ADDON_NAME on/off",
	Short:  "Manually enables or disables an addon globally",
	Long:   "Manually enables or disables an addon globally for security reasons",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			usage()
		}

		addon := args[0]
		toggle := args[1]
		if toggle != "on" && toggle != "off" {
			usage()
		}

		// Reason is required when toggling off
		if toggle == "off" && toggleReason == "" {
			usage()
		}

		enable := toggle == "on"

		status, err := download.AddonStatus()
		if err != nil {
			exit.Error(reason.InternalAddonScan, "downloading addon status file", err)
		}

		addonStatus := status[addon]
		addonStatus.Enabled = enable
		// If enable is true, then we can reset manual to false since we're turning the addon back on
		// If enable is false, then we're disabling the addon manually and manual should be true
		addonStatus.Manual = !enable
		addonStatus.ManualReason = toggleReason
		status[addon] = addonStatus
		writeStatusYAML(status)
	},
}

var (
	toggleReason string
)

func usage() {
	exit.Message(reason.Usage, "Usage: minikube addons toggle dashboard off --reason=\"I don't like it\"")
}

func init() {
	addonsToggle.Flags().StringVar(&toggleReason, "reason", "", "Reasons the addon was toggled. Required of toggle off.")
	AddonsCmd.AddCommand(addonsToggle)
}
